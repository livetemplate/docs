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
)

// UploadController is a singleton holding dependencies (none needed here).
type UploadController struct{}

// UploadState is pure data, cloned per session. It carries nothing the upload
// itself doesn't already track via .lvt.Uploads, keeping the render
// hydrate-idempotent.
type UploadState struct{}

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

	// The template renders {{lvtClientScriptURL}} — the pinned CDN bundle — by
	// default. For the #453 regression the browser MUST instead load a locally
	// built bundle (published @latest predates the fix): when LVT_LOCAL_CLIENT_JS
	// points at a built dist, serve it same-origin with no caching and repoint
	// the framework's lvtClientScriptURL func at it (funcs merge by name, so this
	// override wins over the pinned default).
	if bundle := os.Getenv("LVT_LOCAL_CLIENT_JS"); bundle != "" {
		lt.Funcs(map[string]any{
			"lvtClientScriptURL": func() string { return "/livetemplate-client.js" },
		})
		http.HandleFunc("/livetemplate-client.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			w.Header().Set("Cache-Control", "no-store")
			http.ServeFile(w, r, bundle)
		})
	}

	controller := &UploadController{}
	handler := lt.Handle(controller, livetemplate.AsState(&UploadState{}))

	http.Handle("/", handler)

	addr := ":" + port
	log.Printf("upload-autoupload example running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
