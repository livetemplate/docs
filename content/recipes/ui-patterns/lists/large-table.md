---
title: "Large Table"
description: "10k-row table with filter, sort, append, update, delete, reset (streaming range)."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/lists/large-table.tmpl"
---

# Large Table

A 10,000-row table that stays interactive through filter, sort, append, per-row
update, delete, and reset — every range op the streaming-range diff can emit, in one
demo. The row dataset lives process-wide on the controller; filter and sort are
per-session view state, so each connection sorts and filters its own view without
touching the shared rows. Stable `data-key`s let the diff engine move, update, or
drop individual rows instead of redrawing the table.

```embed-lvt path="/apps/ui-patterns/lists/large-table" upstream="http://localhost:9091" height="480px"
```

## Template

A debounced `filter` input, sortable column-header buttons (each carrying its key in
`value`), and Append/Update/Reset controls. Every `<tr>` is keyed by `.ID`.

```html include="/examples/patterns/templates/lists/large-table.tmpl"
```

## Handler & state

`applyView` filters then sorts a snapshot purely; `refreshView` recomputes the
displayed slice plus the total. `Change`, `Sort`, `AppendN`, `UpdateRandomRow`,
`Delete`, and `Reset` each mutate state or the shared rows and re-run the view.

```go include="/examples/patterns/handlers_lists.go" region="large-table"
```

```go include="/examples/patterns/state_lists.go" region="large-table-state"
```

## When to use

- Data-grid screens with thousands of rows that need live filter, sort, and edits
  without paging out to a separate route.
- When you want per-session view state (filter/sort) over a shared dataset, so users
  don't fight over each other's view.

For simpler single-row deletion, see
[Delete Row](/recipes/ui-patterns/lists/delete-row); for growing a list on demand,
see [Click To Load](/recipes/ui-patterns/lists/click-to-load).
