---
title: "Pubsub"
description: "Share a multi-author message log across every connection in a session via Publish to SelfTopic()."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/realtime/pubsub.tmpl"
---

# Pubsub

A message sent in one tab appears in every other tab that joined. The shared log
lives on the controller behind a mutex; `Mount` subscribes each connection with
`ctx.Subscribe(ctx.SelfTopic())` and snapshots the log into local state. `Send`
appends under the lock, releases, then `ctx.Publish(ctx.SelfTopic(), "NewMessage", nil)`
fans the action out so every subscribed peer re-reads the log. `Username` is
per-connection — deliberately not persisted — so two tabs can join as different users.

```embed-lvt path="/apps/ui-patterns/realtime/pubsub" upstream="http://localhost:9091" height="400px"
```

## Template

A join form swaps for the message list plus a send form once the user has a name.

```html include="/examples/patterns/templates/realtime/pubsub.tmpl" region="pubsub"
```

## Handler & state

`Mount` subscribes and seeds the log; `Send` mutates-then-publishes after releasing
the lock, and peers run `NewMessage` to converge.

```go include="/examples/patterns/handlers_realtime.go" region="pubsub-controller"
```

`Send` appends under the lock, releases it, then publishes so every peer converges:

```go include="/examples/patterns/handlers_realtime.go" region="pubsub-send"
```

```go include="/examples/patterns/handlers_realtime.go" region="pubsub-newmessage"
```

```go include="/examples/patterns/state_realtime.go" region="pubsub-state"
```

## When to use

- A shared, append-only feed — chat, an activity log, live comments — that every
  connection in the session should see grow.
- Each connection keeps its own identity while reading one shared source of truth.

For the full deep-dive on the mutex rules and pub/sub scope, see
[Pubsub](/recipes/pubsub). Use
[Presence Tracking](/recipes/ui-patterns/realtime/presence) when you only need to
know who is currently connected.
