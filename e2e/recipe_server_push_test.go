package e2e

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestRecipeServerPushFanoutEmbed verifies the live-dashboard embed on the
// Server Push recipe: a single background goroutine pushes to every connected
// browser via handler.Publish, so the embed's counter advances with NO user
// interaction. This is the browser gate for the "fan out to many sessions"
// section (the app's own suite lives in examples/live-dashboard).
func TestRecipeServerPushFanoutEmbed(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)

	const embed = `.tinkerdown-embed-lvt[data-embed-path="/apps/live-dashboard/"]`
	ticksJS := `((document.querySelector('` + embed + ` #ticks'))||{}).textContent || ''`

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/server-push"),
		chromedp.WaitVisible(embed+` #ticks`, chromedp.ByQuery),
		chromedp.Sleep(1400*time.Millisecond), // WS connect
	); err != nil {
		t.Fatalf("load recipe / connect embed: %v", err)
	}

	readTicks := func() int {
		var s string
		if err := chromedp.Run(ctx, chromedp.Evaluate(ticksJS, &s)); err != nil {
			t.Fatalf("read #ticks: %v", err)
		}
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			t.Fatalf("#ticks %q not an int: %v", s, err)
		}
		return n
	}

	start := readTicks()
	// The one-second server ticker must advance the counter with no user action.
	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		if readTicks() > start {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if got := readTicks(); got <= start {
		t.Fatalf("embed #ticks did not advance from %d (got %d) — the background handler.Publish never reached the browser", start, got)
	}

	if errs := consoleErrs(); len(errs) > 0 {
		t.Errorf("browser console errors on the recipe embed:\n%s", strings.Join(errs, "\n"))
	}
}
