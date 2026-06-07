---
title: "Lazy Loading"
description: "Render the page instantly, then let the server push slow content in over the live connection — no client fetch, no loading route."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/loading/lazy-loading.tmpl"
---

# Lazy Loading

Send the shell immediately and fill in the slow part once the data is ready. `Mount`
returns `Loading=true` so the first paint shows a spinner; when the live connection
opens, `OnConnect` spawns a goroutine that simulates a slow API and pushes the payload
back with `session.TriggerAction("dataLoaded", …)`, which clears `Loading` and
re-renders the region with the content.

```embed-lvt path="/apps/ui-patterns/loading/lazy-loading" upstream="http://localhost:9091" height="360px"
```

## Template

One `{{if .Loading}}` branch: an `aria-busy` spinner while the data is in flight, then
the loaded content plus a **Reload** button. The `<noscript>` note is honest about the
trade-off — with JavaScript off there is no WebSocket, so the spinner never resolves.

```html include="/examples/patterns/templates/loading/lazy-loading.tmpl"
```

## Handler & state

`OnConnect` does the lazy fetch off the connect path and pushes the result with
`TriggerAction`; `DataLoaded` writes it into state. It skips re-spawning when the data
has already arrived (e.g. a reconnect), and `Reload` re-runs the same flow on demand.

```go include="/examples/patterns/handlers_loading.go" region="lazy-loading"
```

```go include="/examples/patterns/state_loading.go" region="lazy-loading-state"
```

## When to use

- The page has a fast core and a slow accessory (a feed, a recommendation block) you
  don't want to block the first paint on.
- The work belongs on the server — a slow query or upstream API — and you'd rather
  push the result than expose a separate fetch endpoint.
- The user should see structure immediately and watch the gap fill in.

Reach for [Async Operations](/recipes/ui-patterns/loading/async-operations) instead
when the load is user-triggered and can fail, so you need explicit success and error
states.
