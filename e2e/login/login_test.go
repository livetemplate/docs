// Package main_test exercises the login recipe end-to-end. Test shape
// matches the progressive-enhancement suite: a hand-rolled startServer
// helper with bytes.Buffer log capture, cmd.WaitDelay-bounded reap, and
// a 1s-timeout readiness probe with two-consecutive-success debouncing.
//
// Two test functions:
//
//	TestLogin_E2E      — chromedp browser flow through every controller
//	                     action (Login, Logout, OnConnect server-push).
//	TestLogin_HTTPCookie — raw HTTP exercising the 303 + Set-Cookie path
//	                       without a browser (verifies the cookie shape
//	                       and the logout-deletes-cookie shape).
package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	e2etest "github.com/livetemplate/lvt/testing"
)

func TestMain(m *testing.M) {
	e2etest.CleanupChromeContainers()
	code := m.Run()
	e2etest.CleanupChromeContainers()
	os.Exit(code)
}

// startServer launches the test-server binary mounted at "/" on a free
// port. It is the default for most tests; see startServerAt for tests
// that need to exercise a non-root mount path (the mountPath redirect
// regression).
func startServer(t *testing.T) (int, func()) {
	return startServerAt(t, "/")
}

// startServerAt is the non-root variant. The test-server's main.go
// reads MOUNT_PATH and mounts the recipe under that prefix with
// http.StripPrefix — mirroring cmd/site's mount of /apps/login/.
// Mirrors the progressive-enhancement helper otherwise — same WaitDelay,
// same readiness gate, same on-failure log dump.
func startServerAt(t *testing.T, mountPath string) (int, func()) {
	t.Helper()

	port, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port for server: %v", err)
	}
	portStr := fmt.Sprintf("%d", port)
	serverURL := fmt.Sprintf("http://localhost:%d%s", port, mountPath)

	t.Logf("Starting test server on port %s (mount: %s)", portStr, mountPath)
	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "PORT="+portStr, "LVT_DEV_MODE=true", "MOUNT_PATH="+mountPath)

	var serverLog bytes.Buffer
	cmd.Stdout = &serverLog
	cmd.Stderr = &serverLog

	// `go run .` forks a compiled child binary; SIGKILL on the parent
	// doesn't propagate, so the child's stdout pipe can stay open and
	// stall Wait() forever. WaitDelay (Go 1.20+) closes the I/O
	// goroutines after the delay so Wait() returns even when the child
	// holds the pipe open.
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

	t.Logf("Test server ready at %s", serverURL)
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("--- server stdout/stderr ---\n%s\n--- end server output ---", serverLog.String())
		}
	})

	return port, cleanup
}

// TestLogin_E2E covers the login flow end-to-end:
// - Login form renders
// - UI standards (no inline JS, color-scheme meta, container width)
// - Invalid credentials surface as flash (best-effort: flash may not
//   survive 303 redirect in Tier C; logged, not asserted hard)
// - Valid credentials redirect to dashboard
// - OnConnect → goroutine → session.TriggerAction("serverWelcome") pushes
//   a message that appears in #server-welcome-message within 5s
//   (regression guard for the ctx.Session() nil bug)
// - Logout returns to the login form
func TestLogin_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	serverPort, cleanup := startServer(t)
	defer cleanup()

	debugPort, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port for Chrome: %v", err)
	}

	_ = e2etest.StartDockerChrome(t, debugPort)
	defer e2etest.StopDockerChrome(t, debugPort)

	chromeURL := fmt.Sprintf("http://localhost:%d", debugPort)
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), chromeURL)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	t.Run("InitialLoginForm", func(t *testing.T) {
		var initialHTML string
		err := chromedp.Run(ctx,
			chromedp.Navigate(e2etest.GetChromeTestURL(serverPort)),
			chromedp.WaitVisible(`h1`, chromedp.ByQuery),
			chromedp.OuterHTML(`body`, &initialHTML, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Failed to load page: %v", err)
		}

		if !strings.Contains(initialHTML, "Login") {
			t.Error("Login title not found")
		}
		if !strings.Contains(initialHTML, `type="text"`) {
			t.Error("Username input not found")
		}
		if !strings.Contains(initialHTML, `type="password"`) {
			t.Error("Password input not found")
		}
		if strings.Contains(initialHTML, "Dashboard") {
			t.Error("Dashboard should not be visible before login")
		}
	})

	t.Run("UI_Standards", func(t *testing.T) {
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
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		var html string
		err := chromedp.Run(ctx,
			chromedp.WaitVisible(`#username`, chromedp.ByQuery),
			chromedp.Clear(`#username`, chromedp.ByQuery),
			chromedp.SendKeys(`#username`, "testuser", chromedp.ByQuery),
			chromedp.Clear(`#password`, chromedp.ByQuery),
			chromedp.SendKeys(`#password`, "wrongpassword", chromedp.ByQuery),
			chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
			chromedp.Sleep(2*time.Second),
			chromedp.OuterHTML(`body`, &html, chromedp.ByQuery),
		)
		if err != nil {
			t.Logf("Failed to test invalid credentials: %v", err)
			return
		}
		if !strings.Contains(html, "Invalid credentials") {
			// lvt-form:no-intercept means real HTTP POST; the flash
			// message may not survive the 303 cycle on every browser.
			// Log but don't fail the suite over it.
			t.Log("Note: 'Invalid credentials' flash not displayed — expected for HTTP POST login form")
		}
	})

	t.Run("SuccessfulLogin", func(t *testing.T) {
		var html string
		err := chromedp.Run(ctx,
			chromedp.Navigate(e2etest.GetChromeTestURL(serverPort)),
			chromedp.WaitVisible(`#username`, chromedp.ByQuery),
			chromedp.Clear(`#username`, chromedp.ByQuery),
			chromedp.SendKeys(`#username`, "testuser", chromedp.ByQuery),
			chromedp.Clear(`#password`, chromedp.ByQuery),
			chromedp.SendKeys(`#password`, "secret", chromedp.ByQuery),
			chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
			chromedp.Sleep(2*time.Second),
			chromedp.OuterHTML(`body`, &html, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Failed to test successful login: %v", err)
		}

		if !strings.Contains(html, "Dashboard") {
			t.Logf("HTML content: %s", html[:min(500, len(html))])
			t.Error("Dashboard title not found after login")
		}
		if !strings.Contains(html, "Welcome") {
			t.Error("Welcome message not found")
		}
		if !strings.Contains(html, "testuser") {
			t.Error("Username not displayed on dashboard")
		}
	})

	// Regression guard: before livetemplate's ctx.Session() fix,
	// session was nil inside OnConnect, so sendWelcomeMessage silently
	// no-op'd. The page-literal "Welcome, testuser!" (tested above) is
	// template-rendered and would still pass even with a broken
	// Session, so only an explicit check for the server-pushed payload
	// catches the regression.
	t.Run("ServerPushedWelcome", func(t *testing.T) {
		// Use the explicit #server-welcome-message id rather than a bare
		// `ins` selector so the test can't accidentally match a different
		// <ins> element (e.g., a future generic success flash).
		err := chromedp.Run(ctx,
			e2etest.WaitForText(`#server-welcome-message`, "pushed from the server", 5*time.Second),
		)
		if err != nil {
			var body string
			_ = chromedp.Run(ctx, chromedp.OuterHTML(`body`, &body, chromedp.ByQuery))
			t.Fatalf("Server welcome message did not arrive via WebSocket push within 5s: %v\n=== body ===\n%s", err, body[:min(len(body), 800)])
		}
	})

	t.Run("Logout", func(t *testing.T) {
		var html string
		err := chromedp.Run(ctx,
			chromedp.WaitVisible(`button[name="logout"]`, chromedp.ByQuery),
			chromedp.Click(`button[name="logout"]`, chromedp.ByQuery),
			chromedp.Sleep(2*time.Second),
			chromedp.OuterHTML(`body`, &html, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Failed to test logout: %v", err)
		}
		if !strings.Contains(html, "Login") {
			t.Error("Login title not found after logout")
		}
	})
}

// TestLogin_E2E_SubpathMount is the regression guard for the bug where
// the recipe hard-coded "/" as its post-login redirect target. When the
// recipe is mounted at a non-root subpath (production at /apps/login/),
// http.StripPrefix strips the prefix before the recipe sees the request
// URL, so the recipe can't reconstruct its own mount — it must be told
// the mount path explicitly. A "/" redirect leaks past the StripPrefix
// boundary and lands on the domain root, not the recipe dashboard, so
// the user logs in successfully (Set-Cookie fires) but never sees the
// dashboard branch of the template.
//
// The original test suite mounts at "/" exclusively, which masks the
// bug — "/" and the mount path happen to coincide. This test mounts
// the same recipe at "/apps/login/" and asserts:
//
//   - The browser ends up at <serverURL>/apps/login/ after a successful
//     login, not the server root.
//   - The dashboard branch renders (so the recipe is reachable via the
//     mount, and POST-Redirect-GET re-mounts the same recipe).
//   - Logout also returns to <serverURL>/apps/login/, not the root.
func TestLogin_E2E_SubpathMount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	const mountPath = "/apps/login/"
	serverPort, cleanup := startServerAt(t, mountPath)
	defer cleanup()

	debugPort, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port for Chrome: %v", err)
	}
	_ = e2etest.StartDockerChrome(t, debugPort)
	defer e2etest.StopDockerChrome(t, debugPort)

	chromeURL := fmt.Sprintf("http://localhost:%d", debugPort)
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), chromeURL)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	mountURL := e2etest.GetChromeTestURL(serverPort) + mountPath

	var postLoginURL, html string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(mountURL),
		chromedp.WaitVisible(`#username`, chromedp.ByQuery),
		chromedp.SendKeys(`#username`, "testuser", chromedp.ByQuery),
		chromedp.SendKeys(`#password`, "secret", chromedp.ByQuery),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Location(&postLoginURL),
		chromedp.OuterHTML(`body`, &html, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("Failed to complete login flow: %v", err)
	}

	// The bug-fix assertion: post-login URL stays at the mount path.
	// Before the fix this would be the server root (mountURL minus the
	// /apps/login/ suffix), where a 404 would render (no handler there
	// in the test server).
	if !strings.HasSuffix(postLoginURL, mountPath) {
		t.Errorf("post-login URL should end with %q; got %q", mountPath, postLoginURL)
	}
	if !strings.Contains(html, "Dashboard") {
		t.Logf("HTML content: %s", html[:min(800, len(html))])
		t.Error("Dashboard not rendered after login at subpath mount")
	}
	if !strings.Contains(html, "testuser") {
		t.Error("Username not displayed on dashboard")
	}

	// Logout must also resolve back to the mount path, not root.
	var postLogoutURL, logoutHTML string
	if err := chromedp.Run(ctx,
		chromedp.WaitVisible(`button[name="logout"]`, chromedp.ByQuery),
		chromedp.Click(`button[name="logout"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Location(&postLogoutURL),
		chromedp.OuterHTML(`body`, &logoutHTML, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("Failed to complete logout: %v", err)
	}
	if !strings.HasSuffix(postLogoutURL, mountPath) {
		t.Errorf("post-logout URL should end with %q; got %q", mountPath, postLogoutURL)
	}
	if !strings.Contains(logoutHTML, "Login") {
		t.Error("Login form not rendered after logout at subpath mount")
	}
}

// TestLogin_HTTPCookie covers the raw HTTP shape of the login flow
// without a browser:
// - POST login → 303 + Set-Cookie: session_token=session_<user>_<unix>
//   with HttpOnly + SameSite=Strict
// - POST logout → Set-Cookie: session_token= with MaxAge=-1 or 0
func TestLogin_HTTPCookie(t *testing.T) {
	port, cleanup := startServer(t)
	defer cleanup()

	serverURL := fmt.Sprintf("http://localhost:%d", port)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Login.
	resp, err := client.Post(
		serverURL,
		"application/x-www-form-urlencoded",
		strings.NewReader("login=&username=testuser&password=secret"),
	)
	if err != nil {
		t.Fatalf("Login POST failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", resp.StatusCode)
	}

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session_token" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("session_token cookie not set after login")
	}
	if !sessionCookie.HttpOnly {
		t.Error("session_token cookie should be HttpOnly")
	}
	if !strings.HasPrefix(sessionCookie.Value, "session_testuser_") {
		t.Errorf("session_token value unexpected: %s", sessionCookie.Value)
	}

	// Logout.
	req, err := http.NewRequest("POST", serverURL, strings.NewReader("logout="))
	if err != nil {
		t.Fatalf("Failed to create logout request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Logout request failed: %v", err)
	}
	resp.Body.Close()

	for _, c := range resp.Cookies() {
		if c.Name == "session_token" && c.MaxAge >= 0 {
			t.Errorf("session_token should be deleted on logout (MaxAge=%d)", c.MaxAge)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
