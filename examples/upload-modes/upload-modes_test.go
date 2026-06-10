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

// TestUploadModes_E2E drives all four modes in a real browser. It is gated on
// LVT_LOCAL_CLIENT (the path to a built client bundle with mode support) and a
// chromium binary, since it exercises unreleased client behaviour.
//
//	LVT_LOCAL_CLIENT=../../../client/.worktrees/upload-modes/dist/livetemplate-client.browser.js \
//	  go test ./examples/upload-modes/ -run E2E -v
func TestUploadModes_E2E(t *testing.T) {
	if os.Getenv("LVT_LOCAL_CLIENT") == "" {
		t.Skip("set LVT_LOCAL_CLIENT to a built client bundle to run the browser e2e")
	}

	_ = os.RemoveAll("storage")
	_ = os.RemoveAll(".uploads")
	t.Cleanup(func() {
		_ = os.RemoveAll("storage")
		_ = os.RemoveAll(".uploads")
	})

	ctrl := &UploadModesController{}
	srv := httptest.NewServer(newApp(ctrl))
	defer srv.Close()
	ctrl.baseURL = srv.URL

	img, err := filepath.Abs("testdata.png")
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	execPath := os.Getenv("LVT_CHROME")
	if execPath == "" {
		execPath = "chromium"
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(execPath),
		chromedp.NoSandbox,
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()
	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	defer cancelTimeout()

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
		chromedp.SetUploadFiles(`input[lvt-upload="proxied"]`, []string{img}, chromedp.ByQuery),
		chromedp.WaitVisible(`#proxied-result`, chromedp.ByQuery),
		chromedp.Text(`#proxied-result`, &proxiedText, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("proxied flow: %v", err)
	}
	if !strings.Contains(proxiedText, "/files/proxied/testdata.png") {
		t.Errorf("proxied result = %q, want it to contain the stored ref", proxiedText)
	}

	// Zero-disk: the Proxied upload staged nothing under .uploads (the dir may
	// exist because the Volume field created a temp manager, but it holds no
	// staged files from this upload).
	if n := countFiles(t, ".uploads"); n != 0 {
		t.Errorf("zero-disk violated: %d staged files under .uploads", n)
	}
	if _, err := os.Stat("storage/proxied/testdata.png"); err != nil {
		t.Errorf("proxied bytes not written to storage: %v", err)
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
	if os.Getenv("LVT_LOCAL_CLIENT") == "" {
		t.Skip("set LVT_LOCAL_CLIENT to a built client bundle to run the browser e2e")
	}

	_ = os.RemoveAll("storage")
	_ = os.RemoveAll(".uploads")
	t.Cleanup(func() {
		_ = os.RemoveAll("storage")
		_ = os.RemoveAll(".uploads")
	})

	ctrl := &UploadModesController{}
	srv := httptest.NewServer(newApp(ctrl))
	defer srv.Close()
	ctrl.baseURL = srv.URL

	img, err := filepath.Abs("testdata.png")
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	execPath := os.Getenv("LVT_CHROME")
	if execPath == "" {
		execPath = "chromium"
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(execPath), chromedp.NoSandbox)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()
	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	defer cancelTimeout()

	// Disable WebSocket before any page script runs: a socket that never opens
	// makes the client fall back to the HTTP transport.
	const killWS = `
		class DeadSocket {
			constructor(url){ this.url=url; this.readyState=3;
				setTimeout(()=>{ this.onerror&&this.onerror(new Event('error'));
					this.onclose&&this.onclose(new CloseEvent('close',{code:1006,wasClean:false})); },0); }
			send(){} close(){}
		}
		DeadSocket.CONNECTING=0; DeadSocket.OPEN=1; DeadSocket.CLOSING=2; DeadSocket.CLOSED=3;
		window.WebSocket = DeadSocket;`

	var proxiedText string
	if err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(killWS).Do(ctx)
			return err
		}),
		chromedp.Navigate(srv.URL),
		chromedp.WaitVisible(`input[lvt-upload="proxied"]`, chromedp.ByQuery),
		chromedp.SetUploadFiles(`input[lvt-upload="proxied"]`, []string{img}, chromedp.ByQuery),
		chromedp.WaitVisible(`#proxied-result`, chromedp.ByQuery),
		chromedp.Text(`#proxied-result`, &proxiedText, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("proxied flow (WS disabled): %v", err)
	}
	if !strings.Contains(proxiedText, "/files/proxied/testdata.png") {
		t.Errorf("proxied result = %q, want the stored ref", proxiedText)
	}
	if _, err := os.Stat("storage/proxied/testdata.png"); err != nil {
		t.Errorf("proxied bytes not written: %v", err)
	}
	if n := countFiles(t, ".uploads"); n != 0 {
		t.Errorf("zero-disk violated: %d staged files under .uploads", n)
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
