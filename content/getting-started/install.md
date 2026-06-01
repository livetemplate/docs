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

That's it for the server side. The TypeScript client (loaded from CDN by default) wires up the browser side automatically — you'll add a single `<script>` tag in your template; no npm step needed for getting started.

## Client CDN versioning

Docs examples may use `@latest` when they are demonstrating the current client quickly. For production apps, prefer a version range that matches your LiveTemplate core release policy, or pin an exact client version when you need reproducible archived behavior. The important part is to keep the server library and browser client on compatible major/minor versions.

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
