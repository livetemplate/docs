// Command custom-attributes starts the custom-attributes recipe as a
// standalone HTTP server, so it can be run on its own with
// `go run ./examples/custom-attributes/cmd`.
//
//	PORT              listen port (default 8080)
//	LVT_DEV_MODE=true alias for --dev
//	--dev             relax origin checks for localhost development
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"

	customattributes "github.com/livetemplate/docs/examples/custom-attributes"
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

	log.Printf("custom-attributes listening on :%s", port)
	if err := http.ListenAndServe(":"+port, customattributes.Handler(opts...)); err != nil {
		log.Fatal(err)
	}
}
