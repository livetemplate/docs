// Package redactedform_test exercises Preview-mode field redaction end-to-end
// in a real browser. It is the acceptance test for the data-lvt-redact feature
// (livetemplate client) + the lvt.Redact Go helper.
//
// Unlike the other docs recipes, this suite serves the LOCALLY-BUILT client
// bundle (client/dist/livetemplate-client.browser.js) rather than the CDN one:
// the feature under test is unreleased, so the shared CDN-fetching test helper
// would silently run the wrong code. The bundle is located by walking up from
// the test's working directory; the suite skips with a clear message if it (or
// Docker) is unavailable, so it degrades gracefully off the dev box.
//
// What it proves:
//   - The raw passport value is written to browser localStorage and NEVER
//     appears in any WebSocket frame the browser sends — only a
//     {redacted:true,field:"passport"} sentinel does.
//   - The server-side state reflects "provided, value never received".
//   - A non-redacted field (note) round-trips its real value normally.
//   - On render, the lvt.Redact span data-lvt-redact="passport" is filled back from
//     localStorage so the user sees their own value.
//
// Per project E2E standards the test captures browser console logs, server
// stderr, WebSocket frames, rendered HTML, and a screenshot on the run.
package redactedform_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"github.com/livetemplate/livetemplate"
	e2etest "github.com/livetemplate/lvt/testing"

	redactedform "github.com/livetemplate/docs/examples/redacted-form"
)

const rawPassport = "X1234567" // the secret that must never reach the server

func TestMain(m *testing.M) {
	e2etest.CleanupChromeContainers()
	code := m.Run()
	e2etest.CleanupChromeContainers()
	os.Exit(code)
}

// findRepoAsset walks up from cwd looking for client/<rel> (e.g.
// client/dist/livetemplate-client.browser.js). Returns "" if not found.
func findRepoAsset(rel string) string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, "client", rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// startServer serves the recipe's LiveHandler plus the LOCAL client bundle on
// an all-interfaces port (so the Dockerized Chrome can reach it via
// host.docker.internal). Returns the port and a cleanup func.
func startServer(t *testing.T, clientJS, clientCSS string) (int, func()) {
	t.Helper()

	port, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("free port: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", redactedform.LiveHandler(
		livetemplate.WithDevMode(true),
		livetemplate.WithPermissiveOriginCheck(),
	))
	mux.HandleFunc("/livetemplate-client.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, clientJS)
	})
	mux.HandleFunc("/livetemplate.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, clientCSS)
	})

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()

	// Readiness: two consecutive 200s.
	ready := false
	client := &http.Client{Timeout: time.Second}
	for i := 0; i < 50 && !ready; i++ {
		ok := 0
		for j := 0; j < 2; j++ {
			resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					ok++
				}
			}
		}
		if ok == 2 {
			ready = true
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if !ready {
		_ = srv.Close()
		t.Fatal("server failed to become ready")
	}

	return port, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
}

func TestRedactedForm_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E in short mode")
	}

	clientJS := findRepoAsset(filepath.Join("dist", "livetemplate-client.browser.js"))
	clientCSS := findRepoAsset("livetemplate.css")
	if clientJS == "" || clientCSS == "" {
		t.Skip("local client bundle not found (run `npm run build` in client/); skipping unreleased-client E2E")
	}
	t.Logf("serving local client bundle: %s", clientJS)

	serverPort, cleanup := startServer(t, clientJS, clientCSS)
	defer cleanup()

	debugPort, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("free port for chrome: %v", err)
	}
	if err := e2etest.StartDockerChrome(t, debugPort); err != nil {
		t.Skipf("Docker Chrome unavailable: %v", err)
	}
	defer e2etest.StopDockerChrome(t, debugPort)

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(
		context.Background(),
		fmt.Sprintf("http://localhost:%d", debugPort),
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Capture WebSocket frames the browser SENDS, plus console logs.
	var mu sync.Mutex
	var sentFrames []string
	var consoleLogs []string
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventWebSocketFrameSent:
			mu.Lock()
			sentFrames = append(sentFrames, e.Response.PayloadData)
			mu.Unlock()
		case *runtime.EventConsoleAPICalled:
			parts := make([]string, 0, len(e.Args))
			for _, a := range e.Args {
				parts = append(parts, string(a.Value))
			}
			mu.Lock()
			consoleLogs = append(consoleLogs, fmt.Sprintf("[%s] %s", e.Type, strings.Join(parts, " ")))
			mu.Unlock()
		}
	})

	url := e2etest.GetChromeTestURL(serverPort)
	var serverPassportHTML, echoPassportHTML, serverNoteHTML string
	var localStorageDump string
	var screenshot []byte

	err = chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`#passport`, chromedp.ByQuery),
		e2etest.WaitForWebSocketReady(10*time.Second),
		// Fill both fields and submit.
		chromedp.SendKeys(`#passport`, rawPassport, chromedp.ByQuery),
		chromedp.SendKeys(`#note`, "hello world", chromedp.ByQuery),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(1500*time.Millisecond),
		// Read back rendered state + localStorage.
		chromedp.OuterHTML(`#server-passport`, &serverPassportHTML, chromedp.ByQuery),
		chromedp.OuterHTML(`#server-note`, &serverNoteHTML, chromedp.ByQuery),
		chromedp.OuterHTML(`#echo-passport`, &echoPassportHTML, chromedp.ByQuery),
		chromedp.Evaluate(`JSON.stringify(Object.fromEntries(Object.entries(localStorage)))`, &localStorageDump),
		chromedp.FullScreenshot(&screenshot, 90),
	)

	// Always dump diagnostics on failure (per E2E standards).
	dumpDiag := func() {
		mu.Lock()
		defer mu.Unlock()
		t.Logf("--- WebSocket frames sent (%d) ---", len(sentFrames))
		for i, f := range sentFrames {
			t.Logf("  [%d] %s", i, f)
		}
		t.Logf("--- console logs (%d) ---", len(consoleLogs))
		for _, l := range consoleLogs {
			t.Logf("  %s", l)
		}
		t.Logf("--- localStorage ---\n%s", localStorageDump)
		t.Logf("--- #server-passport ---\n%s", serverPassportHTML)
		t.Logf("--- #echo-passport ---\n%s", echoPassportHTML)
	}

	if err != nil {
		dumpDiag()
		t.Fatalf("chromedp run: %v", err)
	}

	// Save screenshot artifact next to the test for visual inspection.
	if len(screenshot) > 0 {
		shot := filepath.Join(t.TempDir(), "redacted-form.png")
		if werr := os.WriteFile(shot, screenshot, 0o644); werr == nil {
			t.Logf("screenshot: %s", shot)
		}
	}

	mu.Lock()
	frames := append([]string(nil), sentFrames...)
	mu.Unlock()

	// 1. The browser must have sent at least one frame carrying the redact
	//    sentinel for the passport field.
	sawSentinel := false
	for _, f := range frames {
		if strings.Contains(f, `"redacted":true`) && strings.Contains(f, `"field":"passport"`) {
			sawSentinel = true
			break
		}
	}
	if !sawSentinel {
		dumpDiag()
		t.Error("no WebSocket frame carried the passport redact sentinel — redaction did not run on the send path")
	}

	// 2. The raw passport value must NEVER appear in any sent frame.
	for i, f := range frames {
		if strings.Contains(f, rawPassport) {
			dumpDiag()
			t.Errorf("raw passport value leaked in WebSocket frame [%d]: %s", i, f)
		}
	}

	// 3. The server stored only the presence flag, not the value.
	if !strings.Contains(serverPassportHTML, "provided (value never received)") {
		dumpDiag()
		t.Errorf("server-passport panel = %q; expected the never-received marker", serverPassportHTML)
	}
	if strings.Contains(serverPassportHTML, rawPassport) {
		dumpDiag()
		t.Errorf("raw passport value rendered in server panel: %s", serverPassportHTML)
	}

	// 4. The non-redacted note round-tripped its real value.
	if !strings.Contains(serverNoteHTML, "hello world") {
		dumpDiag()
		t.Errorf("server-note panel = %q; expected the real note value", serverNoteHTML)
	}

	// 5. The lvt.Redact <span data-lvt-redact="passport"> echo was filled from
	//    localStorage (textContent), so the user sees their own value even though
	//    the server emitted an empty marked span. The attribute being the trust
	//    signal — not a free token — is what prevents user-posted content from
	//    triggering substitution.
	if !strings.Contains(echoPassportHTML, rawPassport) {
		dumpDiag()
		t.Errorf("echo-passport = %q; expected the hydrated real value %q", echoPassportHTML, rawPassport)
	}
	if !strings.Contains(echoPassportHTML, `data-lvt-redact="passport"`) {
		dumpDiag()
		t.Errorf("echo-passport = %q; expected the server-emitted data-lvt-redact span", echoPassportHTML)
	}

	// 6. localStorage holds the raw value under a redact-scoped key.
	if !strings.Contains(localStorageDump, rawPassport) {
		dumpDiag()
		t.Errorf("localStorage does not hold the passport value: %s", localStorageDump)
	}
}
