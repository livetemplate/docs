// Package loginrecipe is the docs-native fold of examples/login. The
// recipe demonstrates LiveTemplate's form-based session-auth shape:
//
//   - HTTP POST login (lvt-form:no-intercept) that issues an HttpOnly
//     session cookie via ctx.SetCookie, then 303-redirects.
//   - OnConnect server-side lifecycle hook that fires a goroutine to
//     push a welcome message back over the WebSocket via
//     session.TriggerAction (no client poll, no second request).
//   - Symmetric logout: HTTP POST logout (lvt-form:no-intercept) that
//     deletes the cookie via ctx.DeleteCookie and 303-redirects.
//
// There is no main() here. Production runs via the docs single-binary
// container, mounted by cmd/site at /apps/login/. The example is also
// linked from the recipe page; opening it in a browser does an
// honest-to-goodness POST + cookie + redirect flow, which an iframe
// can't fairly demonstrate — so the recipe links out rather than
// embedding inline.
//
// Architecture notes:
//
//   - The auth.html template ships as embed.FS and extracts once to a
//     tmpdir at first Handler() call. livetemplate parses templates by
//     filesystem path, so the extract is required.
//
//   - No handlerOnce singleton. The controller is cheap (no DB), and
//     keeping the Handler() shape consistent with progressive-enhancement
//     means future per-mount option overrides come free.
//
//   - The controller holds a sessions map keyed by username for the
//     server-push demo. It's intentionally trivial — a real app would
//     plumb session storage through SessionStore.
package loginrecipe

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/livetemplate/livetemplate"
	e2etest "github.com/livetemplate/lvt/testing"
)

//go:embed auth.html
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "login-tmpl-*")
		if err != nil {
			log.Fatalf("login recipe: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("auth.html")
		if err != nil {
			log.Fatalf("login recipe: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "auth.html")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("login recipe: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the login recipe app as an http.Handler ready to
// mount. Production callers (cmd/site) supply WithAllowedOrigins;
// test-server callers (docs/e2e/login) supply WithDevMode +
// WithPermissiveOriginCheck for random-port test setups.
func Handler(opts ...livetemplate.Option) http.Handler {
	controller := &AuthController{
		sessions: make(map[string]livetemplate.Session),
	}
	initialState := &AuthState{}

	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
	}
	baseOpts = append(baseOpts, opts...)

	tmpl := livetemplate.Must(livetemplate.New("auth", baseOpts...))

	mux := http.NewServeMux()
	mux.Handle("/", tmpl.Handle(controller, livetemplate.AsState(initialState)))
	mux.HandleFunc("/livetemplate-client.js", e2etest.ServeClientLibrary)
	mux.HandleFunc("/livetemplate.css", e2etest.ServeCSS)
	return mux
}
