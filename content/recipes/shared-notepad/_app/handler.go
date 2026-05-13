// Package notepad is the docs-native fold of examples/shared-notepad.
// The recipe demonstrates LiveTemplate's authenticated per-user state
// shape:
//
//   - BasicAuthenticator turns the Authorization header username into
//     ctx.UserID() and into the session-group ID. Different users get
//     different groups; same user across tabs/devices gets the same.
//
//   - A controller-owned map keyed by ctx.UserID() holds the notepad
//     state under a sync.RWMutex. Mount + Refresh both read from it;
//     Save writes.
//
//   - ctx.BroadcastAction("Refresh", nil) is the explicit peer-refresh
//     primitive. After Save commits, every peer connection in the same
//     session group (other tabs of the same user) runs Refresh and
//     re-reads from the map.
//
// There is no main() here. Production runs via the docs single-binary
// container, mounted by cmd/site at /apps/shared-notepad/. The example
// is also linked from the recipe page; opening it in a browser triggers
// the BasicAuth prompt and exercises the real auth path — which an
// iframe can't fairly demonstrate, so the recipe links out instead.
//
// Architecture notes:
//
//   - The .tmpl ships as embed.FS and extracts once to a tmpdir at
//     first Handler() call (livetemplate parses templates by filesystem
//     path).
//
//   - No handlerOnce singleton. The controller is cheap; consistency
//     with progressive-enhancement keeps the Handler() shape uniform.
//
//   - The default authenticator is wired into the recipe's defaults —
//     BasicAuth with password "demo" — so the recipe runs identically
//     in cmd/site and the e2e harness. Callers can override the auth
//     via opts (e.g. tests that need a fixed userID).
package notepad

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

//go:embed notepad.tmpl
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "notepad-tmpl-*")
		if err != nil {
			log.Fatalf("shared-notepad recipe: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("notepad.tmpl")
		if err != nil {
			log.Fatalf("shared-notepad recipe: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "notepad.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("shared-notepad recipe: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// >>> region:basicauth
// NewDemoBasicAuth returns the authenticator the recipe text teaches:
// BasicAuth with password "demo", any username. The username becomes
// both ctx.UserID() (per-user state map key) and the session-group ID
// (BroadcastAction routing). This is the production-shaped wiring and
// what examples/shared-notepad + the e2e suite use.
//
// The docs-site mount (cmd/site) uses AnonymousAuthenticator instead
// of this — see the Handler doc comment for the reason.
func NewDemoBasicAuth() livetemplate.Authenticator {
	return livetemplate.NewBasicAuthenticator(func(_, password string) (bool, error) {
		return password == "demo", nil
	})
}

// <<< region:basicauth

// Handler returns the shared-notepad app as an http.Handler ready to
// mount. No default authenticator — the caller picks. Two flavours
// ship with the recipe:
//
//   - NewDemoBasicAuth(): production-shaped, password "demo". Used by
//     the e2e suite and examples/shared-notepad. ctx.UserID() comes
//     from the Authorization header.
//
//   - livetemplate.AnonymousAuthenticator: cookie-bound, no prompt.
//     Used by the docs-site mount (cmd/site) because tinkerdown's
//     embed-lvt block does a server-side prefetch of the upstream to
//     extract the LiveTemplate wrapper, and that prefetch can't
//     forward Authorization headers — a BasicAuth mount would degrade
//     to "live demo unavailable" in the docs page. Same-browser tabs
//     share the cookie, so the multi-tab BroadcastAction demo still
//     works; different browsers get different identities for isolation.
func Handler(opts ...livetemplate.Option) http.Handler {
	controller := &NotepadController{
		notes: make(map[string]NotepadState),
	}
	initialState := &NotepadState{}

	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
	}
	baseOpts = append(baseOpts, opts...)

	tmpl := livetemplate.Must(livetemplate.New("notepad", baseOpts...))

	mux := http.NewServeMux()
	mux.Handle("/", tmpl.Handle(controller, livetemplate.AsState(initialState)))
	mux.HandleFunc("/livetemplate-client.js", e2etest.ServeClientLibrary)
	mux.HandleFunc("/livetemplate.css", e2etest.ServeCSS)
	return mux
}
