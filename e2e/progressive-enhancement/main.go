// Test-server binary for docs/e2e/progressive-enhancement. The
// upstream test framework (e2etest.StartTestServer) shells out to
// `go run .` from the test cwd, so the directory must contain a main
// package.
//
// We mount the recipe twice:
//
//	/        — Tier A (default, JS+WS)
//	/no-ws/  — Tier B (WithWebSocketDisabled)
//
// Tier A is at root rather than /apps/progressive-enhancement/ (the
// production path) so the template's absolute /livetemplate-client.js
// reference resolves without registering the dev-mode static handlers
// at the outer level. Tier B's template makes the same reference;
// ServeMux longest-prefix-wins routes those requests to Tier A's mux,
// which serves the same client lib. Tier C (no-JS) is exercised via
// raw HTTP POST against Tier A — no separate mount needed.
//
// WithDevMode + WithPermissiveOriginCheck:
//   - DevMode wires the test framework's local client library
//   - PermissiveOriginCheck waives the production allowlist; e2e
//     spawns the server on a random port the allowlist can't anticipate
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	pe "github.com/livetemplate/docs/content/recipes/progressive-enhancement/_app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	devOpts := []livetemplate.Option{
		livetemplate.WithDevMode(true),
		livetemplate.WithPermissiveOriginCheck(),
	}

	mux := http.NewServeMux()
	mux.Handle("/no-ws/", http.StripPrefix("/no-ws", pe.Handler(
		append(devOpts, livetemplate.WithWebSocketDisabled())...,
	)))
	mux.Handle("/", pe.Handler(devOpts...))

	log.Printf("progressive-enhancement test-server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
