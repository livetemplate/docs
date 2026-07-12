// Command notepad starts the shared-notepad recipe as a standalone
// HTTP server. Production cmd/site mounts the same handler at
// /apps/shared-notepad/ with AnonymousAuthenticator (because
// tinkerdown's embed-lvt server-side prefetch can't forward
// Authorization headers in inline embeds). This entry point uses
// notepad.NewDemoBasicAuth so tests can authenticate as alice/bob/etc
// via http://<user>:demo@host/... and exercise the per-user state map
// the recipe teaches.
//
// Flags / environment:
//
//	PORT    listen port (default 8080)
//	LVT_DEV_MODE=true   alias for --dev (set by e2etest.StartTestServer so the
//	                    subprocess inherits dev-mode without needing to pass
//	                    flags through the test harness)
//	--dev   relax origin checks for localhost development and enable
//	        livetemplate's DevMode. Production allowlist applies when
//	        absent.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	notepad "github.com/livetemplate/docs/examples/shared-notepad"
)

func main() {
	dev := flag.Bool("dev", false, "enable dev mode (permissive origin checks, dev-mode logging)")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	opts := []livetemplate.Option{
		livetemplate.WithAuthenticator(notepad.NewDemoBasicAuth()),
	}
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

	mux := http.NewServeMux()
	mux.Handle("/", notepad.Handler(opts...))

	log.Printf("shared-notepad listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
