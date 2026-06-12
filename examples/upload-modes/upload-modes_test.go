package main

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// killWS stubs window.WebSocket with a socket that never opens, before any page
// script runs, so the client falls back to the HTTP transport — the WS-disabled
// path the *WSDisabled_E2E tests exercise.
const killWS = `
	class DeadSocket {
		constructor(url){ this.url=url; this.readyState=3;
			setTimeout(()=>{ this.onerror&&this.onerror(new Event('error'));
				this.onclose&&this.onclose(new CloseEvent('close',{code:1006,wasClean:false})); },0); }
		send(){} close(){}
	}
	DeadSocket.CONNECTING=0; DeadSocket.OPEN=1; DeadSocket.CLOSING=2; DeadSocket.CLOSED=3;
	window.WebSocket = DeadSocket;`

// requireUploadModesE2E skips unless the opt-in env vars are set. These tests
// drive a locally-installed Chromium via ExecAllocator and exercise unreleased
// client behaviour, so they're verified locally; wiring them into the docs
// Docker-chrome CI harness is tracked in livetemplate/docs#67.
func requireUploadModesE2E(t *testing.T) {
	t.Helper()
	if os.Getenv("LVT_UPLOAD_MODES_E2E") == "" || os.Getenv("LVT_LOCAL_CLIENT") == "" {
		t.Skip("set LVT_UPLOAD_MODES_E2E=1 and LVT_LOCAL_CLIENT to run the browser e2e (docs#67)")
	}
}

// newChromiumCtx starts a local Chromium via ExecAllocator and returns a chromedp
// context with a 60s deadline; all cancels are registered on t.Cleanup.
func newChromiumCtx(t *testing.T) context.Context {
	t.Helper()
	execPath := os.Getenv("LVT_CHROME")
	if execPath == "" {
		execPath = "chromium"
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(execPath), chromedp.NoSandbox)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	t.Cleanup(cancelAlloc)
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	t.Cleanup(cancelCtx)
	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	t.Cleanup(cancelTimeout)
	return ctx
}

// newUploadModesApp clears the storage dirs, starts the example server, and
// returns it with its base URL wired into the controller's presigner.
func newUploadModesApp(t *testing.T) (*httptest.Server, *UploadModesController) {
	t.Helper()
	_ = os.RemoveAll("storage")
	_ = os.RemoveAll(".uploads")
	t.Cleanup(func() {
		_ = os.RemoveAll("storage")
		_ = os.RemoveAll(".uploads")
	})
	ctrl := &UploadModesController{}
	srv := httptest.NewServer(newApp(ctrl))
	t.Cleanup(srv.Close)
	ctrl.baseURL = srv.URL
	return srv, ctrl
}

// TestUploadModes_E2E drives all four modes in a real browser. It is gated on
// LVT_LOCAL_CLIENT (the path to a built client bundle with mode support) and a
// chromium binary, since it exercises unreleased client behaviour.
//
//	LVT_LOCAL_CLIENT=../../../client/.worktrees/upload-modes/dist/livetemplate-client.browser.js \
//	  go test ./examples/upload-modes/ -run E2E -v
func TestUploadModes_E2E(t *testing.T) {
	requireUploadModesE2E(t)

	srv, _ := newUploadModesApp(t)

	img, err := filepath.Abs("testdata.png")
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	ctx := newChromiumCtx(t)

	// Surface browser console output in the test log for debugging.
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*runtime.EventConsoleAPICalled); ok {
			var parts []string
			for _, a := range e.Args {
				parts = append(parts, strings.Trim(string(a.Value), `"`))
			}
			t.Logf("console.%s: %s", e.Type, strings.Join(parts, " "))
		}
	})

	// Proxied: selecting the file auto-uploads via a multipart POST that streams
	// straight to OnUpload; the page shows the resulting reference.
	var proxiedText string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(srv.URL),
		chromedp.WaitVisible(`input[lvt-upload="proxied"]`, chromedp.ByQuery),
		// Set the record id before selecting the file: the client serializes it
		// into the multipart POST ahead of the file part, and OnUpload reads it.
		chromedp.SetValue(`#proxied-record-id`, "invoice-42", chromedp.ByQuery),
		chromedp.SetUploadFiles(`input[lvt-upload="proxied"]`, []string{img}, chromedp.ByQuery),
		chromedp.WaitVisible(`#proxied-result`, chromedp.ByQuery),
		chromedp.Text(`#proxied-result`, &proxiedText, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("proxied flow: %v", err)
	}
	// The record id reached OnUpload, so the bytes are routed under its folder —
	// proof the form field rode the streaming POST and was readable mid-stream.
	if !strings.Contains(proxiedText, "/files/proxied/invoice-42/testdata.png") {
		t.Errorf("proxied result = %q, want the ref routed under record id invoice-42", proxiedText)
	}

	// Zero-disk: the Proxied upload staged nothing under .uploads (the dir may
	// exist because the Volume field created a temp manager, but it holds no
	// staged files from this upload).
	if n := countFiles(t, ".uploads"); n != 0 {
		t.Errorf("zero-disk violated: %d staged files under .uploads", n)
	}
	if _, err := os.Stat("storage/proxied/invoice-42/testdata.png"); err != nil {
		t.Errorf("proxied bytes not written under the record id: %v", err)
	}

	// Preview: selecting the file shows a local object URL and uploads nothing.
	var previewSrc string
	if err := chromedp.Run(ctx,
		chromedp.SetUploadFiles(`input[lvt-upload="preview"]`, []string{img}, chromedp.ByQuery),
		chromedp.Sleep(750*time.Millisecond),
		chromedp.AttributeValue(`img[data-lvt-upload-preview="preview"]`, "src", &previewSrc, nil, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("preview flow: %v", err)
	}
	if !strings.HasPrefix(previewSrc, "blob:") {
		t.Errorf("preview src = %q, want a blob: URL", previewSrc)
	}

	// Direct: selecting the file PUTs straight to the presigned sink; the page
	// shows the stored reference.
	var directText string
	if err := chromedp.Run(ctx,
		chromedp.SetUploadFiles(`input[lvt-upload="direct"]`, []string{img}, chromedp.ByQuery),
		chromedp.WaitVisible(`#direct-result`, chromedp.ByQuery),
		chromedp.Text(`#direct-result`, &directText, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("direct flow: %v", err)
	}
	if !strings.Contains(directText, "/files/direct/testdata.png") {
		t.Errorf("direct result = %q, want the stored ref", directText)
	}
	if _, err := os.Stat("storage/direct/testdata.png"); err != nil {
		t.Errorf("direct bytes not written to sink: %v", err)
	}
}

// TestUploadModes_ProxiedWSDisabled_E2E proves Proxied works with the WebSocket
// unavailable: a stubbed window.WebSocket that never opens forces the client onto
// the HTTP transport, so the upload_start handshake and the multipart POST both
// go over HTTP. Same gating as the main e2e.
func TestUploadModes_ProxiedWSDisabled_E2E(t *testing.T) {
	requireUploadModesE2E(t)

	srv, _ := newUploadModesApp(t)

	img, err := filepath.Abs("testdata.png")
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	ctx := newChromiumCtx(t)

	var proxiedText string
	if err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(killWS).Do(ctx)
			return err
		}),
		chromedp.Navigate(srv.URL),
		chromedp.WaitVisible(`input[lvt-upload="proxied"]`, chromedp.ByQuery),
		// The form field rides the multipart POST on the HTTP-fallback path too.
		chromedp.SetValue(`#proxied-record-id`, "offline-9", chromedp.ByQuery),
		chromedp.SetUploadFiles(`input[lvt-upload="proxied"]`, []string{img}, chromedp.ByQuery),
		chromedp.WaitVisible(`#proxied-result`, chromedp.ByQuery),
		chromedp.Text(`#proxied-result`, &proxiedText, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("proxied flow (WS disabled): %v", err)
	}
	if !strings.Contains(proxiedText, "/files/proxied/offline-9/testdata.png") {
		t.Errorf("proxied result = %q, want the ref routed under record id offline-9", proxiedText)
	}
	if _, err := os.Stat("storage/proxied/offline-9/testdata.png"); err != nil {
		t.Errorf("proxied bytes not written under the record id: %v", err)
	}
	if n := countFiles(t, ".uploads"); n != 0 {
		t.Errorf("zero-disk violated: %d staged files under .uploads", n)
	}
}

// TestUploadModes_DirectWSDisabled_E2E proves Direct completes with the WebSocket
// unavailable (livetemplate#448): upload_start presigns over HTTP, the browser
// PUTs to the sink, then the client re-sends the entry metadata + ref over an
// HTTP upload_complete handshake so UploadDirectComplete runs and renders the ref.
func TestUploadModes_DirectWSDisabled_E2E(t *testing.T) {
	requireUploadModesE2E(t)

	srv, _ := newUploadModesApp(t)

	img, err := filepath.Abs("testdata.png")
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	ctx := newChromiumCtx(t)

	var directText string
	if err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(killWS).Do(ctx)
			return err
		}),
		chromedp.Navigate(srv.URL),
		chromedp.WaitVisible(`input[lvt-upload="direct"]`, chromedp.ByQuery),
		chromedp.SetUploadFiles(`input[lvt-upload="direct"]`, []string{img}, chromedp.ByQuery),
		// #direct-result only renders after the HTTP upload_complete handshake
		// reconstructs the entry and UploadDirectComplete sets the ref.
		chromedp.WaitVisible(`#direct-result`, chromedp.ByQuery),
		chromedp.Text(`#direct-result`, &directText, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("direct flow (WS disabled): %v", err)
	}
	if !strings.Contains(directText, "/files/direct/testdata.png") {
		t.Errorf("direct result = %q, want the stored ref", directText)
	}
	if _, err := os.Stat("storage/direct/testdata.png"); err != nil {
		t.Errorf("direct bytes not PUT to the sink: %v", err)
	}
}

// TestUploadModes_VolumeWSDisabled_E2E proves Volume completes with the WebSocket
// unavailable (livetemplate#449): selecting the file falls back to a single
// multipart POST that the server stages to the field's Dir (retained), and
// UploadVolumeComplete renders the on-disk path.
func TestUploadModes_VolumeWSDisabled_E2E(t *testing.T) {
	requireUploadModesE2E(t)

	srv, _ := newUploadModesApp(t)

	img, err := filepath.Abs("testdata.png")
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	ctx := newChromiumCtx(t)

	var volumeText string
	if err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(killWS).Do(ctx)
			return err
		}),
		chromedp.Navigate(srv.URL),
		chromedp.WaitVisible(`input[lvt-upload="volume"]`, chromedp.ByQuery),
		chromedp.SetUploadFiles(`input[lvt-upload="volume"]`, []string{img}, chromedp.ByQuery),
		chromedp.WaitVisible(`#volume-result`, chromedp.ByQuery),
		chromedp.Text(`#volume-result`, &volumeText, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("volume flow (WS disabled): %v", err)
	}
	// The retained path is under the configured Dir (storage/volume), not the
	// ephemeral .uploads tree.
	if !strings.Contains(volumeText, "storage/volume") {
		t.Errorf("volume result = %q, want a path under the Dir storage/volume", volumeText)
	}
	matches, _ := filepath.Glob(filepath.Join("storage", "volume", "volume", "*"))
	if len(matches) != 1 {
		t.Errorf("expected exactly 1 retained file under storage/volume/volume, got %d", len(matches))
	}
}

// countFiles returns the number of regular files under root (0 if absent).
func countFiles(t *testing.T, root string) int {
	t.Helper()
	count := 0
	_ = filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info != nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}
