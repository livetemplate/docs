// Package greet is the tiny "hello, name" recipe shown live in the docs
// homepage hero, right beside its own app.tmpl / app.go source. Handler
// returns the mountable http.Handler; the docs-site cmd/site aggregator
// mounts it at /apps/greet/ so the hero embed-lvt block can run it.
package greet

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
// process (mirrors examples/counter-basic).
func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "greet-tmpl-*")
		if err != nil {
			log.Fatalf("greet: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("greet.tmpl")
		if err != nil {
			log.Fatalf("greet: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "greet.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("greet: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the greet app as an http.Handler ready to mount.
// AnonymousAuthenticator gives each browser its own session group. The initial
// state greets "there" so the hero shows "Hello, there" before any input.
// Callers supply environment-specific options (origin allowlists, dev mode)
// via opts so the recipe itself stays origin-agnostic.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("greet", append(baseOpts, opts...)...))
	return tmpl.Handle(&Controller{}, livetemplate.AsState(&State{Name: "there"}))
}
