---
title: "Keyboard Shortcuts"
description: "Bind global keys to server actions with lvt-on:window:keydown and an lvt-key filter."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/navigation/keyboard-shortcuts.tmpl"
---

# Keyboard Shortcuts

Bind global keys directly to server actions with `lvt-on:window:keydown` plus an
`lvt-key` filter — press `/` to open a command panel and `Escape` to close it, no
client JavaScript. The bindings are scoped: the `/` listener is attached only while
the panel is closed and the `Escape` listener only while it is open, so stray
keypresses never fire a no-op round-trip.

```embed-lvt path="/apps/ui-patterns/navigation/keyboard-shortcuts" upstream="http://localhost:9091" height="360px"
```

## Template

The `<article>` swaps its `lvt-on:window:keydown`/`lvt-key` pair based on
`.PanelOpen`, so exactly one shortcut is live at a time.

```html include="/examples/patterns/templates/navigation/keyboard-shortcuts.tmpl"
```

## Handler & state

`Open` and `Close` toggle `PanelOpen` and append a timestamped line to a capped
activity log.

```go include="/examples/patterns/handlers_navigation.go" region="keyboard-shortcuts"
```

```go include="/examples/patterns/state_navigation.go" region="keyboard-shortcuts-state"
```

## When to use

- Power-user shortcuts (command palette, quick actions) that should work from anywhere
  on the page.
- You want key bindings to be a function of server state — only live when the relevant
  UI is rendered.
- A no-JavaScript binding is enough; the action runs server-side over the WebSocket.

Pair this with a [Modal Dialog](/recipes/ui-patterns/navigation/modal-dialog) when the
shortcut should open a richer overlay than a simple panel.
