---
title: "Async Operations"
description: "A loading → success / error state machine for any async server call, driven entirely by server state."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/loading/async-operations.tmpl"
---

# Async Operations

The minimal shape for any async RPC — a database query, an HTTP call, a queued job.
**Fetch** sets `Status="loading"` synchronously and spawns a goroutine that waits, then
pushes a `fetchResult` action with either a success payload or an error (this demo fails
~33% of the time at random). The `FetchResult` action moves `Status` to `success` or
`error` and stashes the result or message, so the four states — idle, loading, success,
error — are all just values the template branches on.

```embed-lvt path="/apps/ui-patterns/loading/async-operations" upstream="http://localhost:9091" height="360px"
```

## Template

The button disables itself and shows "Fetching..." while `Status` is `loading`. Success
and error each get a `FlashTag` alert; the result renders in a `<blockquote>` and the
error string in an `aria-live` `<mark>` for assistive tech.

```html include="/examples/patterns/templates/loading/async-operations.tmpl"
```

## Handler & state

`Fetch` guards against a second in-flight request, checks the session before mutating,
then pushes a `fetchResult` from the goroutine; `FetchResult` resolves the state machine
and sets the matching flash.

```go include="/examples/patterns/handlers_loading.go" region="async-operations"
```

```go include="/examples/patterns/state_loading.go" region="async-operations-state"
```

## When to use

- A user-triggered server call that can fail — you need to show loading, then commit to
  a success or error branch.
- The work is one-shot with no measurable progress, so a spinner-to-result transition is
  enough.
- You want the loading and outcome states modeled as plain server state, not juggled in
  the client.

Reach for [Progress Bar](/recipes/ui-patterns/loading/progress-bar) when the operation
has a measurable percentage to report, or
[Lazy Loading](/recipes/ui-patterns/loading/lazy-loading) when the content should load
automatically after first paint rather than on a click.
