// Test-server binary for docs/e2e/recipes/ui-patterns/patterns_test.go. The
// upstream test framework (e2etest.StartTestServer) shells out to
// `go run .` from the test cwd, so the test directory must contain a
// main package.
//
// The upstream examples/patterns/main.go served the catalog index at
// "/" and individual patterns at "/recipes/ui-patterns/<cat>/<slug>". The tests
// were written against that shape — they navigate to bare root for
// the index and to "/recipes/ui-patterns/<cat>/<slug>" for individual pages.
//
// The B1 fold refactored patterns into a basePath-aware single mount:
// internal routes are at "/", "/forms/<slug>", "/api/index.json", etc.,
// no "/recipes/ui-patterns/" prefix. cmd/site mounts that handler at "/recipes/ui-patterns/"
// (with StripPrefix) — so cmd/site's request "/recipes/ui-patterns/forms/..."
// becomes "/forms/..." inside the handler.
//
// To make the unmodified upstream tests work against the docs handler,
// we mount it at BOTH "/" (for tests that navigate to bare root for
// the catalog) AND "/recipes/ui-patterns/" (for tests that hardcode pattern
// pages at "/recipes/ui-patterns/<cat>/<slug>"). Calls to patterns.Handler with
// the same basePath return the same instance via handlerOnce — so
// this is one logical handler reachable via two URL forms.
//
// basePath is "/recipes/ui-patterns" so the templates render hrefs as
// "/recipes/ui-patterns/forms/click-to-edit" — matching what tests assert (e.g.
// TestIndexPage/Pattern_Links queries a[href^="/recipes/ui-patterns/forms/"]).
// The root mount at "/" routes to the index handler inside h's mux
// without any path rewriting; the "/recipes/ui-patterns/" mount strips the
// prefix so individual pattern URLs resolve to h's internal routes.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	patterns "github.com/livetemplate/docs/content/recipes/patterns/_app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// WithPermissiveOriginCheck — the e2e framework spins up servers on
	// random ports and Docker Chrome reaches them via host.docker.internal
	// with that port. The production allowlist can't anticipate those
	// origins; permissive-origin gates this binary to test use only.
	// WithDevMode enables verbose logging the test framework can capture.
	h := patterns.Handler("/recipes/ui-patterns",
		livetemplate.WithDevMode(true),
		livetemplate.WithPermissiveOriginCheck(),
	)
	mux := http.NewServeMux()
	mux.Handle("/recipes/ui-patterns/", http.StripPrefix("/recipes/ui-patterns", h))
	mux.Handle("/", h)
	log.Printf("patterns test-server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
