// Verifies the landing-page iframe-embedded LiveTemplate demo loads and
// is interactive. The iframe points at /demo/counter/ which is proxied
// same-origin to the deployed lt-landing-demo app. Runs against
// E2E_BASE_URL — point it at a local tinkerdown serve to test before
// deploy.
package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func TestLandingDemoIframeLoadsAndIsInteractive(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)

	var iframeSrc string
	var beforeCount, afterCount string

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible(`iframe[src="/demo/counter/"]`, chromedp.ByQuery),
		chromedp.AttributeValue(`iframe[src="/demo/counter/"]`, "src", &iframeSrc, nil, chromedp.ByQuery),
		// Read the initial counter value from inside the iframe. Same-origin
		// proxy means we can reach the iframe's contentDocument; with cross-
		// origin this would be blocked, so this assertion also catches a
		// proxy misconfiguration.
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(15 * time.Second)
			for time.Now().Before(deadline) {
				_ = chromedp.Evaluate(`document.querySelector('iframe[src="/demo/counter/"]').contentDocument.querySelector('.count')?.textContent || ""`, &beforeCount).Do(ctx)
				if beforeCount != "" {
					return nil
				}
				time.Sleep(300 * time.Millisecond)
			}
			return nil
		}),
		// Click the increment button inside the iframe.
		chromedp.Evaluate(`document.querySelector('iframe[src="/demo/counter/"]').contentDocument.querySelector('button[name=increment]').click()`, nil),
		chromedp.Sleep(800*time.Millisecond),
		chromedp.Evaluate(`document.querySelector('iframe[src="/demo/counter/"]').contentDocument.querySelector('.count').textContent`, &afterCount),
	); err != nil {
		t.Fatalf("chromedp run failed: %v\nconsole errors: %v\niframe src: %q",
			err, consoleErrs(), iframeSrc)
	}

	if beforeCount == "" {
		t.Fatalf("iframe counter did not render within 15s; iframe src = %q; console errors: %v",
			iframeSrc, consoleErrs())
	}
	if beforeCount == afterCount {
		t.Errorf("counter did not change after Increment click (before=%q, after=%q); console errors: %v",
			beforeCount, afterCount, consoleErrs())
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
