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
// examples/<slug>/.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/livetemplate/livetemplate"

	"github.com/livetemplate/docs/examples/counter"
	counterbasic "github.com/livetemplate/docs/examples/counter-basic"
	draftform "github.com/livetemplate/docs/examples/draft-form"
	"github.com/livetemplate/docs/examples/greet"
	greetloading "github.com/livetemplate/docs/examples/greet-loading"
	greetloadingserver "github.com/livetemplate/docs/examples/greet-loading-server"
	greetnojs "github.com/livetemplate/docs/examples/greet-nojs"
	greetvalidate "github.com/livetemplate/docs/examples/greet-validate"
	greetwall "github.com/livetemplate/docs/examples/greet-wall"
	loginrecipe "github.com/livetemplate/docs/examples/login"
	"github.com/livetemplate/docs/examples/patterns"
	pe "github.com/livetemplate/docs/examples/progressive-enhancement"
	seatpicker "github.com/livetemplate/docs/examples/seat-picker"
	notepad "github.com/livetemplate/docs/examples/shared-notepad"
	"github.com/livetemplate/docs/examples/todos"
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
		"http://localhost:8099",
		"http://localhost:9091",
		"http://localhost:9191",
		"http://127.0.0.1:8080",
		"http://127.0.0.1:8084",
		"http://127.0.0.1:8099",
		"http://127.0.0.1:9091",
		"http://127.0.0.1:9191",
		"http://devbox:8084",
		"http://devbox:9091",
		"http://devbox:9191",
		"http://100.123.67.113:8084", // devbox tailscale IP (preview over IP)
		"http://100.123.67.113:9091",
		"http://100.123.67.113:9191",
	}

	mux := http.NewServeMux()
	// Recipes are mounted under /apps/<slug>/ to match the embed-lvt
	// `path=` attribute on docs pages. Tinkerdown's auto-proxy
	// concatenates upstream + path, so the loopback URL is identity:
	//   page embed-lvt path="/apps/counter/" upstream="http://localhost:9091"
	//   → tinkerdown fetches http://localhost:9091/apps/counter/
	//   → mux routes to counter.Handler()
	mux.Handle("/apps/counter/", http.StripPrefix("/apps/counter", counter.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))

	// seat-picker is the cross-user real-time recipe. Unlike counter (whose
	// state is per-browser), its Controller is a process-wide singleton, so
	// every visitor shares one seat hall — that shared state is the demo:
	// open it in two windows and watch selections appear live across them.
	// Same shared-state shape as the realtime UI patterns.
	mux.Handle("/apps/seat-picker/", http.StripPrefix("/apps/seat-picker", seatpicker.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))

	// counter-basic is the no-pubsub counter — the homepage "Try it" demo
	// and Steps 2-5 of Your First App. Same template/handler as counter, but
	// its controller omits Subscribe/Publish, so the docs can introduce
	// single-session reactivity first and layer cross-tab sync on top via
	// /apps/counter/ as the "next level."
	mux.Handle("/apps/counter-basic/", http.StripPrefix("/apps/counter-basic", counterbasic.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))

	// greet is the tiny "hello, name" app shown live in the homepage hero
	// (Step 1 of the progressive-narrative spine), beside its own app.tmpl /
	// app.go source. WebSocket-enabled: the hero's "under the hood" animation
	// reveals the real WS round-trip, so the app runs WS to match.
	mux.Handle("/apps/greet/", http.StripPrefix("/apps/greet", greet.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))

	// The progressive-narrative spine's middle steps (2-4) each run a live
	// app, but with WebSocket DISABLED — request/response is exactly what they
	// teach. The client falls back to HTTP fetch (Step 4's "works without JS"
	// degrades further to a plain form POST when JS itself is off).
	//
	//   Step 2 — greet-validate: server-side field validation (NewFieldError),
	//            WS-disabled (a single submit; HTTP fetch is reliable here).
	//   Step 4 — greet-nojs:     the SAME greet handler, remounted WS-disabled,
	//                            proving transport is a config flag, not a rewrite.
	mux.Handle("/apps/greet-validate/", http.StripPrefix("/apps/greet-validate", greetvalidate.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
		livetemplate.WithWebSocketDisabled(),
	)))
	// greet-nojs is Step 4 ("works without JavaScript"): mounted WS-disabled and
	// shown live in a script-disabled iframe so the reader can watch the plain
	// HTTP form-POST round-trip. mountStripped also rewrites the POST-Redirect-
	// GET Location (the handler emits "/") back under this prefix, so the no-JS
	// submit lands on the greet page instead of bouncing to the site root.
	mux.Handle("/apps/greet-nojs/", mountStripped("/apps/greet-nojs", greetnojs.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
		livetemplate.WithWebSocketDisabled(),
	)))

	// draft-form demonstrates server-side formnovalidate: "Publish" validates
	// the required title; "Save draft" carries formnovalidate so the SAME
	// ctx.ValidateForm() call is skipped. The default mount keeps WebSocket on;
	// the /no-js/ mount is WS-disabled (mountStripped rewrites the PRG Location
	// back under the prefix) so the reader can watch the skip hold over a plain
	// HTTP form POST — the tier where the kebab-case save-draft button name also
	// routes to SaveDraft verbatim.
	mux.Handle("/apps/draft-form/", http.StripPrefix("/apps/draft-form", draftform.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))
	mux.Handle("/apps/draft-form/no-js/", mountStripped("/apps/draft-form/no-js", draftform.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
		livetemplate.WithWebSocketDisabled(),
	)))

	// Step 3 — greet-loading: the HTML-declared loading spinner (a class the
	// client toggles on pending/done). WS-disabled like the other middle
	// steps; the loading lifecycle resolves over either transport.
	mux.Handle("/apps/greet-loading/", http.StripPrefix("/apps/greet-loading", greetloading.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
		livetemplate.WithWebSocketDisabled(),
	)))
	// Step 3 also shows the server-owned loading variant. This one keeps
	// Loading in server state and clears it with a follow-up server push, so it
	// keeps WebSocket enabled.
	mux.Handle("/apps/greet-loading-server/", http.StripPrefix("/apps/greet-loading-server", greetloadingserver.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))

	// greet-wall is the spine's climax (Steps 5-7) and the one shared,
	// WebSocket-enabled real-time app on the landing: per-user tab sync
	// (SelfTopic), a cross-user shared wall, and server-initiated push. Like
	// seat-picker its controller is a process-wide singleton, so every visitor
	// shares one wall — open it in two windows to watch greetings appear live.
	mux.Handle("/apps/greet-wall/", http.StripPrefix("/apps/greet-wall", greetwall.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))

	// UI pattern live apps are mounted under /apps/ (like every other embedded
	// recipe app), so the docs URL space /recipes/ui-patterns/<cat>/<slug> is
	// free for the per-pattern markdown pages. The existing /apps/ proxy route
	// forwards here; no per-category proxy routes are needed. The per-pattern
	// docs pages embed these via embed-lvt path="/apps/ui-patterns/<cat>/<slug>".
	//
	// Production options: AnonymousAuthenticator (default — per-browser
	// session group), explicit origin allowlist for the docs deploy
	// targets. The handler package's Handler signature accepts opts so
	// the e2e test-server (docs/e2e/patterns/main.go) can supply
	// WithPermissiveOriginCheck for random-port test setups.
	mux.Handle("/apps/ui-patterns/", http.StripPrefix("/apps/ui-patterns", patterns.Handler("/apps/ui-patterns",
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

	// login — form-based session auth (lvt-form:no-intercept POST + 303
	// + Set-Cookie + OnConnect server-push). The recipe page links to
	// this URL ("Launch demo →") rather than embedding it inline,
	// because lvt-form:no-intercept posts to the current URL — which on
	// the docs page is the markdown route, not the recipe handler. Auth
	// is intrinsic (password "secret"). The first argument is the mount
	// path the recipe redirects to after Login/Logout — http.StripPrefix
	// strips it before the handler sees the request URL, so the handler
	// can't reconstruct it.
	mux.Handle("/apps/login/", http.StripPrefix("/apps/login", loginrecipe.Handler("/apps/login/",
		livetemplate.WithAllowedOrigins(allowedOrigins),
	)))

	// shared-notepad — per-user state map + explicit peer refresh via
	// ctx.Publish(ctx.SelfTopic(), "Refresh", nil). The recipe TEACHES BasicAuth
	// (the e2e suite + examples/shared-notepad use notepad.NewDemoBasicAuth);
	// the embed here uses AnonymousAuthenticator because tinkerdown's
	// embed-lvt server-side prefetch can't forward Authorization headers.
	// Same-browser tabs share the cookie, so the Publish-to-SelfTopic
	// multi-tab refresh story still works in the embed; cross-browser
	// users get different identities for the isolation demo.
	mux.Handle("/apps/shared-notepad/", http.StripPrefix("/apps/shared-notepad", notepad.Handler(
		livetemplate.WithAllowedOrigins(allowedOrigins),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
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

// mountStripped is http.StripPrefix plus redirect-Location rewriting. A recipe
// mounted under a prefix sees its root as "/", so the POST-Redirect-GET path
// (the no-JS form-submit transport) emits a root-absolute Location like "/" —
// which, unstripped, points at the site root rather than back under the mount.
// This re-adds the prefix to such Locations so a no-JS submit returns to the
// recipe. Locations already under the prefix, or absolute URLs, pass through.
func mountStripped(prefix string, h http.Handler) http.Handler {
	stripped := http.StripPrefix(prefix, h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stripped.ServeHTTP(&locationPrefixer{ResponseWriter: w, prefix: prefix}, r)
	})
}

// locationPrefixer rewrites a root-absolute redirect Location to sit under
// prefix, just before the status line is written.
type locationPrefixer struct {
	http.ResponseWriter
	prefix  string
	written bool
}

func (lp *locationPrefixer) WriteHeader(code int) {
	if !lp.written {
		lp.written = true
		if loc := lp.Header().Get("Location"); strings.HasPrefix(loc, "/") &&
			!strings.HasPrefix(loc, "//") && // protocol-relative URL, leave it
			!strings.HasPrefix(loc, lp.prefix+"/") && loc != lp.prefix {
			lp.Header().Set("Location", lp.prefix+loc)
		}
	}
	lp.ResponseWriter.WriteHeader(code)
}
