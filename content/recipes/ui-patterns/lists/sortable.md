---
title: "Sortable List"
description: "Drag-and-drop reordering with native HTML5 drag events."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/lists/sortable.tmpl"
---

# Sortable List

Drag a row onto another to reorder. Each `<li>` is `draggable` and binds native HTML5
`dragstart`/`dragover`/`drop` events with `lvt-on:`; on drop the client sends the
source and target `data-key`s as `dragSourceKey`/`dragTargetKey`, and the `Reorder`
handler moves the item in a process-wide slice. The new order persists across
reloads, and Reset Order restores the original sequence.

```embed-lvt path="/apps/ui-patterns/lists/sortable" upstream="http://localhost:9091" height="360px"
```

## Template

`draggable="true"` plus the three `lvt-on:` drag bindings make each item a drop
target; `data-key` is what the handler reads to know what moved where.

```html include="/examples/patterns/templates/lists/sortable.tmpl"
```

## Handler & state

`Reorder` reads `dragSourceKey`/`dragTargetKey`, finds both items under a mutex, and
splices the source into the target's position; it always re-reads the shared slice so
the order stays authoritative. `Reset` restores the initial items.

```go include="/examples/patterns/handlers_lists.go" region="sortable"
```

```go include="/examples/patterns/state_lists.go" region="sortable-state"
```

## When to use

- Reorderable lists — task priorities, playlists, kanban-style ordering — where the
  order is the data.
- When you want native drag-and-drop without a drag library; the framework forwards
  the source/target keys for you.

For removing rather than reordering rows, see
[Delete Row](/recipes/ui-patterns/lists/delete-row).
