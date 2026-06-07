---
title: "Presence Tracking"
description: "Show who's online with explicit Join/Leave actions and a shared, mutex-guarded user map."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/realtime/presence.tmpl"
---

# Presence Tracking

Track how many users are currently online. Explicit `Join` and `Leave` actions
mutate a mutex-guarded user map on the controller, then each calls
`ctx.Publish(ctx.SelfTopic(), "PresenceChanged", nil)` so every subscribed peer
recomputes its `OnlineCount` from the shared map. `Username` and `Joined` stay
per-connection — peers update only the count, never another connection's identity.

```embed-lvt path="/apps/ui-patterns/realtime/presence" upstream="http://localhost:9091" height="400px"
```

## Template

A live count plus a join form that swaps for a Leave button once you're in.

```html include="/examples/patterns/templates/realtime/presence.tmpl"
```

## Handler & state

`Mount` subscribes and seeds the count; `Join`/`Leave` mutate the shared map and
publish; `PresenceChanged` refreshes only the count on peers.

```go include="/examples/patterns/handlers_realtime.go" region="presence"
```

```go include="/examples/patterns/state_realtime.go" region="presence-state"
```

## When to use

- An online indicator, a "who's here" list, or a typing/active status for a shared
  space.
- You want explicit join/leave semantics rather than inferring presence from raw
  connection lifecycle.

Reach for [Pubsub](/recipes/ui-patterns/realtime/pubsub) when peers
need to exchange messages, not just presence.
