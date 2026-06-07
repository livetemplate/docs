---
title: "Animations"
description: "Play a one-shot entry animation on freshly added elements with lvt-fx:animate — fade, slide, or scale, no JavaScript."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/feedback/animations.tmpl"
---

# Animations

Add `lvt-fx:animate="fade|slide|scale"` to an element and it plays its entry effect
exactly once, when it first appears in a render. A `data-key` identity plus an
internal WeakSet guard keep existing rows still on later renders, so only the genuinely
new element animates in. Tune the timing with the `--lvt-animate-duration` CSS variable
(default 500ms).

```embed-lvt path="/apps/ui-patterns/feedback/animations" upstream="http://localhost:9091" height="340px"
```

## Template

A `<select>` chooses the effect and each list item carries `lvt-fx:animate="{{.Mode}}"`
with a stable `data-key` so the framework can tell new rows from existing ones.

```html include="/examples/patterns/templates/feedback/animations.tmpl"
```

## Handler & state

`Add` records the chosen mode and appends one item; the new element is what animates.

```go include="/examples/patterns/handlers_feedback.go" region="animations"
```

```go include="/examples/patterns/state_feedback.go" region="animations-state"
```

## When to use

- Drawing the eye to a newly inserted row, card, or notification without writing any
  animation JavaScript.
- One-shot entry effects where re-animating unchanged elements would be distracting.

For a transient flash on elements that *change* rather than *appear*, reach for
[Highlight on Change](/recipes/ui-patterns/feedback/highlight) instead.
