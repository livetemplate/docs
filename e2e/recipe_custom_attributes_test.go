// Verifies the /recipes/custom-attributes page: the embed mounts, and the
// app's own lvt-x:* handlers still work once tinkerdown has inlined the app
// into the docs page. The handlers are registered from an inline <script> in
// the app's <head>, so this is the check that inlining keeps them. The
// handlers' behaviour in depth (one frame per click, the sweep when the server
// removes lvt-x:hotkey) is covered by examples/custom-attributes' own e2e.
package e2e

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func TestRecipeCustomAttributesEmbed(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)
	// The census warns rather than errors, so captureConsoleErrors misses it.
	var warnings []string
	// Handlers attach when the first render arrives over the WebSocket, so a
	// click before then does nothing. On a remote site the button is visible
	// well before that.
	firstFrame := make(chan struct{})
	var once sync.Once
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			if e.Type == runtime.APITypeWarning {
				for _, a := range e.Args {
					warnings = append(warnings, string(a.Value))
				}
			}
		case *network.EventWebSocketFrameReceived:
			once.Do(func() { close(firstFrame) })
		}
	})

	const mount = `.tinkerdown-embed-lvt[data-embed-path="/apps/custom-attributes/"]`
	rating := func(n string) string {
		return `(() => { const m = document.querySelector('` + mount + `');` +
			` const r = m && m.querySelector('#rating');` +
			` return !!r && r.textContent.trim() === 'Rated ` + n + ` of 5.'; })()`
	}

	if err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(baseURL()+"/recipes/custom-attributes"),
		chromedp.WaitVisible(mount+` [data-star="3"]`, chromedp.ByQuery),
		chromedp.ActionFunc(func(context.Context) error {
			select {
			case <-firstFrame:
				return nil
			case <-time.After(15 * time.Second):
				return errors.New("no WebSocket frame from the embed")
			}
		}),
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

// The guide is synced from livetemplate, and tinkerdown serves only the pages
// tinkerdown.yaml lists, falling back to the landing page for the rest. So a
// synced guide without a nav entry renders the home page at its own URL.
func TestGuideCustomAttributesIsRouted(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()
	consoleErrs := captureConsoleErrors(ctx)

	var h1s []string
	var navLinks, recipeLinks int
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/guides/custom-attributes"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Evaluate(`[...document.querySelectorAll('h1')].map(h => h.textContent.trim())`, &h1s),
		chromedp.Evaluate(`document.querySelectorAll('nav a[href*="guides/custom-attributes"], aside a[href*="guides/custom-attributes"]').length`, &navLinks),
		chromedp.Evaluate(`document.querySelectorAll('a[href*="/recipes/custom-attributes"]').length`, &recipeLinks),
	); err != nil {
		t.Fatalf("load guide: %v\nconsole errors: %v", err, consoleErrs())
	}
	if len(h1s) == 0 || h1s[0] != "Custom attributes" {
		t.Errorf("guide h1s = %q, want the first to be %q", h1s, "Custom attributes")
	}
	if navLinks == 0 {
		t.Error("sidebar has no link to the guide")
	}
	if recipeLinks == 0 {
		t.Error("guide has no link to the recipe")
	}
}
