// Verifies the /recipes/pubsub recipe page renders and the embed-lvt block
// for the Pubsub pattern successfully inlines its widget. The page carries
// several embeds (two synced counters in the "Watch it in action" demo plus
// the Pubsub pattern widget), so this test targets the pattern embed by its
// data-embed-path rather than the first mount on the page. The
// Send/NewMessage cross-tab fan-out flow is exercised by the upstream
// patterns_test.go (in examples/patterns/); this test asserts the page-level
// integration:
//
//   - the recipe markdown renders without error
//   - the Pubsub embed mount div is present with the correct upstream path
//   - tinkerdown server-side inlines the widget (Pubsub markup appears inside
//     the mount), proving cmd/site returned 200 for
//     /apps/ui-patterns/realtime/pubsub
//   - no CSP violations or unexpected console errors
//
// Cross-tab fan-out is NOT verified here — that lives in the patterns app's
// own e2e (where Docker Chrome + per-test server lifecycle are already wired).
package e2e

import (
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

func TestRecipePubsubRenders(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)

	const pubsubEmbed = `.tinkerdown-embed-lvt[data-embed-path="/apps/ui-patterns/realtime/pubsub"]`

	var (
		title      string
		hasPubsubH bool // the inlined widget renders <h3>Pubsub</h3>
	)

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/pubsub"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Title(&title),
		// tinkerdown renders embed-lvt as a div with class
		// "tinkerdown-embed-lvt" and a data-embed-path attribute.
		// Server-side, tinkerdown also fetches the upstream and inlines
		// its rendered HTML inside the mount — no iframe involved. The page
		// has multiple embeds; wait for the Pubsub pattern one specifically
		// (WaitVisible Fatals if it never mounts).
		chromedp.WaitVisible(pubsubEmbed, chromedp.ByQuery),
		// The Pubsub upstream renders <h3>Pubsub</h3>; if the inline
		// succeeded, that heading appears inside the mount.
		chromedp.Evaluate(`(() => {
			const m = document.querySelector('`+pubsubEmbed+`');
			if (!m) return false;
			const h = m.querySelector('h3');
			return h != null && h.textContent.trim() === 'Pubsub';
		})()`, &hasPubsubH),
	); err != nil {
		t.Fatalf("chromedp run failed: %v\nconsole errors: %v", err, consoleErrs())
	}

	if !strings.Contains(title, "Pubsub") {
		t.Errorf("title = %q, want to contain %q", title, "Pubsub")
	}
	if !hasPubsubH {
		t.Fatalf("Pubsub widget did not inline (no <h3>Pubsub</h3> inside mount); console errors: %v", consoleErrs())
	}

	// Surface CSP / frame-src / fetch failures explicitly so a regression
	// (e.g. forgetting to allow the upstream in tinkerdown.yaml) fails
	// loud instead of silent.
	for _, e := range consoleErrs() {
		low := strings.ToLower(e)
		if strings.Contains(low, "content security policy") ||
			strings.Contains(low, "frame-src") ||
			strings.Contains(low, "refused to frame") ||
			strings.Contains(low, "failed to load") {
			t.Errorf("CSP/load error: %s", e)
		}
	}
}
