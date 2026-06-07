---
title: "Delete Row"
description: "Animated row removal."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/lists/delete-row.tmpl"
---

# Delete Row

Each row carries a Delete button whose `value` is its ID; the handler removes that
item from a process-wide in-memory table and re-renders. New rows slide in via
`lvt-fx:animate="slide"`, and the stable `data-key` on every `<tr>` lets the diff
engine drop only the deleted row instead of redrawing the table. Deletions persist
across reloads, with a Restore button to refill the demo.

```embed-lvt path="/apps/ui-patterns/lists/delete-row" upstream="http://localhost:9091" height="380px"
```

## Template

The `data-key` keys each row for the diff engine, and `lvt-fx:animate="slide"` gives
new rows an entry animation. When the list empties, a Restore button replaces it.

```html include="/examples/patterns/templates/lists/delete-row.tmpl"
```

## Handler & state

`Delete` reads the clicked button's `value`, removes that item under a mutex, and
copies a fresh snapshot back into session state; `Restore` refills the table.

```go include="/examples/patterns/handlers_lists.go" region="delete-row"
```

```go include="/examples/patterns/state_lists.go" region="delete-row-state"
```

## When to use

- Lists where rows are removed one at a time and the rest should stay put — the
  keyed diff keeps every surviving row's DOM intact.
- When deletions need to outlive a reload — back the list with a shared store rather
  than per-session state.

See [Large Table](/recipes/ui-patterns/lists/large-table) for delete alongside
filter, sort, and append on a 10k-row dataset.
