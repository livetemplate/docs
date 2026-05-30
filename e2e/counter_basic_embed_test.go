// Verifies the homepage "Try it" embed loads the no-pubsub counter.
// The landing page now introduces reactivity progressively: the first
// embed is /apps/counter-basic/ (single-session reactivity), and a
// later "Next level" section embeds /apps/counter/ (cross-tab pubsub,
// covered by landing_demo_test.go). This test asserts the basic embed
// is present and server-side inlined, mirroring the broadcasting and
// landing-counter smoke tests.
//
// Interaction (clicking +1, WebSocket patch, peer fan-out) is verified
// against a local cmd/site stack during development, not this remote
// staging smoke suite — cold-start staging makes WS-bootstrap timing
// flaky, so the established pattern (see recipe_patterns_broadcasting_test.go)
// keeps CI assertions to render/inline + console-error checks.
package e2e

import (
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

func TestLandingCounterBasicEmbedLoads(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)

	var (
		mountPath string
		hasMount  bool
		hasCount  bool
		mountHTML string
	)

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible(`.tinkerdown-embed-lvt[data-embed-path="/apps/counter-basic/"]`, chromedp.ByQuery),
		chromedp.Evaluate(`!!document.querySelector('.tinkerdown-embed-lvt[data-embed-path="/apps/counter-basic/"]')`, &hasMount),
		chromedp.AttributeValue(`.tinkerdown-embed-lvt[data-embed-path="/apps/counter-basic/"]`, "data-embed-path", &mountPath, nil, chromedp.ByQuery),
		chromedp.Evaluate(`(() => {
			const m = document.querySelector('.tinkerdown-embed-lvt[data-embed-path="/apps/counter-basic/"]');
			if (!m) return false;
			const h = m.querySelector('h1');
			return h != null && h.textContent.trim().startsWith('Counter:');
		})()`, &hasCount),
		// Capture the inlined markup so a failed inline assertion reports
		// what the server actually returned, not just "false".
		chromedp.Evaluate(`(() => {
			const m = document.querySelector('.tinkerdown-embed-lvt[data-embed-path="/apps/counter-basic/"]');
			return m ? m.innerHTML : '';
		})()`, &mountHTML),
	); err != nil {
		t.Fatalf("chromedp run failed: %v\nconsole errors: %v", err, consoleErrs())
	}

	if !hasMount {
		t.Fatalf("no basic-counter embed mount on landing page")
	}
	if mountPath != "/apps/counter-basic/" {
		t.Errorf("mount path = %q, want /apps/counter-basic/", mountPath)
	}
	if !hasCount {
		t.Fatalf("basic counter widget did not inline (no Counter heading inside mount)\nmount HTML: %s\nconsole errors: %v", mountHTML, consoleErrs())
	}

	for _, e := range consoleErrs() {
		low := strings.ToLower(e)
		if strings.Contains(low, "content security policy") ||
			strings.Contains(low, "failed to load") {
			t.Errorf("embed load error: %s", e)
		}
	}
}
