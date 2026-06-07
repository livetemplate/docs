---
title: "Active Search"
description: "Debounced live search."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/search/active-search.tmpl"
---

# Active Search

Type in the search box and the result list re-renders on the server as you go — no
client-side filtering. A single `Change` action carries the current `query`, the handler
re-runs the search, and the table updates. The client debounces input (300ms) so each
keystroke doesn't fire a round trip.

```embed-lvt path="/apps/ui-patterns/search/active-search" upstream="http://localhost:9091" height="400px"
```

## Template

A plain `<form method="POST">` with one search input whose `name` is `query`; the table
ranges over `.Results`. The input's `value="{{.Query}}"` echoes the server's view so the
box stays in sync after each render.

```html include="/examples/patterns/templates/search/active-search.tmpl"
```

## Handler & state

`Change` reads `query` when present, re-runs `searchContacts`, and stores both the query
and the matching contacts back on state.

```go include="/examples/patterns/handlers_search.go" region="active-search"
```

```go include="/examples/patterns/state_search.go" region="active-search-state"
```

## When to use

- A list or directory that's easier to filter by typing than to page through.
- The match logic belongs on the server — search a database, normalize input, or rank
  results — instead of duplicating it in the browser.
- You want the search box and results to share one server-rendered region.

Reach for [URL-Preserved Filters](/recipes/ui-patterns/search/url-filters) when the
filter state should be bookmarkable and survive a reload.
