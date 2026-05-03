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

// TestRecipe1_CatalogHydratesFromREST verifies Phase 5 recipe 1: the
// /patterns/ catalog is a tinkerdown <div lvt-source="patterns"> block
// that renders a "Connecting..." placeholder server-side, then hydrates
// over WebSocket from https://lt-patterns.fly.dev/api/index.json.
//
// We must NOT inspect the curl body — it would only show the loading
// shell. Instead, wait for a [data-test="pattern-row"] element (only
// produced by the post-hydration template) to appear, then count rows.
//
// This test also pulls double duty as a Phase 4 regression check: if
// the upstream pattern count drops below 30 (someone deleted patterns
// without updating data.go) OR the WS bind silently breaks (sources
// config typo, network, CORS), we catch it here.
func TestRecipe1_CatalogHydratesFromREST(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var rowCount int
	var summary string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/patterns/"),
		// Hydration is gated on a successful WS source fetch; if the
		// REST source ever fails, this WaitVisible deadlocks until the
		// per-test timeout — exactly the observable failure we want.
		chromedp.WaitVisible(`[data-test="pattern-row"]`, chromedp.ByQuery),
		chromedp.Text(`[data-test="catalog-summary"]`, &summary, chromedp.ByQuery),
		chromedp.Evaluate(
			`document.querySelectorAll('[data-test="pattern-row"]').length`,
			&rowCount,
		),
	); err != nil {
		t.Fatalf("hydrate: %v\nsummary so far: %q", err, summary)
	}

	if rowCount < 30 {
		t.Errorf("catalog hydrated with %d pattern rows; want >= 30 (drift between upstream API and docs catalog?)", rowCount)
	}
	if !strings.Contains(summary, "categories from the upstream API") {
		t.Errorf("catalog summary did not render expected text: %q", summary)
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
		"/recipes/architecture-flow",
		"/recipes/progressive-complexity-tree",
		"/recipes/sync-and-broadcast",
		"/recipes/patterns-stats",
		"/recipes/live-releases",
		"/recipes/how-this-site-works",
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

// TestRecipe2_ArchitectureFlowDiagramRenders verifies Phase 5 recipe 2:
// the architecture page contains a Mermaid sequence diagram that
// renders to inline SVG client-side, and the presentation-mode chrome
// that lets the page be walked as a slide deck.
//
// Mermaid hydrates after page load — wait for an <svg> to appear
// inside the rendered block. The presence of the .presentation-btn
// in the page chrome proves the second feature is wired.
func TestRecipe2_ArchitectureFlowDiagramRenders(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var svgCount, presentBtnCount int
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/architecture-flow"),
		chromedp.WaitVisible("h1", chromedp.ByQuery),
		// Mermaid initializes asynchronously after DOMContentLoaded.
		// Poll until at least one rendered <svg> appears or the
		// per-test timeout fires.
		chromedp.Poll(
			`document.querySelectorAll('.mermaid svg, [data-tinkerdown-block] svg').length > 0`,
			nil,
			chromedp.WithPollingTimeout(20*time.Second),
		),
		chromedp.Evaluate(
			`document.querySelectorAll('.mermaid svg, [data-tinkerdown-block] svg').length`,
			&svgCount,
		),
		chromedp.Evaluate(
			`document.querySelectorAll('.presentation-btn').length`,
			&presentBtnCount,
		),
	); err != nil {
		t.Fatalf("hydrate: %v", err)
	}

	if svgCount == 0 {
		t.Errorf("no Mermaid SVG rendered on architecture-flow recipe; mermaid bundle wired?")
	}
	if presentBtnCount == 0 {
		t.Errorf("no .presentation-btn in chrome; presentation-mode feature regressed?")
	}
}

// TestRecipe3_StatsViewSharesPatternsSource verifies Phase 5 recipe 3:
// the stats page binds to the SAME patterns source as the catalog and
// hydrates a different view (table with category rows) from it. This
// proves the "one source, many views" pattern is wired — the source
// is reachable from a second page without redeclaration.
//
// If this passes but TestRecipe1 fails, something is wrong with the
// catalog template specifically. If both fail, the source itself is
// likely broken (network, sources config, CORS).
func TestRecipe3_StatsViewSharesPatternsSource(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var rowCount int
	var summary string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/patterns-stats"),
		chromedp.WaitVisible(`[data-test="cat-row"]`, chromedp.ByQuery),
		chromedp.Text(`[data-test="stats-summary"]`, &summary, chromedp.ByQuery),
		chromedp.Evaluate(
			`document.querySelectorAll('[data-test="cat-row"]').length`,
			&rowCount,
		),
	); err != nil {
		t.Fatalf("hydrate: %v", err)
	}

	if rowCount < 5 {
		t.Errorf("stats page hydrated with %d category rows; want >= 5", rowCount)
	}
	if !strings.Contains(summary, "categories") {
		t.Errorf("stats summary did not render expected text: %q", summary)
	}
}

// TestRecipe4_ProgressiveComplexityTreeRenders verifies the
// progressive-complexity decision-tree recipe (mermaid flowchart).
// Same client-side mermaid hydration assertion as recipe 2 — a single
// rendered <svg> proves the bundle wired and the markdown parsed.
func TestRecipe4_ProgressiveComplexityTreeRenders(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var svgCount int
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/progressive-complexity-tree"),
		chromedp.WaitVisible("h1", chromedp.ByQuery),
		chromedp.Poll(
			`document.querySelectorAll('.mermaid svg, [data-tinkerdown-block] svg').length > 0`,
			nil,
			chromedp.WithPollingTimeout(20*time.Second),
		),
		chromedp.Evaluate(
			`document.querySelectorAll('.mermaid svg, [data-tinkerdown-block] svg').length`,
			&svgCount,
		),
	); err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	if svgCount == 0 {
		t.Errorf("no Mermaid SVG on progressive-complexity tree recipe")
	}
}

// TestRecipe5_SyncBroadcastDiagramRenders verifies the sync+broadcast
// recipe renders both of its mermaid sequence diagrams. The page has
// two ```mermaid blocks, so we expect at least 2 SVGs once hydrated.
func TestRecipe5_SyncBroadcastDiagramRenders(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var svgCount int
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/sync-and-broadcast"),
		chromedp.WaitVisible("h1", chromedp.ByQuery),
		chromedp.Poll(
			`document.querySelectorAll('.mermaid svg, [data-tinkerdown-block] svg').length >= 2`,
			nil,
			chromedp.WithPollingTimeout(20*time.Second),
		),
		chromedp.Evaluate(
			`document.querySelectorAll('.mermaid svg, [data-tinkerdown-block] svg').length`,
			&svgCount,
		),
	); err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	if svgCount < 2 {
		t.Errorf("sync-and-broadcast recipe rendered %d SVGs; expected 2 (sync + broadcast)", svgCount)
	}
}

// TestRecipe6_LiveReleasesHydratesFromGitHub verifies the live-releases
// recipe binds to the GitHub Releases API and renders rows. GitHub may
// rate-limit (60/h unauth) — if THIS test starts flaking specifically
// (and other recipes hold), the cause is likely shared-IP rate limit
// from CI runners or fly's edge IPs, not a regression.
func TestRecipe6_LiveReleasesHydratesFromGitHub(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var rowCount int
	var summary string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/live-releases"),
		chromedp.WaitVisible(`[data-test="release-row"], [data-test="releases-summary"]`, chromedp.ByQuery),
		chromedp.Text(`[data-test="releases-summary"]`, &summary, chromedp.ByQuery, chromedp.AtLeast(0)),
		chromedp.Evaluate(
			`document.querySelectorAll('[data-test="release-row"]').length`,
			&rowCount,
		),
	); err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	if rowCount == 0 {
		t.Errorf("live-releases recipe rendered 0 rows; GitHub API rate-limited or upstream changed?\n  summary: %q", summary)
	}
}

// TestRecipe7_MetaPageHasMermaidAndLiveCount verifies the meta recipe
// renders both its mermaid diagram AND the live patterns-source count
// (proving its embedded source bind works alongside the diagram).
func TestRecipe7_MetaPageHasMermaidAndLiveCount(t *testing.T) {
	ctx, cancel := newCtx(t)
	defer cancel()

	var svgCount int
	var summary string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL()+"/recipes/how-this-site-works"),
		chromedp.WaitVisible(`[data-test="meta-summary"]`, chromedp.ByQuery),
		chromedp.Poll(
			`document.querySelectorAll('.mermaid svg, [data-tinkerdown-block] svg').length > 0`,
			nil,
			chromedp.WithPollingTimeout(20*time.Second),
		),
		chromedp.Text(`[data-test="meta-summary"]`, &summary, chromedp.ByQuery),
		chromedp.Evaluate(
			`document.querySelectorAll('.mermaid svg, [data-tinkerdown-block] svg').length`,
			&svgCount,
		),
	); err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	if svgCount == 0 {
		t.Errorf("meta recipe missing mermaid SVG")
	}
	if !strings.Contains(summary, "pattern categories") {
		t.Errorf("meta summary did not include live patterns count: %q", summary)
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
