package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// TestGreetAsync exercises the Async + {{.lvt.Pending}} example: clicking
// the button triggers a two-frame WebSocket sequence:
//
//  1. Pending render: button shows "Loading..." with disabled + aria-busy
//  2. Completion render: headline shows the name, button reverts to "Say hi"
//
// Unlike TestSpineLoadingServerEmbed (which only checks final DOM), this test
// captures WS frames and asserts the intermediate pending state actually
// reached the client.
func TestGreetAsync(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	consoleErrs := captureConsoleErrors(ctx)
	ws := recordWSFrames(ctx)

	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		t.Fatalf("enable network domain: %v", err)
	}

	const name = "Ada"

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/apps/greet-async/"),
		chromedp.WaitVisible(`h1`, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("navigate: %v\nconsole: %v", err, consoleErrs())
	}

	// Wait for WS to connect and the initial mount render to arrive.
	if err := ws.waitForReceivedWithTreeCount(1, 10*time.Second); err != nil {
		t.Fatalf("WS mount frame never arrived: %v\nframes:\n%s", err, ws.dump())
	}

	// Baseline: count tree-bearing frames before the action so we can
	// isolate the pending + completion pair (ignores heartbeats/acks).
	baselineCount := len(ws.receivedWithTree())

	// Type name and click.
	if err := chromedp.Run(ctx,
		chromedp.SendKeys(`input[name="name"]`, name, chromedp.ByQuery),
		chromedp.Click(`button.greet-btn`, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("send keys / click: %v\nconsole: %v", err, consoleErrs())
	}

	// Wait for the completion render (tree frame #2 after the action). The
	// async work takes ~700ms, so allow up to 5s. Filtering by tree key
	// ensures heartbeats/acks don't shift the index.
	if err := ws.waitForReceivedWithTreeCount(baselineCount+2, 5*time.Second); err != nil {
		t.Fatalf("expected 2 tree frames after action (pending + completion): %v\nframes:\n%s",
			err, ws.dump())
	}

	actionFrames := ws.receivedWithTreeSince(baselineCount)

	// --- Frame #1: pending render ---
	// The pending frame's tree should contain "Loading..." (the button text
	// when .lvt.Pending is true).
	if len(actionFrames) < 1 {
		t.Fatal("no frames after action")
	}
	pendingData := actionFrames[0].Data
	if !strings.Contains(pendingData, "Loading...") {
		t.Errorf("frame #1 (pending) should contain \"Loading...\", got:\n%s", truncateForLog(pendingData))
	}

	// --- Frame #2: completion render ---
	// The completion frame should contain the applied name and NOT contain
	// "Loading..." (pending cleared).
	if len(actionFrames) < 2 {
		t.Fatal("only 1 frame after action, expected 2 (pending + completion)")
	}
	completionData := actionFrames[1].Data
	if !strings.Contains(completionData, name) {
		t.Errorf("frame #2 (completion) should contain %q, got:\n%s", name, truncateForLog(completionData))
	}
	if strings.Contains(completionData, "Loading...") {
		t.Errorf("frame #2 (completion) should NOT contain \"Loading...\" (pending should be cleared)")
	}

	// --- Final DOM state ---
	var headline, button, disabled, busy string
	if err := chromedp.Run(ctx,
		chromedp.Text(`h1`, &headline, chromedp.ByQuery),
		chromedp.Text(`button.greet-btn`, &button, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('button.greet-btn')?.getAttribute('disabled') || ''`, &disabled),
		chromedp.Evaluate(`document.querySelector('button.greet-btn')?.getAttribute('aria-busy') || ''`, &busy),
	); err != nil {
		t.Fatalf("DOM query: %v\nconsole: %v", err, consoleErrs())
	}

	if !strings.Contains(headline, name) {
		t.Errorf("headline = %q, want to contain %q", headline, name)
	}
	if strings.TrimSpace(button) != "Say hi" {
		t.Errorf("button text = %q, want \"Say hi\" after completion", button)
	}
	if disabled != "" {
		t.Errorf("button still disabled after completion")
	}
	if busy != "" {
		t.Errorf("button still aria-busy after completion")
	}

	// No console errors.
	if errs := consoleErrs(); len(errs) > 0 {
		t.Errorf("browser console errors: %v", errs)
	}

	// Dump frames on failure for diagnostics.
	if t.Failed() {
		t.Logf("WS frames:\n%s", ws.dump())
	}
}

func truncateForLog(s string) string {
	if len(s) <= 500 {
		return s
	}
	return s[:500] + "..."
}
