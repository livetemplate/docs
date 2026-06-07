package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

// uiPatternCategories is the locked nesting: the 33 UI patterns group under
// these 7 collapsible category sub-groups inside the "UI Patterns" section.
var uiPatternCategories = []string{
	"Forms & Editing", "Lists & Data", "Search & Filtering",
	"Loading & Progress", "Dialogs, Tabs & Navigation",
	"Visual Feedback", "Real-Time & Multi-User",
}

// detailsOpenExpr builds JS that returns whether the <details> whose summary
// (of class cls) reads exactly title is open.
func detailsOpenExpr(cls, title string) string {
	t, _ := json.Marshal(title)
	return fmt.Sprintf(
		`(() => { const el=[...document.querySelectorAll(%q)].find(e=>e.textContent.trim()===%s); return el ? el.closest('details').open : null })()`,
		"."+cls, string(t))
}

// TestUIPatternsSidebarNesting verifies the 33 patterns render as nested,
// collapsible category groups (not a flat list), that the section + group
// holding the active page auto-open while a sibling group stays closed, and
// that the top-level IA section count is unchanged. Captures a sidebar
// screenshot on a non-pattern page so Learn/Concepts/Recipes can be eyeballed
// for collapse-styling regressions.
func TestUIPatternsSidebarNesting(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()
	consoleErrs := captureConsoleErrors(ctx)

	var h1 string
	var sectionTitles, groupTitles []string
	var uiPatternsOpen, formsOpen, listsOpen any

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/ui-patterns/forms/click-to-edit"),
		chromedp.WaitVisible(".nav-section-title", chromedp.ByQuery),
		chromedp.Text("h1", &h1, chromedp.ByQuery),
		chromedp.Evaluate(`[...document.querySelectorAll('.nav-section-title')].map(e=>e.textContent.trim())`, &sectionTitles),
		chromedp.Evaluate(`[...document.querySelectorAll('.nav-group-title')].map(e=>e.textContent.trim())`, &groupTitles),
		chromedp.Evaluate(detailsOpenExpr("nav-section-title", "UI Patterns"), &uiPatternsOpen),
		chromedp.Evaluate(detailsOpenExpr("nav-group-title", "Forms & Editing"), &formsOpen),
		chromedp.Evaluate(detailsOpenExpr("nav-group-title", "Lists & Data"), &listsOpen),
	); err != nil {
		t.Fatalf("load nested pattern page: %v", err)
	}

	// Page actually served (not 303'd back to home).
	if !strings.Contains(h1, "Click to Edit") {
		t.Errorf("h1 = %q, want it to contain \"Click to Edit\" (nested leaf may not be served)", h1)
	}

	// Top-level IA unchanged: "UI Patterns" is still one section, not 33.
	if !slices.Contains(sectionTitles, "UI Patterns") {
		t.Errorf("sidebar sections %v missing \"UI Patterns\"", sectionTitles)
	}
	if len(sectionTitles) > 10 {
		t.Errorf("sidebar has %d top-level sections (%v) — patterns leaked to top level instead of nesting", len(sectionTitles), sectionTitles)
	}

	// All 7 categories present as groups.
	for _, cat := range uiPatternCategories {
		if !slices.Contains(groupTitles, cat) {
			t.Errorf("category group %q missing from sidebar groups %v", cat, groupTitles)
		}
	}

	// Auto-open: the active page's section + group open; sibling group closed.
	if uiPatternsOpen != true {
		t.Errorf("UI Patterns section open = %v, want true (holds active page)", uiPatternsOpen)
	}
	if formsOpen != true {
		t.Errorf("Forms & Editing group open = %v, want true (holds active page)", formsOpen)
	}
	if listsOpen != false {
		t.Errorf("Lists & Data group open = %v, want false (does not hold active page)", listsOpen)
	}

	// Screenshot the sidebar on a plain page to inspect section collapse styling.
	var png []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/getting-started/introduction"),
		chromedp.WaitVisible(".nav-section-title", chromedp.ByQuery),
		chromedp.Screenshot("#tinkerdown-sidebar", &png, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("sidebar screenshot: %v", err)
	}
	out := filepath.Join(t.TempDir(), "sidebar-nested.png")
	if err := os.WriteFile(out, png, 0o644); err != nil {
		t.Fatalf("write screenshot: %v", err)
	}
	t.Logf("sidebar screenshot: %s (%d bytes)", out, len(png))

	for _, e := range consoleErrs() {
		low := strings.ToLower(e)
		if strings.Contains(low, "content security policy") || strings.Contains(low, "failed to load") {
			t.Errorf("console error on nested pattern page: %s", e)
		}
	}
}

// TestUIPatternsAllLeavesServed is the registration invariant: every one of
// the 33 nested pattern pages must resolve to 200. A nested leaf missed by the
// recursive nav builder would silently 303 to home in site mode.
func TestUIPatternsAllLeavesServed(t *testing.T) {
	leaves := []string{
		"forms/click-to-edit", "forms/edit-row", "forms/inline-validation",
		"forms/bulk-update", "forms/reset-input", "forms/file-upload", "forms/preserve-inputs",
		"lists/delete-row", "lists/click-to-load", "lists/infinite-scroll",
		"lists/value-select", "lists/sortable", "lists/large-table",
		"search/active-search", "search/url-filters",
		"loading/lazy-loading", "loading/progress-bar", "loading/async-operations",
		"navigation/modal-dialog", "navigation/confirm-dialog", "navigation/tabs",
		"navigation/spa-navigation", "navigation/keyboard-shortcuts",
		"feedback/animations", "feedback/loading-states", "feedback/highlight", "feedback/flash-messages",
		"realtime/multi-user-sync", "realtime/broadcasting", "realtime/presence",
		"realtime/reconnection", "realtime/live-preview", "realtime/server-push",
	}
	if len(leaves) != 33 {
		t.Fatalf("expected 33 leaves, listed %d", len(leaves))
	}

	// Don't follow redirects — a 303 to home is the failure mode we're hunting.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	for _, leaf := range leaves {
		url := baseURL() + "/recipes/ui-patterns/" + leaf
		resp, err := client.Get(url)
		if err != nil {
			t.Errorf("GET %s: %v", url, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s = %d, want 200 (nested leaf not registered/served)", url, resp.StatusCode)
		}
	}
}
