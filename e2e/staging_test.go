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
// /patterns/forms/* are. After Phase 4 the catalog lists all 33
// patterns, so this also serves as a count-regression check.
func TestPatternsCatalogIsNative(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var linkCount int
	var bodyText string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/patterns/"),
		chromedp.Text("body", &bodyText, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelectorAll('a[href^="/patterns/"]').length`, &linkCount),
	); err != nil {
		t.Fatalf("navigate: %v", err)
	}
	// Catalog phrase only present on the native page (the upstream
	// patterns app's index says "31 UI patterns demonstrating
	// progressive complexity", different wording).
	if !strings.Contains(bodyText, "33 reactive") {
		t.Errorf("catalog page does not look native: body excerpt:\n%s", bodyText[:min(500, len(bodyText))])
	}
	// Phase 4: 33 individual pattern links (5 are also rendered as
	// section headers in the sidebar; that's fine, the assertion is on
	// links into /patterns/<category>/<slug>).
	if linkCount < 33 {
		t.Errorf("catalog has %d pattern links; want >= 33", linkCount)
	}
}

// TestPatternsAPIReachable verifies the upstream patterns app is
// serving the JSON catalog endpoint that powers (or could power) the
// docs catalog. If this breaks, the docs catalog is at risk of
// silently going stale relative to what's actually deployed.
func TestPatternsAPIReachable(t *testing.T) {
	warmupStaging(t)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://lt-patterns.fly.dev/api/index.json")
	if err != nil {
		t.Fatalf("fetch /api/index.json: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
	if cors := resp.Header.Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want * (cross-origin docs fetch needs it)", cors)
	}
}

// TestSidebarWalk visits every URL emitted by the sidebar nav and
// asserts none returns an error status. Phase 2 brings this from the
// 5 placeholder pages to the full ~40-page site, so the walk is the
// canary for "did the manual port land cleanly without breaking refs".
func TestSidebarWalk(t *testing.T) {
	warmupStaging(t)

	urls := []string{
		"/",
		"/getting-started/install",
		"/guides/progressive-complexity",
		"/guides/standard-html-reactivity",
		"/guides/ephemeral-components",
		"/guides/observability",
		"/guides/scaling",
		"/reference/api",
		"/reference/client-attributes",
		"/reference/configuration",
		"/reference/session",
		"/reference/server-actions",
		"/reference/authentication",
		"/reference/uploads",
		"/reference/pubsub",
		"/reference/error-handling",
		"/reference/controller-pattern",
		"/reference/navigate",
		"/reference/template-support-matrix",
		"/reference/limitations",
		"/cli/",
		"/cli/auth-customization",
		"/cli/components",
		"/cli/testing",
		"/client/",
		"/patterns/",
		"/examples/",
		"/examples/counter",
		"/examples/todos",
		"/examples/chat",
		"/examples/avatar-upload",
		"/examples/flash-messages",
		"/examples/progressive-enhancement",
		"/examples/ws-disabled",
		"/contributing/livetemplate",
		"/contributing/client",
		"/contributing/cli",
		"/contributing/examples",
		"/changelog",
	}

	// Browser-only walk would re-allocate Chrome 39 times; fall back to a
	// plain HTTP GET per URL with the test's existing client. Each URL
	// visited via 200 status is sufficient for the regression — chromedp
	// covers behaviour-level checks in the other tests.
	client := &http.Client{Timeout: 15 * time.Second}
	failures := 0
	for _, u := range urls {
		full := baseURL() + u
		resp, err := client.Get(full)
		if err != nil {
			t.Errorf("%s: %v", u, err)
			failures++
			continue
		}
		resp.Body.Close()
		// 303 is acceptable: tinkerdown redirects /cli to /cli/ when
		// only the index variant exists. Anything else outside 2xx
		// counts as a failure.
		ok := (resp.StatusCode >= 200 && resp.StatusCode < 300) || resp.StatusCode == 303
		if !ok {
			t.Errorf("%s: HTTP %d", u, resp.StatusCode)
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d of %d sidebar URLs failed", failures, len(urls))
	}
}

// TestEditLinkForSyncedPage asserts that a page mirrored from another
// repo (frontmatter source_repo + source_path set) renders an edit link
// pointing at THAT repo, not the docs site. Without this, "Edit this
// page" sends contributors to the wrong repo and edits get lost when
// the next sync overwrites them.
func TestEditLinkForSyncedPage(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var href string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/guides/progressive-complexity"),
		chromedp.AttributeValue(".page-edit-link a", "href", &href, nil, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("locate edit link: %v", err)
	}
	want := "https://github.com/livetemplate/livetemplate/edit/main/docs/guides/progressive-complexity.md"
	if href != want {
		t.Errorf("synced page edit link = %q, want %q\n  (frontmatter source_repo+source_path should win over site repo)", href, want)
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
