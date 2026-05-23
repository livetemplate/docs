// Command todos starts the todos recipe as a standalone HTTP server.
// Production deployments mount the same handler via cmd/site at
// /apps/todos/ with StripPrefix; this entry point exists so the recipe
// can be exercised in isolation (local dev, e2e tests, demo).
//
// Tests navigate to bare root with BasicAuth in the URL:
//
//	http://alice:password@host.docker.internal:<port>/
//
// — todos is a single-page recipe (one mount, no internal routing).
//
// Flags / environment:
//
//	PORT    listen port (default 8080)
//	LVT_DEV_MODE=true   alias for --dev (set by e2etest.StartTestServer so the
//	                    subprocess inherits dev-mode without needing to pass
//	                    flags through the test harness)
//	--dev   relax origin checks for localhost development and enable
//	        livetemplate's DevMode (verbose logging the e2e framework
//	        captures). The production allowlist applies when absent.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	"github.com/livetemplate/docs/examples/todos"
)

func main() {
	dev := flag.Bool("dev", false, "enable dev mode (permissive origin checks, dev-mode logging)")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var opts []livetemplate.Option
	if *dev || os.Getenv("LVT_DEV_MODE") == "true" {
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

	log.Printf("todos listening on :%s", port)
	if err := http.ListenAndServe(":"+port, todos.Handler(opts...)); err != nil {
		log.Fatal(err)
	}
}
