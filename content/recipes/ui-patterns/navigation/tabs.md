---
title: "Tabs (HATEOAS)"
description: "Server-driven tabs as query-param links, routed over the WebSocket with no HTTP round-trip."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/navigation/tabs.tmpl"
---

# Tabs (HATEOAS)

Render tabs as plain `?tab=…` links and let the framework intercept the click and
route it over the WebSocket via the in-band `__navigate__` action — no HTTP round-trip
and no client-side tab state. `Mount()` re-runs with the new query param, picks the
active tab from a `validTabs` allowlist, and re-renders the panel server-side, so deep
links like `?tab=settings` work on a cold load too.

```embed-lvt path="/apps/ui-patterns/navigation/tabs" upstream="http://localhost:9091" height="360px"
```

## Template

Each tab is an anchor with `?tab=`; the active one gets `aria-current="page"`. The
panel body is a server-rendered `{{if eq .ActiveTab …}}` branch — switching tabs is a
re-render, not a partial fragment.

```html include="/examples/patterns/templates/navigation/tabs.tmpl"
```

## Handler & state

A single `Mount` handler reads the `tab` param, validates it against an allowlist, and
falls back to overview for unknown or stale values.

```go include="/examples/patterns/handlers_navigation.go" region="tabs"
```

```go include="/examples/patterns/state_navigation.go" region="tabs-state"
```

## When to use

- Sectioned content where each section should be bookmarkable and shareable by URL.
- You want zero client tab state — the server is the single source of truth for which
  panel is active.
- The tab set is small and known, so an allowlist keeps stale links safe.

This is [SPA Navigation](/recipes/ui-patterns/navigation/spa-navigation) applied to one
link group; reach for that pattern when whole-page links need the same interception.
