// Self-contained chromedp e2e for examples/draft-form — the post editor that
// demonstrates livetemplate v0.15.0's server-side formnovalidate honoring
// (#239). Unlike the staging-targeted tests in this package, this file boots
// its own in-process httptest server hosting draftform.Handler, so it needs no
// external harness or deployed site.
//
// Per project policy every failure surfaces all four debugging signals:
//  1. browser console errors      (captureConsoleErrors)
//  2. server logs                 (slog default + httptest ErrorLog → syncBuf)
//  3. websocket frames            (captureWSFrames via the Network domain)
//  4. final rendered HTML         (chromedp.OuterHTML)
//
// Run:
//
//	GOWORK=off go test ./e2e/ -run TestDraftForm -v -count=1
package e2e

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/livetemplate/livetemplate"

	draftform "github.com/livetemplate/docs/examples/draft-form"
)

// chromiumPath is the chromium binary available in this environment; there is
// no google-chrome, so the default exec allocator's lookup would fail.
const chromiumPath = "/run/current-system/sw/bin/chromium"

// clientJSURL is the CDN bundle the draft.tmpl loads. The test needs it to
// reach the browser for the WS tier to come alive; if the CDN is unreachable
// we skip rather than hang on a never-connecting client.
const clientJSURL = "https://cdn.jsdelivr.net/npm/@livetemplate/client@latest/dist/livetemplate-client.browser.js"

// syncBuf is a mutex-guarded buffer so slog (which may write from background
// goroutines) and the test reader don't race.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// newDraftCtx builds a chromedp context pinned to the local chromium, modeled
// on staging_test.go's newCtx but self-contained (no staging warmup). Each
// scenario gets its own allocator so it runs in a fresh browser — and thus a
// fresh livetemplate session, since cookies don't carry across allocators.
func newDraftCtx(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromiumPath),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.WSURLReadTimeout(45*time.Second),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 45*time.Second)
	return timeoutCtx, func() {
		timeoutCancel()
		cancel()
		allocCancel()
	}
}

// captureWSFrames records every WebSocket frame (both directions) for the
// target so a failed assertion can show exactly what the live session
// exchanged. Requires network.Enable() to be run as an action, otherwise the
// CDP Network domain never emits these events. Returns a snapshot closure.
func captureWSFrames(ctx context.Context) func() []string {
	var mu sync.Mutex
	var frames []string
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventWebSocketFrameSent:
			mu.Lock()
			frames = append(frames, "SENT: "+e.Response.PayloadData)
			mu.Unlock()
		case *network.EventWebSocketFrameReceived:
			mu.Lock()
			frames = append(frames, "RECV: "+e.Response.PayloadData)
			mu.Unlock()
		}
	})
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		out := make([]string, len(frames))
		copy(out, frames)
		return out
	}
}

// cdnReachable verifies the @livetemplate/client bundle the template loads is
// fetchable; if not, the WS tier can't hydrate and the test should skip.
func cdnReachable() bool {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, clientJSURL, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func TestDraftForm(t *testing.T) {
	if !cdnReachable() {
		t.Skipf("client CDN %s unreachable — cannot exercise the WS tier in this env", clientJSURL)
	}

	// (2) server logs: livetemplate exposes no logger option in v0.15.0, so we
	// redirect the default slog handler into a synchronized buffer and also
	// point the httptest server's ErrorLog at it. Restored in cleanup.
	logBuf := &syncBuf{}
	prevLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prevLogger) })

	// Boot the app. The origin allowlist is the chicken-and-egg case: the
	// handler needs the server URL, but the URL isn't known until the listener
	// exists. NewUnstartedServer allocates the listener up front, so we can
	// read its address, build the handler, then Start.
	ts := httptest.NewUnstartedServer(nil)
	serverURL := "http://" + ts.Listener.Addr().String()
	ts.Config.Handler = draftform.Handler(
		livetemplate.WithAllowedOrigins([]string{serverURL}),
	)
	ts.Config.ErrorLog = log.New(logBuf, "httptest: ", log.LstdFlags)
	ts.Start()
	t.Cleanup(ts.Close)

	// dumpSignals prints all four signals; called from each subtest's cleanup
	// on failure so debugging never requires a re-run.
	dumpSignals := func(t *testing.T, label, html string, console, ws []string) {
		t.Helper()
		t.Logf("\n========== FAILURE SIGNALS [%s] ==========", label)
		t.Logf("--- (1) browser console errors (%d) ---\n%s", len(console), strings.Join(console, "\n"))
		t.Logf("--- (2) server logs ---\n%s", logBuf.String())
		t.Logf("--- (3) websocket frames (%d) ---\n%s", len(ws), strings.Join(ws, "\n"))
		t.Logf("--- (4) rendered HTML ---\n%s", html)
	}

	// Scenario 1 — the heart of #239: "Save draft" carries formnovalidate, so
	// the server-side ctx.ValidateForm() is skipped even though the title is
	// empty and `required`. Success (data-status="draft") proves the framework
	// honored the formnovalidate submitter; otherwise an empty title would have
	// produced a field error and no draft.
	t.Run("SaveDraftSkipsValidation", func(t *testing.T) {
		ctx, cancel := newDraftCtx(t)
		defer cancel()
		consoleErrs := captureConsoleErrors(ctx)
		wsFrames := captureWSFrames(ctx)

		var html, status, errText, aria string
		t.Cleanup(func() {
			if t.Failed() {
				dumpSignals(t, "SaveDraftSkipsValidation", html, consoleErrs(), wsFrames())
			}
		})

		if err := chromedp.Run(ctx,
			network.Enable(),
			chromedp.Navigate(serverURL),
			chromedp.WaitVisible(`input[name="title"]`, chromedp.ByQuery),
			chromedp.Sleep(1500*time.Millisecond), // WS connect + hydrate
			chromedp.Click(`button[name="save-draft"]`, chromedp.ByQuery),
			chromedp.Sleep(1200*time.Millisecond), // action round-trip
			chromedp.Evaluate(`(document.querySelector('[data-status]')||{}).getAttribute?.('data-status')||''`, &status),
			chromedp.Evaluate(`(document.querySelector('form small')||{}).innerText||''`, &errText),
			chromedp.Evaluate(`(document.querySelector('input[name="title"]')||{}).getAttribute?.('aria-invalid')||''`, &aria),
			chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		); err != nil {
			t.Fatalf("save-draft run: %v", err)
		}

		if status != "draft" {
			t.Errorf("data-status = %q, want \"draft\" — formnovalidate must skip server validation for an empty title (#239)", status)
		}
		if strings.TrimSpace(errText) != "" {
			t.Errorf("validation error text = %q, want empty — save-draft must not surface a title error", errText)
		}
		if aria == "true" {
			t.Errorf("aria-invalid = %q, want unset — the title field must not be marked invalid on a skipped validation", aria)
		}
	})

	// Scenario 2 — "Publish" has no formnovalidate, so the empty required title
	// is rejected. WHERE it's rejected is decided empirically by the WS-frame
	// capture: the browser's native constraint validation blocks the submit
	// before any round-trip (so no frame is sent and no server ErrorTag
	// renders), which is correct HTML semantics. We assert the invariant that
	// always holds — nothing gets published — and branch on the frame count.
	t.Run("PublishEmptyIsRejected", func(t *testing.T) {
		ctx, cancel := newDraftCtx(t)
		defer cancel()
		consoleErrs := captureConsoleErrors(ctx)
		wsFrames := captureWSFrames(ctx)

		var html string
		t.Cleanup(func() {
			if t.Failed() {
				dumpSignals(t, "PublishEmptyIsRejected", html, consoleErrs(), wsFrames())
			}
		})

		if err := chromedp.Run(ctx,
			network.Enable(),
			chromedp.Navigate(serverURL),
			chromedp.WaitVisible(`input[name="title"]`, chromedp.ByQuery),
			chromedp.Sleep(1500*time.Millisecond), // WS connect + hydrate
		); err != nil {
			t.Fatalf("publish-empty connect: %v", err)
		}

		framesBefore := len(wsFrames())

		var publishedCount int
		var valueMissing bool
		var aria, errText string
		if err := chromedp.Run(ctx,
			chromedp.Click(`button[name="publish"]`, chromedp.ByQuery),
			chromedp.Sleep(1200*time.Millisecond),
			chromedp.Evaluate(`document.querySelectorAll('[data-status="published"]').length`, &publishedCount),
			chromedp.Evaluate(`!!(document.querySelector('input[name="title"]')||{}).validity?.valueMissing`, &valueMissing),
			chromedp.Evaluate(`(document.querySelector('input[name="title"]')||{}).getAttribute?.('aria-invalid')||''`, &aria),
			chromedp.Evaluate(`(document.querySelector('form small')||{}).innerText||''`, &errText),
			chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		); err != nil {
			t.Fatalf("publish-empty run: %v", err)
		}
		framesDuringClick := len(wsFrames()) - framesBefore

		// Invariant for both branches: nothing is ever published.
		if publishedCount != 0 {
			t.Errorf("found %d [data-status=published] elements, want 0 — an empty title must never publish", publishedCount)
		}

		if framesDuringClick == 0 {
			// Native client-side block: the browser refused to submit a
			// required-empty field, so the server never saw it.
			if !valueMissing {
				t.Errorf("no WS frame was sent on Publish yet input.validity.valueMissing = false; "+
					"expected the browser to block the empty required submit (aria=%q errText=%q)", aria, errText)
			}
			t.Logf("Publish+empty was blocked client-side (no WS frame sent); valueMissing=%v — server enforcement is not browser-observable here", valueMissing)
		} else {
			// Submit reached the server: it must have produced a field error.
			if aria != "true" && strings.TrimSpace(errText) == "" {
				t.Errorf("Publish sent %d WS frame(s) but no server validation error surfaced (aria=%q errText=%q)", framesDuringClick, aria, errText)
			}
		}
	})

	// Scenario 3 — happy path: a real title + Publish stores it and renders the
	// published status with the title echoed back.
	t.Run("PublishWithTitleSucceeds", func(t *testing.T) {
		ctx, cancel := newDraftCtx(t)
		defer cancel()
		consoleErrs := captureConsoleErrors(ctx)
		wsFrames := captureWSFrames(ctx)

		const title = "Hello World"
		var html, status, statusText string
		t.Cleanup(func() {
			if t.Failed() {
				dumpSignals(t, "PublishWithTitleSucceeds", html, consoleErrs(), wsFrames())
			}
		})

		if err := chromedp.Run(ctx,
			network.Enable(),
			chromedp.Navigate(serverURL),
			chromedp.WaitVisible(`input[name="title"]`, chromedp.ByQuery),
			chromedp.Sleep(1500*time.Millisecond), // WS connect + hydrate
			chromedp.SendKeys(`input[name="title"]`, title, chromedp.ByQuery),
			chromedp.Click(`button[name="publish"]`, chromedp.ByQuery),
			chromedp.Sleep(1200*time.Millisecond), // action round-trip
			chromedp.Evaluate(`(document.querySelector('[data-status]')||{}).getAttribute?.('data-status')||''`, &status),
			chromedp.Evaluate(`(document.querySelector('[data-status]')||{}).innerText||''`, &statusText),
			chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		); err != nil {
			t.Fatalf("publish-with-title run: %v", err)
		}

		if status != "published" {
			t.Errorf("data-status = %q, want \"published\" after submitting a valid title", status)
		}
		if !strings.Contains(statusText, title) {
			t.Errorf("published status text = %q, want it to echo the title %q", statusText, title)
		}
	})
}
