package main

import (
	"strings"
	"testing"
)

func TestStripFrontmatter(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"no frontmatter", "# Hello\n\nbody", "# Hello\n\nbody"},
		{"empty body", "", ""},
		{"basic frontmatter", "---\ntitle: x\n---\n\n# H1\nbody", "\n# H1\nbody"},
		{"frontmatter only", "---\ntitle: x\n---\n", ""},
		{"crlf", "---\r\ntitle: x\r\n---\r\nbody", "body"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := stripFrontmatter(c.in); got != c.want {
				t.Errorf("stripFrontmatter(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestExtractFrontmatterTitle(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"no frontmatter", "# Hello\nbody", ""},
		{"frontmatter without title", "---\ntype: guide\n---\nbody", ""},
		{"quoted title", "---\ntitle: \"Hello World\"\n---\nbody", "Hello World"},
		{"unquoted title", "---\ntitle: Hello World\n---\nbody", "Hello World"},
		{"single-quoted title", "---\ntitle: 'Hello'\n---\nbody", "Hello"},
		{"title not in frontmatter", "title: not really\nbody", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := extractFrontmatterTitle(c.in); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestExtractFirstH1(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"basic", "# Hello\nbody", "Hello"},
		{"with frontmatter", "---\nfoo: bar\n---\n# Hello\n", "Hello"},
		{"first of multiple", "# First\nbody\n# Second\n", "First"},
		{"trims trailing whitespace", "#  Hello   \n", "Hello"},
		{"no h1", "## H2 only\nbody", ""},
		{"h1 not at line start (skipped)", "    # not really\n# Real\n", "Real"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := extractFirstH1(c.in); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestExtractTitlePrecedence(t *testing.T) {
	// frontmatter title wins over H1
	got := extractTitle("---\ntitle: From FM\n---\n# From H1\n", "ignored.md")
	if got != "From FM" {
		t.Errorf("frontmatter should win: got %q", got)
	}
	// H1 wins over filename when no frontmatter title
	got = extractTitle("# From H1\nbody", "ignored.md")
	if got != "From H1" {
		t.Errorf("H1 should win over filename: got %q", got)
	}
	// Filename fallback when neither
	got = extractTitle("body without headers", "some-file-name.md")
	if got != "some file name" {
		t.Errorf("filename fallback: got %q", got)
	}
}

func TestDestFor(t *testing.T) {
	cases := []struct {
		siteURL, want string
	}{
		{"/", "site/content/index.md"},
		{"/getting-started/install", "site/content/getting-started/install.md"},
		{"/cli/", "site/content/cli/index.md"},
		{"/recipes/apps/", "site/content/recipes/apps/index.md"},
		{"/recipes/apps/counter", "site/content/recipes/apps/counter.md"},
	}
	for _, c := range cases {
		got := destFor("site", c.siteURL)
		if got != c.want {
			t.Errorf("destFor(site, %q) = %q, want %q", c.siteURL, got, c.want)
		}
	}
}

func TestComposeWithFrontmatter(t *testing.T) {
	got := composeWithFrontmatter("My Title", "https://github.com/x/y", "docs/foo.md", "v1.2.3", "abc123", nil, "Body content\n")

	wantLines := []string{
		`---`,
		`title: "My Title"`,
		`source_repo: "https://github.com/x/y"`,
		`source_path: "docs/foo.md"`,
		`source_ref: "v1.2.3"`,
		`source_commit: "abc123"`,
		`---`,
		``,
		`Body content`,
	}
	want := strings.Join(wantLines, "\n") + "\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestComposeStripsLeadingBlankLines(t *testing.T) {
	// After stripFrontmatter the remaining body often has a leading
	// blank line. Compose should swallow it so the page doesn't render
	// with an awkward gap above its first heading.
	got := composeWithFrontmatter("T", "r", "p", "v0", "c", nil, "\n\n# Hello\n")
	if !strings.Contains(got, "---\n\n# Hello") {
		t.Errorf("leading blanks not stripped: %q", got)
	}
	if strings.Contains(got, "---\n\n\n# Hello") {
		t.Errorf("too many leading blanks: %q", got)
	}
}

func TestComposeWithFrontmatter_PreservesUpstreamLvtShowSource(t *testing.T) {
	// lvt_show_source must round-trip as a YAML bool (true), not a
	// quoted string ("true") — tinkerdown reads it with bool semantics.
	upstream := map[string]any{"lvt_show_source": true}
	got := composeWithFrontmatter("T", "r", "p", "v0", "c", upstream, "body\n")
	if !strings.Contains(got, "lvt_show_source: true\n") {
		t.Errorf("expected unquoted bool, got:\n%s", got)
	}
	if strings.Contains(got, "lvt_show_source: \"true\"") {
		t.Errorf("emitted as string instead of bool:\n%s", got)
	}
}

func TestComposeWithFrontmatter_PreservesUpstreamDescription(t *testing.T) {
	upstream := map[string]any{"description": "A walkthrough of the counter app"}
	got := composeWithFrontmatter("T", "r", "p", "v0", "c", upstream, "body\n")
	if !strings.Contains(got, `description: "A walkthrough of the counter app"`+"\n") {
		t.Errorf("description not preserved:\n%s", got)
	}
}

func TestComposeWithFrontmatter_PreservesUpstreamSidebar(t *testing.T) {
	upstream := map[string]any{"sidebar": false}
	got := composeWithFrontmatter("T", "r", "p", "v0", "c", upstream, "body\n")
	if !strings.Contains(got, "sidebar: false\n") {
		t.Errorf("sidebar bool not preserved:\n%s", got)
	}
}

func TestComposeWithFrontmatter_DropsUnknownUpstreamFrontmatter(t *testing.T) {
	// Anything outside the explicit allowlist is dropped — proves the
	// docs site stays in control of its frontmatter contract.
	upstream := map[string]any{
		"description":     "kept",
		"random_key":      "dropped",
		"weight":          42,
		"lvt_show_source": true,
	}
	got := composeWithFrontmatter("T", "r", "p", "v0", "c", upstream, "body\n")
	if !strings.Contains(got, `description: "kept"`) {
		t.Errorf("allowlisted description not preserved")
	}
	if !strings.Contains(got, "lvt_show_source: true") {
		t.Errorf("allowlisted lvt_show_source not preserved")
	}
	if strings.Contains(got, "random_key") {
		t.Errorf("non-allowlisted key leaked through:\n%s", got)
	}
	if strings.Contains(got, "weight") {
		t.Errorf("non-allowlisted weight leaked through:\n%s", got)
	}
}

func TestComposeWithFrontmatter_OverridesUpstreamProvenanceKeys(t *testing.T) {
	// If upstream sets title/source_repo/source_path/source_ref/source_commit
	// (e.g. because it was previously synced from another mirror), sync's
	// values still win — those four are sync-owned.
	upstream := map[string]any{
		"title":         "stale upstream title",
		"source_repo":   "https://github.com/wrong/repo",
		"source_path":   "wrong/path.md",
		"source_ref":    "v0.0.0",
		"source_commit": "deadbeef",
	}
	got := composeWithFrontmatter("Sync Title", "https://github.com/right/repo", "right/path.md", "v9.9.9", "abc123", upstream, "body\n")
	if !strings.Contains(got, `title: "Sync Title"`) {
		t.Errorf("sync title should override upstream")
	}
	if strings.Contains(got, "stale upstream title") {
		t.Errorf("upstream title leaked through:\n%s", got)
	}
	if strings.Contains(got, "wrong/repo") || strings.Contains(got, "wrong/path.md") {
		t.Errorf("upstream provenance leaked through:\n%s", got)
	}
	if !strings.Contains(got, `source_ref: "v9.9.9"`) {
		t.Errorf("sync source_ref should override upstream")
	}
}

func TestParseFrontmatter(t *testing.T) {
	cases := []struct {
		name, in       string
		wantKeys       []string
		wantBodyPrefix string
	}{
		{
			name:           "no frontmatter",
			in:             "# Hello\nbody",
			wantKeys:       nil,
			wantBodyPrefix: "# Hello\nbody",
		},
		{
			name:           "basic frontmatter",
			in:             "---\ntitle: Hello\nlvt_show_source: true\n---\n# H1\n",
			wantKeys:       []string{"title", "lvt_show_source"},
			wantBodyPrefix: "# H1",
		},
		{
			name:           "frontmatter with description",
			in:             "---\ndescription: \"A page\"\n---\n\nbody\n",
			wantKeys:       []string{"description"},
			wantBodyPrefix: "\nbody",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, body, err := parseFrontmatter(c.in)
			if err != nil {
				t.Fatalf("parseFrontmatter: %v", err)
			}
			for _, k := range c.wantKeys {
				if _, ok := m[k]; !ok {
					t.Errorf("expected key %q in parsed map; got %v", k, m)
				}
			}
			if !strings.HasPrefix(body, c.wantBodyPrefix) {
				t.Errorf("body prefix mismatch: got %q want prefix %q", body, c.wantBodyPrefix)
			}
		})
	}
}

func TestLinkRewriter_PreservesIncludeFenceAttribute(t *testing.T) {
	// The link rewriter MUST NOT touch tinkerdown fence attributes.
	// `include="./_app/foo.go"` is byte-identical post-sync.
	cfg := &SourceOfTruth{
		Pages: []PageEntry{
			{SiteURL: "/x", SourceRepo: "https://github.com/livetemplate/livetemplate", SourcePath: "docs/x.md"},
		},
	}
	r := newLinkRewriter(cfg)
	body := "```go include=\"./_app/counter.go\" lines=\"5-15\" highlight=\"7\"\n```\n"
	if got := r.Rewrite(body); got != body {
		t.Errorf("include= fence attribute mutated:\nbefore: %q\nafter:  %q", body, got)
	}
}

func TestLinkRewriter_PreservesEmbedLvtBlock(t *testing.T) {
	cfg := &SourceOfTruth{
		Pages: []PageEntry{
			{SiteURL: "/x", SourceRepo: "https://github.com/livetemplate/livetemplate", SourcePath: "docs/x.md"},
		},
	}
	r := newLinkRewriter(cfg)
	body := "```embed-lvt path=\"/apps/counter/\" upstream=\"https://lt-firstapp.fly.dev\" session=\"counter-tour\" height=\"220px\"\n```\n"
	if got := r.Rewrite(body); got != body {
		t.Errorf("embed-lvt block mutated:\nbefore: %q\nafter:  %q", body, got)
	}
}

func TestLinkRewriter_PreservesShowSourceFlag(t *testing.T) {
	cfg := &SourceOfTruth{
		Pages: []PageEntry{
			{SiteURL: "/x", SourceRepo: "https://github.com/livetemplate/livetemplate", SourcePath: "docs/x.md"},
		},
	}
	r := newLinkRewriter(cfg)
	body := "```lvt show-source\n{{define \"main\"}}<p>{{.X}}</p>{{end}}\n```\n"
	if got := r.Rewrite(body); got != body {
		t.Errorf("show-source fence mutated:\nbefore: %q\nafter:  %q", body, got)
	}
}

func TestFilterByRepo(t *testing.T) {
	pages := []PageEntry{
		{SiteURL: "/a", SourceRepo: "https://github.com/livetemplate/livetemplate"},
		{SiteURL: "/b", SourceRepo: "https://github.com/livetemplate/lvt"},
		{SiteURL: "/c", SourceRepo: "https://github.com/livetemplate/livetemplate/"},
		{SiteURL: "/d", SourceRepo: "https://github.com/livetemplate/client"},
	}
	got := filterByRepo(pages, "https://github.com/livetemplate/livetemplate")
	if len(got) != 2 {
		t.Errorf("expected 2 livetemplate matches, got %d: %+v", len(got), got)
	}
	for _, p := range got {
		if p.SiteURL != "/a" && p.SiteURL != "/c" {
			t.Errorf("unexpected match: %+v", p)
		}
	}
}

func TestLinkRewriter(t *testing.T) {
	cfg := &SourceOfTruth{
		Pages: []PageEntry{
			{SiteURL: "/guides/x", SourceRepo: "https://github.com/livetemplate/livetemplate", SourcePath: "docs/guides/x.md"},
			{SiteURL: "/cli/y", SourceRepo: "https://github.com/livetemplate/lvt", SourcePath: "docs/y.md"},
		},
	}
	r := newLinkRewriter(cfg)

	body := `See [the X guide](https://github.com/livetemplate/livetemplate/blob/main/docs/guides/x.md) and [Y guide](https://github.com/livetemplate/lvt/blob/main/docs/y.md). External: https://github.com/golang/go.`
	got := r.Rewrite(body)

	// Mapped URLs become site-relative
	if !strings.Contains(got, "[the X guide](/guides/x)") {
		t.Errorf("X link not rewritten: %q", got)
	}
	if !strings.Contains(got, "[Y guide](/cli/y)") {
		t.Errorf("Y link not rewritten: %q", got)
	}
	// Unmapped URLs survive unchanged
	if !strings.Contains(got, "https://github.com/golang/go") {
		t.Errorf("external link should survive: %q", got)
	}
}

func TestLinkRewriter_AlsoHandlesEditURLs(t *testing.T) {
	// /edit/main/<path> URLs (used by edit-this-page links) should
	// rewrite to the same site URL as /blob/main/<path>.
	cfg := &SourceOfTruth{
		Pages: []PageEntry{
			{SiteURL: "/guides/x", SourceRepo: "https://github.com/livetemplate/livetemplate", SourcePath: "docs/guides/x.md"},
		},
	}
	r := newLinkRewriter(cfg)

	body := "Edit at https://github.com/livetemplate/livetemplate/edit/main/docs/guides/x.md please."
	got := r.Rewrite(body)
	if !strings.Contains(got, "/guides/x") || strings.Contains(got, "edit/main/docs/guides/x.md") {
		t.Errorf("edit URL not rewritten: %q", got)
	}
}
