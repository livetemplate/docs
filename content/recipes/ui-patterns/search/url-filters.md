---
title: "URL-Preserved Filters"
description: "Bookmarkable filter state via query params."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/search/url-filters.tmpl"
---

# URL-Preserved Filters

The filter and sort state lives in the URL's query string, so a filtered view is
bookmarkable and survives a reload. Filter links are SPA navigations that update the
history entry; `Mount` reads `status` and `sort` from the query params on every GET and
recomputes the list, falling back to defaults for unknown values so stale bookmarks
still render.

```embed-lvt path="/apps/ui-patterns/search/url-filters" upstream="http://localhost:9091" height="400px"
```

## Template

The filter and sort controls are ordinary `<a href="?status=…&sort=…">` links — each one
carries the full query string and marks itself `aria-current="page"` when active. The
table ranges over `.Items`.

```html include="/examples/patterns/templates/search/url-filters.tmpl"
```

## Handler & state

`Mount` only reads query params on a GET navigation (`ctx.Action() == ""`), validates
them against allow-lists, and always recomputes `Items` so both the initial render and
later actions see fresh data.

```go include="/examples/patterns/handlers_search.go" region="url-filters"
```

```go include="/examples/patterns/state_search.go" region="url-filters-state"
```

## When to use

- The current view should be shareable or bookmarkable — a link reproduces the exact
  filtered, sorted state.
- Filter changes should land in browser history so Back/Forward work as expected.
- You want stale or hand-edited URLs to degrade gracefully via validated defaults.

Reach for [Active Search](/recipes/ui-patterns/search/active-search) when filtering is
driven by free-text input rather than a fixed set of links.
