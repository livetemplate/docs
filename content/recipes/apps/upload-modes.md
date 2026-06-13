---
title: "Upload Modes"
description: "One lvt-upload input, four destinations — Volume, Direct, Proxied, and Preview — chosen entirely by server config."
source_repo: "https://github.com/livetemplate/docs"
source_path: "content/recipes/apps/upload-modes.md"
---

# Upload Modes — one input, four destinations

A file upload has to decide *where the bytes go*: staged on the server, sent
straight to cloud storage, streamed through the server with nothing kept on
disk, or never uploaded at all. LiveTemplate makes that a **server config**
choice, not a markup or client-code choice. The HTML is the same plain
`<input lvt-upload>` in every case; only `UploadConfig.Mode` differs. The full
source is
[`examples/upload-modes/`](https://github.com/livetemplate/docs/tree/main/examples/upload-modes).

## The four modes

| Mode | Bytes path | Server sees bytes? | Local disk? |
|------|------------|--------------------|-------------|
| **Volume** *(default)* | browser → server → retained directory | yes | yes |
| **Direct** | browser → storage via presigned URL | no | no |
| **Proxied** | browser → server → storage (streamed) | yes | **no** |
| **Preview** | stays on the device | metadata only | no |

When `Mode` is omitted it defaults to **Volume** (server-side staging). For
backward compatibility, a config that sets `External` without an explicit
`Mode` is treated as **Direct**.

## One declaration per mode

Each field is the same `WithUpload` call with a different `Mode` — the example
wires all four on one template:

```go
livetemplate.WithUpload("volume",  livetemplate.UploadConfig{Mode: livetemplate.UploadModeVolume,  Dir: "storage/volume"}),
livetemplate.WithUpload("direct",  livetemplate.UploadConfig{Mode: livetemplate.UploadModeDirect,  External: presigner}),
livetemplate.WithUpload("proxied", livetemplate.UploadConfig{Mode: livetemplate.UploadModeProxied}), // controller implements OnUpload
livetemplate.WithUpload("preview", livetemplate.UploadConfig{Mode: livetemplate.UploadModePreview}),
```

The markup is identical across all four — the mode is invisible to the
template:

```html
<input type="file" lvt-upload="volume"  accept="image/*" />
<input type="file" lvt-upload="direct"  accept="image/*" />
<input type="file" lvt-upload="proxied" accept="image/*" />
<input type="file" lvt-upload="preview" accept="image/*" />
```

And consumption is uniform too — every mode surfaces its result through
`ctx.GetCompletedUploads(name)`, whatever path the bytes took.

## Volume — staged on the server

The default. Bytes land on the server's disk; with `Dir` set they are
**retained** there and your app owns the path (read it from
`entry.TempPath`). This is the [Avatar Upload](/recipes/apps/avatar-upload)
recipe's mode. Use it when the server needs to see and keep the bytes.

## Direct — browser uploads straight to storage

With an `External` presigner, the browser PUTs bytes straight to S3/GCS/etc.
via a presigned URL — they never touch the server. Read the stored reference
from `entry.ExternalRef`. To keep the example self-contained, its presigner
points at the server's own `/sink` route, so no real cloud is needed.

## Proxied — stream through the server, zero local disk

`UploadModeProxied` streams the in-flight bytes straight to a handler with no
local-disk staging — ideal for forwarding to remote object storage. The
controller implements `UploadStreamer`:

```go
func (c *Controller) OnUpload(part *livetemplate.UploadPart, ctx *livetemplate.Context) error {
    recordID := filepath.Base(ctx.GetString("record_id")) // a field ordered before the file input
    dst := filepath.Join("storage/proxied", recordID, filepath.Base(part.Filename))
    // ... os.MkdirAll + os.Create ...
    if _, err := io.Copy(f, part); err != nil {
        return err
    }
    part.SetResult("/files/proxied/" + recordID + "/" + filepath.Base(part.Filename))
    return nil
}
```

Because multipart parts stream in body order, a form field is readable
mid-stream via `ctx.GetString` **only if its input precedes the file input** —
which is how the example routes each upload to its record's folder. The result
recorded with `part.SetResult` is later read back from `entry.ExternalRef`.

## Preview — the file never leaves the device

`UploadModePreview` keeps the file in the browser; only its metadata
(name/type/size) reaches the server. Render the on-device preview with a
template helper:

```html
<input type="file" lvt-upload="preview" accept="image/*" />
{{.lvt.UploadPreview "preview"}}
```

The client fills the placeholder from a local `URL.createObjectURL` and never
uploads the bytes. The server records a metadata-only entry
(`entry.Preview == true`, no `TempPath` / `ExternalRef`).

## Works with the WebSocket disabled

Every mode completes over plain HTTP when the socket is down. **Volume** falls
back to a single multipart POST that the server stages to `Dir`
([#449](https://github.com/livetemplate/livetemplate/issues/449)); **Direct**
presigns over HTTP, the browser PUTs, then the client re-sends the entry metadata
over an HTTP completion handshake so `upload_<field>_complete` still runs
([#448](https://github.com/livetemplate/livetemplate/issues/448)). **Proxied**
and **Preview** are single requests and were already WS-independent. No app code
changes — the same controller works on either transport.

## Run it

```bash
cd examples/upload-modes
GOWORK=off ./run.sh
```

Open <http://localhost:8087> and upload into each of the four cards in turn —
each one stores (or previews) the file a different way while the markup stays
identical. The
[end-to-end test](https://github.com/livetemplate/docs/blob/main/examples/upload-modes/upload-modes_test.go)
drives all four modes in a real browser and asserts the Proxied upload stages
**zero** files on local disk.

## See also

- [Upload reference](/reference/uploads#upload-modes) — the full mode matrix, config, and helpers.
- [Avatar Upload recipe](/recipes/apps/avatar-upload) — Volume mode end to end.
- [File Upload pattern](/recipes/ui-patterns/forms/file-upload) — the minimal two-tier form.
