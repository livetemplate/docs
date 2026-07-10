// Package greet is the tiny "hello, name" recipe shown live in the docs
// homepage hero, right beside its own app.tmpl / app.go source. Handler
// returns the mountable http.Handler; the docs-site cmd/site aggregator
// mounts it at /apps/greet/ so the hero embed-lvt block can run it.
package greet

import (
	"embed"
	"net/http"

	"github.com/livetemplate/livetemplate"
)

//go:embed greet.tmpl
var templateFS embed.FS

// Handler returns the greet app as an http.Handler ready to mount.
// AnonymousAuthenticator gives each browser its own session group. The initial
// state greets "there" so the hero shows "Hello, there" before any input.
// Callers supply environment-specific options (origin allowlists, dev mode)
// via opts so the recipe itself stays origin-agnostic.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFS(templateFS, "greet.tmpl"),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("greet", append(baseOpts, opts...)...))
	return tmpl.Handle(&Controller{}, livetemplate.AsState(&State{Name: "there"}))
}
