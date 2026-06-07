---
title: "Infinite Scroll"
description: "Auto-load on scroll with IntersectionObserver."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/lists/infinite-scroll.tmpl"
---

# Infinite Scroll

A single `<div lvt-scroll-sentinel>` at the end of the list is watched by the
client's IntersectionObserver. When it scrolls into view the client dispatches the
`load_more` action on its own — no client JS to wire up. The handler is nearly
identical to [Click To Load](/recipes/ui-patterns/lists/click-to-load) — it just
pages a larger dataset; only the trigger really differs. When the last page arrives,
`HasMore` goes false and the sentinel is removed so it stops firing.

```embed-lvt path="/apps/ui-patterns/lists/infinite-scroll" upstream="http://localhost:9091" height="440px"
```

## Template

The sentinel `<div lvt-scroll-sentinel>` doubles as the loading indicator; once
`HasMore` is false it is replaced by an "End of list" note.

```html include="/examples/patterns/templates/lists/infinite-scroll.tmpl"
```

## Handler & state

`LoadMore` is dispatched automatically by the sentinel. It bumps the page, appends
the next slice, and updates `HasMore`.

```go include="/examples/patterns/handlers_lists.go" region="infinite-scroll"
```

```go include="/examples/patterns/state_lists.go" region="infinite-scroll-state"
```

## When to use

- Feeds and long lists where seamless scrolling matters more than an explicit
  pagination control.
- When you want auto-pagination without writing or maintaining any client-side
  IntersectionObserver code.

Use [Click To Load](/recipes/ui-patterns/lists/click-to-load) instead when the user
should decide when more rows load.
