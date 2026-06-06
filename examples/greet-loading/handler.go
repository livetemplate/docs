// Package greetloading is Step 3 of the docs homepage's progressive-narrative
// spine — the greet app with an HTML-declared loading button. Handler returns
// the mountable http.Handler; cmd/site mounts it at /apps/greet-loading/ so the
// landing's "Loading states, no JS state machine" section can run it live.
package greetloading

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/livetemplate/livetemplate"
)

//go:embed greet.tmpl
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

// extractTemplate writes the embedded template to a temp file so
// livetemplate's file-based loader can parse it at runtime. Done once per
// process (mirrors examples/greet).
func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "greet-loading-tmpl-*")
		if err != nil {
			log.Fatalf("greet-loading: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("greet.tmpl")
		if err != nil {
			log.Fatalf("greet-loading: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "greet.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("greet-loading: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the greet-loading app as an http.Handler ready to mount.
// The initial state greets "there"; the Greet action sleeps briefly so the
// button's pending state is visible. Callers supply environment-specific
// options (origin allowlists, WithWebSocketDisabled) via opts.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("greet-loading", append(baseOpts, opts...)...))
	return tmpl.Handle(&Controller{}, livetemplate.AsState(&State{Name: "there"}))
}
