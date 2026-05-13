// Test-server binary for docs/e2e/shared-notepad. The test suite
// shells out to `go run .` from this directory, so it must contain a
// main package.
//
// Mounts the shared-notepad recipe at /, wired with the BasicAuth
// flavour from notepad.NewDemoBasicAuth — password "demo", any
// username — so tests authenticate as alice/bob/etc via the standard
// http://<user>:demo@host/... URL form and exercise the production-
// shaped per-user state map that the recipe text teaches.
//
// WithDevMode wires the test framework's local client library;
// WithPermissiveOriginCheck waives the production allowlist (the
// suite spawns on random ports the allowlist can't anticipate).
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	notepad "github.com/livetemplate/docs/content/recipes/shared-notepad/_app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.Handle("/", notepad.Handler(
		livetemplate.WithDevMode(true),
		livetemplate.WithPermissiveOriginCheck(),
		livetemplate.WithAuthenticator(notepad.NewDemoBasicAuth()),
	))

	log.Printf("shared-notepad test-server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
