// Command upload-autoupload is the minimal reproduction for livetemplate
// issue #453: a file input with lvt-upload + AutoUpload that is server-rendered
// in the initial HTML and re-rendered identically on WebSocket connect
// (hydrate-idempotent). Before the client fix, the per-input change handler was
// bound only from updateDOM's post-render block, which is skipped when a render
// adds no nodes — so selecting a file did nothing. This app renders the input
// from Mount on every request, so its first WebSocket render matches the SSR
// DOM exactly, exercising that path.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/livetemplate/livetemplate"
	lvttest "github.com/livetemplate/lvt/testing"
)

// UploadController is a singleton holding dependencies (none needed here).
type UploadController struct{}

// UploadState is pure data, cloned per session. It carries nothing the upload
// itself doesn't already track via .lvt.Uploads, keeping the render
// hydrate-idempotent.
type UploadState struct{}

// serveClientJS serves the LiveTemplate browser client. For the #453 regression
// the browser MUST load a locally built bundle (the published @latest predates
// the fix), so when LVT_LOCAL_CLIENT_JS points at a built dist we serve that
// file directly with no caching. Serving it ourselves — rather than via
// LVT_CLIENT_CDN_URL — sidesteps lvttest's filename-keyed disk cache, which can
// otherwise return a stale CDN copy. Without the env var we fall back to the
// CDN-backed helper so the example still runs standalone.
func serveClientJS(w http.ResponseWriter, r *http.Request) {
	if path := os.Getenv("LVT_LOCAL_CLIENT_JS"); path != "" {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, path)
		return
	}
	lvttest.ServeClientLibrary(w, r)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	lt := livetemplate.Must(livetemplate.New("upload-autoupload",
		livetemplate.WithParseFiles("upload-autoupload.tmpl"),
		livetemplate.WithDevMode(true),
		livetemplate.WithUpload("avatar", livetemplate.UploadConfig{
			Accept:      []string{"image/jpeg", "image/png", "image/gif"},
			MaxFileSize: 5 * 1024 * 1024, // 5MB
			MaxEntries:  1,
			AutoUpload:  true, // upload starts on file select — needs the bound change handler
		}),
	))

	controller := &UploadController{}
	handler := lt.Handle(controller, livetemplate.AsState(&UploadState{}))

	http.HandleFunc("/livetemplate-client.js", serveClientJS)
	http.HandleFunc("/livetemplate.css", lvttest.ServeCSS)
	http.Handle("/", handler)

	addr := ":" + port
	log.Printf("upload-autoupload example running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
