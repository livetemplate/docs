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
		{"/examples/", "site/content/examples/index.md"},
		{"/examples/counter", "site/content/examples/counter.md"},
	}
	for _, c := range cases {
		got := destFor("site", c.siteURL)
		if got != c.want {
			t.Errorf("destFor(site, %q) = %q, want %q", c.siteURL, got, c.want)
		}
	}
}

func TestComposeWithFrontmatter(t *testing.T) {
	got := composeWithFrontmatter("My Title", "https://github.com/x/y", "docs/foo.md", "abc123", "Body content\n")

	wantLines := []string{
		`---`,
		`title: "My Title"`,
		`source_repo: "https://github.com/x/y"`,
		`source_path: "docs/foo.md"`,
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
	got := composeWithFrontmatter("T", "r", "p", "c", "\n\n# Hello\n")
	if !strings.Contains(got, "---\n\n# Hello") {
		t.Errorf("leading blanks not stripped: %q", got)
	}
	if strings.Contains(got, "---\n\n\n# Hello") {
		t.Errorf("too many leading blanks: %q", got)
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
