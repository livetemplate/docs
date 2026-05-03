// Package e2e runs browser-driven smoke tests against the live staging
// docs site. They verify that features which only manifest in a real
// browser (theme toggle persistence, deep-link anchors, copy buttons,
// proxied pattern interactivity) actually work — pages that look
// correct in curl can still be broken in a browser.
//
// Run with:
//
//	cd e2e && go test -v -timeout 120s ./...
//
// Override target with E2E_BASE_URL (default: staging).
package e2e

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func baseURL() string {
	if v := os.Getenv("E2E_BASE_URL"); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return "https://livetemplate-docs-staging.fly.dev"
}

// warmup is performed once before any chromedp test runs. The staging
// fly app uses auto_stop_machines, so the first request after idle can
// take 10-25s for the machine to wake up. Issuing a plain HTTP GET is
// much cheaper than letting Chrome eat that latency on its first
// navigation.
var warmupOnce sync.Once

func warmupStaging(t *testing.T) {
	warmupOnce.Do(func() {
		client := &http.Client{Timeout: 60 * time.Second}
		for i := 0; i < 3; i++ {
			resp, err := client.Get(baseURL() + "/")
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(2 * time.Second)
		}
		t.Logf("warmup: staging didn't respond OK after 3 attempts; tests may flake")
	})
}

// newCtx returns a chromedp context with a sane per-test timeout so a
// hung browser doesn't wedge the entire test run.
func newCtx(t *testing.T) (context.Context, context.CancelFunc) {
	warmupStaging(t)
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
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

// TestHomeRenders is the canary — if this fails the rest can't trust
// anything else. Uses ByQuery (not NodeVisible) because tinkerdown's
// presentation-mode CSS may visually hide the toolbar's h1 without
// removing it from the DOM, which would deadlock NodeVisible.
func TestHomeRenders(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var title, h1 string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Title(&title),
		chromedp.Text("h1", &h1, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	if !strings.Contains(title, "LiveTemplate") {
		t.Errorf("title %q does not contain LiveTemplate", title)
	}
	if !strings.Contains(strings.ToLower(h1), "livetemplate") {
		t.Errorf("h1 %q does not mention LiveTemplate", h1)
	}
}

// TestEditOnGitHubLinkPresent verifies PR-C: every page footer has the
// edit-on-github link with a correct URL pattern.
func TestEditOnGitHubLinkPresent(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var href string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/getting-started/install"),
		chromedp.AttributeValue(".page-edit-link a", "href", &href, nil, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("locate edit link: %v", err)
	}
	want := "https://github.com/livetemplate/docs/edit/main/getting-started/install.md"
	if href != want {
		t.Errorf("edit link href = %q, want %q", href, want)
	}
}

// TestThemeTogglePersists verifies PR-A + the existing dark mode UI:
// clicking the dark theme button persists the choice across reloads via
// localStorage, and the html data-theme attribute updates accordingly.
func TestThemeTogglePersists(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var initial, afterClick, afterReload string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.AttributeValue("html", "data-theme", &initial, nil, chromedp.ByQuery),
		chromedp.Click("#theme-dark", chromedp.ByID),
		chromedp.AttributeValue("html", "data-theme", &afterClick, nil, chromedp.ByQuery),
		chromedp.Reload(),
		chromedp.AttributeValue("html", "data-theme", &afterReload, nil, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("toggle theme: %v", err)
	}

	if afterClick != "dark" {
		t.Errorf("after dark-toggle click: data-theme = %q, want dark", afterClick)
	}
	if afterReload != "dark" {
		t.Errorf("after reload: data-theme = %q, want dark (localStorage persistence)", afterReload)
	}
	t.Logf("initial=%q afterClick=%q afterReload=%q", initial, afterClick, afterReload)
}

// TestThemeAccentInjected verifies PR-A end-to-end: the user's
// primary_color from tinkerdown.yaml ("#5a67d8") is injected into the
// CSS custom property the rest of the styling reads.
func TestThemeAccentInjected(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var accent string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/"),
		chromedp.Evaluate(
			`getComputedStyle(document.documentElement).getPropertyValue('--accent').trim()`,
			&accent,
		),
	); err != nil {
		t.Fatalf("read --accent: %v", err)
	}
	want := "#5a67d8"
	if !strings.EqualFold(accent, want) {
		t.Errorf("--accent = %q, want %q (set in tinkerdown.yaml)", accent, want)
	}
}

// TestPatternsCatalogIsNative verifies the docs site's catalog page is
// served from local markdown (NOT proxied), even though sub-paths like
// /patterns/forms/* are.
func TestPatternsCatalogIsNative(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var bodyText string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/patterns/"),
		chromedp.Text("body", &bodyText, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	// The native markdown page contains this phrase; the upstream patterns
	// app does not.
	if !strings.Contains(bodyText, "Phase 4 builds the full") {
		t.Errorf("catalog page does not look native: body excerpt:\n%s", bodyText[:min(500, len(bodyText))])
	}
}

// TestPatternProxiedAndInteractive verifies PR-D end-to-end. Visiting
// /patterns/forms/click-to-edit on the docs site reverse-proxies to
// the lt-patterns app, the upstream's interactive UI loads, and we can
// see its expected DOM (an Edit button and a name field).
func TestPatternProxiedAndInteractive(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var html string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/patterns/forms/click-to-edit"),
		chromedp.OuterHTML("body", &html, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	// Markers from the lt-patterns app's click-to-edit page.
	for _, want := range []string{"<form", "Edit"} {
		if !strings.Contains(html, want) {
			t.Errorf("proxied pattern body missing %q\nbody excerpt:\n%s",
				want, html[:min(800, len(html))])
		}
	}
}
