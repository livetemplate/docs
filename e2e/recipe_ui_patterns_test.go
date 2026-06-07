// Verifies the per-pattern "tiny recipe" docs pages: each renders a brief
// description, embeds and inlines the live pattern app, and shows the template +
// Go source snippets pulled from the examples/patterns package via region=
// includes.
package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestUIPatternDocPageRenders walks one representative recipe page per category
// and asserts, for each:
//   - the live app embeds and INLINES (its <h3> is present in the embed region),
//   - the embed is rendered, not raw (no leaked {{…}} template tokens),
//   - a known handler symbol from the region= include is on the page, and
//   - no region: marker leaked into the rendered snippet.
//
// The handler-symbol assertion is content-based (symbol text, not line numbers),
// so it catches region-extraction drift in each category's grouped
// handlers_*.go file — if a region marker is moved, renamed, or dropped, the
// snippet stops containing its symbol and this fails.
func TestUIPatternDocPageRenders(t *testing.T) {
	cases := []struct {
		page    string // path under /recipes/ui-patterns/ (also the /apps/ui-patterns/ embed path)
		wantSym string // a symbol the region= handler include must surface
	}{
		{"forms/click-to-edit", "func clickToEditHandler"},
		{"lists/delete-row", "func deleteRowHandler"},
		{"search/active-search", "func activeSearchHandler"},
		{"loading/lazy-loading", "func lazyLoadingHandler"},
		{"navigation/modal-dialog", "func modalDialogHandler"},
		{"feedback/animations", "func animationsHandler"},
		{"realtime/multi-user-sync", "func multiUserSyncHandler"},
	}

	for _, tc := range cases {
		t.Run(tc.page, func(t *testing.T) {
			ctx, cancel := newCtx(t)
			defer cancel()
			consoleErrs := captureConsoleErrors(ctx)

			embed := `.tinkerdown-embed-lvt[data-embed-path="/apps/ui-patterns/` + tc.page + `"]`
			var embedText, bodyText string
			if err := chromedp.Run(ctx,
				chromedp.Navigate(baseURL()+"/recipes/ui-patterns/"+tc.page),
				chromedp.WaitVisible(`h1`, chromedp.ByQuery),
				// The embed mounts + inlines the live app asynchronously. Poll on
				// the real condition (its <h3> present) rather than a fixed sleep —
				// faster on green, no flake on a slow box. Inlining is thus
				// guaranteed by the time we read embedText below.
				chromedp.Poll(
					`!!(document.querySelector('`+embed+`')||{}).querySelector?.('h3')`,
					nil,
					chromedp.WithPollingTimeout(20*time.Second),
				),
				chromedp.Evaluate(`((document.querySelector('`+embed+`')||{}).innerText)||''`, &embedText),
				chromedp.Text(`body`, &bodyText, chromedp.ByQuery),
			); err != nil {
				t.Fatalf("embed did not inline: %v\nconsole: %v", err, consoleErrs())
			}

			if strings.Contains(embedText, "{{") {
				t.Errorf("embed shows raw template tokens (not rendered):\n%s", embedText)
			}
			if !strings.Contains(bodyText, tc.wantSym) {
				t.Errorf("missing %q (region include drift?)", tc.wantSym)
			}
			if strings.Contains(bodyText, "region:") {
				t.Errorf("a region: marker leaked into the rendered page")
			}
		})
	}
}
