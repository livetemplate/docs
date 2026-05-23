// Test-server binary for docs/e2e/login. The test suite shells out to
// `go run .` from this directory, so it must contain a main package.
//
// Mounts the login recipe at MOUNT_PATH (default "/"). WithDevMode wires
// the test framework's local client library; WithPermissiveOriginCheck
// waives the production allowlist (the suite spawns on random ports the
// allowlist can't anticipate).
//
// The MOUNT_PATH env var lets the test suite re-run the same recipe at
// a non-root subpath ("/apps/login/") to catch the redirect-target bug
// that root-only testing misses — http.StripPrefix strips the mount
// before the recipe sees the request URL, so a recipe that hard-codes
// "/" as its redirect target appears to work at "/" and breaks under
// any subpath mount.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/livetemplate/livetemplate"

	loginrecipe "github.com/livetemplate/docs/content/recipes/login/_app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mountPath := os.Getenv("MOUNT_PATH")
	if mountPath == "" {
		mountPath = "/"
	}
	if !strings.HasSuffix(mountPath, "/") {
		mountPath += "/"
	}

	devOpts := []livetemplate.Option{
		livetemplate.WithDevMode(true),
		livetemplate.WithPermissiveOriginCheck(),
	}

	mux := http.NewServeMux()
	if mountPath == "/" {
		mux.Handle("/", loginrecipe.Handler(mountPath, devOpts...))
	} else {
		// Mirrors cmd/site: StripPrefix removes the mount before the
		// recipe sees the request URL. The recipe's mountPath argument
		// preserves the absolute prefix for redirect targets.
		stripped := strings.TrimSuffix(mountPath, "/")
		mux.Handle(mountPath, http.StripPrefix(stripped, loginrecipe.Handler(mountPath, devOpts...)))
	}

	log.Printf("login test-server listening on :%s (mount: %s)", port, mountPath)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
