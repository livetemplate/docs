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

## Client-owned or server-owned?

All three tiers here are **client-owned**: the pending state lives in the browser, costs
no Go code, and clears on `done`. Since `@livetemplate/client` v0.18.1 the classes and
attributes applied by `lvt-el:` survive a server re-render — the client re-applies them
after each morph pass — so you do not need `lvt-ignore-attrs` to protect them.

What client-owned pending cannot do is outlive the browser: it does not fan out to peer
tabs, does not survive a reconnect, and cannot drive server-side logic. When the pending
state has to do any of that, move it to the server:

| Need | Reach for |
|---|---|
| Visual feedback only, single action | the three tiers on this page |
| Slow work off the event loop, spinner rendered by the framework | [`Async` + `{{.lvt.Pending}}`](/reference/api#async) — one method, no state field |
| Spinner must survive a reconnect, or fan out to peers | `Async` plus a `Loading bool` field tagged `lvt:"persist"` |
| Work starts in `Mount`/`OnConnect`, or reports progress repeatedly | the manual two-action pattern — `Async` is action-handler-only and one-shot |

The landing page's Step 4 demos the `Async` + `{{.lvt.Pending}}` version. The
[`greet-loading-server`](/apps/greet-loading-server/) example keeps the manual
two-action shape, which is still what you want when the spinner has to be recovered
after a reconnect — that requires a real state field, because `{{.lvt.Pending}}` is
per-render and is false on any render that did not itself start the work.
