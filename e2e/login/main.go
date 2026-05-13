// Test-server binary for docs/e2e/login. The test suite shells out to
// `go run .` from this directory, so it must contain a main package.
//
// Mounts the login recipe at /. WithDevMode wires the test framework's
// local client library; WithPermissiveOriginCheck waives the production
// allowlist (the suite spawns on random ports the allowlist can't
// anticipate).
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	loginrecipe "github.com/livetemplate/docs/content/recipes/login/_app"
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
	mux.Handle("/", loginrecipe.Handler(devOpts...))

	log.Printf("login test-server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
