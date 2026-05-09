// Package patterns is the docs-native fold of examples/patterns —
// 33 reactive UI pattern handlers exposed as a single Go package
// consumable by docs/cmd/site. There is no main() here: production
// runs via the docs single-binary container, mounted by cmd/site at
// /apps/patterns/ with StripPrefix.
//
// Architecture notes:
//
//   - Templates ship as embed.FS and extract once to a tmpdir at
//     Handler() time, mirroring the counter recipe pattern. livetemplate
//     parses templates by filesystem path, so the extract is required.
//
//   - All hrefs that need to point inside this package render via
//     {{basePath}}/<route>. basePath is registered as a template func
//     (via Template.Funcs after New). The closure reads pkgBasePath,
//     set once at Handler() invocation time.
//
//   - Inner mux registrations DROP the "/patterns/" prefix that the
//     upstream examples/patterns/ uses. cmd/site StripPrefix strips
//     "/apps/patterns" so a request "/apps/patterns/realtime/broadcasting"
//     reaches this mux as "/realtime/broadcasting". Public catalog
//     URLs (in data.go's PatternEntry.Path and api_index.go's JSON)
//     keep the "/patterns/" shape — they're a separate contract.
package patterns

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/livetemplate/livetemplate"
	e2etest "github.com/livetemplate/lvt/testing"
)

// stripPatternsPrefix turns "/patterns/forms/click-to-edit" into
// "/forms/click-to-edit". Used by the relPath template func so the
// catalog index can navigate within the docs container at any mount
// prefix without the data.go entries hardcoding any specific prefix.
func stripPatternsPrefix(p string) string {
	return strings.TrimPrefix(p, "/patterns")
}

//go:embed templates
var templateFS embed.FS

var (
	tmplRoot     string
	tmplOnce     sync.Once
	pkgBasePath  string
	pkgBaseOpts  []livetemplate.Option
	pkgFuncs     template.FuncMap
	handlerOnce  sync.Once
	rootHandler  http.Handler
)

// Handler returns an http.Handler that serves all 33 pattern routes
// plus /api/index.json + the dev-mode static asset routes
// (/livetemplate-client.js, /livetemplate.css). Mount it with
// http.StripPrefix; the basePath argument is the prefix the docs
// container exposes externally (e.g. "/apps/patterns").
//
// basePath has no trailing slash. Templates were authored with the
// literal token "{{basePath}}" which extractTemplates rewrites to the
// runtime value before livetemplate parses them — the substitution
// has to happen at extract time because livetemplate.New parses
// immediately and rejects unknown template funcs.
//
// Calling Handler more than once with a different basePath returns
// the same first-call handler; the package-level state is one-shot.
// (B2's eventual second mount at /patterns/ would need a different
// approach — likely a second package init or relaxing handlerOnce
// to scope per-basePath. Out of scope for B1.)
func Handler(basePath string) http.Handler {
	handlerOnce.Do(func() {
		pkgBasePath = basePath
		pkgBaseOpts = []livetemplate.Option{
			livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
			livetemplate.WithDevMode(true),
			livetemplate.WithPermissiveOriginCheck(),
		}
		// pkgFuncs is unused for now — every template-time helper this
		// package needs (relPath) is now a method on PatternLink so
		// livetemplate.New's eager parse doesn't need to know about it.
		// Kept declared so a future func-based helper has a slot.
		pkgFuncs = template.FuncMap{}
		extractTemplates()
		rootHandler = buildMux()
	})
	return rootHandler
}

// extractTemplates walks templateFS and writes every embedded file to
// a tmpdir whose path is captured in tmplRoot. The literal token
// "{{basePath}}" is substituted with pkgBasePath as each file is
// written — html/template's Parse rejects unknown funcs, so we have
// to bake basePath into the template source before livetemplate sees
// it. relPath stays a template func because it takes a .Path arg
// known only at render time. Idempotent — guarded by tmplOnce. The
// tmpdir survives until the OS reaps /tmp; the binary's lifecycle
// equals the container's.
func extractTemplates() {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "patterns-tmpl-*")
		if err != nil {
			log.Fatalf("patterns: mkdtemp: %v", err)
		}
		err = fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			outPath := filepath.Join(dir, path)
			if d.IsDir() {
				return os.MkdirAll(outPath, 0o755)
			}
			data, readErr := templateFS.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			out := strings.ReplaceAll(string(data), "{{basePath}}", pkgBasePath)
			return os.WriteFile(outPath, []byte(out), 0o644)
		})
		if err != nil {
			log.Fatalf("patterns: extract templates: %v", err)
		}
		tmplRoot = dir
	})
}

// newLayoutTmpl is the canonical template constructor used by every
// handler factory. It joins relative template paths to tmplRoot,
// builds with pkgBaseOpts + WithParseFiles, then registers pkgFuncs
// so {{basePath}} resolves at render time.
func newLayoutTmpl(files ...string) *livetemplate.Template {
	return newLayoutTmplWithOpts(files)
}

// newLayoutTmplWithOpts is the variant for handlers that need extra
// options beyond WithParseFiles — file-upload + preserve-inputs use
// it to add WithUpload configs.
func newLayoutTmplWithOpts(files []string, extra ...livetemplate.Option) *livetemplate.Template {
	full := make([]string, len(files))
	for i, f := range files {
		full[i] = filepath.Join(tmplRoot, f)
	}
	opts := append(slices.Clone(pkgBaseOpts),
		livetemplate.WithParseFiles(full...),
	)
	opts = append(opts, extra...)
	tmpl := livetemplate.Must(livetemplate.New("layout", opts...))
	tmpl.Funcs(pkgFuncs)
	return tmpl
}

// buildMux registers every pattern route, the API index, and the
// dev-mode static asset routes. Routes are registered without the
// "/patterns/" prefix — cmd/site's StripPrefix layer handles the
// external mount path.
func buildMux() http.Handler {
	mux := http.NewServeMux()

	// Index — catalog landing page (renders all 33 patterns).
	mux.Handle("/", indexHandler())

	// Forms & Editing
	mux.Handle("/forms/click-to-edit", clickToEditHandler())
	mux.Handle("/forms/edit-row", editRowHandler())
	mux.Handle("/forms/inline-validation", inlineValidationHandler())
	mux.Handle("/forms/bulk-update", bulkUpdateHandler())
	mux.Handle("/forms/reset-input", resetInputHandler())
	mux.Handle("/forms/file-upload", fileUploadHandler())
	mux.Handle("/forms/preserve-inputs", preserveInputsHandler())

	// Lists & Data
	mux.Handle("/lists/delete-row", deleteRowHandler())
	mux.Handle("/lists/click-to-load", clickToLoadHandler())
	mux.Handle("/lists/infinite-scroll", infiniteScrollHandler())
	mux.Handle("/lists/value-select", valueSelectHandler())
	mux.Handle("/lists/sortable", sortableHandler())
	mux.Handle("/lists/large-table", largeTableHandler())

	// Search & Filtering
	mux.Handle("/search/active-search", activeSearchHandler())
	mux.Handle("/search/url-filters", urlFiltersHandler())

	// Loading & Progress
	mux.Handle("/loading/lazy-loading", lazyLoadingHandler())
	mux.Handle("/loading/progress-bar", progressBarHandler())
	mux.Handle("/loading/async-operations", asyncOperationsHandler())

	// Dialogs, Tabs & Navigation
	mux.Handle("/navigation/modal-dialog", modalDialogHandler())
	mux.Handle("/navigation/confirm-dialog", confirmDialogHandler())
	mux.Handle("/navigation/tabs", tabsHandler())
	mux.Handle("/navigation/spa-navigation", spaNavigationHandler())
	mux.Handle("/navigation/keyboard-shortcuts", keyboardShortcutsHandler())

	// Visual Feedback
	mux.Handle("/feedback/animations", animationsHandler())
	mux.Handle("/feedback/loading-states", loadingStatesHandler())
	mux.Handle("/feedback/highlight", highlightHandler())
	mux.Handle("/feedback/flash-messages", flashMessagesHandler())

	// Real-Time & Multi-User
	mux.Handle("/realtime/multi-user-sync", multiUserSyncHandler())
	mux.Handle("/realtime/broadcasting", broadcastingHandler())
	mux.Handle("/realtime/presence", presenceHandler())
	mux.Handle("/realtime/reconnection", reconnectionHandler())
	mux.Handle("/realtime/live-preview", livePreviewHandler())
	mux.Handle("/realtime/server-push", serverPushHandler())

	// JSON catalog index — same shape as upstream examples/patterns
	// served (kept for forward-compat consumers; unused in B1).
	mux.Handle("/api/index.json", apiIndexHandler())

	// Dev-mode client + CSS, served from the inner mount so templates
	// render {{basePath}}/livetemplate-client.js. lvt/testing fetches
	// these from CDN once and caches in-memory.
	mux.HandleFunc("/livetemplate-client.js", e2etest.ServeClientLibrary)
	mux.HandleFunc("/livetemplate.css", e2etest.ServeCSS)

	return mux
}

// indexHandler renders the catalog landing page that lists all 33
// patterns grouped by category. Templates use {{$.BasePath}}-style
// access to the categories slice via state.
func indexHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/index.tmpl")
	return tmpl.Handle(&IndexController{}, livetemplate.AsState(&IndexState{
		Categories: allPatterns(),
	}))
}

// IndexController serves the pattern catalog index page.
type IndexController struct{}

// IndexState holds the categorized pattern list for the index page.
type IndexState struct {
	Title      string
	Category   string
	Categories []PatternCategory
}
