// Package greetvalidate is Step 2 of the docs homepage's progressive-narrative
// spine — the greet app with server-side field validation. Handler returns the
// mountable http.Handler; cmd/site mounts it at /apps/greet-validate/ so the
// landing's "Validation lives in Go" section can run it live.
package greetvalidate

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
		dir, err := os.MkdirTemp("", "greet-validate-tmpl-*")
		if err != nil {
			log.Fatalf("greet-validate: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("greet.tmpl")
		if err != nil {
			log.Fatalf("greet-validate: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "greet.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("greet-validate: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the greet-validate app as an http.Handler ready to mount.
// The initial state greets "there"; an empty submit returns a field error
// rather than changing the greeting. Callers supply environment-specific
// options (origin allowlists, WithWebSocketDisabled) via opts.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("greet-validate", append(baseOpts, opts...)...))
	return tmpl.Handle(&Controller{}, livetemplate.AsState(&State{Name: "there"}))
}
