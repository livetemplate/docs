// Verifies the per-pattern "tiny recipe" docs pages: each renders a brief
// description, embeds and inlines the live pattern app, and shows the template +
// Go source snippets pulled from the examples/patterns package.
package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestUIPatternDocPageRenders checks the pilot page (Forms › Click to Edit):
//   - the live app embeds and INLINES (its <h3> is present in the embed region),
//   - the embed is rendered, not raw (no leaked {{…}} template tokens),
//   - the template snippet (raw {{if .Editing}}) and the Go snippet
//     (func clickToEditHandler) are both visible on the page.
//
// The snippet assertions are content-based (symbol text, not line numbers), so
// they also catch include line-range drift in the grouped handlers_*.go files.
func TestUIPatternDocPageRenders(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()
	consoleErrs := captureConsoleErrors(ctx)

	embed := `.tinkerdown-embed-lvt[data-embed-path="/apps/ui-patterns/forms/click-to-edit"]`
	var embedText, bodyText string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/ui-patterns/forms/click-to-edit"),
		chromedp.WaitVisible(`h1`, chromedp.ByQuery),
		// The embed mounts + inlines the live app asynchronously. Poll on the
		// real condition (its <h3> present) rather than a fixed sleep — faster
		// on green, no flake on a slow box. Inlining is thus guaranteed by the
		// time we read embedText below.
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
	for _, want := range []string{"func clickToEditHandler", "{{if .Editing}}", "Click to Edit"} {
		if !strings.Contains(bodyText, want) {
			t.Errorf("doc page missing %q (snippet/include drift?)", want)
		}
	}
}
