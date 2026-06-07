---
title: "Highlight on Change"
description: "Flash a temporary background on any element whose subtree a render touched, using lvt-fx:highlight."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/feedback/highlight.tmpl"
---

# Highlight on Change

Add `lvt-fx:highlight="flash"` to an element and it briefly flashes a background color
whenever a render updates anything inside it. The effect is per-element, not per-value:
two elements bound to the same counter both flash when it changes. Customize the look
with `--lvt-highlight-duration` (default 500ms) and `--lvt-highlight-color` (default
amber `#ffc107`).

```embed-lvt path="/apps/ui-patterns/feedback/highlight" upstream="http://localhost:9091" height="340px"
```

## Template

One button increments a shared counter; two separate `lvt-fx:highlight="flash"` blocks
mirror it, so both flash on every change.

```html include="/examples/patterns/templates/feedback/highlight.tmpl"
```

## Handler & state

`Increment` bumps a single counter — the highlight is entirely client-side, triggered by
the resulting DOM update.

```go include="/examples/patterns/handlers_feedback.go" region="highlight"
```

```go include="/examples/patterns/state_feedback.go" region="highlight-state"
```

## When to use

- Drawing attention to values that just changed — prices, counts, statuses — without
  managing any timers or classes yourself.
- Cross-user or background updates where the viewer didn't trigger the change and needs a
  cue that something moved.

For a one-shot effect on elements that newly *appear* rather than *change*, see
[Animations](/recipes/ui-patterns/feedback/animations).
