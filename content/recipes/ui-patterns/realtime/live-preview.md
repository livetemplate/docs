---
title: "Live Preview"
description: "Render a real-time preview as the user types, using the auto-bound Change() method."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/realtime/live-preview.tmpl"
---

# Live Preview

When the controller defines a `Change()` method, the framework auto-binds it to
`input`/`change` events on form fields with a 300ms debounce — no extra attribute
needed. `Change` reads `ctx.GetString("input")` and updates `state.Preview` only; it
deliberately never writes back to `state.Input`, because patching the input's `value`
mid-typing would reset the cursor. An explicit `Submit` commits the value.

```embed-lvt path="/apps/ui-patterns/realtime/live-preview" upstream="http://localhost:9091" height="400px"
```

## Template

One input bound to `Change()` and an `<output>` that mirrors the live preview.

```html include="/examples/patterns/templates/realtime/live-preview.tmpl"
```

## Handler & state

`Change` builds the debounced preview; `Submit` commits the value to the persisted
`Input` field.

```go include="/examples/patterns/handlers_realtime.go" region="live-preview"
```

```go include="/examples/patterns/state_realtime.go" region="live-preview-state"
```

## When to use

- A preview of formatted output — Markdown, a slug, a greeting, a computed total —
  that should update as the user types.
- You want the live feedback without committing or persisting on every keystroke.

Reach for [Reconnection Recovery](/recipes/ui-patterns/realtime/reconnection) when
the typed value itself must survive a reload.
