---
title: "Loading States"
description: "Show that a slow action is in flight — automatic aria-busy, custom button text, or a reactive attribute toggle."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/feedback/loading-states.tmpl"
---

# Loading States

While an action is in flight the framework automatically marks the submitting form
`aria-busy="true"` and disables its `<fieldset>` — no directives required. When you want
more, `lvt-form:disable-with` swaps the button's text for the pending duration, and
`lvt-el:setAttr:on:pending` / `:on:done` toggle any attribute reactively across the
action lifecycle. All three tiers here call the same 2-second `slowSave` action.

```embed-lvt path="/apps/ui-patterns/feedback/loading-states" upstream="http://localhost:9091" height="340px"
```

## Template

Three forms, one action — each demonstrates a different feedback tier, from
zero-directive automatic busy state up to a reactive attribute toggle.

```html include="/examples/patterns/templates/feedback/loading-states.tmpl"
```

## Handler & state

`SlowSave` sleeps two seconds to make the pending window visible, then stamps the time.

```go include="/examples/patterns/handlers_feedback.go" region="loading-states"
```

```go include="/examples/patterns/state_feedback.go" region="loading-states-state"
```

## When to use

- Any action with perceptible latency — saving, uploading, calling a slow upstream.
- When you want accessible busy feedback (`aria-busy`) for free and richer cues only
  where they earn their keep.

To run that slow work without blocking the rest of the page, see
[Async Operations](/recipes/ui-patterns/loading/async-operations).
