// Package counter is the literate-counter recipe deployable. It exposes
// Handler() so the docs-site cmd/site aggregator can mount it at
// /apps/counter/. There is no main() here — production runs via the
// docs binary, not a standalone process.
package counter

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/livetemplate/livetemplate"
)

//go:embed counter.tmpl
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

// extractTemplate writes the embedded template to a temp file so
// livetemplate's file-based loader can parse it at runtime. Done once
// per process. The temp dir survives until the OS reaps /tmp — this
// program does not delete it explicitly, which is fine because it's a
// few-KB file and the binary's lifecycle is the container's lifecycle.
func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "counter-tmpl-*")
		if err != nil {
			log.Fatalf("counter: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("counter.tmpl")
		if err != nil {
			log.Fatalf("counter: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "counter.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("counter: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the counter app as an http.Handler ready to mount.
// AnonymousAuthenticator gives each browser its own session group so
// public visitors get clean state on first visit; multi-tab broadcast
// still works within a single browser via the shared cookie.
func Handler() http.Handler {
	tmpl := livetemplate.Must(livetemplate.New("counter",
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
		livetemplate.WithAllowedOrigins([]string{
			"https://livetemplate.fly.dev",
			"https://livetemplate-docs-staging.fly.dev",
			"http://localhost:8080",
			"http://localhost:8084",
			"http://devbox:8084",
		}),
	))
	return tmpl.Handle(&CounterController{}, livetemplate.AsState(&CounterState{}))
}
