---
title: "Skip validation with formnovalidate"
description: "Let one button on a form bypass server-side validation — a 'Save draft' next to 'Publish' — using the standard HTML formnovalidate attribute, honored on every tier with no client code."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/formnovalidate.md
---

# Skip validation with `formnovalidate`

A form often has two submit buttons that mean different things. **Publish** should
enforce every rule — the title is required, the URL must parse. **Save draft**
should save whatever you have so far, half-finished and all, so you can come back
later. HTML already has an attribute for exactly this: `formnovalidate` on a submit
control tells the browser to skip its native constraint check for that button.

LiveTemplate honors `formnovalidate` **on the server too**. `ctx.ValidateForm()`
checks which control submitted the form, and when that control carried
`formnovalidate`, validation is skipped — so the same call enforces the rules for
**Publish** and waves them through for **Save draft**. No client code, no second
code path, and it works whether the submit arrived over the WebSocket, an HTTP
`fetch()`, or a plain no-JS form POST.

```embed-lvt path="/apps/draft-form/" upstream="http://localhost:9091" height="320px"
```

Leave the title empty and press **Publish** — you get a "required" error. Leave it
empty and press **Save draft** — it saves as a draft. Same field, same validation
call, different button.

## One template, two buttons

The only markup that matters is the `formnovalidate` attribute on the second button:

```html include="/examples/draft-form/draft.tmpl" lines="14-19"
```

`required` on the input drives the browser's native check (and the server's — the
framework infers the rule from the same attribute). `formnovalidate` on
`save-draft` opts that button out of both.

## One validation call, honored by submitter

Both actions call `ctx.ValidateForm()`. It enforces for `Publish` and skips for
`SaveDraft` — not because the handlers differ, but because the framework knows
which button was clicked:

```go include="/examples/draft-form/draft.go" lines="21-48"
```

At parse time the framework records every submit control that carries
`formnovalidate` into the form schema (`FormSchema.NoValidateSubmitters`, keyed by
the control's `name`). At request time it compares the form's *submitter* — the
clicked button — against that set, and `ctx.ValidateForm()` returns `nil` early
when it matches. Because the decision is keyed on the submitter rather than the
action, it holds even under `lvt-on:submit` routing, where the action is the
handler and the submitter is a separate button.

## It works without JavaScript

The skip is enforced on the server, so it survives all the way down to a plain
form POST. The mount below is the same app with `WithWebSocketDisabled()`; with JS
off, the browser submits natively and the server still skips validation for the
`save-draft` button.

```embed-lvt path="/apps/draft-form/no-js/" upstream="http://localhost:9091" height="320px"
```

On the no-JS tier the button's `name` *is* the action verbatim, so the kebab-case
`name="save-draft"` routes to the Go method `SaveDraft` (LiveTemplate matches
kebab-case, camelCase, snake_case, and PascalCase action names). That is why the
documented `<button name="save-draft">` pattern works end to end without a hidden
`lvt-action` field.

## A note on trust

`formnovalidate` is a **convenience, not a security boundary**. It lets a *cooperating*
client say "this submit is a draft, don't nag me about required fields" — and the
server obliges only when the clicked button is one your own template marked
`formnovalidate`. It does not authorize skipping authorization or integrity checks:
a draft still belongs to a session, and any rule that must always hold (ownership,
quotas, sanitization) belongs in the handler, unconditionally, not in
`ValidateForm()`. For the CSRF posture of these no-JS form posts, see
[Progressive enhancement](/recipes/progressive-enhancement/).
