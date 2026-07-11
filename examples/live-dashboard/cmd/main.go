// Command live-dashboard starts the live-dashboard recipe as a standalone HTTP
// server. Production cmd/site mounts the same Handler at /apps/live-dashboard/.
//
// Flags / environment:
//
//	PORT               listen port (default 8080)
//	LVT_DEV_MODE=true  alias for --dev (set by e2etest.StartTestServer so the
//	                   subprocess inherits dev-mode without flag plumbing)
//	--dev              relax origin checks for localhost development and enable
//	                   livetemplate's DevMode. Production allowlist applies when
//	                   absent.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	livedashboard "github.com/livetemplate/docs/examples/live-dashboard"
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

	mux := http.NewServeMux()
	mux.Handle("/", livedashboard.Handler(opts...))

	log.Printf("live-dashboard listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
