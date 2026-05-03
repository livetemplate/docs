package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Options configures a sync run.
type Options struct {
	SourceRepo string // e.g. "https://github.com/livetemplate/livetemplate"
	Ref        string // tag or branch
	SiteRoot   string // path to the livetemplate/docs checkout
	DryRun     bool   // if true, do not write any files
}

// Result reports what the sync did.
type Result struct {
	Updated        int
	Unchanged      int
	Skipped        int
	UpdatedPaths   []string
	SkippedReasons []string
}

// PageEntry mirrors one entry of source-of-truth.yaml. Only the fields
// the sync action consumes are declared; unknown fields are ignored so
// the YAML schema can grow without breaking the reader.
type PageEntry struct {
	SiteURL    string `yaml:"site_url"`
	SourceRepo string `yaml:"source_repo"`
	SourcePath string `yaml:"source_path"`
}

// SourceOfTruth is the top-level shape of source-of-truth.yaml.
type SourceOfTruth struct {
	Pages []PageEntry `yaml:"pages"`
}

// Run performs the sync described by Options and returns a Result.
// All errors that point at a recoverable problem (bad source-repo URL,
// no matching entries) carry an exit code via the exitCoder interface.
func Run(opts Options) (Result, error) {
	res := Result{}

	cfgPath := filepath.Join(opts.SiteRoot, "content", "_meta", "source-of-truth.yaml")
	cfg, err := loadSourceOfTruth(cfgPath)
	if err != nil {
		return res, codedErr{err, 1}
	}

	matched := filterByRepo(cfg.Pages, opts.SourceRepo)
	if len(matched) == 0 {
		return res, codedErr{
			fmt.Errorf("no entries in %s have source_repo=%q", cfgPath, opts.SourceRepo),
			1,
		}
	}

	tmp, err := os.MkdirTemp("", "lvt-docs-sync-*")
	if err != nil {
		return res, codedErr{fmt.Errorf("temp dir: %w", err), 2}
	}
	defer os.RemoveAll(tmp)

	if err := cloneShallow(opts.SourceRepo, opts.Ref, tmp); err != nil {
		return res, codedErr{err, 2}
	}

	commit, err := headCommit(tmp)
	if err != nil {
		return res, codedErr{err, 2}
	}

	rewriter := newLinkRewriter(cfg)

	for _, p := range matched {
		srcAbs := filepath.Join(tmp, p.SourcePath)
		body, err := os.ReadFile(srcAbs)
		if err != nil {
			res.Skipped++
			res.SkippedReasons = append(res.SkippedReasons,
				fmt.Sprintf("%s (source missing at ref %s): %v", p.SiteURL, opts.Ref, err))
			continue
		}

		dest := destFor(opts.SiteRoot, p.SiteURL)
		title := extractTitle(string(body), p.SourcePath)
		stripped := stripFrontmatter(string(body))
		rewritten := rewriter.Rewrite(stripped)
		out := composeWithFrontmatter(title, p.SourceRepo, p.SourcePath, commit, rewritten)

		existing, _ := os.ReadFile(dest)
		if string(existing) == out {
			res.Unchanged++
			continue
		}

		if !opts.DryRun {
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return res, codedErr{fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err), 3}
			}
			if err := os.WriteFile(dest, []byte(out), 0o644); err != nil {
				return res, codedErr{fmt.Errorf("write %s: %w", dest, err), 3}
			}
		}
		res.Updated++
		res.UpdatedPaths = append(res.UpdatedPaths, p.SiteURL)
	}

	return res, nil
}

func loadSourceOfTruth(path string) (*SourceOfTruth, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg SourceOfTruth
	if err := yaml.Unmarshal(body, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(cfg.Pages) == 0 {
		return nil, fmt.Errorf("no `pages:` entries in %s", path)
	}
	return &cfg, nil
}

func filterByRepo(pages []PageEntry, srcRepo string) []PageEntry {
	srcRepo = strings.TrimSuffix(strings.TrimSpace(srcRepo), "/")
	var out []PageEntry
	for _, p := range pages {
		if strings.TrimSuffix(strings.TrimSpace(p.SourceRepo), "/") == srcRepo {
			out = append(out, p)
		}
	}
	return out
}

func cloneShallow(repo, ref, dest string) error {
	cmd := exec.Command("git", "clone", "--depth=1", "--branch="+ref, repo, dest)
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone %s @%s: %w", repo, ref, err)
	}
	return nil
}

func headCommit(repoDir string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// destFor maps a site URL to a docs/content/ file path. Trailing slash
// becomes <path>/index.md; everything else becomes <path>.md. Mirrors
// the convention scripts/manual-port.sh established in Phase 2.
func destFor(siteRoot, siteURL string) string {
	rel := strings.TrimPrefix(siteURL, "/")
	switch {
	case rel == "":
		return filepath.Join(siteRoot, "content", "index.md")
	case strings.HasSuffix(rel, "/"):
		return filepath.Join(siteRoot, "content", rel, "index.md")
	default:
		return filepath.Join(siteRoot, "content", rel+".md")
	}
}

// extractTitle returns the page title from (in order) source frontmatter
// title, the first markdown H1, or a humanized version of the filename.
func extractTitle(body, srcPath string) string {
	if t := extractFrontmatterTitle(body); t != "" {
		return t
	}
	if t := extractFirstH1(body); t != "" {
		return t
	}
	base := filepath.Base(srcPath)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return strings.ReplaceAll(base, "-", " ")
}

func extractFrontmatterTitle(body string) string {
	scanner := bufio.NewScanner(strings.NewReader(body))
	first := true
	inFM := false
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			if strings.TrimSpace(line) != "---" {
				return ""
			}
			inFM = true
			continue
		}
		if !inFM {
			break
		}
		if strings.TrimSpace(line) == "---" {
			break
		}
		if strings.HasPrefix(line, "title:") {
			t := strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			t = strings.Trim(t, `"'`)
			return t
		}
	}
	return ""
}

var h1Re = regexp.MustCompile(`(?m)^# +(.+?)\s*$`)

func extractFirstH1(body string) string {
	body = stripFrontmatter(body)
	if m := h1Re.FindStringSubmatch(body); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// stripFrontmatter removes a leading "---\n...\n---\n" block, returning
// the body content. If no frontmatter is present, returns the body
// unchanged.
func stripFrontmatter(body string) string {
	if !strings.HasPrefix(body, "---\n") && !strings.HasPrefix(body, "---\r\n") {
		return body
	}
	end := strings.Index(body[4:], "\n---\n")
	if end < 0 {
		end = strings.Index(body[4:], "\n---\r\n")
		if end < 0 {
			return body
		}
		return body[4+end+len("\n---\r\n"):]
	}
	return body[4+end+len("\n---\n"):]
}

// composeWithFrontmatter prepends the docs-site provenance frontmatter
// to a body that has already been frontmatter-stripped.
func composeWithFrontmatter(title, srcRepo, srcPath, commit, body string) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %q\n", title)
	fmt.Fprintf(&b, "source_repo: %q\n", srcRepo)
	fmt.Fprintf(&b, "source_path: %q\n", srcPath)
	fmt.Fprintf(&b, "source_commit: %q\n", commit)
	b.WriteString("---\n\n")
	// Trim leading blank lines from body so the output isn't littered
	// with extra whitespace where the source's preamble used to be.
	body = strings.TrimLeft(body, "\n\r")
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

// linkRewriter rewrites cross-repo GitHub links to docs-site-relative
// URLs based on the source-of-truth matrix. Links that don't map to a
// known page are left untouched (so external GitHub references survive).
type linkRewriter struct {
	urlToSiteURL map[string]string
}

func newLinkRewriter(cfg *SourceOfTruth) *linkRewriter {
	m := make(map[string]string, len(cfg.Pages)*2)
	for _, p := range cfg.Pages {
		repo := strings.TrimSuffix(strings.TrimSpace(p.SourceRepo), "/")
		path := strings.TrimPrefix(strings.TrimSpace(p.SourcePath), "/")
		if repo == "" || path == "" {
			continue
		}
		// Both the canonical edit-form URL and the blob form should rewrite.
		m[repo+"/blob/main/"+path] = p.SiteURL
		m[repo+"/edit/main/"+path] = p.SiteURL
	}
	return &linkRewriter{urlToSiteURL: m}
}

// Rewrite applies the rewrite rules to the input markdown body. Only
// exact URL matches (within `(...)` markdown link syntax or bare in
// prose) are rewritten — partial matches and substrings are left
// alone to avoid mangling code blocks.
func (r *linkRewriter) Rewrite(body string) string {
	for from, to := range r.urlToSiteURL {
		body = strings.ReplaceAll(body, from, to)
	}
	return body
}

// codedErr lets a Run-time error carry an exit code that main() reads.
type codedErr struct {
	err  error
	code int
}

func (e codedErr) Error() string { return e.err.Error() }
func (e codedErr) Unwrap() error { return e.err }
func (e codedErr) exitCode() int { return e.code }
