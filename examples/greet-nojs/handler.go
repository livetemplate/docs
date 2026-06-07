package greetnojs

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/livetemplate/livetemplate"
)

//go:embed nojs.tmpl
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

// extractTemplate writes the embedded template to a temp file so livetemplate's
// file-based loader can parse it at runtime (mirrors examples/greet).
func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "greet-nojs-tmpl-*")
		if err != nil {
			log.Fatalf("greet-nojs: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("nojs.tmpl")
		if err != nil {
			log.Fatalf("greet-nojs: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "nojs.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("greet-nojs: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the no-JS greet app as an http.Handler ready to mount.
// AnonymousAuthenticator gives each browser its own session group (the
// livetemplate-id cookie), which keys the per-session name store. cmd/site
// mounts this with WithWebSocketDisabled() so the live demo proves the plain
// HTTP form-POST path.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("greet-nojs", append(baseOpts, opts...)...))
	return tmpl.Handle(newController(), livetemplate.AsState(&State{Name: "there"}))
}
