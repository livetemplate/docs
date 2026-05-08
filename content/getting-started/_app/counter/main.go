package main

import (
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"
)

// sharedAuth puts every connection in the same session group so
// BroadcastAction calls reach all clients, including the side-by-side
// embeds on the docs page. Real apps use a per-user authenticator;
// here a constant groupID is what makes the linked-embed demo work.
type sharedAuth struct{}

func (sharedAuth) Identify(r *http.Request) (string, error) {
	return "shared", nil
}

func (sharedAuth) GetSessionGroup(r *http.Request, userID string) (string, error) {
	return "shared", nil
}

func main() {
	tmpl := livetemplate.Must(livetemplate.New("counter",
		livetemplate.WithParseFiles("counter.tmpl"),
		livetemplate.WithAuthenticator(sharedAuth{}),
		// Browser WS Origin stays as the docs origin (tinkerdown's
		// reverse-proxy rewrites Host but not Origin). Allowlist the
		// docs site origins plus common local-dev hosts; a permissive
		// check would let any third-party site drive the shared demo.
		livetemplate.WithAllowedOrigins([]string{
			"https://livetemplate.fly.dev",
			"https://livetemplate-docs-staging.fly.dev",
			"http://localhost:8080",
			"http://localhost:8084",
			"http://devbox:8084",
		}),
	))
	handler := tmpl.Handle(&CounterController{}, livetemplate.AsState(&CounterState{}))

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	log.Printf("firstapp-counter listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
