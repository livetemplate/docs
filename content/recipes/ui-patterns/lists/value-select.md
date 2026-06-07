---
title: "Value Select"
description: "Cascading dependent selects."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/lists/value-select.tmpl"
---

# Value Select

Picking a Make repopulates the Model select server-side. The `Change` handler
auto-fires when the `make` select changes, looks up that make's models, and
auto-selects the first one so the cascade is visible — no client JS. The Model select
stays `disabled` until a Make is chosen, and `Mount` seeds the make list (and any
pre-selected models) on connect.

```embed-lvt path="/apps/ui-patterns/lists/value-select" upstream="http://localhost:9091" height="340px"
```

## Template

Both `<select>` elements live in one form; `Change` fires on either. The Model
options come from `.Models`, which the server refills whenever `make` changes.

```html include="/examples/patterns/templates/lists/value-select.tmpl"
```

## Handler & state

`Change` checks which field changed: a new `make` reloads `Models` and resets the
selection; a new `model` just records it. `Mount` populates the initial lists.

```go include="/examples/patterns/handlers_lists.go" region="value-select"
```

```go include="/examples/patterns/state_lists.go" region="value-select-state"
```

## When to use

- Dependent dropdowns where one selection determines the options of the next
  (country → state, make → model).
- When the option sets are large or computed server-side and you would rather not
  ship them all to the client up front.

For a non-cascading list that grows on demand, see
[Click To Load](/recipes/ui-patterns/lists/click-to-load).
