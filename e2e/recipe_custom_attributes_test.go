// Verifies the /recipes/custom-attributes page: the embed mounts, and the
// app's own lvt-x:* handlers still work once tinkerdown has inlined the app
// into the docs page. The handlers are registered from an inline <script> in
// the app's <head>, so this is the check that inlining keeps them. The
// handlers' behaviour in depth (one frame per click, the sweep when the server
// removes lvt-x:hotkey) is covered by examples/custom-attributes' own e2e.
package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func TestRecipeCustomAttributesEmbed(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)
	// The census warns rather than errors, so captureConsoleErrors misses it.
	var warnings []string
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*runtime.EventConsoleAPICalled); ok && e.Type == runtime.APITypeWarning {
			for _, a := range e.Args {
				warnings = append(warnings, string(a.Value))
			}
		}
	})

	const mount = `.tinkerdown-embed-lvt[data-embed-path="/apps/custom-attributes/"]`
	rating := func(n string) string {
		return `(() => { const m = document.querySelector('` + mount + `');` +
			` const r = m && m.querySelector('#rating');` +
			` return !!r && r.textContent.trim() === 'Rated ` + n + ` of 5.'; })()`
	}

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/custom-attributes"),
		chromedp.WaitVisible(mount+` [data-star="3"]`, chromedp.ByQuery),
		chromedp.Click(mount+` [data-star="3"]`, chromedp.ByQuery),
		chromedp.Poll(rating("3"), nil, chromedp.WithPollingTimeout(15*time.Second)),
	); err != nil {
		t.Fatalf("lvt-x:rating did not rate inside the embed: %v\nconsole errors: %v", err, consoleErrs())
	}

	// The low-level handler listens on document, so a key press anywhere on
	// the docs page reaches it.
	if err := chromedp.Run(ctx,
		chromedp.KeyEvent("5"),
		chromedp.Poll(rating("5"), nil, chromedp.WithPollingTimeout(15*time.Second)),
	); err != nil {
		t.Fatalf("lvt-x:hotkey did not rate inside the embed: %v\nconsole errors: %v", err, consoleErrs())
	}

	// The census runs after window load; give it a moment, then expect no
	// warning that names one of this app's attributes.
	time.Sleep(500 * time.Millisecond)
	for _, w := range warnings {
		if strings.Contains(w, "no attribute handler claims it") && strings.Contains(w, "lvt-x:") {
			t.Errorf("census warning for the app's own attribute: %s", w)
		}
	}
}
