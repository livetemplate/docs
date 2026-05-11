// Verifies the landing-page embedded LiveTemplate counter loads.
// The home page used to iframe a separate cross-origin landing demo; it
// now uses tinkerdown's inline embed-lvt proxy against the docs-site
// recipes binary. This test asserts the mount is present and the counter
// widget is server-side inlined.
package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func TestLandingCounterEmbedLoads(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)

	var (
		mountPath string
		hasMount  bool
		hasCount  bool
	)

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible(`.tinkerdown-embed-lvt[data-embed-path="/apps/counter/"]`, chromedp.ByQuery),
		chromedp.Evaluate(`!!document.querySelector('.tinkerdown-embed-lvt[data-embed-path="/apps/counter/"]')`, &hasMount),
		chromedp.AttributeValue(`.tinkerdown-embed-lvt[data-embed-path="/apps/counter/"]`, "data-embed-path", &mountPath, nil, chromedp.ByQuery),
		chromedp.Evaluate(`(() => {
			const m = document.querySelector('.tinkerdown-embed-lvt[data-embed-path="/apps/counter/"]');
			if (!m) return false;
			const h = m.querySelector('h1');
			return h != null && h.textContent.trim().startsWith('Counter:');
		})()`, &hasCount),
	); err != nil {
		t.Fatalf("chromedp run failed: %v\nconsole errors: %v", err, consoleErrs())
	}

	if !hasMount {
		t.Fatalf("no landing counter embed mount on page")
	}
	if mountPath != "/apps/counter/" {
		t.Errorf("mount path = %q, want /apps/counter/", mountPath)
	}
	if !hasCount {
		t.Fatalf("counter widget did not inline (no Counter heading inside mount); console errors: %v", consoleErrs())
	}

	for _, e := range consoleErrs() {
		low := strings.ToLower(e)
		if strings.Contains(low, "content security policy") ||
			strings.Contains(low, "failed to load") {
			t.Errorf("embed load error: %s", e)
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
