// Package filetree is the recursive-template recipe deployable — a directory
// tree rendered by a template that invokes itself, which is the shape
// {{template}} inlining cannot express and runtime invocation can (livetemplate
// v0.19.0 and later). Handler returns the mountable http.Handler; cmd/main.go
// wraps it in a standalone listener. The docs-site cmd/site aggregator mounts
// the same Handler at /apps/file-tree/ in the docs container.
package filetree

import (
	"embed"
	"net/http"

	"github.com/livetemplate/livetemplate"
)

//go:embed file-tree.tmpl
var templateFS embed.FS

// Handler returns the file-tree app as an http.Handler ready to mount.
// AnonymousAuthenticator gives each browser its own session group, so one
// visitor expanding a folder does not move anyone else's tree. Callers supply
// environment-specific options (origin allowlists, dev mode) via opts — the
// recipe itself stays origin-agnostic so cmd/site can pass production hosts
// and cmd/main.go can pass localhost-permissive settings under --dev.
//
// No depth option is set: the fixture is a handful of levels deep and the
// default cap is 128. WithMaxTemplateDepth matters when recursion runs over
// user-supplied data, where a cycle in the data is possible.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFS(templateFS, "file-tree.tmpl"),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("file-tree", append(baseOpts, opts...)...))
	return tmpl.Handle(&Controller{}, livetemplate.AsState(&State{}))
}
