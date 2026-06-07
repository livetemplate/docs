---
title: "Server Push"
description: "Push updates to the browser from a background goroutine with session.TriggerAction."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/realtime/server-push.tmpl"
---

# Server Push

Drive the UI from the server with no client-side polling. `StartTimer` spawns a
background goroutine that calls `session.TriggerAction(name, data)` once per second,
firing an action on the originating connection so the re-rendered tree is pushed to
the browser. `TriggerAction` returns an error when the session group has no live
connections — checking it each tick is the documented cancellation pattern, so the
goroutine exits cleanly if the user closes the tab.

```embed-lvt path="/apps/ui-patterns/realtime/server-push" upstream="http://localhost:9091" height="400px"
```

## Template

A Start button that flips to a live "running" view as the server pushes each tick.

```html include="/examples/patterns/templates/realtime/server-push.tmpl"
```

## Handler & state

`StartTimer` launches the ticker goroutine; `Tick` and `TimerDone` are the actions
the goroutine fires to update state.

```go include="/examples/patterns/handlers_realtime.go" region="server-push"
```

```go include="/examples/patterns/state_realtime.go" region="server-push-state"
```

## When to use

- Server-originated updates — progress bars, countdowns, job status, live metrics —
  where the client has nothing to send.
- A long-running task whose progress should stream to the browser without the user
  clicking anything.

Reach for [Multi-User Refresh](/recipes/ui-patterns/realtime/multi-user-sync) when
the trigger is a peer action rather than a background goroutine.
