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

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

// spineEmbeds are the embed-lvt mounts the spine relies on, in scroll order.
// Step 4 uses real iframes instead, so greet-nojs is asserted separately.
var spineEmbeds = []string{
	"/apps/greet/",
	"/apps/greet-validate/",
	"/apps/greet-loading/",
	"/apps/greet-loading-server/",
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

	// Step 4 no longer uses embed-lvt: both cards render the real greet-nojs
	// app in iframes, one script-enabled and one script-disabled.
	var nojsFrames int
	var nojsScriptless bool
	const nojsJS = `(() => {
		const frames = [...document.querySelectorAll('iframe.nojs-frame[src="/apps/greet-nojs/"]')];
		return [frames.length, frames.some(f => !(f.getAttribute('sandbox') || '').includes('allow-scripts'))];
	})()`
	var nojsRes []any
	if err := chromedp.Run(ctx, chromedp.Evaluate(nojsJS, &nojsRes)); err != nil {
		t.Fatalf("eval greet-nojs iframes: %v", err)
	}
	if len(nojsRes) == 2 {
		if v, ok := nojsRes[0].(float64); ok {
			nojsFrames = int(v)
		}
		if v, ok := nojsRes[1].(bool); ok {
			nojsScriptless = v
		}
	}
	if nojsFrames < 2 {
		t.Errorf("step 4 greet-nojs iframes = %d, want both JS-on and JS-off cards present", nojsFrames)
	}
	if !nojsScriptless {
		t.Errorf("step 4 greet-nojs iframes missing a script-disabled sandboxed variant")
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

// TestSpineLoadingServerEmbed exercises Step 3's server-owned loading demo on
// the landing page itself. It must leave the pending state, clear aria-busy
// and disabled, and render the final greeting after the follow-up action.
func TestSpineLoadingServerEmbed(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	const (
		path = `/apps/greet-loading-server/`
		name = "Ada"
	)
	sel := `.tinkerdown-embed-lvt[data-embed-path="` + path + `"]`

	var headline, button, disabled, busy string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.ScrollIntoView(sel, chromedp.ByQuery),
		chromedp.Sleep(1500*time.Millisecond),
		chromedp.SendKeys(sel+` input[name="name"]`, name, chromedp.ByQuery),
		chromedp.Click(sel+` button`, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		chromedp.Text(sel+` h1`, &headline, chromedp.ByQuery),
		chromedp.Text(sel+` button`, &button, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector(`+"`"+sel+` button`+"`"+`)?.getAttribute('disabled') || ''`, &disabled),
		chromedp.Evaluate(`document.querySelector(`+"`"+sel+` button`+"`"+`)?.getAttribute('aria-busy') || ''`, &busy),
	); err != nil {
		t.Fatalf("step 3 server loading run: %v", err)
	}
	if !strings.Contains(headline, name) {
		t.Errorf("headline = %q, want %q after the server-owned loading demo finishes", headline, name)
	}
	if disabled != "" {
		t.Errorf("disabled = %q, want cleared after follow-up action completes", disabled)
	}
	if busy != "" {
		t.Errorf("aria-busy = %q, want cleared after follow-up action completes", busy)
	}
	if button != "Say hi" {
		t.Errorf("button = %q, want \"Say hi\" after loading clears", button)
	}
}

// TestSpineNoJSFormPost is the heart of step 4's "show, don't tell": it runs a
// real browser with JavaScript EXECUTION DISABLED, so the greet-nojs client
// can't intercept the form. Submitting therefore does a plain HTTP POST; the
// server replies 303 (POST-Redirect-GET) and the followed GET must still greet
// the typed name — proving the no-JS transport genuinely round-trips (the
// per-session name store survives it, and cmd/site rewrites the redirect back
// under the mount so it doesn't bounce to the site root).
func TestSpineNoJSFormPost(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	const name = "Nomad"
	var headline string
	if err := chromedp.Run(ctx,
		emulation.SetScriptExecutionDisabled(true), // the whole point: no client JS
		chromedp.Navigate(baseURL()+"/apps/greet-nojs/"),
		chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="name"]`, name, chromedp.ByQuery),
		chromedp.Click(`button[name="greet"]`, chromedp.ByQuery), // native form submit
		chromedp.Sleep(1500*time.Millisecond),                    // POST -> 303 -> GET render
		chromedp.Text(`h1`, &headline, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("no-JS run: %v", err)
	}
	if !strings.Contains(headline, name) {
		t.Errorf("no-JS headline = %q, want %q after a plain form POST with JavaScript disabled", headline, name)
	}
}

// TestSpineNoJSIframe verifies step 4's right-hand card on the LANDING — the
// surface the reader actually uses. It confirms the card is a script-disabled
// sandboxed frame of the real app, then proves the framed no-JS round-trip:
// a JS-disabled greeting is recorded against the session, and the embedded
// iframe — a separate document — renders it. The risky half is whether the
// livetemplate-id cookie (SameSite=Lax, HttpOnly) reaches the framed request;
// it does, because allow-same-origin keeps the same-origin frame first-party.
// (We seed via a top-level JS-disabled submit and observe the frame because
// chromedp's input dispatch INTO a sandboxed frame is unreliable, while frame
// DOM reads are not; TestSpineNoJSFormPost covers the POST half top-level.)
func TestSpineNoJSIframe(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	const name = "Framed"
	// Record a greeting for this browser's session over the no-JS path.
	if err := chromedp.Run(ctx,
		emulation.SetScriptExecutionDisabled(true),
		chromedp.Navigate(baseURL()+"/apps/greet-nojs/"),
		chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="name"]`, name, chromedp.ByQuery),
		chromedp.Click(`button[name="greet"]`, chromedp.ByQuery),
		chromedp.Sleep(1500*time.Millisecond), // POST -> 303 -> GET
	); err != nil {
		t.Fatalf("seed no-JS greeting: %v", err)
	}

	// Load the landing; the embedded iframe's GET must carry the same-origin
	// cookie and render the recorded greeting.
	var src, sandbox string
	if err := chromedp.Run(ctx,
		emulation.SetScriptExecutionDisabled(false),
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitVisible(`iframe.nojs-frame[sandbox="allow-forms allow-same-origin"]`, chromedp.ByQuery),
		chromedp.AttributeValue(`iframe.nojs-frame[sandbox="allow-forms allow-same-origin"]`, "src", &src, nil),
		chromedp.AttributeValue(`iframe.nojs-frame[sandbox="allow-forms allow-same-origin"]`, "sandbox", &sandbox, nil),
		chromedp.ScrollIntoView(`iframe.nojs-frame[sandbox="allow-forms allow-same-origin"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),                                 // frame loads + renders
	); err != nil {
		t.Fatalf("iframe attrs: %v", err)
	}
	if !strings.Contains(src, "/apps/greet-nojs/") {
		t.Errorf("iframe src = %q, want the greet-nojs app", src)
	}
	if strings.Contains(sandbox, "allow-scripts") {
		t.Errorf("iframe sandbox = %q, must NOT allow-scripts — it's the no-JS demo", sandbox)
	}
	if !strings.Contains(sandbox, "allow-forms") {
		t.Errorf("iframe sandbox = %q, must allow-forms so the no-JS POST works", sandbox)
	}

	var frame []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(`iframe.nojs-frame[sandbox="allow-forms allow-same-origin"]`, &frame,
		chromedp.ByQuery, chromedp.Populate(-1, true))); err != nil {
		t.Fatalf("populate iframe: %v", err)
	}
	if len(frame) == 0 {
		t.Fatal("greet-nojs iframe not found")
	}
	var headline string
	if err := chromedp.Run(ctx,
		chromedp.Text(`h1`, &headline, chromedp.ByQuery, chromedp.FromNode(frame[0])),
	); err != nil {
		t.Fatalf("read iframe headline: %v", err)
	}
	if !strings.Contains(headline, name) {
		t.Errorf("iframe headline = %q, want %q — the framed no-JS request must carry the session cookie", headline, name)
	}
}

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

// TestSpineServerPush exercises step 7, "the server speaks first", on the
// LANDING embed — the exact surface the bug report showed. With NO user action,
// the server's heartbeat populates an in-place ".from-server" slot inside the
// embedded greet-wall, and crucially it does NOT append rows to the embed's
// shared wall (the regression this change fixes; server lines used to pile up
// and crowd out real greetings). The stack under test must run a fast heartbeat
// (GREET_WALL_SERVER_INTERVAL); the Makefile e2e targets and the CI docker run
// set it so this never silently idles.
func TestSpineServerPush(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	embed := `.tinkerdown-embed-lvt[data-embed-path="/apps/greet-wall/"]`
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitVisible(embed+` input[name="name"]`, chromedp.ByQuery),
		chromedp.Sleep(1400*time.Millisecond), // WS connect
	); err != nil {
		t.Fatalf("connect: %v", err)
	}

	// Poll the embed's in-place slot, with no user action of any kind.
	serverJS := `((document.querySelector('` + embed + ` .from-server'))||{}).innerText || ''`
	deadline := time.Now().Add(serverPushTimeout())
	var serverLine string
	for time.Now().Before(deadline) {
		if err := chromedp.Run(ctx, chromedp.Evaluate(serverJS, &serverLine)); err != nil {
			t.Fatalf("read server slot: %v", err)
		}
		if strings.Contains(serverLine, "the server said hi") {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !strings.Contains(serverLine, "the server said hi") {
		t.Fatalf("landing embed server slot = %q, want a server-initiated 'the server said hi at …' within %v "+
			"(is GREET_WALL_SERVER_INTERVAL set fast on the stack under test?)", serverLine, serverPushTimeout())
	}

	// The embed's wall must stay free of server rows — that's the whole point.
	var wallText string
	if err := chromedp.Run(ctx, chromedp.Evaluate(
		`((document.querySelector('`+embed+` ul.wall'))||{}).innerText || ''`, &wallText),
	); err != nil {
		t.Fatalf("read wall: %v", err)
	}
	if strings.Contains(wallText, "the server") {
		t.Errorf("landing embed wall contains a server line:\n%s\nserver pushes must replace the in-place slot, not append to the wall", wallText)
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
