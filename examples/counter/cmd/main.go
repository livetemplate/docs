// Command counter starts the counter recipe as a standalone HTTP
// server. Production deployments embed the recipe via the docs-site
// cmd/site aggregator; this entry point exists so developers can run
// `go run ./examples/counter/cmd` to iterate on the recipe in
// isolation, and so the cross-repo test harness can drive a real
// browser against a real process.
//
// Flags / environment:
//
//	PORT    listen port (default 8080)
//	LVT_DEV_MODE=true   alias for --dev (set by e2etest.StartTestServer so the
//	                    subprocess inherits dev-mode without needing to pass
//	                    flags through the test harness)
//	--dev   relax origin checks for localhost development (so the
//	        WebSocket upgrader accepts requests from arbitrary
//	        localhost ports); the production allowlist applies when
//	        the flag is absent.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	"github.com/livetemplate/docs/examples/counter"
)

func main() {
	dev := flag.Bool("dev", false, "enable dev mode (permissive origin checks, dev-mode template reload)")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var opts []livetemplate.Option
	if *dev || os.Getenv("LVT_DEV_MODE") == "true" {
		// Dev mode also relaxes the WebSocket origin check (allows all
		// origins), so localhost on any port works during development.
		opts = append(opts, livetemplate.WithDevMode(true))
	} else {
		opts = append(opts, livetemplate.WithAllowedOrigins([]string{
			"https://livetemplate.fly.dev",
			"https://livetemplate-docs-staging.fly.dev",
			"http://localhost:8080",
			"http://localhost:8084",
			"http://devbox:8084",
		}))
	}

	log.Printf("counter listening on :%s", port)
	if err := http.ListenAndServe(":"+port, counter.Handler(opts...)); err != nil {
		log.Fatal(err)
	}
}
