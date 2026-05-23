// Command patterns starts the UI patterns recipe as a standalone HTTP
// server. Production deployments mount this handler via cmd/site at
// /recipes/ui-patterns/; this entry point exists so the recipe can be
// exercised in isolation (local dev, browser e2e tests).
//
// The recipe handler is mounted twice in the test/dev server:
//
//   - "/" — for tests that navigate to bare root for the catalog index
//     (TestIndexPage et al.)
//   - "/recipes/ui-patterns/" — for tests that hardcode individual
//     pattern URLs like "/recipes/ui-patterns/forms/click-to-edit"
//
// patterns.Handler is idempotent across calls with the same basePath
// (handlerOnce inside the package returns the cached instance), so
// both mounts route to the same logical app instance.
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

	"github.com/livetemplate/docs/examples/patterns"
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

	h := patterns.Handler("/recipes/ui-patterns", opts...)
	mux := http.NewServeMux()
	mux.Handle("/recipes/ui-patterns/", http.StripPrefix("/recipes/ui-patterns", h))
	mux.Handle("/", h)

	log.Printf("patterns listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
