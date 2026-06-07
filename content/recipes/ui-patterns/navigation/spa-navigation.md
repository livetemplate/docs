---
title: "SPA Navigation"
description: "Auto-intercept every link inside the LiveTemplate wrapper — same-pathname over WebSocket, cross-pathname via fetch + reconnect."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/navigation/spa-navigation.tmpl"
---

# SPA Navigation

Every `<a>` inside the LiveTemplate wrapper is auto-intercepted, so ordinary links
behave like a single-page app with no router and no client code. Same-pathname links
(here, `?step=`) route through the in-band `__navigate__` action over the WebSocket and
re-render in place; cross-pathname links fetch the new page and reconnect the socket
transparently — both update history with `pushState` and neither does a hard reload.
External links opt out with `lvt-nav:no-intercept`.

```embed-lvt path="/apps/ui-patterns/navigation/spa-navigation" upstream="http://localhost:9091" height="360px"
```

## Template

The included snippet shows the same-pathname case: `?step=` links the framework
routes over the WebSocket. The full template also demonstrates cross-pathname links
(fetch + reconnect) and an `lvt-nav:no-intercept` external opt-out.

```html include="/examples/patterns/templates/navigation/spa-navigation.tmpl" region="spa-navigation"
```

## Handler & state

`Mount` reads and range-checks the `step` param; out-of-range or non-integer values
fall back to step 1.

```go include="/examples/patterns/handlers_navigation.go" region="spa-navigation"
```

```go include="/examples/patterns/state_navigation.go" region="spa-navigation-state"
```

## When to use

- You want SPA-feel navigation across a server-rendered app without adopting a
  client-side router or writing navigation JavaScript.
- Some links should stay in-band (query-param re-renders) while others cross to new
  handlers — the framework picks the right transport automatically.
- External or download links need to opt out and let the browser handle them.

For a single group of in-band links rather than whole-page navigation, see
[Tabs (HATEOAS)](/recipes/ui-patterns/navigation/tabs).
