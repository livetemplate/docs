---
title: "Click to Load"
description: "Append-only pagination."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/lists/click-to-load.tmpl"
---

# Click To Load

Load More asks the server for the next page and appends it to the list. The handler
bumps `CurrentPage`, fetches that page, and appends the new items; because every row
has a stable `data-key`, the diff engine adds only the new rows and leaves the
existing ones untouched. A `HasMore` flag hides the button at the end of the list.

```embed-lvt path="/apps/ui-patterns/lists/click-to-load" upstream="http://localhost:9091" height="420px"
```

## Template

The button's `name="loadMore"` names the action; `{{if .HasMore}}` swaps it for an
"End of list" note once the last page is loaded.

```html include="/examples/patterns/templates/lists/click-to-load.tmpl"
```

## Handler & state

`LoadMore` increments the page counter, appends the next slice, and sets `HasMore`
from whether a full page came back.

```go include="/examples/patterns/handlers_lists.go" region="click-to-load"
```

```go include="/examples/patterns/state_lists.go" region="click-to-load-state"
```

## When to use

- Pagination where the user controls when more loads — an explicit button avoids
  surprise network traffic.
- Long lists where redrawing everything would be wasteful; the keyed append touches
  only new rows.

Reach for [Infinite Scroll](/recipes/ui-patterns/lists/infinite-scroll) when pages
should load automatically as the user scrolls.
