// Verifies the /recipes/broadcasting recipe page renders and
// the embed-lvt block successfully inlines the broadcasting widget.
// The Send/NewMessage cross-tab broadcast flow is exercised by the
// upstream patterns_test.go (in examples/patterns/); this test asserts
// the page-level integration that B1 adds:
//
//   - the recipe markdown renders without error
//   - the embed mount div is present with the correct upstream path
//   - tinkerdown server-side inlines the widget (broadcasting markup
//     appears inside the mount), proving cmd/site returned 200 for
//     /recipes/ui-patterns/realtime/broadcasting
//   - no CSP violations or unexpected console errors
//
// Cross-tab broadcast is NOT verified here — that lives in the patterns
// app's own e2e (where Docker Chrome + per-test server lifecycle are
// already wired). When B2 relocates patterns_test.go into docs/, this
// test gets that coverage too. For B1, manual iPhone testing per the
// CLAUDE.md feedback covers the user-visible broadcast.
package e2e

import (
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

func TestRecipePatternsBroadcastingRenders(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)

	var (
		title         string
		hasMount      bool
		mountPath     string
		hasBroadcastH bool // the inlined widget renders <h3>Broadcasting</h3>
	)

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/broadcasting"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Title(&title),
		// tinkerdown renders embed-lvt as a div with class
		// "tinkerdown-embed-lvt" and a data-embed-path attribute.
		// Server-side, tinkerdown also fetches the upstream and inlines
		// its rendered HTML inside the mount — no iframe involved.
		chromedp.WaitVisible(`.tinkerdown-embed-lvt`, chromedp.ByQuery),
		chromedp.Evaluate(`!!document.querySelector('.tinkerdown-embed-lvt')`, &hasMount),
		chromedp.AttributeValue(`.tinkerdown-embed-lvt`, "data-embed-path", &mountPath, nil, chromedp.ByQuery),
		// The broadcasting upstream renders <h3>Broadcasting</h3>; if
		// the inline succeeded, that heading appears inside the mount.
		chromedp.Evaluate(`(() => {
			const m = document.querySelector('.tinkerdown-embed-lvt');
			if (!m) return false;
			const h = m.querySelector('h3');
			return h != null && h.textContent.trim() === 'Broadcasting';
		})()`, &hasBroadcastH),
	); err != nil {
		t.Fatalf("chromedp run failed: %v\nconsole errors: %v", err, consoleErrs())
	}

	if !strings.Contains(title, "Broadcasting") {
		t.Errorf("title = %q, want to contain %q", title, "Broadcasting")
	}
	if !hasMount {
		t.Fatalf("no .tinkerdown-embed-lvt mount on page")
	}
	if mountPath != "/recipes/ui-patterns/realtime/broadcasting" {
		t.Errorf("mount path = %q, want /recipes/ui-patterns/realtime/broadcasting", mountPath)
	}
	if !hasBroadcastH {
		t.Fatalf("broadcasting widget did not inline (no <h3>Broadcasting</h3> inside mount); console errors: %v", consoleErrs())
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
