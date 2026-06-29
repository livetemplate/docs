// Package draftform demonstrates server-side formnovalidate honoring: a post
// editor where "Publish" validates the required title and "Save draft" (a
// formnovalidate button) skips validation — the same ctx.ValidateForm() call,
// different outcome by submitter, on every tier. cmd/site mounts it at
// /apps/draft-form/ and (WS-disabled) at /apps/draft-form/no-js/.
package draftform

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/livetemplate/livetemplate"
)

//go:embed draft.tmpl
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

// extractTemplate writes the embedded template to a temp file so livetemplate's
// file-based loader can parse it at runtime (mirrors examples/greet-validate).
func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "draft-form-tmpl-*")
		if err != nil {
			log.Fatalf("draft-form: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("draft.tmpl")
		if err != nil {
			log.Fatalf("draft-form: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "draft.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("draft-form: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the draft-form app as an http.Handler ready to mount. The
// initial state is an empty editor; callers supply environment-specific options
// (origin allowlists, WithWebSocketDisabled) via opts.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("draft-form", append(baseOpts, opts...)...))
	return tmpl.Handle(&Controller{}, livetemplate.AsState(&State{}))
}
