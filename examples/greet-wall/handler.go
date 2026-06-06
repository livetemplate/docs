// Package greetwall is the climax of the docs homepage's progressive-narrative
// spine (Steps 5-7) — the greet app grown into a live, shared greeting wall. One
// app layers three real-time capabilities: per-user tab sync (SelfTopic), a
// cross-user shared wall ("wall" topic), and server-initiated push (a ticker
// calling Session.TriggerAction). cmd/site mounts it at /apps/greet-wall/ — the
// one WebSocket-enabled embed on the landing.
package greetwall

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/livetemplate/livetemplate"
)

//go:embed wall.tmpl
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

// extractTemplate writes the embedded template to a temp file so
// livetemplate's file-based loader can parse it at runtime. Done once per
// process (mirrors examples/greet).
func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "greet-wall-tmpl-*")
		if err != nil {
			log.Fatalf("greet-wall: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("wall.tmpl")
		if err != nil {
			log.Fatalf("greet-wall: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "wall.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("greet-wall: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// Handler returns the greet-wall app as an http.Handler ready to mount. The
// WithTopicACL admits only the shared "wall" topic; every other developer
// topic stays denied, while each user's own SelfTopic() is always permitted
// (ACL-exempt). Callers supply environment-specific options (origin
// allowlists) via opts.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
		livetemplate.WithTopicACL(func(topic, _ string, _ *http.Request) (bool, error) {
			return topic == wallTopic, nil
		}),
	}
	tmpl := livetemplate.Must(livetemplate.New("greet-wall", append(baseOpts, opts...)...))
	return tmpl.Handle(newController(), livetemplate.AsState(&State{Name: "there"}))
}
