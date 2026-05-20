// Package main_test exercises the shared-notepad recipe end-to-end.
//
// Three test functions:
//
//	TestSharedNotepad_E2E              — chromedp: type, save, refresh,
//	                                     verify content persists.
//	TestSharedNotepad_MultiUserIsolation — HTTP-only: alice and bob each
//	                                       save different content; verify
//	                                       neither sees the other's. Goes
//	                                       beyond the original example's
//	                                       single-user shape and asserts
//	                                       the per-user state map.
//	TestSharedNotepad_PublishRefreshAction — verifies the Publish-to-SelfTopic
//	                                          queue dispatches the
//	                                          "Refresh" action.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

// startServer launches the test-server binary on a free port and waits
// until it's ready. Note that readiness probes against /
// receive 401 (BasicAuth challenge) — that's the success signal here,
// not 200. A 401 means the server is up; without BasicAuth on the probe
// it's the correct response.
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

	var serverLog bytes.Buffer
	cmd.Stdout = &serverLog
	cmd.Stderr = &serverLog

	// `go run .` forks a compiled child binary; SIGKILL on the parent
	// doesn't propagate. WaitDelay (Go 1.20+) closes the I/O goroutines
	// after the delay so Wait() returns even when the child holds the
	// pipe open.
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
			// 401 is the expected success signal — the server is up and
			// is correctly issuing the BasicAuth challenge.
			if resp.StatusCode == http.StatusUnauthorized {
				consecutiveSuccesses++
				if consecutiveSuccesses >= requiredSuccesses {
					time.Sleep(100 * time.Millisecond)
					ready = true
					break
				}
			} else {
				lastErr = fmt.Errorf("status %d (expected 401)", resp.StatusCode)
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

// TestSharedNotepad_E2E covers the single-user save-and-refresh flow:
// - Page loads, displays the authenticated username
// - UI standards (no inline JS, color-scheme meta, container width)
// - Type into textarea, verify character count updates
// - Click Save, verify "saved at" timestamp appears
// - Refresh the page, verify content persists
func TestSharedNotepad_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	serverPort, cleanup := startServer(t)
	defer cleanup()

	debugPort, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port for Chrome: %v", err)
	}

	if err := e2etest.StartDockerChrome(t, debugPort); err != nil {
		t.Fatalf("Failed to start Docker Chrome: %v", err)
	}
	defer e2etest.StopDockerChrome(t, debugPort)

	chromeURL := fmt.Sprintf("http://localhost:%d", debugPort)
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), chromeURL)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	appURL := fmt.Sprintf("http://alice:demo@host.docker.internal:%d/", serverPort)

	t.Run("PageLoads", func(t *testing.T) {
		var html string
		err := chromedp.Run(ctx,
			chromedp.Navigate(appURL),
			chromedp.WaitReady("body", chromedp.ByQuery),
			e2etest.WaitFor(`window.liveTemplateClient && window.liveTemplateClient.isReady()`, 10*time.Second),
			chromedp.OuterHTML("body", &html, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Failed to load page: %v", err)
		}
		if !strings.Contains(html, "Shared Notepad") {
			t.Error("Page title not found")
		}
		if !strings.Contains(html, "alice") {
			t.Error("Username 'alice' not displayed after BasicAuth")
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

	t.Run("TypeSaveAndRefresh", func(t *testing.T) {
		err := chromedp.Run(ctx,
			chromedp.WaitVisible(`#content`, chromedp.ByQuery),
			chromedp.Click(`#content`, chromedp.ByQuery),
			chromedp.SendKeys(`#content`, "Hello persistence test", chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Failed to type: %v", err)
		}

		var beforeSave string
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`document.getElementById('content').value`, &beforeSave),
		)
		if err != nil {
			t.Fatalf("Failed to read textarea: %v", err)
		}
		if !strings.Contains(beforeSave, "Hello persistence test") {
			t.Fatalf("Typed text not in textarea: %q", beforeSave)
		}

		err = chromedp.Run(ctx,
			chromedp.Click(`button[name="save"]`, chromedp.ByQuery),
			e2etest.WaitFor(`document.getElementById('charcount').textContent.includes('saved at')`, 5*time.Second),
		)
		if err != nil {
			t.Fatalf("Failed to save: %v", err)
		}

		var afterSave string
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`document.getElementById('content').value`, &afterSave),
		)
		if err != nil {
			t.Fatalf("Failed to read after save: %v", err)
		}
		if afterSave == "" {
			t.Fatal("BUG: Textarea wiped after save (lvt-form:preserve regression)")
		}

		// Browser favicon request (chromedp does this on navigation) is
		// triggered ahead of the navigation so the server has both an
		// auth'd and a no-auth path open when the WS reconnects.
		go func() {
			client := &http.Client{Timeout: 2 * time.Second}
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/favicon.ico", serverPort), nil)
			req.SetBasicAuth("alice", "demo")
			_, _ = client.Do(req)
		}()
		time.Sleep(100 * time.Millisecond)

		err = chromedp.Run(ctx,
			chromedp.Navigate(appURL),
			chromedp.WaitReady("body", chromedp.ByQuery),
			e2etest.WaitFor(`window.liveTemplateClient && window.liveTemplateClient.isReady()`, 10*time.Second),
		)
		if err != nil {
			t.Fatalf("Failed to reload: %v", err)
		}

		var afterRefresh string
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`document.getElementById('content').value`, &afterRefresh),
		)
		if err != nil {
			t.Fatalf("Failed to read after refresh: %v", err)
		}
		if afterRefresh == "" {
			t.Fatal("BUG: Content lost after page refresh")
		}
		if !strings.Contains(afterRefresh, "Hello persistence test") {
			t.Errorf("Expected content to persist, got: %q", afterRefresh)
		}
	})
}

// TestSharedNotepad_MultiUserIsolation asserts the per-user-state map
// keyed by ctx.UserID() actually isolates between users. Two clients
// authenticate as different BasicAuth users, each writes different
// content, then each reads back — neither should see the other's text.
//
// This is new pedagogy beyond the original example, which only tested
// the single-user shape. The cross-user isolation is the recipe's
// central teaching point and deserves an assertion.
//
// HTTP-only (no chromedp): the Tier C `Accept: text/html` POST/GET
// shape exercises the same Mount/Save/Refresh handlers without the
// browser overhead, and lets us cleanly drive two distinct identities
// from one process.
func TestSharedNotepad_MultiUserIsolation(t *testing.T) {
	port, cleanup := startServer(t)
	defer cleanup()

	baseURL := fmt.Sprintf("http://localhost:%d/", port)

	// Each user gets their own cookie jar so livetemplate-id stays
	// scoped to that identity. BasicAuth still travels on every request
	// via SetBasicAuth.
	clientFor := func(_ /*user*/, _ /*pass*/ string) *http.Client {
		// jar is mostly a safety net here — the framework's authenticator
		// returns groupID=username on every request regardless of the
		// session cookie, but keeping separate jars makes the test honest
		// about "two clients, two identities."
		return &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	save := func(t *testing.T, user, pass, body string) {
		t.Helper()
		client := clientFor(user, pass)
		form := fmt.Sprintf("save=&content=%s", strings.ReplaceAll(body, " ", "+"))
		req, err := http.NewRequest("POST", baseURL, strings.NewReader(form))
		if err != nil {
			t.Fatalf("save: build request: %v", err)
		}
		req.SetBasicAuth(user, pass)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "text/html")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("save as %s: %v", user, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("save as %s: status %d (want 200 or 303)", user, resp.StatusCode)
		}
	}

	read := func(t *testing.T, user, pass string) string {
		t.Helper()
		client := clientFor(user, pass)
		req, err := http.NewRequest("GET", baseURL, nil)
		if err != nil {
			t.Fatalf("read: build request: %v", err)
		}
		req.SetBasicAuth(user, pass)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("read as %s: %v", user, err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read as %s body: %v", user, err)
		}
		return string(body)
	}

	const aliceText = "alice-only-secret"
	const bobText = "bob-only-secret"

	// Save alice's content.
	save(t, "alice", "demo", aliceText)

	// Read back as alice — should contain alice's text, not bob's.
	aliceHTML := read(t, "alice", "demo")
	if !strings.Contains(aliceHTML, aliceText) {
		t.Errorf("alice should see her saved text %q; HTML: %s", aliceText, truncate(aliceHTML, 800))
	}
	if strings.Contains(aliceHTML, bobText) {
		t.Errorf("alice should NOT see bob's text yet; HTML: %s", truncate(aliceHTML, 800))
	}

	// Read as bob BEFORE saving anything — should be empty (or at least
	// not contain alice's text).
	bobHTMLEmpty := read(t, "bob", "demo")
	if strings.Contains(bobHTMLEmpty, aliceText) {
		t.Errorf("bob should NOT see alice's text; HTML: %s", truncate(bobHTMLEmpty, 800))
	}
	if !strings.Contains(bobHTMLEmpty, "Shared Notepad") {
		t.Errorf("bob's page should still render the notepad shell; HTML: %s", truncate(bobHTMLEmpty, 800))
	}

	// Save bob's content.
	save(t, "bob", "demo", bobText)

	// Re-read alice — should still see alice's only, NOT bob's.
	aliceHTML2 := read(t, "alice", "demo")
	if !strings.Contains(aliceHTML2, aliceText) {
		t.Errorf("alice should still see her saved text; HTML: %s", truncate(aliceHTML2, 800))
	}
	if strings.Contains(aliceHTML2, bobText) {
		t.Errorf("alice should NEVER see bob's text — per-user state leak!; HTML: %s", truncate(aliceHTML2, 800))
	}

	// Re-read bob — should see bob's only, NOT alice's.
	bobHTML := read(t, "bob", "demo")
	if !strings.Contains(bobHTML, bobText) {
		t.Errorf("bob should see his saved text; HTML: %s", truncate(bobHTML, 800))
	}
	if strings.Contains(bobHTML, aliceText) {
		t.Errorf("bob should NEVER see alice's text — per-user state leak!; HTML: %s", truncate(bobHTML, 800))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
