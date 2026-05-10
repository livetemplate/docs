// Package todos is the docs-native fold of examples/todos — a Tier 1
// CRUD recipe with SQLite persistence, BasicAuth, and the modal/toast
// components from lvt/components. There is no main() here: production
// runs via the docs single-binary container, mounted by cmd/site at
// /apps/todos/ with StripPrefix.
//
// Architecture notes:
//
//   - The .tmpl ships as embed.FS and extracts once to a tmpdir at
//     Handler() time, mirroring the counter and patterns recipes.
//     livetemplate parses templates by filesystem path, so the extract
//     is required.
//
//   - SQLite runs in-memory (":memory:") — the docs deploy has no
//     persistence requirements (visitors test alice/bob, data is
//     ephemeral by design). User-scoped data via the user_id column
//     keeps alice's todos separate from bob's even though they share
//     one process-wide DB.
//
//   - BasicAuth and the modal/toast components are intrinsic to the
//     recipe — they're the teaching surface, not deployment config.
//     Callers (cmd/site, test-server) supply origin policy and
//     dev-mode opts via the variadic opts argument.
package todos

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/livetemplate/livetemplate"
	"github.com/livetemplate/lvt/components/base"
	"github.com/livetemplate/lvt/components/modal"
	"github.com/livetemplate/lvt/components/toast"
	e2etest "github.com/livetemplate/lvt/testing"
)

//go:embed todos.tmpl
var templateFS embed.FS

var (
	validate = validator.New()

	tmplPath    string
	tmplOnce    sync.Once
	handlerOnce sync.Once
	rootHandler http.Handler
)

// extractTemplate writes the embedded template to a tmpdir so
// livetemplate's file-based loader can parse it. Idempotent.
func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "todos-tmpl-*")
		if err != nil {
			log.Fatalf("todos: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("todos.tmpl")
		if err != nil {
			log.Fatalf("todos: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "todos.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("todos: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the todos app as an http.Handler ready to mount. The
// basePath argument is reserved for parity with the patterns recipe;
// todos.tmpl uses absolute paths and current-URL form posts, so the
// value is unused here — kept in the signature so future cross-recipe
// hrefs slot in without a breaking API change.
//
// Production callers (cmd/site) supply WithAllowedOrigins; test-server
// callers supply WithPermissiveOriginCheck + WithDevMode for random-port
// setups.
//
// Calling Handler more than once returns the same first-call handler
// (handlerOnce); the package-level state (DB, template) is one-shot
// per process.
func Handler(basePath string, opts ...livetemplate.Option) http.Handler {
	_ = basePath
	handlerOnce.Do(func() {
		queries, err := InitDB(":memory:")
		if err != nil {
			log.Fatalf("todos: init DB: %v", err)
		}

		controller := &TodoController{Queries: queries}

		auth := livetemplate.NewBasicAuthenticator(func(username, password string) (bool, error) {
			users := map[string]string{
				"alice": "password",
				"bob":   "password",
			}
			pass, ok := users[username]
			return ok && pass == password, nil
		})

		componentSets := []*base.TemplateSet{
			modal.Templates(),
			toast.Templates(),
		}
		ltSets := make([]*livetemplate.TemplateSet, len(componentSets))
		for i, set := range componentSets {
			ltSets[i] = convertTemplateSet(set)
		}

		baseOpts := []livetemplate.Option{
			livetemplate.WithParseFiles(extractTemplate()),
			livetemplate.WithAuthenticator(auth),
			livetemplate.WithComponentTemplates(ltSets...),
		}
		baseOpts = append(baseOpts, opts...)

		tmpl := livetemplate.Must(livetemplate.New("todos", baseOpts...))

		initialState := &TodoState{
			Title:       "Todo App",
			CurrentPage: DefaultPage,
			PageSize:    DefaultPageSize,
			LastUpdated: formatTime(),
		}

		mux := http.NewServeMux()
		mux.Handle("/", tmpl.Handle(controller, livetemplate.AsState(initialState)))
		mux.HandleFunc("/livetemplate-client.js", e2etest.ServeClientLibrary)
		mux.HandleFunc("/livetemplate.css", e2etest.ServeCSS)
		rootHandler = mux
	})
	return rootHandler
}

// convertTemplateSet bridges the components library's TemplateSet shape
// to livetemplate.TemplateSet — the duplicate exists in the components
// library to avoid an import cycle.
func convertTemplateSet(set *base.TemplateSet) *livetemplate.TemplateSet {
	return &livetemplate.TemplateSet{
		FS:        set.FS,
		Pattern:   set.Pattern,
		Namespace: set.Namespace,
		Funcs:     set.Funcs,
	}
}
