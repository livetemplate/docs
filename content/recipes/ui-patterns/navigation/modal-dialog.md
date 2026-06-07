---
title: "Modal Dialog"
description: "Open a native dialog element with the Invoker Commands API — no inline JS — and keep it open on validation errors."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/navigation/modal-dialog.tmpl"
---

# Modal Dialog

Open a native `<dialog>` declaratively with the Invoker Commands API
(`command="show-modal"` + `commandfor`) or a plain `#id` hash link — no inline
JavaScript. A valid submit closes the dialog and shows a success flash; an invalid
submit keeps it open with field errors rendered inside, because the form posts
through LiveTemplate and re-renders the same region.

```embed-lvt path="/apps/ui-patterns/navigation/modal-dialog" upstream="http://localhost:9091" height="360px"
```

## Template

The trigger button carries `command`/`commandfor`; the Cancel button closes via
`command="close"`. `BindAndValidate` populates `AriaInvalid` and `ErrorTag` so a
failed save renders errors next to each field without closing the dialog.

```html include="/examples/patterns/templates/navigation/modal-dialog.tmpl"
```

## Handler & state

One `Save` action binds and validates the form; on error it returns the validation
error and the still-open dialog re-renders with the messages.

```go include="/examples/patterns/handlers_navigation.go" region="modal-dialog"
```

```go include="/examples/patterns/state_navigation.go" region="modal-dialog-state"
```

## When to use

- A focused edit or confirmation that should overlay the page without a route change.
- You want the browser's built-in modal semantics (focus trap, backdrop, `Escape`)
  for free, driven by HTML attributes rather than JavaScript.
- Validation should keep the dialog open and surface errors in place.

Reach for [Confirm Dialog](/recipes/ui-patterns/navigation/confirm-dialog) when the
dialog only needs a yes/no confirmation for a destructive action.
