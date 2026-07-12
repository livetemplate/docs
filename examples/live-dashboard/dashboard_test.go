// Package livedashboard_test exercises the live-dashboard recipe end-to-end.
//
//	TestLiveDashboard_SmokeHTTP — no Chrome: the initial HTTP render carries the
//	                              metrics markup (runs in -short).
//	TestLiveDashboard_E2E       — chromedp: the background goroutine advances the
//	                              counter with no user interaction, and TWO tabs
//	                              are refreshed by that ONE goroutine in lockstep.
//
// Per project rules, the chromedp tests capture browser console logs, WebSocket
// frames, the server's stdout/stderr, and the rendered HTML — all dumped on
// failure so a regression is diagnosable without a rerun.
package livedashboard_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/livetemplate/livetemplate"
	e2etest "github.com/livetemplate/lvt/testing"

	livedashboard "github.com/livetemplate/docs/examples/live-dashboard"
)

func TestMain(m *testing.M) {
	e2etest.CleanupChromeContainers()
	code := m.Run()
	e2etest.CleanupChromeContainers()
	os.Exit(code)
}

// TestLiveDashboard_SmokeHTTP checks the recipe wires up and the initial render
// carries the metrics markup — no Chrome required, so it runs in -short.
func TestLiveDashboard_SmokeHTTP(t *testing.T) {
	handler := livedashboard.Handler(
		livetemplate.WithDevMode(true), // also relaxes the WS origin check
	)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body := new(bytes.Buffer)
	if _, err := body.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	for _, want := range []string{`id="ticks"`, `id="jobs"`, `id="updated"`, "Live Ops Dashboard"} {
		if !strings.Contains(body.String(), want) {
			t.Errorf("initial render missing %q", want)
		}
	}
}

// startServer launches the recipe binary on a free port and waits until ready.
// Anonymous auth means / returns 200 (no BasicAuth challenge).
func startServer(t *testing.T) (int, func()) {
	t.Helper()

	port, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("free port: %v", err)
	}
	cmd := exec.Command("go", "run", "./cmd")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port), "LVT_DEV_MODE=true")

	var serverLog bytes.Buffer
	cmd.Stdout = &serverLog
	cmd.Stderr = &serverLog
	cmd.WaitDelay = 2 * time.Second

	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	cleanup := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}

	e2etest.WaitForServer(t, fmt.Sprintf("http://localhost:%d/", port), 15*time.Second)
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("--- server stdout/stderr ---\n%s\n--- end server output ---", serverLog.String())
		}
	})
	return port, cleanup
}

// attachCapture wires browser console logs and WebSocket frames to t.Logf so a
// failing run is diagnosable. Console entries surface client-side errors; WS
// frames show the tree updates the background Publish produced.
func attachCapture(ctx context.Context, t *testing.T, tab string) {
	t.Helper()
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			var parts []string
			for _, arg := range e.Args {
				parts = append(parts, strings.Trim(string(arg.Value), `"`))
			}
			t.Logf("[%s console.%s] %s", tab, e.Type, strings.Join(parts, " "))
		case *runtime.EventExceptionThrown:
			t.Logf("[%s console.exception] %s", tab, e.ExceptionDetails.Text)
		case *network.EventWebSocketFrameReceived:
			t.Logf("[%s ws recv] %s", tab, e.Response.PayloadData)
		case *network.EventWebSocketFrameSent:
			t.Logf("[%s ws sent] %s", tab, e.Response.PayloadData)
		}
	})
}

// readTicks reads the integer in #ticks for the given chromedp context.
func readTicks(t *testing.T, ctx context.Context, tab string) int {
	t.Helper()
	var text string
	if err := chromedp.Run(ctx, chromedp.Text("#ticks", &text, chromedp.ByQuery)); err != nil {
		t.Fatalf("%s: read #ticks: %v", tab, err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		t.Fatalf("%s: #ticks %q not an int: %v", tab, text, err)
	}
	return n
}

func newTab(t *testing.T, allocCtx context.Context, appURL, tab string) (context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	attachCapture(ctx, t, tab)
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(appURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		e2etest.WaitFor(`window.liveTemplateClient && window.liveTemplateClient.isReady()`, 10*time.Second),
	)
	if err != nil {
		var html string
		_ = chromedp.Run(ctx, chromedp.OuterHTML("body", &html, chromedp.ByQuery))
		t.Logf("[%s] rendered HTML on failure:\n%s", tab, html)
		cancel()
		t.Fatalf("%s: load %s: %v", tab, appURL, err)
	}
	return ctx, cancel
}

func TestLiveDashboard_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	serverPort, cleanup := startServer(t)
	defer cleanup()

	debugPort, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("free port for Chrome: %v", err)
	}
	if err := e2etest.StartDockerChrome(t, debugPort); err != nil {
		t.Fatalf("start Docker Chrome: %v", err)
	}
	defer e2etest.StopDockerChrome(t, debugPort)

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(),
		fmt.Sprintf("http://localhost:%d", debugPort))
	defer allocCancel()

	appURL := fmt.Sprintf("http://host.docker.internal:%d/", serverPort)

	// One tab: the counter must advance with NO user interaction — proof the
	// background goroutine drives the render.
	ctxA, cancelA := newTab(t, allocCtx, appURL, "tabA")
	defer cancelA()

	startA := readTicks(t, ctxA, "tabA")
	if err := chromedp.Run(ctxA, e2etest.WaitFor(
		fmt.Sprintf(`parseInt(document.getElementById('ticks').textContent) > %d`, startA),
		5*time.Second,
	)); err != nil {
		t.Fatalf("tabA: background push never advanced #ticks past %d: %v", startA, err)
	}

	t.Run("UI_Standards", func(t *testing.T) {
		var violations string
		if err := chromedp.Run(ctxA, chromedp.Evaluate(`(() => {
			const v = [];
			['onclick','onchange','oninput','onsubmit'].forEach(h => {
				document.querySelectorAll('[' + h + ']').forEach(el => v.push('inline ' + h));
			});
			document.querySelectorAll('[style]').forEach(el => v.push('inline style on <' + el.tagName.toLowerCase() + '>'));
			if (!document.querySelector('meta[name="color-scheme"]')) v.push('missing color-scheme meta');
			if (document.documentElement.lang !== 'en') v.push('missing lang=en');
			return v.join('; ');
		})()`, &violations)); err != nil {
			t.Fatalf("UI standards check: %v", err)
		}
		if violations != "" {
			t.Errorf("UI standard violations: %s", violations)
		}
	})

	// Second tab (a distinct session group under AnonymousAuthenticator): the
	// SAME one goroutine must refresh it too. Both tabs advance from their own
	// starting points with no interaction.
	t.Run("TwoTabs_OneGoroutineRefreshesBoth", func(t *testing.T) {
		ctxB, cancelB := newTab(t, allocCtx, appURL, "tabB")
		defer cancelB()

		startA := readTicks(t, ctxA, "tabA")
		startB := readTicks(t, ctxB, "tabB")

		advanced := func(ctx context.Context, tab string, from int) chromedp.Action {
			return e2etest.WaitFor(
				fmt.Sprintf(`parseInt(document.getElementById('ticks').textContent) > %d`, from),
				5*time.Second,
			)
		}
		if err := chromedp.Run(ctxA, advanced(ctxA, "tabA", startA)); err != nil {
			t.Errorf("tabA not refreshed by background goroutine past %d: %v", startA, err)
		}
		if err := chromedp.Run(ctxB, advanced(ctxB, "tabB", startB)); err != nil {
			t.Errorf("tabB not refreshed by background goroutine past %d: %v", startB, err)
		}
	})
}
