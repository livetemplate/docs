// Command progressive-enhancement starts the recipe as a standalone
// HTTP server with dual mounts:
//
//	/        — Tier A (default, JS + WebSocket)
//	/no-ws/  — Tier B (WithWebSocketDisabled, HTTP-only fallback)
//
// Tier C (no-JS) is exercised against Tier A by disabling JavaScript
// in the browser; no separate mount.
//
// Production cmd/site mounts the same handler twice with explicit
// origins at /apps/progressive-enhancement/ and /apps/progressive-
// enhancement/no-ws/. This entry point exists so the recipe can be
// exercised in isolation (local dev, e2e tests).
//
// Tier A is at root rather than /apps/progressive-enhancement/ (the
// production path) so the template's absolute /livetemplate-client.js
// reference resolves without registering dev-mode static handlers at
// the outer level. Tier B's template makes the same reference;
// ServeMux longest-prefix-wins routes those requests to Tier A's mux.
//
// Flags / environment:
//
//	PORT    listen port (default 8080)
//	LVT_DEV_MODE=true   alias for --dev (set by e2etest.StartTestServer so the
//	                    subprocess inherits dev-mode without needing to pass
//	                    flags through the test harness)
//	--dev   relax origin checks for localhost development and enable
//	        livetemplate's DevMode (verbose logging tests capture).
//	        The production allowlist applies when absent.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	pe "github.com/livetemplate/docs/examples/progressive-enhancement"
)

func main() {
	dev := flag.Bool("dev", false, "enable dev mode (permissive origin checks, dev-mode logging)")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var baseOpts []livetemplate.Option
	if *dev || os.Getenv("LVT_DEV_MODE") == "true" {
		baseOpts = append(baseOpts,
			livetemplate.WithDevMode(true),
			livetemplate.WithPermissiveOriginCheck(),
		)
	} else {
		baseOpts = append(baseOpts, livetemplate.WithAllowedOrigins([]string{
			"https://livetemplate.fly.dev",
			"https://livetemplate-docs-staging.fly.dev",
			"http://localhost:8080",
			"http://localhost:8084",
			"http://devbox:8084",
		}))
	}

	mux := http.NewServeMux()
	mux.Handle("/no-ws/", http.StripPrefix("/no-ws", pe.Handler(
		append(baseOpts, livetemplate.WithWebSocketDisabled())...,
	)))
	mux.Handle("/", pe.Handler(baseOpts...))

	log.Printf("progressive-enhancement listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
