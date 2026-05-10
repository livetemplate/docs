// Command site hosts the docs-site recipe apps on an internal HTTP
// listener. The public-facing docs server is the tinkerdown binary,
// running in the same container; tinkerdown auto-proxies embed-lvt
// blocks here. The block usage on docs pages is:
//
//	embed-lvt path="/apps/<slug>/" upstream="http://localhost:9091"
//
// tinkerdown concatenates upstream + path and fetches
// http://localhost:9091/apps/<slug>/, which this mux serves.
//
// Recipes are imported as Go packages — each exposes `Handler() http.Handler`.
// Adding a recipe is two lines here plus a Go package under
// content/recipes/<slug>/_app/.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	counter "github.com/livetemplate/docs/content/recipes/counter/_app"
	patterns "github.com/livetemplate/docs/content/recipes/patterns/_app"
	pe "github.com/livetemplate/docs/content/recipes/progressive-enhancement/_app"
	todos "github.com/livetemplate/docs/content/recipes/todos/_app"
)

func main() {
	// Origin allowlist shared by every recipe — the docs binary serves
	// from one of these hosts in production (Fly prod, Fly staging) or
	// localhost / devbox during dev. Defining it once avoids drift when
	// new origins (e.g. preview deploys) are added.
	allowedOrigins := []string{
		"https://livetemplate.fly.dev",
		"https://livetemplate-docs-staging.fly.dev",
		"http://localhost:8080",
		"http://localhost:8084",
		"http://devbox:8084",
	}

	mux := http.NewServeMux()
	// Recipes are mounted under /apps/<slug>/ to match the embed-lvt
	// `path=` attribute on docs pages. Tinkerdown's auto-proxy
	// concatenates upstream + path, so the loopback URL is identity:
	//   page embed-lvt path="/apps/counter/" upstream="http://localhost:9091"
	//   → tinkerdown fetches http://localhost:9091/apps/counter/
	//   → mux routes to counter.Handler()
	mux.Handle("/apps/counter/", http.StripPrefix("/apps/counter", counter.Handler()))

	// patterns is mounted at /patterns/ (no /apps/ prefix) because it's
	// also the public catalog URL space — tinkerdown's proxy routes for
	// /patterns/forms/, /patterns/lists/, etc. forward there. The recipe
	// page's embed-lvt block uses the same /patterns/... path so a
	// single mount handles both surfaces. Templates render absolute
	// hrefs as {{basePath}}/realtime/broadcasting → /patterns/realtime/broadcasting.
	//
	// Production options: AnonymousAuthenticator (default — per-browser
	// session group), explicit origin allowlist for the docs deploy
	// targets. The handler package's Handler signature accepts opts so
	// the e2e test-server (docs/e2e/patterns/main.go) can supply
	// WithPermissiveOriginCheck for random-port test setups.
	mux.Handle("/patterns/", http.StripPrefix("/patterns", patterns.Handler("/patterns",
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))

	// todos is mounted at /apps/todos/ — recipe-only (no public catalog
	// like patterns). Auth is intrinsic to the recipe (BasicAuth with
	// alice/bob inside todos.Handler), so cmd/site only supplies the
	// origin allowlist for the docs deploy targets.
	mux.Handle("/apps/todos/", http.StripPrefix("/apps/todos", todos.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))

	// progressive-enhancement is mounted twice from one handler package
	// — the only difference is WithWebSocketDisabled on the /no-ws/
	// mount. Tier A (default) demonstrates JS+WS; Tier B (no-ws) shows
	// the client falling back to HTTP fetch when the server rejects WS
	// upgrades; Tier C (no-JS) is the same Tier A URL viewed with
	// JavaScript disabled in the browser — the recipe page describes
	// how to try it.
	mux.Handle("/apps/progressive-enhancement/", http.StripPrefix("/apps/progressive-enhancement", pe.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))
	mux.Handle("/apps/progressive-enhancement/no-ws/", http.StripPrefix("/apps/progressive-enhancement/no-ws", pe.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
		livetemplate.WithWebSocketDisabled(),
	)))

	addr := ":" + getenv("RECIPES_PORT", "9091")
	log.Printf("docs-site recipes listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
