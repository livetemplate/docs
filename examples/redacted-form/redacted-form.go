// Package redactedform demonstrates Preview-mode field redaction: an app can
// let a visitor fill sensitive fields whose values stay in the browser's
// localStorage and never reach the server.
//
// Two framework pieces make this work, both exercised here:
//
//   - The Go template helper {{.lvt.Redact "field"}} renders a "[[field]]"
//     placeholder token. The server-side State only ever holds these tokens for
//     redacted fields, never the real values.
//   - The TypeScript client, on an input tagged data-lvt-redact="field",
//     persists the typed value to localStorage and swaps it for a redact
//     sentinel ({redacted:true,field}) in the outgoing action payload; on each
//     render it substitutes the localStorage value back into the input and into
//     any [[field]] token in page content, before the DOM patch is applied.
//
// The demo: a two-field form (passport number + a non-sensitive note). The
// passport field is redacted; the note is a normal field. A live echo area
// shows the redacted value via {{.lvt.Redact}} so you can see the client
// substituting it back from localStorage. Whatever the server actually holds is
// shown verbatim in the "server sees" panel — for the passport that's the
// "[[passport]]" token, proving the raw value never left the browser.
package redactedform

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

//go:embed redacted-form.html
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

// extractTemplate writes the embedded template to a temp file because
// livetemplate parses templates by filesystem path. Done once per process.
func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "redacted-form-tmpl-*")
		if err != nil {
			log.Fatalf("redacted-form: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("redacted-form.html")
		if err != nil {
			log.Fatalf("redacted-form: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "redacted-form.html")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("redacted-form: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// FormState is pure data, cloned per session.
//
// Note what is and isn't here: Note holds the real note text, but Passport
// holds only whatever the action payload carried. Because the client redacts
// the passport input, the server receives the sentinel and stores the
// placeholder token — the real passport number is never assigned to any field.
type FormState struct {
	// Note is a normal (non-redacted) field — its real value lives server-side.
	Note string `lvt:"persist"`
	// PassportProvided records that a redacted passport value was submitted,
	// without storing the value. The client sends a sentinel, so the server
	// learns presence (for validation / structural state) but not content.
	PassportProvided bool `lvt:"persist"`
}

// FormController holds shared dependencies (none here).
type FormController struct{}

// Save handles the "save" form submission.
//
// ctx.GetString("note") returns the real note text. ctx.Get("passport")
// returns the redact sentinel map ({redacted:true, field:"passport"}) rather
// than a string, so we record presence without ever seeing the value.
func (c *FormController) Save(state FormState, ctx *livetemplate.Context) (FormState, error) {
	state.Note = ctx.GetString("note")

	// A redacted field arrives as a sentinel object, not a string. Detect it
	// structurally; never attempt to read a value that by design isn't here.
	if raw := ctx.Get("passport"); raw != nil {
		if m, ok := raw.(map[string]interface{}); ok {
			if redacted, _ := m["redacted"].(bool); redacted {
				state.PassportProvided = true
			}
		}
	}
	return state, nil
}

// LiveHandler returns just the livetemplate "/" handler for the recipe (no
// client-asset routes). AnonymousAuthenticator gives each browser its own
// session so the demo is safe to expose publicly. The e2e suite mounts this
// directly so it can serve the locally-built client bundle instead of the CDN
// one — the only way to exercise unreleased client changes end-to-end.
func LiveHandler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("redacted-form", append(baseOpts, opts...)...))
	return tmpl.Handle(&FormController{}, livetemplate.AsState(&FormState{}))
}

// Handler returns the redacted-form app as an http.Handler ready to mount,
// including the client JS + CSS routes (served from CDN via the shared test
// helpers, matching the other docs recipes). Callers supply environment-
// specific options.
func Handler(opts ...livetemplate.Option) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", LiveHandler(opts...))
	mux.HandleFunc("/livetemplate-client.js", e2etest.ServeClientLibrary)
	mux.HandleFunc("/livetemplate.css", e2etest.ServeCSS)
	return mux
}
