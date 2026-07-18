package filetree

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/livetemplate/livetemplate"
	e2etest "github.com/livetemplate/lvt/testing"
)

// ctxWithValue builds the Context an action sees when a button carrying
// value="<path>" is clicked, which is how the template tells an action which
// of the many rendered nodes was hit.
func ctxWithValue(path string) *livetemplate.Context {
	return livetemplate.NewContext(context.Background(), "", map[string]interface{}{
		"value": path,
	})
}

// ---------------------------------------------------------------------------
// White-box logic tests — fast, deterministic, no browser. These pin the tree
// walk (does a click on a deeply nested node reach the right node?) without
// the cost of a real browser. The render test below then proves the recursive
// template itself resolves, which is the thing this recipe exists to show.
// ---------------------------------------------------------------------------

// deepFile is the deepest leaf in the fixture. Tests reference it by name so
// a change to sampleTree's shape fails loudly here rather than silently
// weakening the "deep" in these assertions.
const deepFile = "/internal/store/sql/query.go"

func TestUpdateNode_ReachesDeepestLeaf(t *testing.T) {
	root := sampleTree()

	if ok := updateNode(&root, deepFile, func(n *Node) { n.Starred = true }); !ok {
		t.Fatalf("updateNode did not find %s", deepFile)
	}

	var found *Node
	var walk func(*Node)
	walk = func(n *Node) {
		if n.Path == deepFile {
			found = n
			return
		}
		for i := range n.Children {
			walk(&n.Children[i])
		}
	}
	walk(&root)

	if found == nil {
		t.Fatalf("%s missing from tree after update", deepFile)
	}
	if !found.Starred {
		t.Errorf("expected %s to be starred", deepFile)
	}
}

func TestUpdateNode_UnknownPathIsANoop(t *testing.T) {
	root := sampleTree()
	if updateNode(&root, "/nope/missing.go", func(n *Node) { n.Starred = true }) {
		t.Error("updateNode reported a match for a path that is not in the tree")
	}
}

// TestUpdateNode_StopsAtFirstMatch guards the early return. Paths are unique
// in the fixture, so the observable contract is simply that the walk reports
// a match and does not keep descending afterwards.
func TestUpdateNode_StopsAtFirstMatch(t *testing.T) {
	root := sampleTree()
	calls := 0
	if !updateNode(&root, "/cmd", func(n *Node) { calls++ }) {
		t.Fatal("updateNode did not find /cmd")
	}
	if calls != 1 {
		t.Errorf("expected the mutator to run once, ran %d times", calls)
	}
}

func TestToggle_FlipsExpanded(t *testing.T) {
	c := &Controller{}
	s := State{Root: sampleTree()}

	// /internal starts collapsed in the fixture.
	before := findNode(t, &s.Root, "/internal")
	if before.Expanded {
		t.Fatal("fixture changed: /internal is expected to start collapsed")
	}

	s, err := c.Toggle(s, ctxWithValue("/internal"))
	if err != nil {
		t.Fatalf("Toggle: %v", err)
	}
	if !findNode(t, &s.Root, "/internal").Expanded {
		t.Error("Toggle did not expand /internal")
	}

	s, err = c.Toggle(s, ctxWithValue("/internal"))
	if err != nil {
		t.Fatalf("Toggle (second): %v", err)
	}
	if findNode(t, &s.Root, "/internal").Expanded {
		t.Error("Toggle did not collapse /internal on the second click")
	}
}

func TestStar_FlipsStarredOnALeaf(t *testing.T) {
	c := &Controller{}
	s := State{Root: sampleTree()}

	s, err := c.Star(s, ctxWithValue(deepFile))
	if err != nil {
		t.Fatalf("Star: %v", err)
	}
	if !findNode(t, &s.Root, deepFile).Starred {
		t.Errorf("Star did not mark %s", deepFile)
	}
}

func TestMount_SeedsTreeOnce(t *testing.T) {
	c := &Controller{}

	s, err := c.Mount(State{}, nil)
	if err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if s.Root.Path == "" {
		t.Fatal("Mount did not seed the tree")
	}

	// A second Mount (reconnect) must not discard state the visitor built up.
	s = mustStar(t, c, s, deepFile)
	s, err = c.Mount(s, nil)
	if err != nil {
		t.Fatalf("Mount (second): %v", err)
	}
	if !findNode(t, &s.Root, deepFile).Starred {
		t.Error("Mount reset a tree that was already seeded, losing session state")
	}
}

// TestHandler_RendersRecursiveDepth is the version gate. The template invokes
// itself, which livetemplate could not do before v0.19.0 — under an older core
// the parse fails outright. Asserting on the deepest leaf proves the recursion
// resolved all the way down rather than stopping at the first level.
func TestHandler_RendersRecursiveDepth(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	html := string(body)

	// /cmd/server/main.go sits three directories down and every one of them is
	// expanded in the fixture, so it must be present on first paint.
	for _, want := range []string{
		`data-key="/"`,
		`data-key="/cmd"`,
		`data-key="/cmd/server"`,
		`data-key="/cmd/server/main.go"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("first render missing %s — recursion did not reach that depth:\n%s", want, html)
		}
	}

	// /internal is collapsed, so its children must NOT be rendered. This keeps
	// the test honest: it would still pass if the template rendered every node
	// regardless of Expanded, which is not what the recipe demonstrates.
	if strings.Contains(html, `data-key="/internal/store"`) {
		t.Error("collapsed directory rendered its children")
	}
}

// findNode locates a node by path or fails the test.
func findNode(t *testing.T, root *Node, path string) *Node {
	t.Helper()
	var found *Node
	var walk func(*Node)
	walk = func(n *Node) {
		if n.Path == path {
			found = n
			return
		}
		for i := range n.Children {
			walk(&n.Children[i])
		}
	}
	walk(root)
	if found == nil {
		t.Fatalf("node %s not found", path)
	}
	return found
}

func mustStar(t *testing.T, c *Controller, s State, path string) State {
	t.Helper()
	s, err := c.Star(s, ctxWithValue(path))
	if err != nil {
		t.Fatalf("Star %s: %v", path, err)
	}
	return s
}

// ---------------------------------------------------------------------------
// Browser E2E. The logic tests above prove the tree walk; this proves the
// recipe actually works in a browser against the real published client.
//
// It exists mainly to catch one specific silent failure. A full HTML document
// whose recursive {{define}} blocks fail to resolve does not error — the
// initial tree build falls back to HTML-string diffing, and the page still
// renders. It just stops updating reactively. So rendering correctly proves
// nothing on its own; the server log assertion is what makes this a real gate.
// ---------------------------------------------------------------------------

// fallbackWarning is the log line livetemplate emits when the reactive tree
// build fails and it degrades to HTML-string diffing.
const fallbackWarning = "falling back to HTML structure-based tree"

func TestFileTreeE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Concurrency-safe server-log capture. Do NOT add t.Parallel() — global
	// log capture is incompatible with it.
	serverLogs := e2etest.NewSafeBuffer()
	log.SetOutput(serverLogs)
	defer log.SetOutput(os.Stderr)

	port, err := e2etest.GetFreePort()
	if err != nil {
		t.Fatalf("free port: %v", err)
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: Handler(livetemplate.WithDevMode(true)),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()
	e2etest.WaitForServer(t, fmt.Sprintf("http://localhost:%d", port), 10*time.Second)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			t.Logf("server shutdown warning: %v", err)
		}
	}()

	chromeCtx, cleanup := e2etest.SetupDockerChrome(t, 30*time.Second)
	defer cleanup()
	ctx := chromeCtx.Context

	consoleLog := e2etest.NewConsoleLogger()
	consoleLog.Start(ctx)
	wsLog := e2etest.RecordWSFrames(ctx)

	dumpDiagnostics := func(label string) {
		t.Logf("=== %s ===", label)
		for _, l := range consoleLog.GetLogs() {
			t.Logf("console [%s]: %s", l.Type, l.Message)
		}
		for _, m := range wsLog.GetMessages() {
			t.Logf("ws %s: %s", m.Direction, m.Data)
		}
		t.Logf("server log:\n%s", serverLogs.String())
		var html string
		if err := chromedp.Run(ctx, chromedp.OuterHTML(`html`, &html, chromedp.ByQuery)); err != nil {
			t.Logf("outerHTML: %v", err)
		} else {
			t.Logf("rendered HTML:\n%s", html)
		}
	}

	url := e2etest.GetChromeTestURL(port)

	// Phase 1 — first load must render the recursion all the way down.
	var deepPresent bool
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`li[data-key="/"]`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('li[data-key="/cmd/server/main.go"]') !== null`, &deepPresent),
	); err != nil {
		dumpDiagnostics("initial load failed")
		t.Fatalf("chromedp.Run (initial load): %v", err)
	}
	if !deepPresent {
		dumpDiagnostics("recursion did not reach depth on first load")
		t.Fatal("expected /cmd/server/main.go to render on first load")
	}

	// Phase 2 — expanding a collapsed directory must arrive over the socket,
	// revealing children that were not in the initial HTML at all.
	wsBefore := len(wsLog.GetMessages())
	var childVisible, singleNavigation bool
	if err := chromedp.Run(ctx,
		chromedp.Click(`button[name="toggle"][value="/internal"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`li[data-key="/internal/store"]`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('li[data-key="/internal/store"]') !== null`, &childVisible),
		chromedp.Evaluate(`window.performance.getEntriesByType('navigation').length === 1`, &singleNavigation),
	); err != nil {
		dumpDiagnostics("expand failed")
		t.Fatalf("chromedp.Run (expand /internal): %v", err)
	}
	if !childVisible {
		dumpDiagnostics("expand did not reveal children")
		t.Error("expected /internal/store to appear after expanding /internal")
	}
	if !singleNavigation {
		dumpDiagnostics("page navigated instead of updating over the socket")
		t.Error("expand caused a full navigation; the update must arrive over the WebSocket")
	}
	if len(wsLog.GetMessages()) == wsBefore {
		dumpDiagnostics("no WebSocket traffic on expand")
		t.Error("expand produced no WebSocket frames")
	}

	// Phase 3 — starring a leaf four levels down must update that leaf without
	// disturbing its siblings. Marking a sibling from the browser first gives
	// us a DOM-identity witness: if the branch were rebuilt rather than patched
	// per-leaf, the marker would not survive.
	if err := chromedp.Run(ctx,
		chromedp.Click(`button[name="toggle"][value="/internal/store"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`li[data-key="/internal/store/sql"]`, chromedp.ByQuery),
		chromedp.Click(`button[name="toggle"][value="/internal/store/sql"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`li[data-key="`+deepFile+`"]`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('li[data-key="/internal/store/sql/migrate.go"]').__lvtMark = 'sibling'; true`, nil),
	); err != nil {
		dumpDiagnostics("drilling to the deep leaf failed")
		t.Fatalf("chromedp.Run (drill down): %v", err)
	}

	var starred, siblingSurvived bool
	if err := chromedp.Run(ctx,
		chromedp.Click(`button[name="star"][value="`+deepFile+`"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`li[data-key="`+deepFile+`"] button[name="star"]`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('li[data-key="`+deepFile+`"] button[name="star"]').textContent.includes('★')`, &starred),
		chromedp.Evaluate(`document.querySelector('li[data-key="/internal/store/sql/migrate.go"]').__lvtMark === 'sibling'`, &siblingSurvived),
	); err != nil {
		dumpDiagnostics("star failed")
		t.Fatalf("chromedp.Run (star deep leaf): %v", err)
	}
	if !starred {
		dumpDiagnostics("star did not take effect")
		t.Errorf("expected %s to render as starred", deepFile)
	}
	if !siblingSurvived {
		dumpDiagnostics("sibling DOM node was replaced")
		t.Error("starring a leaf replaced its sibling's DOM node; the update was not per-leaf")
	}

	// Prove the capture works before trusting its silence. If log.SetOutput
	// ever stops reaching livetemplate's logger, serverLogs goes empty and the
	// fallback check below passes for the wrong reason — a green test with no
	// gate behind it. A connected session always logs, so empty means broken.
	if serverLogs.String() == "" {
		t.Fatal("captured no server logs; the fallback assertion below would pass vacuously")
	}

	// The gate: a recursive full-HTML document that silently degrades still
	// renders and still updates via HTML-string diffing, so every assertion
	// above can pass while the reactive path is dead. This is what catches it.
	if logs := serverLogs.String(); strings.Contains(logs, fallbackWarning) {
		dumpDiagnostics("reactive path degraded")
		t.Errorf("server fell back to HTML-string diffing; the recursive region is not reactive:\n%s", logs)
	}

	for _, l := range consoleLog.GetLogs() {
		if l.Type == "error" {
			dumpDiagnostics("console error")
			t.Errorf("browser console error: %s", l.Message)
		}
	}
}
