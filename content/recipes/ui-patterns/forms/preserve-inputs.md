---
title: "Preserving File Inputs"
description: "Keep a file selection (and other inputs) across server re-renders with lvt-form:preserve."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/forms/preserve-inputs.tmpl"
---

# Preserving File Inputs

A file input can't be re-populated by the server for security reasons, so a naive
re-render would wipe the user's selection. `lvt-form:preserve` on the `<form>` tells
the client to retain the live form values (including the chosen file) across a
re-render, so a validation error on another field doesn't lose the attachment.

```embed-lvt path="/apps/ui-patterns/forms/preserve-inputs" upstream="http://localhost:9091" height="360px"
```

## Template

One attribute — `lvt-form:preserve` on the form — keeps the inputs across re-renders.

```html include="/examples/patterns/templates/forms/preserve-inputs.tmpl"
```

## Handler & state

`Submit` validates and flashes; on a validation error the form re-renders but the
client-preserved file stays selected.

```go include="/examples/patterns/handlers_forms.go" region="preserve-inputs"
```

```go include="/examples/patterns/state_forms.go" region="preserve-inputs-state"
```

## When to use

- Any form with a file input plus other fields that can fail validation — preserve
  keeps the user from re-picking the file.
- Pairs naturally with [Inline Validation](/recipes/ui-patterns/forms/inline-validation).
