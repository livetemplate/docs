package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gorilla/websocket"
	e2etest "github.com/livetemplate/lvt/testing"
)

func TestMain(m *testing.M) {
	e2etest.CleanupChromeContainers()
	code := m.Run()
	e2etest.CleanupChromeContainers()
	os.Exit(code)
}

// startServer launches the test-server binary on a free port and waits
// until it's ready. Returns the port and a cleanup func.
func startServer(t *testing.T) (int, func()) {
	t.Helper()

	port, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port for server: %v", err)
	}
	portStr := fmt.Sprintf("%d", port)
	serverURL := fmt.Sprintf("http://localhost:%d/", port)

	t.Logf("Starting test server on port %s", portStr)
	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "PORT="+portStr, "LVT_DEV_MODE=true")

	// Capture stdout/stderr — exec.Cmd defaults stdio to /dev/null
	// otherwise, so a crashed server is silent.
	var serverLog bytes.Buffer
	cmd.Stdout = &serverLog
	cmd.Stderr = &serverLog

	// `go run .` forks a compiled child binary; SIGKILL on the `go run`
	// parent doesn't reliably propagate to that child, so the inherited
	// stdout pipe can stay open and stall a naive Wait() forever.
	// WaitDelay (Go 1.20+) closes the I/O goroutines after the delay so
	// Wait() returns even when the child holds the pipe open — bounded
	// reap, no zombies leaked across the suite's many startServer calls.
	cmd.WaitDelay = 2 * time.Second

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	cleanup := func() {
		if cmd.Process == nil {
			return
		}
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}

	// Cap individual readiness probes at 1s so a hung server can't
	// stall the loop.
	readyClient := &http.Client{Timeout: 1 * time.Second}

	const requiredSuccesses = 2
	consecutiveSuccesses := 0
	var lastErr error
	ready := false
	for i := 0; i < 50; i++ {
		resp, err := readyClient.Get(serverURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				consecutiveSuccesses++
				if consecutiveSuccesses >= requiredSuccesses {
					time.Sleep(100 * time.Millisecond)
					ready = true
					break
				}
			} else {
				lastErr = fmt.Errorf("status %d", resp.StatusCode)
				consecutiveSuccesses = 0
			}
		} else {
			lastErr = err
			consecutiveSuccesses = 0
		}
		time.Sleep(200 * time.Millisecond)
	}

	if !ready {
		cleanup()
		t.Logf("--- server stdout/stderr ---\n%s\n--- end server output ---", serverLog.String())
		t.Fatalf("Server failed to start within 10 seconds. Last error: %v", lastErr)
	}

	t.Logf("✅ Test server ready at %s", serverURL)
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("--- server stdout/stderr ---\n%s\n--- end server output ---", serverLog.String())
		}
	})

	return port, cleanup
}

func waitForWSClient(timeout time.Duration) chromedp.Action {
	return e2etest.WaitFor(
		`window.liveTemplateClient && window.liveTemplateClient.isReady()`,
		timeout,
	)
}

func uiStandardsCheck(t *testing.T, ctx context.Context) {
	t.Helper()
	var violations string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			const v = [];
			['onclick','onchange','oninput','onsubmit','onkeydown','onkeyup'].forEach(h => {
				document.querySelectorAll('[' + h + ']').forEach(el => v.push('inline ' + h + ' on <' + el.tagName.toLowerCase() + '>'));
			});
			document.querySelectorAll('[style]').forEach(el => {
				if (el.tagName !== 'INS' && el.tagName !== 'DEL' && !el.closest('[data-modal]') && !el.closest('[data-lvt-toast-stack]'))
					v.push('inline style on <' + el.tagName.toLowerCase() + '>');
			});
			if (!document.querySelector('meta[name="color-scheme"]')) v.push('missing color-scheme meta');
			if (document.documentElement.lang !== 'en') v.push('missing lang=en');
			const c = document.querySelector('.container');
			if (c && c.offsetWidth > 700) v.push('container too wide: ' + c.offsetWidth + 'px');
			return v.join('; ');
		})()`, &violations),
	)
	if err != nil {
		t.Fatalf("UI standards check failed: %v", err)
	}
	if violations != "" {
		t.Errorf("UI standard violations: %s", violations)
	}
}

// ============================================================================
// Tier A — JS on + WS on (default)
// ============================================================================

func TestPE_TierA_BrowserE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	serverPort, cleanup := startServer(t)
	defer cleanup()

	debugPort, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port for Chrome: %v", err)
	}

	chromeCmd := e2etest.StartDockerChrome(t, debugPort)
	defer e2etest.StopDockerChrome(t, debugPort)
	_ = chromeCmd

	chromeURL := fmt.Sprintf("http://localhost:%d", debugPort)
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), chromeURL)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tierAURL := fmt.Sprintf("http://host.docker.internal:%d/", serverPort)

	t.Run("Initial_Load", func(t *testing.T) {
		var html string
		err := chromedp.Run(ctx,
			chromedp.Navigate(tierAURL),
			e2etest.WaitForWebSocketReady(5*time.Second),
			chromedp.WaitVisible(`h1`, chromedp.ByQuery),
			e2etest.ValidateNoTemplateExpressions("[data-lvt-id]"),
			chromedp.OuterHTML(`html`, &html, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Failed to load page: %v", err)
		}
		if !strings.Contains(html, "Progressive Enhancement") {
			t.Error("Page title not found")
		}
		if !strings.Contains(html, `name="add"`) {
			t.Error("Expected button with name='add' in form")
		}
		if !strings.Contains(html, "livetemplate-client.js") {
			t.Error("Script tag for livetemplate-client.js not found")
		}
		t.Log("✅ Initial page load verified")
	})

	t.Run("UI_Standards", func(t *testing.T) {
		uiStandardsCheck(t, ctx)
	})

	t.Run("WebSocket_Connection", func(t *testing.T) {
		var wsConnected bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(`window.liveTemplateClient && window.liveTemplateClient.isReady()`, &wsConnected),
		)
		if err != nil {
			t.Fatalf("Failed to check WebSocket: %v", err)
		}
		if !wsConnected {
			t.Error("WebSocket not connected")
		}
		t.Log("✅ WebSocket connection established")
	})

	t.Run("Add_Todo", func(t *testing.T) {
		var initialCount int
		err := chromedp.Run(ctx,
			chromedp.Evaluate(`document.querySelectorAll('table tbody tr').length`, &initialCount),
		)
		if err != nil {
			t.Fatalf("Failed to count todos: %v", err)
		}

		expectedCount := initialCount + 1
		err = chromedp.Run(ctx,
			chromedp.SendKeys(`input[name="title"]`, "Tier A Todo", chromedp.ByQuery),
			chromedp.Click(`button[name="add"]`, chromedp.ByQuery),
			e2etest.WaitFor(fmt.Sprintf(`document.querySelectorAll('table tbody tr').length === %d`, expectedCount), 5*time.Second),
		)
		if err != nil {
			t.Fatalf("Failed to add todo: %v", err)
		}

		var inputValue string
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`document.querySelector('input[name="title"]').value`, &inputValue),
		)
		if err != nil {
			t.Fatalf("Failed to get input value: %v", err)
		}
		if inputValue != "" {
			t.Errorf("Input should be cleared after add, got: %q", inputValue)
		}
		t.Logf("✅ Todo added, count: %d → %d, input cleared", initialCount, expectedCount)
	})

	t.Run("Toggle_Todo", func(t *testing.T) {
		err := chromedp.Run(ctx,
			chromedp.Click(`table tbody tr:last-child button[name="toggle"]`, chromedp.ByQuery),
			e2etest.WaitFor(`document.querySelector('table tbody tr:last-child').querySelector('s') !== null`, 5*time.Second),
		)
		if err != nil {
			t.Fatalf("Failed to toggle todo: %v", err)
		}

		err = chromedp.Run(ctx,
			chromedp.Click(`table tbody tr:last-child button[name="toggle"]`, chromedp.ByQuery),
			e2etest.WaitFor(`document.querySelector('table tbody tr:last-child').querySelector('s') === null`, 5*time.Second),
		)
		if err != nil {
			t.Fatalf("Failed to untoggle todo: %v", err)
		}
		t.Log("✅ Toggle and untoggle work correctly")
	})

	t.Run("Delete_Todo", func(t *testing.T) {
		var beforeCount int
		err := chromedp.Run(ctx,
			chromedp.Evaluate(`document.querySelectorAll('table tbody tr').length`, &beforeCount),
		)
		if err != nil {
			t.Fatalf("Failed to count todos: %v", err)
		}

		expectedCount := beforeCount - 1
		err = chromedp.Run(ctx,
			chromedp.Click(`table tbody tr:last-child button[name="delete"]`, chromedp.ByQuery),
			e2etest.WaitFor(fmt.Sprintf(`document.querySelectorAll('table tbody tr').length === %d`, expectedCount), 5*time.Second),
		)
		if err != nil {
			t.Fatalf("Failed to delete todo: %v", err)
		}
		t.Logf("✅ Todo deleted, count: %d → %d", beforeCount, expectedCount)
	})

	t.Run("Validation_Error", func(t *testing.T) {
		err := chromedp.Run(ctx,
			network.ClearBrowserCookies(),
			chromedp.Navigate(tierAURL),
			e2etest.WaitForWebSocketReady(5*time.Second),
			chromedp.WaitVisible(`table`, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Failed to reload page: %v", err)
		}

		var beforeCount int
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`document.querySelectorAll('table tbody tr').length`, &beforeCount),
		)
		if err != nil {
			t.Fatalf("Failed to count todos: %v", err)
		}

		err = chromedp.Run(ctx,
			chromedp.SetValue(`input[name="title"]`, "ab", chromedp.ByQuery),
			chromedp.Click(`button[name="add"]`, chromedp.ByQuery),
			e2etest.WaitFor(`document.querySelector('[aria-invalid="true"]') !== null`, 5*time.Second),
		)
		if err != nil {
			t.Fatalf("Failed to submit invalid form: %v", err)
		}

		var afterCount int
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`document.querySelectorAll('table tbody tr').length`, &afterCount),
		)
		if err != nil {
			t.Fatalf("Failed to count after invalid submit: %v", err)
		}
		if afterCount != beforeCount {
			t.Errorf("Count should be unchanged after validation error: expected %d, got %d", beforeCount, afterCount)
		}
		t.Log("✅ Validation error handled correctly in browser")
	})
}

// ============================================================================
// Tier B — JS on + WS off (WithWebSocketDisabled)
// ============================================================================

func TestPE_TierB_BrowserE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	serverPort, cleanup := startServer(t)
	defer cleanup()

	debugPort, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port for Chrome: %v", err)
	}

	chromeCmd := e2etest.StartDockerChrome(t, debugPort)
	defer e2etest.StopDockerChrome(t, debugPort)
	_ = chromeCmd

	chromeURL := fmt.Sprintf("http://localhost:%d", debugPort)
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), chromeURL)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tierBURL := fmt.Sprintf("http://host.docker.internal:%d/no-ws/", serverPort)

	t.Run("PageLoads", func(t *testing.T) {
		var html string
		err := chromedp.Run(ctx,
			chromedp.Navigate(tierBURL),
			chromedp.WaitReady("body", chromedp.ByQuery),
			waitForWSClient(10*time.Second),
			chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("chromedp failed: %v", err)
		}
		if !strings.Contains(html, "Progressive Enhancement") {
			t.Error("Expected page title in HTML")
		}
		if !strings.Contains(html, `name="add"`) {
			t.Error("Expected button with name='add'")
		}
	})

	t.Run("UI_Standards", func(t *testing.T) {
		uiStandardsCheck(t, ctx)
	})

	t.Run("FormSubmission", func(t *testing.T) {
		err := chromedp.Run(ctx,
			network.ClearBrowserCookies(),
			chromedp.Navigate(tierBURL),
			chromedp.WaitReady("body", chromedp.ByQuery),
			waitForWSClient(10*time.Second),
		)
		if err != nil {
			t.Fatalf("Initial load failed: %v", err)
		}

		var htmlAfter string
		err = chromedp.Run(ctx,
			chromedp.SendKeys(`input[name="title"]`, "Tier B no-WS todo", chromedp.ByQuery),
			chromedp.Click(`button[name="add"]`, chromedp.ByQuery),
			e2etest.WaitFor(`document.body.innerText.includes('Tier B no-WS todo')`, 10*time.Second),
			chromedp.OuterHTML("html", &htmlAfter, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Form submission failed: %v", err)
		}
		if !strings.Contains(htmlAfter, "Tier B no-WS todo") {
			t.Error("Expected 'Tier B no-WS todo' in page after form submission")
		}

		// Form should have auto-reset after successful submit
		var inputVal string
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`document.querySelector('input[name="title"]').value`, &inputVal),
		)
		if err != nil {
			t.Logf("Warning: could not read input field: %v", err)
		} else if inputVal != "" {
			t.Errorf("Title input should be empty after submit, got %q", inputVal)
		}
	})

	t.Run("MultipleSubmissions", func(t *testing.T) {
		err := chromedp.Run(ctx,
			network.ClearBrowserCookies(),
			chromedp.Navigate(tierBURL),
			chromedp.WaitReady("body", chromedp.ByQuery),
			waitForWSClient(10*time.Second),
		)
		if err != nil {
			t.Fatalf("Initial load failed: %v", err)
		}

		err = chromedp.Run(ctx,
			chromedp.SendKeys(`input[name="title"]`, "First B todo", chromedp.ByQuery),
			chromedp.Click(`button[name="add"]`, chromedp.ByQuery),
			e2etest.WaitFor(`document.body.innerText.includes('First B todo')`, 10*time.Second),
		)
		if err != nil {
			t.Fatalf("First submission failed: %v", err)
		}

		var htmlAfterTwo string
		err = chromedp.Run(ctx,
			chromedp.Clear(`input[name="title"]`, chromedp.ByQuery),
			chromedp.SendKeys(`input[name="title"]`, "Second B todo", chromedp.ByQuery),
			chromedp.Click(`button[name="add"]`, chromedp.ByQuery),
			e2etest.WaitFor(`document.body.innerText.includes('Second B todo')`, 10*time.Second),
			chromedp.OuterHTML("html", &htmlAfterTwo, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Second submission failed: %v", err)
		}
		if !strings.Contains(htmlAfterTwo, "First B todo") {
			t.Error("Expected 'First B todo' in page")
		}
		if !strings.Contains(htmlAfterTwo, "Second B todo") {
			t.Error("Expected 'Second B todo' in page")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := chromedp.Run(ctx,
			network.ClearBrowserCookies(),
			chromedp.Navigate(tierBURL),
			chromedp.WaitReady("body", chromedp.ByQuery),
			waitForWSClient(10*time.Second),
			chromedp.SendKeys(`input[name="title"]`, "To delete in B", chromedp.ByQuery),
			chromedp.Click(`button[name="add"]`, chromedp.ByQuery),
			e2etest.WaitFor(`document.body.innerText.includes('To delete in B')`, 10*time.Second),
		)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// Find the row containing our text and click its delete button
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`(() => {
				const rows = document.querySelectorAll('table tbody tr');
				for (const r of rows) {
					if (r.textContent.includes('To delete in B')) {
						r.querySelector('button[name="delete"]').click();
						return true;
					}
				}
				return false;
			})()`, nil),
			e2etest.WaitFor(`!document.body.innerText.includes('To delete in B')`, 10*time.Second),
		)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})
}

// ============================================================================
// Tier B — HTTP-only sentinel: WebSocket upgrade is rejected
// ============================================================================

func TestPE_TierB_WebSocketRejected(t *testing.T) {
	serverPort, cleanup := startServer(t)
	defer cleanup()

	// A real WebSocket dial — gorilla/websocket sets the Sec-WebSocket-Key
	// and Sec-WebSocket-Version headers a valid handshake needs. Just
	// asserting "non-101 from a hand-rolled GET with bare Upgrade
	// headers" would false-pass even when WebSocket was enabled, since
	// the handshake itself is malformed.
	wsURL := fmt.Sprintf("ws://localhost:%d/no-ws/", serverPort)
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, resp, err := dialer.Dial(wsURL, nil)
	if conn != nil {
		conn.Close()
	}
	if err == nil {
		t.Error("WebSocket dial should have failed against the no-ws mount")
	} else if !errors.Is(err, websocket.ErrBadHandshake) {
		t.Logf("non-bad-handshake dial error (still acceptable): %v", err)
	}
	if resp != nil && resp.StatusCode == http.StatusSwitchingProtocols {
		t.Errorf("Expected non-101 response on no-ws mount, got 101")
	}
}

// ============================================================================
// Tier C — JS off, raw HTTP POST (POST-Redirect-GET)
// ============================================================================

func TestPE_TierC_NoJS_PostRedirectsToSee(t *testing.T) {
	serverPort, cleanup := startServer(t)
	defer cleanup()

	tierAURL := fmt.Sprintf("http://localhost:%d/", serverPort)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// GET initial page
	resp, err := client.Get(tierAURL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	cookies := resp.Cookies()
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// POST a new todo simulating a JS-disabled browser
	form := strings.NewReader("add=&title=Tier+C+raw+POST")
	req, err := http.NewRequest("POST", tierAURL, form)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected 303 See Other (PRG), got %d", resp.StatusCode)
	}
	if resp.Header.Get("Location") == "" {
		t.Error("Expected Location header in redirect response")
	}

	// Flash should be carried via lvt-flash cookie, not in the URL
	var flashCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "lvt-flash" {
			flashCookie = c
			break
		}
	}
	if flashCookie == nil {
		t.Fatal("Expected lvt-flash cookie to be set on PRG response")
	}
	if !strings.Contains(flashCookie.Value, "success=") {
		t.Errorf("Expected 'success=' in flash cookie, got: %s", flashCookie.Value)
	}
}

func TestPE_TierC_NoJS_HTTPToggle(t *testing.T) {
	serverPort, cleanup := startServer(t)
	defer cleanup()

	tierAURL := fmt.Sprintf("http://localhost:%d/", serverPort)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(tierAURL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	cookies := resp.Cookies()
	resp.Body.Close()

	form := strings.NewReader("toggle=&id=1")
	req, _ := http.NewRequest("POST", tierAURL, form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("POST toggle failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected 303 (See Other) for HTTP toggle, got %d", resp.StatusCode)
	}
}

func TestPE_TierC_NoJS_HTTPDelete(t *testing.T) {
	serverPort, cleanup := startServer(t)
	defer cleanup()

	tierAURL := fmt.Sprintf("http://localhost:%d/", serverPort)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(tierAURL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	cookies := resp.Cookies()
	resp.Body.Close()

	form := strings.NewReader("delete=&id=2")
	req, _ := http.NewRequest("POST", tierAURL, form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("POST delete failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected 303 (See Other) for HTTP delete, got %d", resp.StatusCode)
	}
}

func TestPE_TierC_ValidationError_JSON(t *testing.T) {
	serverPort, cleanup := startServer(t)
	defer cleanup()

	tierAURL := fmt.Sprintf("http://localhost:%d/", serverPort)
	client := &http.Client{}

	resp, err := client.Get(tierAURL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	cookies := resp.Cookies()
	resp.Body.Close()

	// Empty title — Accept: application/json so the JS client path returns
	// inline field errors as JSON, not the PRG redirect.
	form := strings.NewReader("add=&title=")
	req, _ := http.NewRequest("POST", tierAURL, form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 for JSON validation error, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "title") {
		t.Error("Expected 'title' field error in JSON response")
	}
}
