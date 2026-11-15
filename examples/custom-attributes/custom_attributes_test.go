package customattributes

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/livetemplate/livetemplate"
	e2etest "github.com/livetemplate/lvt/testing"
)

// ---------------------------------------------------------------------------
// White-box tests for the two actions. The attributes themselves are
// client-side, so the browser test below is the only place they run.
// ---------------------------------------------------------------------------

func ctxWithStars(stars any) *livetemplate.Context {
	return livetemplate.NewContext(context.Background(), "SetRating", map[string]interface{}{"stars": stars})
}

func TestSetRating(t *testing.T) {
	c := &Controller{}
	s, err := c.SetRating(State{}, ctxWithStars(4))
	if err != nil || s.Rating != 4 {
		t.Fatalf("SetRating(4) = %d, %v; want 4, nil", s.Rating, err)
	}
	for _, bad := range []int{0, 6, -1} {
		if _, err := c.SetRating(State{Rating: 2}, ctxWithStars(bad)); err == nil {
			t.Errorf("SetRating(%d) accepted an out-of-range rating", bad)
		}
	}
}

func TestShareURLCarriesTheRating(t *testing.T) {
	if got := (State{}).ShareURL(); got != ShareBase {
		t.Errorf("unrated ShareURL = %q, want %q", got, ShareBase)
	}
	if got, want := (State{Rating: 3}).ShareURL(), ShareBase+"?stars=3"; got != want {
		t.Errorf("ShareURL = %q, want %q", got, want)
	}
}

func TestToggleShortcuts(t *testing.T) {
	s, _ := (&Controller{}).ToggleShortcuts(State{Shortcuts: true}, nil)
	if s.Shortcuts {
		t.Error("ToggleShortcuts did not turn shortcuts off")
	}
}

// ---------------------------------------------------------------------------
// Browser test: all three attributes, in a real browser, against the real
// handler.
// ---------------------------------------------------------------------------

const (
	// The client's own wording for the attribute census and the registry.
	censusWarning   = "no attribute handler claims it"
	fallbackWarning = "falling back to HTML structure-based tree"
)

// serveLocalClient makes the browser load the client bundle at path instead of
// the published one that lvtClientScriptURL points at, so this test can run
// against an unreleased client. Set LVT_LOCAL_CLIENT_JS to use it.
func serveLocalClient(t *testing.T, ctx context.Context, path string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("LVT_LOCAL_CLIENT_JS: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(body)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		e, ok := ev.(*fetch.EventRequestPaused)
		if !ok {
			return
		}
		go func() {
			exec := cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Target)
			err := fetch.FulfillRequest(e.RequestID, http.StatusOK).
				WithResponseHeaders([]*fetch.HeaderEntry{{Name: "Content-Type", Value: "application/javascript"}}).
				WithBody(encoded).
				Do(exec)
			if err != nil {
				log.Printf("serveLocalClient: fulfill %s: %v", e.Request.URL, err)
			}
		}()
	})
	if err := chromedp.Run(ctx, fetch.Enable().WithPatterns([]*fetch.RequestPattern{
		{URLPattern: "*@livetemplate/client@*/livetemplate-client.browser.js"},
	})); err != nil {
		t.Fatalf("enable fetch interception: %v", err)
	}
	t.Logf("serving the client from %s", path)
}

// The page runs over plain http in this test, where navigator.clipboard does
// not exist. This stub records what would have been copied.
const clipboardStub = `
Object.defineProperty(navigator, "clipboard", {
  configurable: true,
  value: { writeText: async (text) => { (window.__copied ||= []).push(text); } },
});`

func TestCustomAttributesE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Global log capture: do NOT add t.Parallel().
	serverLogs := e2etest.NewSafeBuffer()
	log.SetOutput(serverLogs)
	defer log.SetOutput(os.Stderr)

	port, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("free port: %v", err)
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: Handler(livetemplate.WithDevMode(true)),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()
	e2etest.WaitForServer(t, fmt.Sprintf("http://localhost:%d", port), 10*time.Second)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			t.Logf("server shutdown warning: %v", err)
		}
	}()

	chromeCtx, cleanup := e2etest.SetupDockerChrome(t, 60*time.Second)
	defer cleanup()
	ctx := chromeCtx.Context

	consoleLog := e2etest.NewConsoleLogger()
	consoleLog.Start(ctx)
	wsLog := e2etest.RecordWSFrames(ctx)
	if p := os.Getenv("LVT_LOCAL_CLIENT_JS"); p != "" {
		serveLocalClient(t, ctx, p)
	}

	dump := func(label string) {
		t.Logf("=== %s ===", label)
		for _, l := range consoleLog.GetLogs() {
			t.Logf("console [%s]: %s", l.Type, l.Message)
		}
		for _, m := range wsLog.GetMessages() {
			t.Logf("ws %s: %s", m.Direction, m.Data)
		}
		t.Logf("server log:\n%s", serverLogs.String())
		var html string
		if err := chromedp.Run(ctx, chromedp.OuterHTML(`html`, &html, chromedp.ByQuery)); err != nil {
			t.Logf("outerHTML: %v", err)
		} else {
			t.Logf("rendered HTML:\n%s", html)
		}
	}
	fail := func(label string, args ...any) {
		t.Helper()
		dump(label)
		t.Fatal(args...)
	}
	waitFor := func(js, label string) {
		t.Helper()
		if err := chromedp.Run(ctx, chromedp.Poll(js, nil, chromedp.WithPollingTimeout(10*time.Second))); err != nil {
			fail(label, label+": ", err)
		}
	}
	ratingIs := func(n int) string {
		return fmt.Sprintf(`document.getElementById("rating").textContent.trim() === "Rated %d of 5."`, n)
	}
	sentRatings := func() int {
		n := 0
		for _, m := range wsLog.GetSent() {
			if strings.Contains(m.Data, `"SetRating"`) {
				n++
			}
		}
		return n
	}
	press := func(key string) {
		t.Helper()
		if err := chromedp.Run(ctx, chromedp.KeyEvent(key)); err != nil {
			fail("key press failed", err)
		}
	}

	if err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(clipboardStub).Do(ctx)
			return err
		}),
		chromedp.Navigate(e2etest.GetChromeTestURL(port)),
		e2etest.WaitForWebSocketReady(30*time.Second),
	); err != nil {
		fail("initial load failed", err)
	}

	// lvt-x:rating: a click dispatches SetRating, the action runs, and the
	// re-render lands. A second click after that render sends exactly one
	// more frame: the listener did not stack.
	if err := chromedp.Run(ctx, chromedp.Click(`[data-star="3"]`, chromedp.ByQuery)); err != nil {
		fail("click star 3", err)
	}
	waitFor(ratingIs(3), "rating never reached 3")
	if err := chromedp.Run(ctx, chromedp.Click(`[data-star="5"]`, chromedp.ByQuery)); err != nil {
		fail("click star 5", err)
	}
	waitFor(ratingIs(5), "rating never reached 5")
	if n := sentRatings(); n != 2 {
		fail("listener count", fmt.Sprintf("want 2 SetRating frames for 2 clicks, got %d", n))
	}

	// lvt-x:copy: copies the link for the CURRENT rating, set by a render
	// after the listener was wired.
	if err := chromedp.Run(ctx, chromedp.Click(`#copy`, chromedp.ByID)); err != nil {
		fail("click copy", err)
	}
	waitFor(fmt.Sprintf(`(window.__copied || []).at(-1) === %q`, ShareBase+"?stars=5"),
		"lvt-x:copy did not copy the current share URL")

	// lvt-x:hotkey: the low-level handler maps 1-5 to the stars.
	press("2")
	waitFor(ratingIs(2), "hotkey 2 did not rate 2")

	// The server removes the attribute. The handler's next setup() rebuilds
	// its map without those elements, so the keys stop doing anything.
	if err := chromedp.Run(ctx, chromedp.Click(`#shortcuts`, chromedp.ByID)); err != nil {
		fail("turn shortcuts off", err)
	}
	waitFor(`!document.querySelector("[lvt-x\\:hotkey]")`, "server never removed lvt-x:hotkey")
	before := sentRatings()
	press("4")
	time.Sleep(500 * time.Millisecond) // a negative: nothing to wait on
	if n := sentRatings(); n != before {
		fail("disarmed hotkey fired", fmt.Sprintf("a key with lvt-x:hotkey removed still sent SetRating (%d -> %d frames)", before, n))
	}
	waitFor(ratingIs(2), "rating changed with shortcuts off")

	// And back on: the keys work again.
	if err := chromedp.Run(ctx, chromedp.Click(`#shortcuts`, chromedp.ByID)); err != nil {
		fail("turn shortcuts on", err)
	}
	waitFor(`!!document.querySelector("[lvt-x\\:hotkey]")`, "server never restored lvt-x:hotkey")
	press("4")
	waitFor(ratingIs(4), "hotkey 4 did not rate 4 after shortcuts came back")

	// Every lvt-x:* name on the page is claimed, so the census stays quiet,
	// and the registry has nothing to complain about.
	time.Sleep(300 * time.Millisecond)
	for _, l := range consoleLog.GetLogs() {
		if strings.Contains(l.Message, censusWarning) || strings.Contains(l.Message, "AttributeRegistry") {
			fail("attribute diagnostics", "unexpected attribute diagnostic: "+l.Message)
		}
		if l.Type == "error" {
			fail("console error", "browser console error: "+l.Message)
		}
	}

	if serverLogs.String() == "" {
		t.Fatal("captured no server logs; the fallback check below would pass vacuously")
	}
	if strings.Contains(serverLogs.String(), fallbackWarning) {
		fail("reactive path degraded", "server fell back to HTML-string diffing")
	}
}
