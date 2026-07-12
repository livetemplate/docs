// Command redacted-form starts the redacted-form recipe as a standalone HTTP
// server. Production cmd/site mounts the same handler; this entry point exists
// so the recipe can be exercised in isolation and so the cross-repo test
// harness can drive a real browser against a real process.
//
// Flags / environment:
//
//	PORT                listen port (default 8080)
//	LVT_DEV_MODE=true   alias for --dev (set by e2etest.StartTestServer so the
//	                    subprocess inherits dev-mode without flag plumbing)
//	--dev               relax origin checks for localhost development and enable
//	                    livetemplate's DevMode. Production allowlist applies when
//	                    absent.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	redactedform "github.com/livetemplate/docs/examples/redacted-form"
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

	log.Printf("redacted-form listening on :%s", port)
	if err := http.ListenAndServe(":"+port, redactedform.Handler(opts...)); err != nil {
		log.Fatal(err)
	}
}
