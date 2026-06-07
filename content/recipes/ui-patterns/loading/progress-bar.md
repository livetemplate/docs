---
title: "Progress Bar"
description: "Drive a determinate progress bar from the server — a goroutine ticks the percentage over the live connection, no client polling."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/loading/progress-bar.tmpl"
---

# Progress Bar

Show real server-side progress without polling. **Start** sets `Running` and spawns a
goroutine that pushes the percentage every 500ms with
`session.TriggerAction("updateProgress", …)`, climbing 10% at a time until the
`UpdateProgress` action hits 100%, flips to `Done`, and emits a success flash.
`Progress` and `Done` are persisted so a finished run survives a brief reconnect, while
`Running` is intentionally not — a stale spinner with no goroutine behind it is the
failure mode to avoid.

```embed-lvt path="/apps/ui-patterns/loading/progress-bar" upstream="http://localhost:9091" height="360px"
```

## Template

A native `<progress>` element bound to `.Progress`, shown whenever the job is running or
done. The success `FlashTag` lives inside the `{{if .Done}}` branch so it renders in the
same pass that completes the job, and the button label switches between **Start Job** and
**Run Again**.

```html include="/examples/patterns/templates/loading/progress-bar.tmpl"
```

## Handler & state

`Start` guards against re-entrancy, then `spawnTicker` runs the bounded loop;
`tickWithRetry` retries each push for ~5s so a brief mobile background doesn't drop a
tick. `UpdateProgress` writes the value and finalizes at 100%.

```go include="/examples/patterns/handlers_loading.go" region="progress-bar"
```

```go include="/examples/patterns/state_loading.go" region="progress-bar-state"
```

## When to use

- A long server job with a measurable percentage — file processing, a batch import, a
  multi-step export — where the user wants to watch it advance.
- You want push-based updates instead of the client hammering a status endpoint.
- The outcome should survive a flaky connection, so completion state is persisted.

Reach for [Async Operations](/recipes/ui-patterns/loading/async-operations) when the
work has no measurable progress and you only need loading / success / error, or
[Lazy Loading](/recipes/ui-patterns/loading/lazy-loading) when content just needs to
arrive after the first paint.
