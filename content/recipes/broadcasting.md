---
title: "Broadcasting, deeper"
description: "How Subscribe(SelfTopic())+Publish routes within a session group, why some state belongs on the controller and not in lvt:\"persist\", and the two mutex rules that keep it from deadlocking."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/broadcasting.md
---

# Broadcasting, deeper

[Counter, deeper](/recipes/counter) showed `ctx.Subscribe(ctx.SelfTopic())` in `Mount` plus `ctx.Publish(ctx.SelfTopic(), "Increment", nil)` keeping every tab in a single browser in sync — a counter clicked in one tab ticks in the others. The scope of "every tab" was the [session group](/reference/session): the browser's cookie pins all its tabs to one group, `SelfTopic()` resolves to `lvt:session:<groupID>` for that group, and `Publish` fans the named action out to every connection that subscribed.

Broadcasting goes further within the same scope. Counter shared one integer; this pattern shares a multi-author message log. Same Subscribe/Publish primitives, two design choices that change everything — which fields are per-connection vs persisted, and where the source of truth lives.

```embed-lvt path="/recipes/ui-patterns/realtime/broadcasting" upstream="http://localhost:9091" height="380px"
```

Open the page in a second tab. Join with a different name. Send a message from either side. Both update. Both tabs are in the same session group (same cookie), so each tab's `SelfTopic()` resolves to the same string, and a Publish from either reaches both — but each tab keeps its own `Username` because identity is per-connection, not persisted.

(For a setup where every visitor — across browsers, across machines — sees the same fan-out, you'd swap [`AnonymousAuthenticator`](/reference/authentication) for one that returns a constant group ID, or define a developer-named topic like `"announcements"` and admit it in `WithTopicACL`. That's an authentication or ACL choice, not a `Publish` choice.)

## Anatomy of the state

```go include="./patterns/_app/state_realtime.go" region="broadcasting-state"
```

Note what's *not* persisted. `Username` looks like a candidate for `lvt:"persist"` — it's user identity, surely you want it to survive a reconnect? But persist storage is keyed by **session group**, so persisting `Username` would force every tab in the same browser to share one identity, defeating the demo where two tabs join as different users.

The pattern that *does* persist state across reconnects is `ReconnectionState` (also in this file) — different recipe, same package. Same fan-out scope (session group), but every connection sees the same value across drops because the field is `lvt:"persist"`-tagged.

## Where the messages live

```go include="./patterns/_app/handlers_realtime.go" region="broadcasting-controller"
```

The message log is on the **controller**, not in state. State is per-connection; the controller is the singleton dependency layer the [Controller+State pattern](/reference/controller-pattern) puts in front of every connection routed to this handler. `c.messages` is the source of truth — every tab reads from it under the same `RWMutex`.

The `Mount` method runs on every initial render — and in v0.10.0 it does **two** things: opt the connection into peer fan-out via `ctx.Subscribe(ctx.SelfTopic())`, *and* snapshot the current log into per-connection state. Without the Subscribe, a Publish from another tab would have no receiver in this session and the demo wouldn't work. Without the snapshot, a tab that opens *after* others have sent messages would render with `Messages: nil` until the next Publish arrives.

## Sending — Publish under the lock-release rule

```go include="./patterns/_app/handlers_realtime.go" region="broadcasting-send"
```

Two non-obvious mutex rules in this method:

1. **`Publish` after the lock release.** Holding the connection registry mutex while queuing publishes can deadlock with peer dispatches taking the same mutex from the other side. The pattern: mutate-and-snapshot under your lock, release, *then* Publish.

2. **`snapshotLocked()` requires the caller hold the lock.** A naked `slices.Clone(c.messages)` reads concurrently with `Send`'s append and races. The `Locked` suffix is documentation: violate it and you get a data race the test suite will catch under `-race`.

The third rule is implicit — `c.messages` is uncapped here. Production apps would ring-buffer, paginate, or persist to a TTL store. This demo skips that to keep the focus on the fan-out machinery itself.

## What peers do

```go include="./patterns/_app/handlers_realtime.go" region="broadcasting-newmessage"
```

`NewMessage` runs on every peer connection that subscribed to `SelfTopic()` when the Publish fires. It reads the shared log under `RLock` and copies into per-connection state. The template re-renders; the diff goes over the wire as patches, not full HTML.

This is why fan-out volume isn't proportional to message size: each peer's wire bytes equal the diff between its local state before and after `NewMessage`, which is roughly "one new message appended to the messages list."

## When this scales

Single process, single replica: works as-shown. The mutex serializes appends; the fan-out is in-process pub/sub.

Multi-replica: swap in-process fan-out for Redis Pub/Sub via [`WithPubSubBroadcaster`](/reference/pubsub). The handler shape stays identical — the `Mount`, `Send`, and `NewMessage` methods don't change. What changes is *where* `c.messages` lives (a shared store instead of a Go slice) and *how* the Publish propagates (Redis publish to `livetemplate:topic_action:<topic>`, replica subscribers fire `NewMessage` on their own subscribed connections; the framework's seen-ring deduplicates the SUBSCRIBE+PSUBSCRIBE double-fire for cross-instance wildcard topics).

## What's next

The reconnection-recovery pattern (live demo at [/recipes/ui-patterns/realtime/reconnection](/recipes/ui-patterns/realtime/reconnection)) is the persist-state companion. Same Subscribe/Publish shape, but the demo state survives a WebSocket drop because the fields are `lvt:"persist"`-tagged. A future recipe will go deep on it; for now the live widget plus its source in the same `_app/` is the reference.

