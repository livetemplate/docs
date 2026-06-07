---
title: "Flash Messages"
description: "Toast-style notifications set on the server with ctx.SetFlash and rendered as accessible output tags."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/feedback/flash-messages.tmpl"
---

# Flash Messages

Set a message server-side with `ctx.SetFlash("key", text)` and render it with
`{{.lvt.FlashTag "key"}}`, which emits an accessible `<output>` tag. Over a WebSocket
connection a flash persists across renders until you `ctx.ClearFlash(key)` it or it
auto-prunes via `FlashExpiry(d)`; over plain HTTP it is one-shot. This example shows a
required-field error that stays put, a success toast that expires after 5s, and a notice
that lingers until dismissed.

```embed-lvt path="/apps/ui-patterns/feedback/flash-messages" upstream="http://localhost:9091" height="340px"
```

## Template

Two forms plus three `FlashTag` slots — one each for the success, error, and info keys
the handler sets.

```html include="/examples/patterns/templates/feedback/flash-messages.tmpl"
```

## Handler & state

`Save` sets either an error or a 5s-expiring success; `Notify` / `DismissNotify` set and
clear a persistent info flash. Because `FlashExpiry` is render-driven, a small goroutine
nudges a re-render at the deadline so the expired toast actually leaves the DOM.

```go include="/examples/patterns/handlers_feedback.go" region="flash-messages"
```

```go include="/examples/patterns/state_feedback.go" region="flash-messages-state"
```

## When to use

- Transient confirmations and errors that belong to the response, not to the page's
  persistent state — "Saved", "Name is required", "Copied".
- Notices that should auto-dismiss after a moment, or stick until the user acknowledges
  them.

For the same feedback during a slow action rather than after it, see
[Loading States](/recipes/ui-patterns/feedback/loading-states).
