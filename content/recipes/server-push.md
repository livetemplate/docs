---
title: "Server push"
description: "Push updates to a session's live connections from server-owned work — goroutines, timers, subscriptions, background jobs — with session.TriggerAction, independent of Subscribe/Publish."
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

| Need | Use |
|---|---|
| A background goroutine / timer / job should push to live connections | `session.TriggerAction("...", data)` (this page) |
| A user action should also update peer tabs after it succeeds | [Pubsub](/recipes/pubsub) — `Subscribe` / `Publish` |
| The current connection should update from its own action | Return the new state from the action |

## What's next

- [Server Push pattern](/recipes/ui-patterns/realtime/server-push) — a live demo:
  a background goroutine calling `TriggerAction` once per second.
- [Pubsub](/recipes/pubsub) — the peer fan-out side: `Subscribe` + `Publish`.
- [Login](/recipes/login) — pushes a completed background result to the browser.
