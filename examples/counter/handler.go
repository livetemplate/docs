// Package counter is the literate-counter recipe deployable. Handler
// returns the mountable http.Handler; cmd/main.go wraps it in a
// standalone listener. The docs-site cmd/site aggregator mounts the
// same Handler at /apps/counter/ in the docs container.
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
// still works within a single browser via the shared cookie. Callers
// supply environment-specific options (origin allowlists, dev mode)
// via opts — the recipe itself stays origin-agnostic so cmd/site can
// pass production hosts and cmd/main.go can pass localhost-permissive
// settings under --dev.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("counter", append(baseOpts, opts...)...))
	return tmpl.Handle(&CounterController{}, livetemplate.AsState(&CounterState{}))
}
