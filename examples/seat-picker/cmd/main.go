// Command seat-picker starts the seat-picker recipe as a standalone HTTP
// server. Production deployments embed the recipe via the docs-site
// cmd/site aggregator (mounted at /apps/seat-picker/); this entry point
// exists so developers can run `go run ./examples/seat-picker/cmd` to
// iterate in isolation, and so the e2e harness can drive a real browser
// against a real process.
//
// Flags / environment:
//
//	PORT    listen port (default 8095)
//	LVT_DEV_MODE=true   alias for --dev (set by e2etest.StartTestServer so the
//	                    subprocess inherits dev-mode without flag plumbing)
//	--dev   relax origin checks for localhost development (so the WebSocket
//	        upgrader accepts arbitrary localhost ports) and enable DevMode,
//	        which makes the template load the local client assets. The
//	        production allowlist applies when the flag is absent.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	seatpicker "github.com/livetemplate/docs/examples/seat-picker"
)

func main() {
	dev := flag.Bool("dev", false, "enable dev mode (permissive origin checks, dev-mode template assets)")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8095"
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

	log.Printf("seat-picker listening on :%s", port)
	log.Println("Open two browsers, join as different names, and pick seats — each sees the other live.")
	if err := http.ListenAndServe(":"+port, seatpicker.Handler(opts...)); err != nil {
		log.Fatal(err)
	}
}
