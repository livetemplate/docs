---
title: "Bulk Update"
description: "Toggle many checkboxes and commit them in one submit, with a summary flash."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/forms/bulk-update.tmpl"
---

# Bulk Update

One form, many rows of checkboxes, a single submit. Each checkbox is named
`active-<id>`; the handler walks the rows, applies the new states, counts what
changed, and sets a flash summarizing the result.

```embed-lvt path="/apps/ui-patterns/forms/bulk-update" upstream="http://localhost:9091" height="360px"
```

## Template

The checkbox `name` encodes the row ID, so one form submits every row's state at once.

```html include="/examples/patterns/templates/forms/bulk-update.tmpl"
```

## Handler & state

`BulkUpdate` reads each `active-<id>` with `ctx.GetBool`, applies it, and flashes a
count (or "No changes").

```go include="/examples/patterns/handlers_forms.go" lines="114-142"
```

```go include="/examples/patterns/state_forms.go" lines="31-35"
```

## When to use

- Editing a set of boolean flags across rows where one commit is clearer than
  per-row saves.
- The `name="field-<id>"` convention scales to any number of rows with no extra
  client code.
