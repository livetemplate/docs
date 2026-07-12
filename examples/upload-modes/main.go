// Command upload-modes demonstrates LiveTemplate's four upload modes, each
// selected purely by server config on an otherwise identical <input lvt-upload>:
//
//   - Direct:  browser uploads straight to storage via a presigned URL.
//   - Proxied: bytes stream through the server to storage with zero local disk.
//   - Volume:  bytes are staged to a retained directory on the server.
//   - Preview: the file stays on the device; only metadata reaches the server.
package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/livetemplate/livetemplate"
)

const (
	proxiedDir = "storage/proxied"
	volumeDir  = "storage/volume"
	sinkDir    = "storage/direct" // where the Direct presigned PUT lands
)

// UploadModesState is pure data, cloned per session.
type UploadModesState struct {
	DirectRef   string `lvt:"persist"`
	ProxiedRef  string `lvt:"persist"`
	VolumePath  string `lvt:"persist"`
	PreviewName string `lvt:"persist"`
}

// UploadModesController is a singleton holding dependencies. baseURL is the
// absolute base for presigned Direct URLs; it is read lazily by the presigner so
// tests can set it after the server's address is known.
type UploadModesController struct {
	baseURL string
}

// OnUpload implements livetemplate.UploadStreamer for the Proxied field: it
// streams the in-flight bytes straight to storage without local-disk staging by
// the framework, then records the resulting reference. The record id arrives as
// a form field ordered before the file part, so it is readable here mid-stream —
// letting the handler route each upload to its record's folder.
func (c *UploadModesController) OnUpload(part *livetemplate.UploadPart, ctx *livetemplate.Context) error {
	recordID := filepath.Base(ctx.GetString("record_id"))
	if recordID == "" || recordID == "." {
		recordID = "unfiled"
	}
	dir := filepath.Join(proxiedDir, recordID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(dir, filepath.Base(part.Filename))
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(f, part); err != nil {
		return err
	}
	part.SetResult("/files/proxied/" + recordID + "/" + filepath.Base(part.Filename))
	return nil
}

// presigner returns a presigned PUT URL pointing at this server's /sink route,
// making the Direct demo self-contained (no external cloud needed). It reads the
// base URL lazily from the controller so the address can be set after startup.
type presigner struct{ ctrl *UploadModesController }

func (p *presigner) Presign(entry *livetemplate.UploadEntry) (livetemplate.UploadMeta, error) {
	return livetemplate.UploadMeta{
		Uploader: "s3",
		URL:      p.ctrl.baseURL + "/sink/" + filepath.Base(entry.ClientName),
		Headers:  map[string]string{"Content-Type": entry.ClientType},
	}, nil
}

// UploadDirectComplete reads the presigned reference after the browser PUTs.
func (c *UploadModesController) UploadDirectComplete(state UploadModesState, ctx *livetemplate.Context) (UploadModesState, error) {
	if ups := ctx.GetCompletedUploads("direct"); len(ups) > 0 {
		state.DirectRef = "/files/direct/" + filepath.Base(ups[0].ClientName)
	}
	return state, nil
}

// UploadProxiedComplete reads the reference OnUpload recorded via SetResult.
func (c *UploadModesController) UploadProxiedComplete(state UploadModesState, ctx *livetemplate.Context) (UploadModesState, error) {
	if ups := ctx.GetCompletedUploads("proxied"); len(ups) > 0 {
		state.ProxiedRef = ups[0].ExternalRef
	}
	return state, nil
}

// UploadVolumeComplete reads the retained on-disk path.
func (c *UploadModesController) UploadVolumeComplete(state UploadModesState, ctx *livetemplate.Context) (UploadModesState, error) {
	if ups := ctx.GetCompletedUploads("volume"); len(ups) > 0 {
		state.VolumePath = ups[0].TempPath
	}
	return state, nil
}

// newApp builds the example's HTTP handler. The controller's baseURL may be set
// after the server starts (the presigner reads it lazily), which lets tests use
// an httptest server with a dynamic address.
func newApp(ctrl *UploadModesController) http.Handler {
	tmpl := livetemplate.Must(livetemplate.New("upload-modes",
		livetemplate.WithParseFiles("upload-modes.tmpl"),
		livetemplate.WithDevMode(true),
		// One field per mode — identical markup, server config picks the mode.
		livetemplate.WithUpload("direct", livetemplate.UploadConfig{
			Mode:        livetemplate.UploadModeDirect,
			AutoUpload:  true,
			External:    &presigner{ctrl: ctrl},
			Accept:      []string{"image/*"},
			MaxFileSize: 10 << 20,
		}),
		livetemplate.WithUpload("proxied", livetemplate.UploadConfig{
			Mode:        livetemplate.UploadModeProxied,
			AutoUpload:  true,
			Accept:      []string{"image/*"},
			MaxFileSize: 10 << 20,
		}),
		livetemplate.WithUpload("volume", livetemplate.UploadConfig{
			Mode:        livetemplate.UploadModeVolume,
			AutoUpload:  true,
			Dir:         volumeDir,
			Accept:      []string{"image/*"},
			MaxFileSize: 10 << 20,
		}),
		livetemplate.WithUpload("preview", livetemplate.UploadConfig{
			Mode:       livetemplate.UploadModePreview,
			AutoUpload: true,
			Accept:     []string{"image/*"},
		}),
	))

	// When LVT_LOCAL_CLIENT points at a locally-built client bundle (for
	// developing against unreleased client changes), repoint the framework's
	// lvtClientScriptURL func at a same-origin route serving it; otherwise the
	// template renders the pinned CDN bundle. Funcs merge by name, so this wins.
	localClient := os.Getenv("LVT_LOCAL_CLIENT")
	if localClient != "" {
		tmpl.Funcs(map[string]any{
			"lvtClientScriptURL": func() string { return "/livetemplate-client.js" },
		})
	}

	handler := tmpl.Handle(ctrl, livetemplate.AsState(&UploadModesState{}))

	mux := http.NewServeMux()

	// Self-contained Direct sink: accept the presigned PUT and store the bytes.
	mux.HandleFunc("/sink/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := filepath.Base(r.URL.Path)
		if err := os.MkdirAll(sinkDir, 0o755); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		f, err := os.Create(filepath.Join(sinkDir, name))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = f.Close() }()
		if _, err := io.Copy(f, r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Serve stored files back for display.
	mux.Handle("/files/direct/", http.StripPrefix("/files/direct/", http.FileServer(http.Dir(sinkDir))))
	mux.Handle("/files/proxied/", http.StripPrefix("/files/proxied/", http.FileServer(http.Dir(proxiedDir))))

	// Client library: only served locally when developing against an unreleased
	// bundle (LVT_LOCAL_CLIENT, wired to the func override above). Production
	// renders the pinned CDN URL via the client-asset template funcs.
	if localClient != "" {
		mux.HandleFunc("/livetemplate-client.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			http.ServeFile(w, r, localClient)
		})
	}

	mux.Handle("/", handler)
	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8087"
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	app := newApp(&UploadModesController{baseURL: baseURL})
	log.Printf("upload-modes example running at %s", baseURL)
	if err := http.ListenAndServe(":"+port, app); err != nil {
		log.Fatal(err)
	}
}
