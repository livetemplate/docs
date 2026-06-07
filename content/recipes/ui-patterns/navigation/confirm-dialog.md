---
title: "Confirm Dialog"
description: "Gate a destructive row action behind a native dialog element confirmation — CSP-friendly, no inline JS."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/navigation/confirm-dialog.tmpl"
---

# Confirm Dialog

Gate a destructive action behind a per-row native `<dialog>` confirmation with no
inline JavaScript, so it works under a strict Content-Security-Policy. Each row owns
its own dialog id (`confirm-{{.ID}}`), opened with `command="show-modal"`; confirming
posts the clicked button's `value` (the item id) to a `Delete` action that removes the
item and re-renders the table.

```embed-lvt path="/apps/ui-patterns/navigation/confirm-dialog" upstream="http://localhost:9091" height="360px"
```

## Template

Dialogs render outside the `<table>` (a `<dialog>` inside `<tbody>` is invalid HTML).
The Delete button carries `value="{{.ID}}"` so the server learns which row to remove
without a hidden input.

```html include="/examples/patterns/templates/navigation/confirm-dialog.tmpl"
```

## Handler & state

`Delete` reads the clicked button's `value` and drops that item; unknown ids are a
tolerated no-op so client and server lists reconcile silently.

```go include="/examples/patterns/handlers_navigation.go" region="confirm-dialog"
```

```go include="/examples/patterns/state_navigation.go" region="confirm-dialog-state"
```

## When to use

- A destructive, irreversible action (delete, archive) that warrants an explicit
  "are you sure?" step.
- You need confirmations under a strict CSP where inline `onclick` handlers are banned.
- Each row needs its own confirmation that can even be deep-linked by hash.

Reach for [Modal Dialog](/recipes/ui-patterns/navigation/modal-dialog) when the
overlay holds a full form rather than a single confirm/cancel choice.
