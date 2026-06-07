---
title: "Inline Validation"
description: "Validate fields on the server as the user types, rendering errors inline."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/forms/inline-validation.tmpl"
---

# Inline Validation

Validate on the server as the user types — no client validation library. The form's
`Change` action re-checks the fields and the template renders each error next to its
input via `{{.lvt.AriaInvalid}}` + `{{.lvt.ErrorTag}}`. Submit is only accepted once
`ctx.ValidateForm()` passes.

```embed-lvt path="/apps/ui-patterns/forms/inline-validation" upstream="http://localhost:9091" height="320px"
```

## Template

The browser enforces HTML rules (`type="email"`, `required`, `minlength`); the server
re-checks the same rules and renders `aria-invalid` + the error message inline.

```html include="/examples/patterns/templates/forms/inline-validation.tmpl"
```

## Handler & state

`Change` updates the touched field and runs `ValidateForm` (errors surface via the
template tags); `Submit` rejects until validation passes.

```go include="/examples/patterns/handlers_forms.go" lines="83-110"
```

```go include="/examples/patterns/state_forms.go" lines="22-28"
```

## When to use

- Fields whose validity depends on server-only knowledge (uniqueness, cross-field
  rules) — the browser can't decide those alone.
- You want one set of rules enforced in both places without duplicating logic.
