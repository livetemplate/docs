---
title: "Reset User Input"
description: "Clear a form automatically after a successful submit — no extra attributes."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/forms/reset-input.tmpl"
---

# Reset User Input

Submit a message and the input clears itself, ready for the next one. There's nothing
special to wire: the server re-renders the form from fresh state after the action, so
the field comes back empty while the submitted values are kept in the list.

```embed-lvt path="/apps/ui-patterns/forms/reset-input" upstream="http://localhost:9091" height="300px"
```

## Template

A plain form plus a list of submitted messages. No reset attribute, no client JS.

```html include="/examples/patterns/templates/forms/reset-input.tmpl"
```

## Handler & state

`Submit` appends the message; because the input isn't bound to retained state, the
re-render starts it empty.

```go include="/examples/patterns/handlers_forms.go" region="reset-input"
```

```go include="/examples/patterns/state_forms.go" region="reset-input-state"
```

## When to use

- Append-style inputs (chat, comments, tags) where the field should reset after each
  successful submit.
- To *keep* a file input across re-renders instead, see
  [Preserving File Inputs](/recipes/ui-patterns/forms/preserve-inputs).
