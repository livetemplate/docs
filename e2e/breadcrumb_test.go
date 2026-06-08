// Verifies that breadcrumb category pills navigate to their category index
// page rather than 303-redirecting to the home/landing page.
//
// Regression test for the bug where a section crumb linked to a bare directory
// path (e.g. "/getting-started", or "/recipes" when the index is served at
// "/recipes/") which tinkerdown's site mode redirects to the landing page. The
// fix resolves each crumb to its real servable route (the trailing-slash index
// route) and the Learn/Concepts/Reference sections gained overview pages so
// every pill has a destination.
//
// Run locally against a tinkerdown serve of content/:
//
//	E2E_BASE_URL=http://localhost:8088 go test ./e2e -run TestBreadcrumbPills
package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

func TestBreadcrumbPillsNavigate(t *testing.T) {
	cases := []struct {
		name     string
		deepPage string // a page nested under the section
		pill     string // breadcrumb pill text to click
		wantPath string // path the click must land on
		wantH1   string // h1 of the landed index page
	}{
		{"Learn", "/getting-started/install", "Learn", "/getting-started/", "Learn"},
		{"Concepts", "/guides/standard-html-reactivity", "Concepts", "/guides/", "Concepts"},
		{"Reference", "/reference/api", "Reference", "/reference/", "Reference"},
		{"Recipes", "/recipes/apps/counter", "Recipes", "/recipes/", "Recipes"},
		{"Apps", "/recipes/apps/counter", "Apps", "/recipes/apps/", "App Recipes"},
		{"UIPatterns", "/recipes/ui-patterns/forms/click-to-edit", "UI Patterns", "/recipes/ui-patterns/", "UI Pattern Recipes"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := newCtx(t)
			defer cancel()
			consoleErrs := captureConsoleErrors(ctx)

			// Match the breadcrumb pill by its visible text.
			pill := fmt.Sprintf(`//nav[contains(@class,'breadcrumbs')]//a[normalize-space()=%q]`, tc.pill)

			var href, gotPath, gotH1 string
			if err := chromedp.Run(ctx,
				chromedp.Navigate(baseURL()+tc.deepPage),
				chromedp.WaitVisible(pill, chromedp.BySearch),
				chromedp.AttributeValue(pill, "href", &href, nil, chromedp.BySearch),
				chromedp.Click(pill, chromedp.BySearch),
				chromedp.WaitVisible("h1", chromedp.ByQuery),
				chromedp.Evaluate(`location.pathname`, &gotPath),
				chromedp.Text("h1", &gotH1, chromedp.ByQuery),
			); err != nil {
				t.Fatalf("%s: click %q breadcrumb pill from %s: %v", tc.name, tc.pill, tc.deepPage, err)
			}

			if href != tc.wantPath {
				t.Errorf("%s pill href = %q, want %q", tc.name, href, tc.wantPath)
			}
			// Landing on "/" is the original bug (303 to the landing page).
			if gotPath != tc.wantPath {
				t.Errorf("%s pill click landed on %q, want %q (a redirect to the landing page is the bug)", tc.name, gotPath, tc.wantPath)
			}
			if strings.TrimSpace(gotH1) != tc.wantH1 {
				t.Errorf("%s landed h1 = %q, want %q", tc.name, strings.TrimSpace(gotH1), tc.wantH1)
			}
			for _, e := range consoleErrs() {
				low := strings.ToLower(e)
				if strings.Contains(low, "content security policy") || strings.Contains(low, "failed to load") {
					t.Errorf("%s console error: %s", tc.name, e)
				}
			}
		})
	}
}
