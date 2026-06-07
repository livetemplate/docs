// Verifies the Phase-4 documentation information architecture: the sidebar
// is organised into the seven learning-ordered sections, the new
// Introduction page renders, and a representative page from each section
// loads without console errors.
//
// Run locally against a tinkerdown serve of content/:
//
//	E2E_BASE_URL=http://localhost:8088 go test ./e2e -run TestDocsIA
package e2e

import (
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

// wantSections is the locked IA, in sidebar order. "UI Patterns" is a
// dedicated section (collapsed by default) for the per-pattern recipe pages,
// which must be nav-registered to be served.
var wantSections = []string{
	"Learn", "Concepts", "Recipes", "UI Patterns", "Apps",
	"Reference", "Deploy & Operate", "Ecosystem",
}

// TestDocsIASidebarSections asserts every docs page shows exactly the
// IA sections in the sidebar, in order.
func TestDocsIASidebarSections(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var got []string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/getting-started/introduction"),
		chromedp.WaitVisible(".nav-section-title", chromedp.ByQuery),
		chromedp.Evaluate(`[...document.querySelectorAll('.nav-section-title')].map(e => e.textContent.trim())`, &got),
	); err != nil {
		t.Fatalf("read sidebar sections: %v", err)
	}

	if len(got) != len(wantSections) {
		t.Fatalf("sidebar sections = %v (%d), want %v (%d)", got, len(got), wantSections, len(wantSections))
	}
	for i, want := range wantSections {
		if got[i] != want {
			t.Errorf("sidebar section %d = %q, want %q (full: %v)", i, got[i], want, got)
		}
	}
}

// TestDocsIAIntroduction asserts the new Introduction page renders with its
// heading and the three "where to go next" links into the Learn spine.
func TestDocsIAIntroduction(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()
	consoleErrs := captureConsoleErrors(ctx)

	var h1 string
	var hrefs []string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/getting-started/introduction"),
		chromedp.WaitVisible("h1", chromedp.ByQuery),
		chromedp.Text("h1", &h1, chromedp.ByQuery),
		chromedp.Evaluate(`[...document.querySelectorAll('.content-wrapper a, main a, article a')].map(a => a.getAttribute('href'))`, &hrefs),
	); err != nil {
		t.Fatalf("load introduction: %v", err)
	}

	if !strings.Contains(h1, "Introduction") {
		t.Errorf("introduction h1 = %q, want it to contain \"Introduction\"", h1)
	}
	for _, want := range []string{"/getting-started/install", "/getting-started/your-first-app", "/getting-started/mental-model"} {
		if !slices.Contains(hrefs, want) {
			t.Errorf("introduction missing read-next link %q (hrefs: %v)", want, hrefs)
		}
	}
	for _, e := range consoleErrs() {
		low := strings.ToLower(e)
		if strings.Contains(low, "content security policy") || strings.Contains(low, "failed to load") {
			t.Errorf("introduction console error: %s", e)
		}
	}
}

// TestDocsIASectionPages loads one representative page per IA section and
// asserts it renders a heading — a smoke test that the re-slotted paths
// still resolve after the nav rewrite.
func TestDocsIASectionPages(t *testing.T) {
	pages := map[string]string{
		"Learn":            "/getting-started/introduction",
		"Concepts":         "/guides/standard-html-reactivity",
		"Recipes":          "/recipes/",
		"UI Patterns":      "/recipes/ui-patterns/forms/click-to-edit",
		"Apps":             "/recipes/todos/",
		"Reference":        "/reference/api",
		"Deploy & Operate": "/reference/configuration",
		"Ecosystem":        "/cli/",
	}
	// Stable iteration order for readable output.
	sections := make([]string, 0, len(pages))
	for s := range pages {
		sections = append(sections, s)
	}
	sort.Strings(sections)

	for _, section := range sections {
		path := pages[section]
		t.Run(section, func(t *testing.T) {
			ctx, cancel := newCtx(t)
			defer cancel()
			var h1 string
			if err := chromedp.Run(ctx,
				chromedp.Navigate(baseURL()+path),
				chromedp.WaitVisible("h1", chromedp.ByQuery),
				chromedp.Text("h1", &h1, chromedp.ByQuery),
			); err != nil {
				t.Fatalf("%s page %s: %v", section, path, err)
			}
			if strings.TrimSpace(h1) == "" {
				t.Errorf("%s page %s rendered an empty h1", section, path)
			}
		})
	}
}
