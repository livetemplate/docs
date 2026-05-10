// Package progressiveenhancement is the docs-native fold of
// examples/progressive-enhancement (with the "no WebSocket upgrade"
// angle from examples/ws-disabled subsumed). The recipe demonstrates
// LiveTemplate's three graceful-degradation tiers from a single
// controller:
//
//	Tier A — JS on + WS on:  default options, instant updates over WS.
//	Tier B — JS on + WS off: append WithWebSocketDisabled(); the client
//	         falls back to HTTP fetch transparently.
//	Tier C — JS off:         browser submits forms via raw POST; server
//	         responds with 303 See Other (POST-Redirect-GET).
//
// There is no main() here. Production runs via the docs single-binary
// container, mounted by cmd/site twice — once at
// /apps/progressive-enhancement/ (Tier A) and once at
// /apps/progressive-enhancement/no-ws/ (Tier B), the only difference
// being the option set passed in.
//
// Architecture notes:
//
//   - The .tmpl ships as embed.FS and extracts once to a tmpdir at
//     first Handler() call, mirroring counter/todos. livetemplate parses
//     templates by filesystem path, so the extract is required.
//
//   - Unlike the todos package, there is no handlerOnce singleton.
//     Two mounts with different option sets cannot share a sync.Once,
//     and PE has no expensive init (no DB) so caching the handler buys
//     nothing. Each Handler() call builds a fresh livetemplate + mux;
//     session state is cloned per session by the framework, so the two
//     mounts stay isolated.
package progressiveenhancement

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/livetemplate/livetemplate"
	e2etest "github.com/livetemplate/lvt/testing"
)

//go:embed progressive-enhancement.tmpl
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "pe-tmpl-*")
		if err != nil {
			log.Fatalf("progressive-enhancement: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("progressive-enhancement.tmpl")
		if err != nil {
			log.Fatalf("progressive-enhancement: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "progressive-enhancement.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("progressive-enhancement: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the progressive-enhancement app as an http.Handler
// ready to mount. Production callers (cmd/site) supply
// WithAllowedOrigins (and optionally WithWebSocketDisabled for the
// Tier B mount). Test-server callers supply WithDevMode +
// WithPermissiveOriginCheck for random-port setups.
func Handler(opts ...livetemplate.Option) http.Handler {
	controller := &TodoController{validate: validator.New()}
	initialState := &TodoState{}

	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
	}
	baseOpts = append(baseOpts, opts...)

	tmpl := livetemplate.Must(livetemplate.New("progressive-enhancement", baseOpts...))

	mux := http.NewServeMux()
	mux.Handle("/", tmpl.Handle(controller, livetemplate.AsState(initialState)))
	mux.HandleFunc("/livetemplate-client.js", e2etest.ServeClientLibrary)
	mux.HandleFunc("/livetemplate.css", e2etest.ServeCSS)
	return mux
}
