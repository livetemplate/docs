// Verifies the landing-page lvt block actually populates and is interactive.
// Runs against E2E_BASE_URL — point it at a local tinkerdown serve to test
// before deploy.
package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func TestLandingLvtDemoPopulatesAndIsInteractive(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)

	var bodyText string
	var firstCheckboxBefore, firstCheckboxAfter bool

	// Wait for the lvt block to populate (initial render is a "Connecting..."
	// placeholder; the WebSocket connect replaces it with the actual list).
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		// Poll body text until our seed content lands. ~5s is generous.
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(8 * time.Second)
			for time.Now().Before(deadline) {
				_ = chromedp.Text("body", &bodyText, chromedp.ByQuery).Do(ctx)
				if strings.Contains(bodyText, "Click me — I'm interactive") {
					return nil
				}
				time.Sleep(250 * time.Millisecond)
			}
			return nil
		}),
		chromedp.Evaluate(`document.querySelector('[lvt-source="landing_tasks"] input[type=checkbox]').checked`, &firstCheckboxBefore),
		chromedp.Click(`[lvt-source="landing_tasks"] input[type=checkbox]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(`document.querySelector('[lvt-source="landing_tasks"] input[type=checkbox]').checked`, &firstCheckboxAfter),
	); err != nil {
		t.Fatalf("chromedp run failed: %v\nconsole errors: %v", err, consoleErrs())
	}

	if !strings.Contains(bodyText, "Click me — I'm interactive") {
		t.Fatalf("lvt block did not populate within timeout; body text = %q", truncate(bodyText, 400))
	}
	if firstCheckboxBefore == firstCheckboxAfter {
		t.Errorf("first checkbox state unchanged after click (was %v, still %v); console errors: %v",
			firstCheckboxBefore, firstCheckboxAfter, consoleErrs())
	}
}

// captureConsoleErrors streams browser console errors so a failed assertion
// can report what the page logged. Returns a closure that yields the captured
// list at call time.
func captureConsoleErrors(ctx context.Context) func() []string {
	var errs []string
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*runtime.EventConsoleAPICalled); ok && e.Type == runtime.APITypeError {
			parts := make([]string, 0, len(e.Args))
			for _, a := range e.Args {
				parts = append(parts, string(a.Value))
			}
			errs = append(errs, strings.Join(parts, " "))
		}
	})
	return func() []string { return errs }
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
