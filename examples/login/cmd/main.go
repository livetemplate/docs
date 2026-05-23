// Command login starts the login recipe as a standalone HTTP server.
// Production cmd/site mounts the same handler at /apps/login/; this
// entry point exists so the recipe can be exercised in isolation.
//
// MOUNT_PATH env var lets the test suite re-run the same recipe at a
// non-root subpath ("/apps/login/") to catch the redirect-target bug
// root-only testing misses — http.StripPrefix removes the mount
// before the recipe sees the request URL, so a recipe that hard-codes
// "/" as its redirect target appears to work at "/" and breaks under
// any subpath mount.
//
// Flags / environment:
//
//	PORT        listen port (default 8080)
//	MOUNT_PATH  mount path (default "/")
//	--dev       relax origin checks for localhost development and enable
//	            livetemplate's DevMode. Production allowlist applies when
//	            absent.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/livetemplate/livetemplate"

	loginrecipe "github.com/livetemplate/docs/examples/login"
)

func main() {
	dev := flag.Bool("dev", false, "enable dev mode (permissive origin checks, dev-mode logging)")
	flag.Parse()

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

	var opts []livetemplate.Option
	if *dev {
		opts = append(opts,
			livetemplate.WithDevMode(true),
			livetemplate.WithPermissiveOriginCheck(),
		)
	} else {
		opts = append(opts, livetemplate.WithAllowedOrigins([]string{
			"https://livetemplate.fly.dev",
			"https://livetemplate-docs-staging.fly.dev",
			"http://localhost:8080",
			"http://localhost:8084",
			"http://devbox:8084",
		}))
	}

	mux := http.NewServeMux()
	if mountPath == "/" {
		mux.Handle("/", loginrecipe.Handler(mountPath, opts...))
	} else {
		stripped := strings.TrimSuffix(mountPath, "/")
		mux.Handle(mountPath, http.StripPrefix(stripped, loginrecipe.Handler(mountPath, opts...)))
	}

	log.Printf("login listening on :%s (mount: %s)", port, mountPath)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
