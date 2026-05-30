// Package counterbasic is the no-pubsub counter recipe deployable — the
// single-session counter shown first on the docs homepage and in Steps 2-5
// of the Your First App tutorial. examples/counter is the same app with
// cross-tab pubsub layered on (the tutorial's Step 6 / "next level").
// Handler returns the mountable http.Handler; cmd/main.go wraps it in a
// standalone listener. The docs-site cmd/site aggregator mounts the same
// Handler at /apps/counter-basic/ in the docs container.
package counterbasic

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
		dir, err := os.MkdirTemp("", "counter-basic-tmpl-*")
		if err != nil {
			log.Fatalf("counter-basic: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("counter.tmpl")
		if err != nil {
			log.Fatalf("counter-basic: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "counter.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("counter-basic: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the basic counter app as an http.Handler ready to mount.
// AnonymousAuthenticator gives each browser its own session group so public
// visitors get clean state on first visit. Unlike examples/counter, this
// recipe does not Subscribe/Publish, so there is no cross-tab sync — that's
// deliberately the "next level" the docs introduce afterward. Callers supply
// environment-specific options (origin allowlists, dev mode) via opts — the
// recipe itself stays origin-agnostic so cmd/site can pass production hosts
// and cmd/main.go can pass localhost-permissive settings under --dev.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("counter-basic", append(baseOpts, opts...)...))
	return tmpl.Handle(&CounterController{}, livetemplate.AsState(&CounterState{}))
}
