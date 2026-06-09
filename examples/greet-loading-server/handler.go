// Package greetloadingserver is the Step 3 variant that keeps loading in
// ordinary server state and clears it via a second server-pushed action.
package greetloadingserver

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/livetemplate/livetemplate"
)

//go:embed greet.tmpl
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "greet-loading-server-tmpl-*")
		if err != nil {
			log.Fatalf("greet-loading-server: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("greet.tmpl")
		if err != nil {
			log.Fatalf("greet-loading-server: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "greet.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("greet-loading-server: write tmpl: %v", err)
		}
	})
	return tmplPath
}

func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
	}
	tmpl := livetemplate.Must(livetemplate.New("greet-loading-server", append(baseOpts, opts...)...))
	return tmpl.Handle(&Controller{}, livetemplate.AsState(&State{Name: "there"}))
}
