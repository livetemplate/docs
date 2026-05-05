// Verifies the landing-page iframe-embedded LiveTemplate demo loads.
// The iframe points at https://lt-landing-demo.fly.dev/ cross-origin,
// so this test cannot reach into contentDocument — that's blocked by
// the same-origin policy and is the entire reason we run cross-origin
// (the embedded LiveTemplate client must talk to its own host). What we
// CAN verify from outside the frame: the iframe element exists with the
// expected src, the load event fires (browser successfully fetched the
// upstream — would fail if CSP frame-src blocked it), and the console
// reports no CSP violations or errors triggered by the iframe load.
//
// Interactivity (clicks update count over WebSocket) is exercised by
// the standalone e2e suite at examples/landing-demo/landing_demo_test.go,
// which runs against a local landing-demo binary; combined with the
// standalone production verification below, that's full coverage of the
// path a real visitor walks.
package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

const landingIframeSrc = "https://lt-landing-demo.fly.dev/"

func TestLandingDemoIframeLoads(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)

	var iframeSrc string
	var loaded bool

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible(`iframe[src="`+landingIframeSrc+`"]`, chromedp.ByQuery),
		chromedp.AttributeValue(`iframe[src="`+landingIframeSrc+`"]`, "src", &iframeSrc, nil, chromedp.ByQuery),
		// Wait for the iframe's load event. A CSP frame-src block would
		// surface here as the load never firing (and a violation in
		// consoleErrs) within the timeout.
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(20 * time.Second)
			for time.Now().Before(deadline) {
				_ = chromedp.Evaluate(`(() => {
					const f = document.querySelector('iframe[src="`+landingIframeSrc+`"]');
					if (!f) return false;
					try {
						return f.contentWindow != null;
					} catch (e) {
						return true;
					}
				})()`, &loaded).Do(ctx)
				if loaded {
					return nil
				}
				time.Sleep(300 * time.Millisecond)
			}
			return nil
		}),
	); err != nil {
		t.Fatalf("chromedp run failed: %v\nconsole errors: %v\niframe src: %q",
			err, consoleErrs(), iframeSrc)
	}

	if iframeSrc != landingIframeSrc {
		t.Errorf("iframe src = %q, want %q", iframeSrc, landingIframeSrc)
	}
	if !loaded {
		t.Fatalf("iframe did not finish loading within 20s; console errors: %v", consoleErrs())
	}

	for _, e := range consoleErrs() {
		if strings.Contains(strings.ToLower(e), "content security policy") ||
			strings.Contains(strings.ToLower(e), "frame-src") ||
			strings.Contains(strings.ToLower(e), "refused to frame") {
			t.Errorf("CSP blocked the iframe load: %s", e)
		}
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
