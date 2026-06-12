# upload-autoupload

Minimal reproduction and regression guard for [livetemplate#453](https://github.com/livetemplate/livetemplate/issues/453):
an `lvt-upload` file input with `AutoUpload: true` that is **server-rendered in
the initial HTML** and re-rendered identically on WebSocket connect
(hydrate-idempotent) must upload as soon as a file is selected.

Before the client fix, the per-input `change` handler was bound only from the
post-render block in `updateDOM`, which is skipped when a render adds no nodes.
A hydrate-idempotent first render therefore left the SSR'd input without a
listener, so selecting a file did nothing — no `upload_start`, no upload. The
fix binds SSR'd file inputs once in `connect()` (the same place document-level
event delegation is wired), independent of later renders.

## Run it

```bash
go run .
# open http://localhost:8080 and pick an image — it uploads on select
```

In dev mode the app serves the browser client from `/livetemplate-client.js`.
By default that comes from the published CDN bundle; set `LVT_LOCAL_CLIENT_JS`
to a locally built bundle to serve that instead (used by the e2e below):

```bash
LVT_LOCAL_CLIENT_JS=/path/to/client/dist/livetemplate-client.browser.js go run .
```

## E2E regression test

`upload-autoupload_test.go` drives a real browser (chromedp + Docker Chrome):
it loads the page, selects a file, and asserts an `upload_start` WebSocket frame
is sent — the precise #453 signal that the change handler was bound. It is gated
on `LVT_LOCAL_CLIENT_JS` so it runs against the exact bundle under test (the
published `@latest` predates the fix):

```bash
(cd ../../../client && npm run build)
LVT_LOCAL_CLIENT_JS=$(cd ../../../client && pwd)/dist/livetemplate-client.browser.js \
  go test -run TestAutoUploadSSRBindsOnConnect -v ./examples/upload-autoupload/...
```
