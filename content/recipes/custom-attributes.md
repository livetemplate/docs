---
title: "Custom Attributes"
description: "Three lvt-x:* attributes written by the app itself: a star rating that sends an action, a copy button, and keyboard shortcuts the server can switch off."
source_repo: "https://github.com/livetemplate/docs"
source_path: "content/recipes/custom-attributes.md"
---

# Custom Attributes

The app below writes its own three attributes, `lvt-x:rating`, `lvt-x:copy` and
`lvt-x:hotkey`, in an inline `<script>`. The [custom attributes guide](/guides/custom-attributes)
explains the interface; the source is
[`examples/custom-attributes/`](https://github.com/livetemplate/docs/tree/main/examples/custom-attributes).

## Try it

```embed-lvt path="/apps/custom-attributes/" upstream="http://localhost:9091" height="420px"
```

Click a star, or press <kbd>1</kbd> to <kbd>5</kbd>. Copy the link and check the
number on the end. Then turn shortcuts off and press a key.

## The markup

```html include="/examples/custom-attributes/custom-attributes.tmpl" lines="91-105"
```

The server renders `lvt-x:rating="SetRating"` like any other attribute, and
renders `lvt-x:hotkey` only while `.Shortcuts` is true.

## A rating that sends an action

```js include="/examples/custom-attributes/custom-attributes.tmpl" lines="22-34"
```

```go include="/examples/custom-attributes/controller.go" lines="34-43"
```

Anyone can send `{"stars": 9000}` from the console, so the action checks the
range like any other input.

## A copy button that reads the current value

```js include="/examples/custom-attributes/custom-attributes.tmpl" lines="35-47"
```

The link includes the rating, which changes after the button is wired, so the
click reads `ctx.value` when it runs.

## Shortcuts the server can take away

```js include="/examples/custom-attributes/custom-attributes.tmpl" lines="49-89"
```

Turning shortcuts off re-renders the stars without `lvt-x:hotkey`. `setup()`
then rebuilds this client's map from the page, finds no keys, and the shortcuts
stop. The server never tells the handler.

## What the test checks

`TestCustomAttributesE2E` drives the page through a real Chrome:

- a star click sends one `SetRating` frame, and a second click after the
  re-render sends exactly one more;
- the copy button copies the link for a rating set after the page loaded;
- <kbd>2</kbd> rates two stars, with shortcuts off <kbd>4</kbd> sends no frame,
  and with shortcuts back on it works again;
- the console has no census warnings, registry warnings or errors.

The test runs over plain HTTP, where `navigator.clipboard` does not exist, so it
stubs the clipboard and records what each click copies.
