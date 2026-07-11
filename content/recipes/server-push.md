---
title: "Server push"
description: "Push updates from server-owned work — goroutines, timers, subscriptions, background jobs — to one session with session.TriggerAction, or to many sessions at once with a shared topic + out-of-band handler.Publish."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/server-push.md
---

# Server push

Most updates start with a client: a user clicks, the action runs, the framework
diffs and patches. **Server push is for the other case** — when *server-owned*
work finishes and needs to reach a session's live connections without any client
request: a background job completes, a subscription delivers an event, a timer
ticks.

`session.TriggerAction(action, data)` dispatches a named action into all of a
session's live connections, exactly as if the client had invoked it — the same
diff-and-patch pipeline, just enqueued from the server side.

## Triggering an action from server-owned work

Grab the session with `ctx.Session()` while you're still on a connection, then
hand it to whatever runs later — a goroutine, a timer callback, a job-result
handler:

```go
func (c *Controller) OnConnect(state State, ctx *livetemplate.Context) (State, error) {
    session := ctx.Session()
    go func() {
        result := fetchSlowData()                 // background work, no client involved
        _ = session.TriggerAction("DataLoaded", map[string]any{"value": result})
    }()
    return state, nil
}

func (c *Controller) DataLoaded(state State, ctx *livetemplate.Context) (State, error) {
    state.Value = ctx.GetString("value")          // ordinary action — re-render from new state
    return state, nil
}
```

`DataLoaded` is a normal action method; the only difference is *who* enqueued it.
Typical sources of a server push: a completed background job, an incoming
subscription or webhook event, a periodic timer.

## Routed by session, not by topic

`TriggerAction` reaches a session's connections by **group ID** — it does not use
topics, and the receiver does **not** need to have called `Subscribe`. That makes
it independent of [Pubsub](/recipes/pubsub): pubsub is "an action fans out to
*peers who opted in* via a topic"; server push is "server-owned work pushes to
*one session's* connections."

## Fanning out to many sessions at once

`TriggerAction` targets **one** session group. When a single background event
must refresh **many** sessions — every viewer of a shared dashboard, every tab of
every user — don't keep a registry of `Session` handles and loop over it. Two
primitives compose into a registry-free fan-out:

1. Every connection **joins a shared topic in `Mount`** with `ctx.Subscribe(topic)`
   (reconnect-durable, because `Mount` re-runs on reconnect).
2. A background goroutine calls the handler's out-of-band
   **`handler.Publish(topic, action, data)`** — no `Context`, safe from anywhere.
   Every subscriber, across every session group, re-runs `action` and re-renders.

The demo below runs **one** `time.Ticker`. On each tick it updates shared state
and calls `handler.Publish("dashboard", "Refresh", nil)`. Open it in two tabs —
both advance in lockstep, driven by that single goroutine, with no per-tab timer.

```embed-lvt path="/apps/live-dashboard/" upstream="http://localhost:9091" height="420px"
```

Because a shared topic is cross-user, subscribing to a developer topic is
**deny-all by default** — the handler authorizes it with `WithTopicACL`, the one
security boundary such a topic has. (Joining your *own* `ctx.SelfTopic()` to reach
just your own tabs needs no ACL.) See the
[server actions reference](https://github.com/livetemplate/livetemplate/blob/main/docs/references/server-actions.md)
for the full pattern and the per-user vs shared-group distinction.

| Need | Use |
|---|---|
| A background goroutine / timer / job should push to **one session's** connections | `session.TriggerAction("...", data)` (this page) |
| A background goroutine should refresh **many** sessions at once | `ctx.Subscribe(topic)` in `Mount` + out-of-band `handler.Publish(topic, ...)` (above) |
| A user action should also update peer tabs after it succeeds | [Pubsub](/recipes/pubsub) — `Subscribe` / `Publish` |
| The current connection should update from its own action | Return the new state from the action |

## What's next

- [Server Push pattern](/recipes/ui-patterns/realtime/server-push) — a live demo:
  a background goroutine calling `TriggerAction` once per second.
- [Live dashboard](https://github.com/livetemplate/docs/tree/main/examples/live-dashboard)
  — the source for the many-sessions demo above: one goroutine, `handler.Publish`,
  every viewer refreshed.
- [Pubsub](/recipes/pubsub) — the peer fan-out side: `Subscribe` + `Publish`.
- [Login](/recipes/login) — pushes a completed background result to the browser.
