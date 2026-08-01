package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
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
		rewritten = rewriter.RewriteRelative(rewritten, p, opts.Ref)
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

// linkRewriter rewrites cross-repo GitHub links to docs-site-relative
// URLs based on the source-of-truth matrix. Links that don't map to a
// known page are left untouched (so external GitHub references survive).
type linkRewriter struct {
	urlToSiteURL map[string]string
	// repoPathToSiteURL maps repo+"\x00"+source_path to a page's site URL.
	// RewriteRelative needs the lookup keyed by repo-relative path rather
	// than by full GitHub URL, since that is what resolving an
	// upstream-relative link produces.
	repoPathToSiteURL map[string]string
}

func newLinkRewriter(cfg *SourceOfTruth) *linkRewriter {
	m := make(map[string]string, len(cfg.Pages)*2)
	byPath := make(map[string]string, len(cfg.Pages))
	for _, p := range cfg.Pages {
		repo := strings.TrimSuffix(strings.TrimSpace(p.SourceRepo), "/")
		srcPath := strings.TrimPrefix(strings.TrimSpace(p.SourcePath), "/")
		if repo == "" || srcPath == "" {
			continue
		}
		// Both the canonical edit-form URL and the blob form should rewrite.
		m[repo+"/blob/main/"+srcPath] = p.SiteURL
		m[repo+"/edit/main/"+srcPath] = p.SiteURL
		byPath[repo+"\x00"+srcPath] = p.SiteURL
	}
	return &linkRewriter{urlToSiteURL: m, repoPathToSiteURL: byPath}
}

// markdownLinkRE matches the target of a `](...)` markdown link. Only that
// form is matched: upstream prose also contains bare relative paths that are
// not links, and rewriting those would corrupt them. isRelativeRef then
// decides which targets are upstream-relative.
var markdownLinkRE = regexp.MustCompile(`\]\(([^)\s]+)\)`)

// isRelativeRef reports whether a markdown link target is a path relative to
// the page's own location upstream — the kind that breaks once mirrored.
//
// Both spellings occur and both break: the dot-prefixed form
// ("../references/api-reference.md") and the bare sibling form
// ("controller-pattern.md", which upstream uses for same-directory links).
// The bare form is the more common of the two and resolves on the docs site
// to a URL ending in .md, which tinkerdown does not serve.
func isRelativeRef(target string) bool {
	switch {
	case target == "":
		return false
	case strings.HasPrefix(target, "#"): // in-page anchor
		return false
	case strings.HasPrefix(target, "/"): // site-absolute, incl. protocol-relative
		return false
	}
	// A scheme (http:, https:, mailto:, tel:) means it is not relative. Look
	// for ':' before the first '/' so that "docs/a:b.md" is not mistaken for
	// one.
	if i := strings.IndexAny(target, ":/"); i >= 0 && target[i] == ':' {
		return false
	}
	return true
}

// RewriteRelative resolves upstream-relative markdown links against the
// mirrored page's own location in its source repo.
//
// Upstream links its siblings relatively — `../references/api-reference.md`
// from `docs/guides/foo.md` — which is correct *in that repo* and dead once
// mirrored, because the docs site has no such path. Resolving against
// page.SourcePath's directory yields the repo-relative target, which either
// maps to a mirrored page (rewrite to its site URL) or is a real upstream
// file that is not mirrored (rewrite to its GitHub URL, so the reader still
// reaches it rather than a 404).
//
// Fenced code blocks are skipped: a relative path inside an example is part
// of the example.
func (r *linkRewriter) RewriteRelative(body string, page PageEntry, ref string) string {
	repo := strings.TrimSuffix(strings.TrimSpace(page.SourceRepo), "/")
	srcPath := strings.TrimPrefix(strings.TrimSpace(page.SourcePath), "/")
	if repo == "" || srcPath == "" || ref == "" {
		return body
	}
	srcDir := path.Dir(srcPath)

	lines := strings.Split(body, "\n")
	inFence := false
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		lines[i] = markdownLinkRE.ReplaceAllStringFunc(ln, func(m string) string {
			target := markdownLinkRE.FindStringSubmatch(m)[1]
			if !isRelativeRef(target) {
				return m
			}
			return "](" + r.resolveRelative(target, repo, srcDir, ref) + ")"
		})
	}
	return strings.Join(lines, "\n")
}

// resolveRelative maps one upstream-relative link target to its docs-site or
// GitHub destination, preserving any #fragment.
func (r *linkRewriter) resolveRelative(target, repo, srcDir, ref string) string {
	frag := ""
	if i := strings.IndexByte(target, '#'); i >= 0 {
		frag, target = target[i:], target[:i]
	}
	if target == "" {
		return target + frag
	}
	isDir := strings.HasSuffix(target, "/")
	resolved := path.Join(srcDir, target)
	// A target that climbs above the repo root cannot be resolved to
	// anything meaningful; leave it exactly as it was.
	if resolved == ".." || strings.HasPrefix(resolved, "../") {
		return target + frag
	}
	if site, ok := r.repoPathToSiteURL[repo+"\x00"+resolved]; ok {
		return site + frag
	}
	kind := "blob"
	if isDir {
		kind = "tree"
	}
	return repo + "/" + kind + "/" + ref + "/" + resolved + frag
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
