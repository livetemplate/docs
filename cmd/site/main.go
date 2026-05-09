// Command site hosts the docs-site recipe apps on an internal HTTP
// listener. The public-facing docs server is the tinkerdown binary,
// running in the same container; tinkerdown auto-proxies embed-lvt
// blocks here. The block usage on docs pages is:
//
//	embed-lvt path="/apps/<slug>/" upstream="http://localhost:9091"
//
// tinkerdown concatenates upstream + path and fetches
// http://localhost:9091/apps/<slug>/, which this mux serves.
//
// Recipes are imported as Go packages — each exposes `Handler() http.Handler`.
// Adding a recipe is two lines here plus a Go package under
// content/recipes/<slug>/_app/.
package main

import (
	"log"
	"net/http"
	"os"

	counter "github.com/livetemplate/docs/content/getting-started/_app/counter"
)

func main() {
	mux := http.NewServeMux()
	// Recipes are mounted under /apps/<slug>/ to match the embed-lvt
	// `path=` attribute on docs pages. Tinkerdown's auto-proxy
	// concatenates upstream + path, so the loopback URL is identity:
	//   page embed-lvt path="/apps/counter/" upstream="http://localhost:9091"
	//   → tinkerdown fetches http://localhost:9091/apps/counter/
	//   → mux routes to counter.Handler()
	mux.Handle("/apps/counter/", http.StripPrefix("/apps/counter", counter.Handler()))

	addr := ":" + getenv("RECIPES_PORT", "9091")
	log.Printf("docs-site recipes listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
