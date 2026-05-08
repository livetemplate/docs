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

// allowlistedFrontmatterKeys are upstream frontmatter keys the sync
// preserves verbatim when mirroring. Anything else upstream sets is
// dropped — sync owns the docs-site frontmatter contract and only
// passes through keys the docs renderer (tinkerdown) actually consumes.
//
// Add new entries with care: each one is a contract between upstream
// content and the docs site renderer.
var allowlistedFrontmatterKeys = []string{
	"description",
	"lvt_show_source",
	"sidebar",
}

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
		upstreamFM, stripped, err := parseFrontmatter(string(body))
		if err != nil {
			res.Skipped++
			res.SkippedReasons = append(res.SkippedReasons,
				fmt.Sprintf("%s (parse frontmatter): %v", p.SiteURL, err))
			continue
		}
		rewritten := rewriter.Rewrite(stripped)
		out := composeWithFrontmatter(title, p.SourceRepo, p.SourcePath, opts.Ref, commit, upstreamFM, rewritten)

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
			if err := mirrorAdjacentApp(srcAbs, dest); err != nil {
				return res, codedErr{fmt.Errorf("mirror _app/ for %s: %w", p.SiteURL, err), 3}
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

// parseFrontmatter parses a leading "---\n...\n---\n" YAML frontmatter
// block into a generic map and returns the body that follows. If no
// frontmatter is present, returns (nil, body, nil) so callers can treat
// it the same as an empty map.
func parseFrontmatter(body string) (map[string]any, string, error) {
	if !strings.HasPrefix(body, "---\n") && !strings.HasPrefix(body, "---\r\n") {
		return nil, body, nil
	}
	var (
		fmText string
		rest   string
	)
	if end := strings.Index(body[4:], "\n---\n"); end >= 0 {
		fmText = body[4 : 4+end]
		rest = body[4+end+len("\n---\n"):]
	} else if end := strings.Index(body[4:], "\n---\r\n"); end >= 0 {
		fmText = body[4 : 4+end]
		rest = body[4+end+len("\n---\r\n"):]
	} else {
		return nil, body, nil
	}
	var m map[string]any
	if err := yaml.Unmarshal([]byte(fmText), &m); err != nil {
		return nil, body, fmt.Errorf("parse upstream frontmatter: %w", err)
	}
	return m, rest, nil
}

// composeWithFrontmatter prepends the docs-site provenance frontmatter
// to a body that has already been frontmatter-stripped, then appends
// allowlisted keys from the upstream frontmatter map.
//
// Provenance keys (sync owns these, always overrides):
//
//	title          — extracted from upstream
//	source_repo    — from source-of-truth.yaml
//	source_path    — from source-of-truth.yaml
//	source_ref     — the human-readable ref (tag/branch) sync was invoked with;
//	                 drives tinkerdown's source-link footer URLs for include= blocks
//	source_commit  — git rev-parse HEAD at sync time (immutable record)
//
// Allowlisted keys from upstream (preserved when present):
//
//	description, lvt_show_source, sidebar
func composeWithFrontmatter(title, srcRepo, srcPath, ref, commit string, upstream map[string]any, body string) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %q\n", title)
	fmt.Fprintf(&b, "source_repo: %q\n", srcRepo)
	fmt.Fprintf(&b, "source_path: %q\n", srcPath)
	fmt.Fprintf(&b, "source_ref: %q\n", ref)
	fmt.Fprintf(&b, "source_commit: %q\n", commit)
	for _, key := range allowlistedFrontmatterKeys {
		v, ok := upstream[key]
		if !ok {
			continue
		}
		writeFrontmatterValue(&b, key, v)
	}
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

// writeFrontmatterValue emits "key: value\n" using the value's Go type
// to choose YAML scalar form. Booleans emit unquoted (true/false);
// strings emit Go-quoted to match the existing emit style for our
// provenance keys; everything else falls back to fmt %v.
func writeFrontmatterValue(b *strings.Builder, key string, v any) {
	switch t := v.(type) {
	case bool:
		fmt.Fprintf(b, "%s: %t\n", key, t)
	case string:
		fmt.Fprintf(b, "%s: %q\n", key, t)
	default:
		fmt.Fprintf(b, "%s: %v\n", key, t)
	}
}

// mirrorAdjacentApp copies an `_app/` directory next to the upstream
// markdown into the same relative position next to the destination
// file. Used for literate authoring (`include="./_app/foo.go"`) where
// the included files live alongside the README.
//
// The destination `_app/` is cleared before repopulation so the docs
// site mirrors upstream state authoritatively (orphaned files removed).
//
// Symlinks inside the upstream `_app/` are rejected — they're a
// path-confinement risk and tinkerdown's include resolver canonicalises
// paths anyway.
func mirrorAdjacentApp(srcReadmeAbs, destReadmeAbs string) error {
	srcAppDir := filepath.Join(filepath.Dir(srcReadmeAbs), "_app")
	st, err := os.Stat(srcAppDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat %s: %w", srcAppDir, err)
	}
	if !st.IsDir() {
		return nil
	}
	destAppDir := filepath.Join(filepath.Dir(destReadmeAbs), "_app")
	if err := os.RemoveAll(destAppDir); err != nil {
		return fmt.Errorf("clear %s: %w", destAppDir, err)
	}
	srcRoot, err := filepath.Abs(srcAppDir)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", srcAppDir, err)
	}
	return filepath.WalkDir(srcAppDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// Reject symlinks anywhere in _app/ — they break path
		// confinement and tinkerdown's include resolver canonicalises
		// paths so symlinked content wouldn't survive the docs render anyway.
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink not allowed in _app/: %s", path)
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(destAppDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("non-regular file in _app/: %s", path)
		}
		return copyFile(path, dest)
	})
}

// copyFile copies a regular file's bytes from src to dst. The destination's
// parent directory must exist (mirrorAdjacentApp creates it via WalkDir).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
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
