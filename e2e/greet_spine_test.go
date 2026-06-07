// Verifies the homepage's progressive-narrative spine — one greet app shown
// across seven steps. Step embeds 1-4 run WebSocket-disabled (HTTP fetch);
// the greet-wall embed (steps 5-7) runs over WebSocket with per-tab sync,
// a shared cross-user wall, and server-initiated push.
//
// Run locally against the live preview:
//
//	E2E_BASE_URL=http://localhost:8084 go test ./e2e -run TestSpine
package e2e

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// spineEmbeds are the live demo mounts the spine relies on, in scroll order.
// greet-wall backs steps 5-7; the rest are one step each.
var spineEmbeds = []string{
	"/apps/greet/",
	"/apps/greet-validate/",
	"/apps/greet-loading/",
	"/apps/greet-nojs/",
	"/apps/greet-wall/",
}

// TestSpineEmbedsMount asserts every step's live app is present in the DOM and
// server-side inlined (its <h1> rendered), with no page console errors.
func TestSpineEmbedsMount(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()
	consoleErrs := captureConsoleErrors(ctx)

	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitVisible(".hero", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // allow every embed to mount
	); err != nil {
		t.Fatalf("navigate: %v\nconsole: %v", err, consoleErrs())
	}

	for _, path := range spineEmbeds {
		var present, inlined bool
		js := `(() => {
			const els = [...document.querySelectorAll('.tinkerdown-embed-lvt[data-embed-path="` + path + `"]')];
			return [els.length > 0, els.some(m => m.querySelector('h1') != null)];
		})()`
		var res []bool
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, &res)); err != nil {
			t.Fatalf("eval %s: %v", path, err)
		}
		if len(res) == 2 {
			present, inlined = res[0], res[1]
		}
		if !present {
			t.Errorf("embed %s not present on landing", path)
		} else if !inlined {
			t.Errorf("embed %s present but did not inline (no <h1>); console: %v", path, consoleErrs())
		}
	}

	for _, e := range consoleErrs() {
		low := strings.ToLower(e)
		if strings.Contains(low, "content security policy") || strings.Contains(low, "failed to load") {
			t.Errorf("page console error: %s", e)
		}
	}
}

// TestSpineValidation exercises step 2's server-side path: typing the reserved
// name "admin" passes the browser's HTML checks but is rejected by a server
// FieldError that renders inline. (The empty case is enforced client-side by
// the input's `required` attribute, so the browser blocks it before any
// round-trip — that's the client half of the both-sides story.)
func TestSpineValidation(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var headline, errText, ariaInvalid string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/apps/greet-validate/"),
		chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
		chromedp.Sleep(800*time.Millisecond),
		chromedp.SendKeys(`input[name="name"]`, "admin", chromedp.ByQuery),
		chromedp.Click(`button[name="greet"]`, chromedp.ByQuery),
		chromedp.Sleep(900*time.Millisecond),
		chromedp.Text(`h1`, &headline, chromedp.ByQuery),
		chromedp.Evaluate(`(document.querySelector('form small')||{}).innerText||''`, &errText),
		chromedp.Evaluate(`(document.querySelector('input[name="name"]')||{}).getAttribute?.('aria-invalid')||''`, &ariaInvalid),
	); err != nil {
		t.Fatalf("validate run: %v", err)
	}

	if !strings.Contains(headline, "there") {
		t.Errorf("headline = %q, want unchanged greeting (Hello, there) on validation error", headline)
	}
	if !strings.Contains(strings.ToLower(errText), "reserved") {
		t.Errorf("error text = %q, want the server FieldError about a reserved name", errText)
	}
	if ariaInvalid != "true" {
		t.Errorf("aria-invalid = %q, want \"true\" on the errored field", ariaInvalid)
	}
}

// NOTE: the "loading spinner stuck on a no-diff repeat click" regression is
// owned by the client repo's unit test (tests/loading-lifecycle-empty-diff),
// since the fix lives in @livetemplate/client and reaches the landing only
// once that client is released and re-vendored into tinkerdown. A landing-side
// e2e guard belongs here after that ships.

// TestSpineSelfTopicSync exercises step 5: two tabs of the SAME session (one
// browser context → one cookie/group) are both open, then a greeting in tab A
// updates tab B's headline live over the self-topic — no reload. This is the
// "your tabs" claim, and it requires a shared session group, which two on-page
// embeds do NOT have (see TestSpineCrossUserWall for that separate case).
func TestSpineSelfTopicSync(t *testing.T) {
	ctxA, cancel := newCtx(t)
	defer cancel()
	url := baseURL() + "/apps/greet-wall/"

	if err := chromedp.Run(ctxA,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
		chromedp.Sleep(1400*time.Millisecond), // WS connect
	); err != nil {
		t.Fatalf("tab A connect: %v", err)
	}

	// Second tab in the same browser → same group. Stays open before the greet.
	tabB, cancelB := newTab(ctxA)
	defer cancelB()
	if err := chromedp.Run(tabB,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`h1`, chromedp.ByQuery),
		chromedp.Sleep(1400*time.Millisecond),
	); err != nil {
		t.Fatalf("tab B connect: %v", err)
	}

	if err := chromedp.Run(ctxA,
		chromedp.SendKeys(`input[name="name"]`, "Tabby", chromedp.ByQuery),
		chromedp.Click(`button[name="greet"]`, chromedp.ByQuery),
		chromedp.Sleep(1600*time.Millisecond),
	); err != nil {
		t.Fatalf("tab A greet: %v", err)
	}

	var headlineA, headlineB string
	if err := chromedp.Run(ctxA, chromedp.Text(`h1`, &headlineA, chromedp.ByQuery)); err != nil {
		t.Fatalf("read tab A: %v", err)
	}
	if err := chromedp.Run(tabB, chromedp.Text(`h1`, &headlineB, chromedp.ByQuery)); err != nil {
		t.Fatalf("read tab B: %v", err)
	}
	if !strings.Contains(headlineA, "Tabby") {
		t.Errorf("tab A headline = %q, want it to greet Tabby", headlineA)
	}
	if !strings.Contains(headlineB, "Tabby") {
		t.Errorf("tab B headline = %q, want LIVE self-topic sync to Tabby (no reload)", headlineB)
	}
}

// TestSpineCrossUserWall exercises step 6: the landing's greet-wall embeds are
// independent sessions (embed-lvt isolates each mount), so they stand in for
// different people. A greeting typed in one embed appears on another embed's
// shared wall live — the cross-user proof. Headlines stay independent (that's
// step 5's self-topic, which doesn't bridge separate sessions).
func TestSpineCrossUserWall(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	const name = "Crosby"
	sel := `.tinkerdown-embed-lvt[data-embed-path="/apps/greet-wall/"]`
	typeJS := `(() => {
		const e = document.querySelectorAll('` + sel + `');
		if (e.length < 2) return false;
		const inp = e[0].querySelector('input[name=name]');
		inp.value = '` + name + `';
		inp.dispatchEvent(new Event('input', {bubbles:true}));
		e[0].querySelector('button[name=greet]').click();
		return true;
	})()`
	readJS := `(() => {
		const e = document.querySelectorAll('` + sel + `');
		return e.length > 1 ? ((e[1].querySelector('ul.wall')||{}).innerText || '') : '';
	})()`

	var clicked bool
	var otherWall string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // all wall embeds connect
		chromedp.Evaluate(typeJS, &clicked),
		chromedp.Sleep(2*time.Second), // let the cross-session broadcast land
		chromedp.Evaluate(readJS, &otherWall),
	); err != nil {
		t.Fatalf("cross-user run: %v", err)
	}
	if !clicked {
		t.Fatal("expected at least two greet-wall embeds on the landing")
	}
	if !strings.Contains(otherWall, name) {
		t.Errorf("second embed wall = %q, want %q to cross from the first embed (shared-topic broadcast)", otherWall, name)
	}
}

// TestSpineServerPush exercises step 7, "the server speaks first": with NO user
// action, the server's heartbeat populates an in-place ".from-server" slot —
// and crucially it does NOT append rows to the shared wall (the regression this
// change fixes; server lines used to pile up and crowd out real greetings). The
// stack under test must run a fast heartbeat (GREET_WALL_SERVER_INTERVAL); the
// Makefile e2e targets and the CI docker run set it so this never silently
// idles.
func TestSpineServerPush(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()
	url := baseURL() + "/apps/greet-wall/"

	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
		chromedp.Sleep(1400*time.Millisecond), // WS connect
	); err != nil {
		t.Fatalf("connect: %v", err)
	}

	// Poll for the server's in-place slot, with no user action of any kind.
	deadline := time.Now().Add(serverPushTimeout())
	var serverLine string
	for time.Now().Before(deadline) {
		if err := chromedp.Run(ctx, chromedp.Evaluate(
			`(document.querySelector('.from-server')||{}).innerText || ''`, &serverLine),
		); err != nil {
			t.Fatalf("read server slot: %v", err)
		}
		if strings.Contains(serverLine, "the server said hi") {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !strings.Contains(serverLine, "the server said hi") {
		t.Fatalf("server-push slot = %q, want a server-initiated 'the server said hi at …' within %v "+
			"(is GREET_WALL_SERVER_INTERVAL set fast on the stack under test?)", serverLine, serverPushTimeout())
	}

	// The wall itself must stay free of server rows — that's the whole point.
	var wallText string
	if err := chromedp.Run(ctx, chromedp.Evaluate(
		`(document.querySelector('ul.wall')||{}).innerText || ''`, &wallText),
	); err != nil {
		t.Fatalf("read wall: %v", err)
	}
	if strings.Contains(wallText, "the server") {
		t.Errorf("wall contains a server line:\n%s\nserver pushes must replace the in-place slot, not append to the wall", wallText)
	}
}

// serverPushTimeout bounds how long TestSpineServerPush waits for a heartbeat,
// derived from the same GREET_WALL_SERVER_INTERVAL the app reads so a fast
// harness tick keeps the test quick while a default stack still passes.
func serverPushTimeout() time.Duration {
	interval := 30 * time.Second
	if v := os.Getenv("GREET_WALL_SERVER_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			interval = d
		}
	}
	return interval*2 + 4*time.Second
}

// newTab opens a second tab in the same browser as parent (shared cookies).
func newTab(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := chromedp.NewContext(parent)
	tctx, tcancel := context.WithTimeout(ctx, 30*time.Second)
	return tctx, func() { tcancel(); cancel() }
}
