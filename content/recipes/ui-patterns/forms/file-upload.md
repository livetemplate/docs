---
title: "File Upload"
description: "Standard multipart upload and chunked upload with live progress."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/forms/file-upload.tmpl"
---

# File Upload

Two tiers from the same handler. **Tier 1** is a plain `multipart/form-data` form —
it just works, no JavaScript. **Tier 2** adds `lvt-upload` to stream the file in
chunks over the WebSocket and render a live `<progress>` bar as it arrives.

```embed-lvt path="/apps/ui-patterns/forms/file-upload" upstream="http://localhost:9091" height="380px"
```

## Template

`{{range .lvt.Uploads "chunked-doc"}}` exposes per-file progress for the chunked input.

```html include="/examples/patterns/templates/forms/file-upload.tmpl"
```

## Handler & state

`WithUpload` declares each named upload (size caps, and a small `ChunkSize` so the
demo's progress is visible); `Upload` flashes the completed file's name. With no
`Mode` set, both fields use the default **Volume** mode — bytes stage on the
server and you read them from `entry.TempPath`.

```go include="/examples/patterns/handlers_forms.go" region="file-upload"
```

```go include="/examples/patterns/state_forms.go" region="file-upload-state"
```

## When to use

- Any file input — start with Tier 1, add `lvt-upload` only when you want progress or
  large-file chunking.
- This recipe stages bytes on the server (Volume mode). To send bytes straight to
  cloud storage, stream them through with zero local disk, or keep them on the
  device, see the [Upload Modes recipe](/recipes/apps/upload-modes).
- See [Uploads](/reference/uploads) for the full upload reference.
