---
title: "Install"
description: "Add LiveTemplate to a Go project — one go get and you have what you need."
---

# Install

LiveTemplate is a Go framework. The minimum to get a reactive page running is one `go get`.

## Add the framework

```bash
go get github.com/livetemplate/livetemplate
```

That's it for the server side. The TypeScript client wires up the browser side automatically — you add two tags to your template's `<head>`/`<body>`; no npm step needed for getting started.

## Loading the browser client

Add these to your template. The `lvtClientScriptURL` / `lvtClientStyleURL` functions are provided by the framework on every template:

```html
<link rel="stylesheet" href="{{lvtClientStyleURL}}">
<script defer src="{{lvtClientScriptURL}}"></script>
```

They render the CDN URL for the client bundle **this server release is wire-compatible with** — pinned, not `@latest`. This matters: there is no runtime handshake between server and client, so an unpinned client can ship a wire-protocol change to browsers still talking to an older server. The pinned version moves only when you upgrade the framework (`go get -u github.com/livetemplate/livetemplate`), keeping the two in lockstep. Every example in these docs uses this pattern.

The URLs are also exported as constants — `livetemplate.ClientVersion`, `livetemplate.ClientScriptURL`, `livetemplate.ClientStyleURL` — if you need them outside a template.

### Self-hosting (offline, air-gapped, or CSP-strict)

If your deployment can't reach a public CDN — or a Content-Security-Policy forbids third-party script origins — vendor the bundle at the pinned version and serve it from your own origin:

```bash
npm install @livetemplate/client@<version>   # match livetemplate.ClientVersion
```

Serve the vendored files same-origin, then either write your own tags pointing at them or repoint the framework functions with a `Funcs` override (they merge by name, so yours wins) — no need to edit every template:

```go
tmpl.Funcs(template.FuncMap{
    "lvtClientScriptURL": func() string { return "/assets/livetemplate-client.browser.js" },
    "lvtClientStyleURL":  func() string { return "/assets/livetemplate.css" },
})
```

## Optional: install the `lvt` CLI

The `lvt` CLI generates LiveTemplate apps with database scaffolding, auth, and a router pre-wired. Optional — you can write apps directly in Go without it.

```bash
go install github.com/livetemplate/lvt/cmd/lvt@latest
lvt new myapp
cd myapp && lvt serve
```

See the [CLI guide](/cli) for the full command reference.

## What next?

- **[Your First App](/getting-started/your-first-app)** — build a counter from scratch in 10 minutes. Walks through the Tier 1 → JS client → multi-tab sync progression.
- [Progressive Complexity](/guides/progressive-complexity) — the framework's two-tier model (start with standard HTML, layer on `lvt-*` attributes only when needed).
- [How a LiveTemplate Update Flows](/recipes/architecture-flow) — interactive walkthrough of what happens between a click and a DOM patch.
- [Server API reference](/reference/api) — `New`, `Handle`, `Context`, action method dispatch.
