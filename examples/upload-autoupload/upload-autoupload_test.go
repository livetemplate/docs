package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	e2etest "github.com/livetemplate/lvt/testing"
)

// setFileJS sets a 1x1 PNG on the SSR'd file input and fires a `change` event.
// AutoUpload triggers on `change`, so this is exactly the gesture that did
// nothing before the #453 fix (the change handler was never bound). We dispatch
// the event manually rather than submitting a form because there is no form —
// the upload must start purely from selection. chromedp.SetUploadFiles is not
// usable here because Chrome runs in Docker and can't see host filesystem paths.
const setFileJS = `
(() => {
  const b64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVQI12P4z8AAAMBBBQAB1x2RAAAASElEQVQI12P4z8BQDwCNAQz/cWMmRQAAAABJRU5ErkJggg==';
  const binary = atob(b64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  const file = new File([bytes], 'auto.png', { type: 'image/png' });
  const input = document.querySelector('#avatar');
  const dt = new DataTransfer();
  dt.items.add(file);
  input.files = dt.files;
  input.dispatchEvent(new Event('change', { bubbles: true }));
  return 'file set (' + input.files.length + ')';
})()
`

// waitForSentFrame polls the captured WebSocket frames for a client-sent frame
// whose payload contains pattern. Asserting the client SENT upload_start is the
// precise #453 signal — it proves the change handler was bound — and is
// independent of whether the upload completes end-to-end (which depends on
// client/server protocol compatibility, orthogonal to this fix).
func waitForSentFrame(wl *e2etest.WSMessageLogger, pattern string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, m := range wl.GetSent() {
			if strings.Contains(m.Data, pattern) {
				return true
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// TestAutoUploadSSRBindsOnConnect is the browser regression for issue #453. It
// loads an app that SSRs an lvt-upload + AutoUpload input and re-renders it
// identically on WebSocket connect (hydrate-idempotent), selects a file, and
// asserts an upload_start frame is sent — proving the change handler was bound
// on connect rather than only from a node-adding render.
//
// The browser must load the locally built client bundle that carries the fix
// (the published @latest predates it), so this test is gated on
// LVT_LOCAL_CLIENT_JS pointing at that bundle. main.go serves it directly. Run:
//
//	(cd ../../../client && npm run build)
//	LVT_LOCAL_CLIENT_JS=$(cd ../../../client && pwd)/dist/livetemplate-client.browser.js \
//	  go test -run TestAutoUploadSSRBindsOnConnect ./examples/upload-autoupload/...
func TestAutoUploadSSRBindsOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	bundle := os.Getenv("LVT_LOCAL_CLIENT_JS")
	if bundle == "" {
		t.Skip("set LVT_LOCAL_CLIENT_JS to a freshly built @livetemplate/client bundle (with the #453 fix) to run this regression e2e")
	}
	if _, err := os.Stat(bundle); err != nil {
		t.Fatalf("LVT_LOCAL_CLIENT_JS=%q is not readable: %v", bundle, err)
	}

	serverPort, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port for server: %v", err)
	}

	// StartTestServer runs `go run main.go` inheriting our env, so the spawned
	// server picks up LVT_LOCAL_CLIENT_JS and serves the bundle under test.
	serverCmd := e2etest.StartTestServer(t, "main.go", serverPort)
	defer func() {
		if serverCmd != nil && serverCmd.Process != nil {
			_ = serverCmd.Process.Kill()
		}
	}()

	chromeCtx, cleanup := e2etest.SetupDockerChrome(t, 120*time.Second)
	defer cleanup()
	ctx := chromeCtx.Context

	// Diagnostics: console logs, WS frames, and rendered HTML on failure.
	wsLogger := e2etest.NewWSMessageLogger()
	consoleLogger := e2etest.NewConsoleLogger()
	wsLogger.Start(ctx)
	consoleLogger.Start(ctx)
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		t.Fatalf("Failed to enable network domain: %v", err)
	}

	defer func() {
		if t.Failed() {
			wsLogger.Print()
			consoleLogger.Print()
			var html string
			_ = chromedp.Run(ctx, chromedp.OuterHTML("body", &html, chromedp.ByQuery))
			t.Logf("Rendered HTML at failure:\n%s", html)
		}
	}()

	url := e2etest.GetChromeTestURL(serverPort)
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`#avatar`, chromedp.ByID),
		e2etest.WaitForWebSocketReady(10*time.Second),
	); err != nil {
		t.Fatalf("Navigation/hydrate failed: %v", err)
	}

	// Select a file — the only gesture. With AutoUpload this must fire the bound
	// change handler and send upload_start.
	var setResult string
	if err := chromedp.Run(ctx, chromedp.Evaluate(setFileJS, &setResult)); err != nil {
		t.Fatalf("Failed to set file on input: %v", err)
	}
	t.Logf("file selection: %s", setResult)

	// The regression assertion: before the fix, no upload_start frame is sent
	// because the SSR'd input's change handler was never bound.
	if !waitForSentFrame(wsLogger, "upload_start", 10*time.Second) {
		t.Fatalf("no upload_start frame after selecting a file — SSR'd lvt-upload change handler was not bound (issue #453)")
	}

	if consoleLogger.HasErrors() {
		t.Errorf("browser console reported errors: %s", fmt.Sprint(consoleLogger.GetErrors()))
	}

	t.Log("✅ SSR'd lvt-upload AutoUpload: change handler bound on connect, upload_start sent on select")
}
