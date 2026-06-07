---
title: "Multi-User Refresh"
description: "Keep a shared counter in sync across every tab in a session by explicitly publishing a peer refresh action."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/realtime/multi-user-sync.tmpl"
---

# Multi-User Refresh

A counter clicked in one tab ticks up in every other tab. `Mount` opts each
connection into peer fan-out with `ctx.Subscribe(ctx.SelfTopic())`; `Increment`
bumps the controller's mutex-guarded counter and then calls
`ctx.Publish(ctx.SelfTopic(), "RefreshCounter", nil)`, so every subscribed peer
runs `RefreshCounter` and reloads the shared value. The publish is explicit — the
counter only converges because the action fans the refresh out.

```embed-lvt path="/apps/ui-patterns/realtime/multi-user-sync" upstream="http://localhost:9091" height="400px"
```

## Template

One button and one rendered counter — all the synchronization happens server-side.

```html include="/examples/patterns/templates/realtime/multi-user-sync.tmpl"
```

## Handler & state

`Mount` subscribes and seeds the initial count; `Increment` mutates and publishes;
`RefreshCounter` is the action peers run to converge.

```go include="/examples/patterns/handlers_realtime.go" region="multi-user-sync"
```

```go include="/examples/patterns/state_realtime.go" region="multi-user-sync-state"
```

## When to use

- A small piece of shared state — a counter, a toggle, a status flag — that every
  open tab should reflect immediately.
- You want explicit control over *when* peers refresh, rather than auto-syncing
  every field.

Reach for [Pubsub](/recipes/ui-patterns/realtime/pubsub) when peers
need to share a growing log rather than a single value.
