---
title: "Broadcasting, deeper"
description: "How ctx.BroadcastAction reaches every connected peer, why some state belongs on the controller and not in lvt:\"persist\", and the two mutex rules that keep it from deadlocking."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/patterns/broadcasting.md
---

# Broadcasting, deeper

[Counter, deeper](/recipes/counter) showed `ctx.BroadcastAction("RefreshState", nil)` reaching every tab a single user has open. This recipe goes one step further: every *connected* peer, across users.

Same primitive, two design choices that change everything — which fields are per-connection vs persisted, and where the source of truth lives.

```embed-lvt path="/apps/patterns/realtime/broadcasting" upstream="http://localhost:9091" height="380px"
```

Open the page in a second tab. Join with a different name. Send a message from either side. Both update.

## Anatomy of the state

```go include="./_app/state_realtime.go" region="broadcasting-state"
```

Note what's *not* persisted. `Username` looks like a candidate for `lvt:"persist"` — it's user identity, surely you want it to survive a reconnect? But persist storage is keyed by **session group**, and the session group is the cookie-bound group the [counter recipe](/recipes/counter) covered. Persisting `Username` would force every tab in the same browser to share one identity, defeating the whole demo.

Per-connection state is what makes two tabs *as two users* work. The pattern that *does* persist user state is [Reconnection Recovery](/recipes/patterns/reconnection) — we'll get there.

## Where the messages live

```go include="./_app/handlers_realtime.go" region="broadcasting-controller"
```

The message log is on the **controller**, not in state. State is per-connection; the controller is the singleton dependency layer the [Controller+State pattern](/reference/controller-pattern) puts in front of every connection routed to this handler. `c.messages` is the source of truth — every tab reads from it under the same `RWMutex`.

The `Mount` method runs on every initial render — without it, a tab that opens *after* others have sent messages would render with `Messages: nil` until the next broadcast arrives. Mount snapshots the current log into per-connection state so each tab starts coherent.

## The broadcast

```go include="./_app/handlers_realtime.go" region="broadcasting-send"
```

Two non-obvious mutex rules in this method:

1. **`BroadcastAction` after the lock release.** Holding the connection registry mutex while queuing broadcasts can deadlock with peer dispatches taking the same mutex from the other side. The pattern: mutate-and-snapshot under your lock, release, *then* broadcast.

2. **`snapshotLocked()` requires the caller hold the lock.** A naked `slices.Clone(c.messages)` reads concurrently with `Send`'s append and races. The `Locked` suffix is documentation: violate it and you get a data race the test suite will catch under `-race`.

The third rule is implicit — `c.messages` is uncapped here. Production apps would ring-buffer, paginate, or persist to a TTL store. This demo skips that to keep the focus on `BroadcastAction` itself.

## What peers do

```go include="./_app/handlers_realtime.go" region="broadcasting-newmessage"
```

`NewMessage` runs on every peer when the broadcast fires. It reads the shared log under `RLock` and copies into per-connection state. The template re-renders; the diff goes over the wire as patches, not full HTML.

This is why broadcast volume isn't proportional to message size: each peer's wire bytes equal the diff between its local state before and after `NewMessage`, which is roughly "one new message appended to the messages list."

## When this scales

Single process, single replica: works as-shown. The mutex serializes appends; the broadcast is in-process Pub/Sub.

Multi-replica: swap in-process broadcast for Redis Pub/Sub via [`WithPubSubBroadcaster`](/reference/pubsub). The handler shape stays identical — the `Send` and `NewMessage` methods don't change. What changes is *where* `c.messages` lives (a shared store instead of a Go slice) and *how* `BroadcastAction` propagates (Redis publish, replica subscribers fire `NewMessage` on their connections).

## What's next

[Reconnection Recovery →](/recipes/patterns/reconnection) — the persist-state case. Same `BroadcastAction` shape, but every connection sees the same identity across reconnects.
