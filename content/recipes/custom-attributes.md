---
title: "Custom Attributes"
description: "Three lvt-x:* attributes written by the app itself: a star rating that sends an action, a copy button, and keyboard shortcuts the server can switch off."
source_repo: "https://github.com/livetemplate/docs"
source_path: "content/recipes/custom-attributes.md"
---

# Custom Attributes — three handlers in one template

None of the attributes on this page ship with LiveTemplate. The app writes all
three in an inline `<script>`, and the client runs them like its own. The
[custom attributes guide](/guides/custom-attributes) explains the interface;
this is the working version, with a browser test behind every claim. The full
source is
[`examples/custom-attributes/`](https://github.com/livetemplate/docs/tree/main/examples/custom-attributes).

## Try it

```embed-lvt path="/apps/custom-attributes/" upstream="http://localhost:9091" height="420px"
```

Click a star, or press <kbd>1</kbd> to <kbd>5</kbd>. Copy the link and look at
the number on the end. Then turn shortcuts off and press a key.

## The markup

```html include="/examples/custom-attributes/custom-attributes.tmpl" lines="91-105"
```

Three attribute names, and the server knows none of them. It renders
`lvt-x:rating="SetRating"` the way it renders `class="on"`. The one conditional
worth noticing is on the stars: `lvt-x:hotkey` is only rendered while
`.Shortcuts` is true.

## A rating that sends an action

```js include="/examples/custom-attributes/custom-attributes.tmpl" lines="22-34"
```

The attribute's value is the action name. `needsServerChannel: true` is what
puts `send` on the context, so a handler that only touches the page never gets
a connection it could misuse. The listener sits on the row of stars and reads
the star from the click, which is one listener per row, not five.

The action it reaches is ordinary:

```go include="/examples/custom-attributes/controller.go" lines="34-43"
```

The range check matters. A custom attribute is client code, and anyone can
send `{"stars": 9000}` from the console. The server checks it like any other
input.

## A copy button that reads the current value

```js include="/examples/custom-attributes/custom-attributes.tmpl" lines="35-47"
```

The link includes the rating, so it changes after `onElementAdded` wired the button.
`ctx.value` is read inside the click, and the click copies whatever the
attribute says now. The test rates five stars after the page loads, clicks
copy, and checks that the copied link ends in `?stars=5`.

Both handlers pass `{ signal: ctx.signal }`. Leave it out and nothing breaks
here, because these attributes never go away. It is still the habit to keep:
the day an attribute does go away, the listener goes with it.

## Shortcuts the server can take away

```js include="/examples/custom-attributes/custom-attributes.tmpl" lines="49-89"
```

This one uses the low-level form, because one `keydown` listener has to serve
every star. Per-element callbacks have nowhere to keep that map.

The interesting part is what happens when you turn shortcuts off. The server
re-renders the stars without `lvt-x:hotkey`. `setup()` runs after that render
and rebuilds this client's map from the page as it is now, which no longer has
those attributes. So the map is empty and the keys do nothing. Nothing had to tell
the handler that shortcuts were off. The rebuild is the sweep.

Each client gets its own map, keyed by its `wrapperRoot`. Every client on a page
shares the same handlers, and a docs page with several live embeds has several
clients. One map would let any of them wipe the others' keys on its next render.

`setup()` adds the listener on every call. That would normally stack
one listener per render. It doesn't here, because the DOM ignores an identical
function added a second time. And `claims` is what keeps the attribute census
quiet: without it, the client would warn that nothing handles `lvt-x:hotkey`.

## What the test checks

`TestCustomAttributesE2E` drives all of this through a real Chrome:

- a star click sends one `SetRating` frame, and a second click after the
  re-render sends exactly one more;
- the copy button copies the link for a rating set after the page loaded;
- <kbd>2</kbd> rates two stars, and with shortcuts off <kbd>4</kbd> sends no
  frame at all;
- with shortcuts back on, <kbd>4</kbd> works again;
- the console stays free of census warnings, registry warnings and errors.

The page runs over plain HTTP in the test, where `navigator.clipboard` does not
exist, so the test stubs it and records the text each click hands it.
