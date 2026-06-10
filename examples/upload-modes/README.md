# Upload Modes

One `<input lvt-upload>`, four destinations — chosen entirely by server config.
The HTML and app code are identical across modes; only `UploadConfig.Mode` differs.

| Mode | Bytes path | Server sees bytes? | Local disk? |
|------|------------|--------------------|-------------|
| **Volume** *(default)* | browser → server → retained directory | yes | yes |
| **Direct** | browser → storage via presigned URL | no | no |
| **Proxied** | browser → server → storage (streamed) | yes | **no** |
| **Preview** | stays on the device | metadata only | no |

When `Mode` is omitted it defaults to **Volume** (server-side staging). For
backward compatibility, a config that sets `External` without an explicit `Mode`
is treated as **Direct**.

```go
livetemplate.WithUpload("direct",  livetemplate.UploadConfig{Mode: livetemplate.UploadModeDirect,  External: presigner})
livetemplate.WithUpload("proxied", livetemplate.UploadConfig{Mode: livetemplate.UploadModeProxied})  // controller implements OnUpload
livetemplate.WithUpload("volume",  livetemplate.UploadConfig{Mode: livetemplate.UploadModeVolume, Dir: "storage/volume"})
livetemplate.WithUpload("preview", livetemplate.UploadConfig{Mode: livetemplate.UploadModePreview})
```

- **Proxied** streams the in-flight bytes straight to the controller's
  `OnUpload(part *livetemplate.UploadPart, ctx)` — zero local-disk staging — and
  records the result with `part.SetResult(ref)`.
- **Direct** is self-contained here: the presigner points at this server's own
  `/sink` route so no external cloud is needed.
- **Preview** uses the `{{.lvt.UploadPreview "preview"}}` helper; the client fills
  it from a local `URL.createObjectURL` and never uploads the bytes.

## Run

```bash
./run.sh                 # http://localhost:8087
```

The example serves the published client by default. To run against an unreleased
client build, point `LVT_LOCAL_CLIENT` at a bundle:

```bash
LVT_LOCAL_CLIENT=/path/to/livetemplate-client.browser.js PORT=8087 go run main.go
```

## Browser test

```bash
LVT_UPLOAD_MODES_E2E=1 LVT_LOCAL_CLIENT=/path/to/livetemplate-client.browser.js \
  go test ./examples/upload-modes/ -run E2E -v
```

Drives all four modes with a locally-installed Chromium and asserts the Proxied
upload stages **zero** files on local disk. Gated on `LVT_UPLOAD_MODES_E2E` so it
skips in the cross-repo CI (which uses a Docker-chrome remote allocator);
integrating it into that harness is tracked in livetemplate/docs#67.
