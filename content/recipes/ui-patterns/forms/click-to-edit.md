---
title: "Click to Edit"
description: "Swap a record between a read-only view and an inline edit form — all server-rendered, no client state."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/forms/click-to-edit.tmpl"
---

# Click to Edit

Show a record read-only, then swap the same region for an edit form on demand — no
client state, no separate edit route. An `Editing` boolean on the server state picks
the branch the template renders; **Save** commits the fields and flips it back,
**Cancel** flips it back untouched.

```embed-lvt path="/apps/ui-patterns/forms/click-to-edit" upstream="http://localhost:9091" height="340px"
```

## Template

The whole pattern is one `{{if .Editing}}` branch — view mode renders a table plus an
Edit button; edit mode renders the form with Save/Cancel. Each button's `name` is the
action it triggers.

```html include="/examples/patterns/templates/forms/click-to-edit.tmpl"
```

## Handler & state

Three tiny actions toggle `Editing`; `Save` also reads the submitted fields.

```go include="/examples/patterns/handlers_forms.go" lines="12-41"
```

```go include="/examples/patterns/state_forms.go" lines="4-11"
```

## When to use

- A record that is read most of the time and edited occasionally — keeps the page
  calm and avoids a separate edit screen.
- The edit affordance and the data share one region, so the swap is local.

Reach for [Edit Row](/recipes/ui-patterns/forms/edit-row) instead when many rows in a
table each need independent inline editing.
