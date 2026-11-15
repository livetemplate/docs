// Package customattributes is the custom-attributes recipe deployable: three
// lvt-x:* attributes written by the app itself, registered from an inline
// script in the template. Handler returns the mountable http.Handler;
// cmd/main.go wraps it in a standalone listener. The docs-site cmd/site
// aggregator mounts the same Handler at /apps/custom-attributes/.
package customattributes

import (
	"embed"
	"net/http"

	"github.com/livetemplate/livetemplate"
)

//go:embed custom-attributes.tmpl
var templateFS embed.FS

// Handler returns the app as an http.Handler ready to mount. Each browser gets
// its own session group, so one visitor's rating is not anyone else's. Callers
// supply environment-specific options (origin allowlists, dev mode) via opts.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFS(templateFS, "custom-attributes.tmpl"),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("custom-attributes", append(baseOpts, opts...)...))
	return tmpl.Handle(&Controller{}, livetemplate.AsState(&State{Shortcuts: true}))
}
