---
title: "Edit Row"
description: "Inline-edit one row of a table in place while the others stay read-only."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/forms/edit-row.tmpl"
---

# Edit Row

A table where any single row flips into an edit form in place. The server tracks one
`EditingID`; the row whose ID matches renders inputs, every other row stays read-only.
Each Edit/Save button carries its row ID in the button `value`, so the handler knows
which record to act on.

```embed-lvt path="/apps/ui-patterns/forms/edit-row" upstream="http://localhost:9091" height="360px"
```

## Template

`{{if eq $.EditingID .ID}}` chooses the edit form vs the read-only cells per row.

```html include="/examples/patterns/templates/forms/edit-row.tmpl"
```

## Handler & state

`Edit` records which row is open (`value` = row ID); `Save` writes the fields back to
that contact and clears `EditingID`.

```go include="/examples/patterns/handlers_forms.go" lines="45-79"
```

```go include="/examples/patterns/state_forms.go" lines="14-19"
```

## When to use

- A list/table where rows are edited in place and you only ever edit one at a time.
- The single `EditingID` keeps it simple — no per-row flags.

For a single standalone record, [Click to Edit](/recipes/ui-patterns/forms/click-to-edit)
is the simpler shape.
