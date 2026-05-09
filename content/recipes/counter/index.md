---
title: "Counter, deeper"
description: "Past the +1 button: how BroadcastAction routes between sessions, why AnonymousAuthenticator is the right default for public demos, and where this pattern stops scaling."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/counter/index.md
---

# Counter, deeper

Most "counter" demos stop at "click +1, see number tick." Useful for proving the framework works; not so useful when you actually have to ship one. This recipe goes past the demo into the production-shaped questions: how `BroadcastAction` routes between sessions, why the cookie-bound session group matters for "multi-tab sync without leaking to other users," and what breaks first when this pattern meets real load.

The code is the same counter from [Your First App](/getting-started/your-first-app) — but the framing is different. Where that walkthrough builds the counter from scratch, this one stares at the four lines that do the actual work and unpacks them.

```embed-lvt path="/apps/counter/" upstream="http://localhost:9091" height="180px"
```

## Anatomy of the handler

The whole thing fits in three files. State + controller in one (the part you'd write):

```go include="./_app/counter.go" lines="9-33"
```

And a wiring file that exposes an `http.Handler`:

```go include="./_app/handler.go" lines="49-66"
```

There's not much to it. The choices that matter for production are the two `livetemplate.With*` options. Everything else is mechanical.

## Why `AnonymousAuthenticator` is the production default

LiveTemplate's `Authenticator` interface answers a single question on every HTTP and WebSocket request: *"who is this client, and which session group do they belong to?"* The session group is what `BroadcastAction` routes between. Two requests with the same group ID share state; different group IDs don't.

`AnonymousAuthenticator` (the framework's default, what this recipe uses) issues a cookie-bound group ID on first contact:

- Same browser, multiple tabs → same cookie → same group → broadcast works
- Different browser → different cookie → different group → isolated state
- Incognito window → its own cookie → its own group → clean slate

For a public docs site, that's the right shape. Every reader gets their own private counter on first visit, can prove broadcast within their own browser, and the demo can't be polluted by a stranger's clicks.

The alternative — a constant-group authenticator that puts every visitor in one shared group — is a demo-flavored shortcut. It makes a global ticker visible to all visitors, which is punchy on a marketing page but fails the "clean slate for thousands of users" test. We used it briefly during early development; the production switch to `AnonymousAuthenticator` was a one-line change with no other code impact:

```go
// Before — every visitor saw the same global counter
livetemplate.WithAuthenticator(sharedAuth{})

// After — each browser gets its own session group
livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{})
```

The `BroadcastAction` calls didn't change. The state struct didn't change. Only the routing rule for "who counts as the same session" changed, and that one swap converted a demo into a production-shaped widget.

## How `BroadcastAction` routes

The two action methods do the obvious thing — bump the counter, return the new state — and then call `ctx.BroadcastAction`:

```go include="./_app/counter.go" lines="22-33" highlight="24,31"
```

`BroadcastAction("Increment", nil)` adds an action to the broadcast queue. It does **not** apply the action immediately to other connections; it queues it. After the current request's response is sent, the framework drains the queue: for every other connection in the same session group, run `Increment` against that connection's local state.

Two consequences worth knowing:

- **Each connection still has its own state copy.** Broadcast doesn't share state — it replays actions. A connection that's been disconnected for a while doesn't get a magical state update; it gets the actions it missed when it reconnects, applied in order.
- **The broadcast is fire-and-forget within a request.** The current request's caller doesn't wait for the broadcast to finish. If you broadcast and then return, the response goes to the originating client immediately; the other clients see the update milliseconds later as the queue drains.

To prove the routing, here are two embeds against the same recipe app, side by side:

<div class="recipe-side-by-side" style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">

```embed-lvt path="/apps/counter/" upstream="http://localhost:9091" session="recipe-counter-deeper" height="200px"
```

```embed-lvt path="/apps/counter/" upstream="http://localhost:9091" session="recipe-counter-deeper" height="200px"
```

</div>

Click `+1` on one. The other ticks too — same browser, same cookie, same group, broadcast routes between them. Open this page in an incognito window: that incognito counter starts at zero and won't see your normal-window clicks. Different cookie, different group.

## Session group lifecycle

Worth pausing on what "session group" actually means in time.

1. **First visit**: the browser has no cookie. `AnonymousAuthenticator.GetSessionGroup` issues a fresh group ID and sets it as a cookie. The connection joins that group.
2. **Subsequent requests** (next tab, page refresh, WebSocket reconnect): the cookie is sent, the same group ID is returned, the connection joins the existing group.
3. **Cookie cleared / different browser**: a new group ID is issued. Old state is unreachable from the new group.
4. **Server restart**: cookies persist but in-memory session state is gone. New connections start fresh; broadcast queue is empty until clients reconnect and trigger new actions.

The group ID is the *only* thing tying a connection to its peers. Two browsers that somehow had the same cookie value would be in the same group. Two tabs from one browser are in the same group not because of the same TCP connection or anything similar — purely because of the shared cookie.

## When this pattern scales — and when it doesn't

This recipe is a deliberately small slice. The scaling story behind it is real:

| Scenario | Works? | Notes |
|---|---|---|
| One user, multiple tabs, single instance | ✅ Trivially. The broadcast queue runs in-process, the cost is one `Increment` call per connected tab. |
| Multiple users, single instance | ✅ Each user has their own session group; broadcasts stay scoped. |
| Multiple users, multiple instances (Fly machines, Kubernetes replicas) | ⚠️ Needs `WithPubSubBroadcaster` — by default a broadcast only reaches connections on the *same* instance. With Redis-backed broadcasting the broadcast fans out across instances. See [PubSub Reference](/reference/pubsub). |
| One group with thousands of connections (everyone broadcasting at high frequency) | ❌ Broadcast cost is O(N) per action; thousand-connection groups broadcasting at 100Hz mean 100k+ in-process calls per second. Either shard the group or use a different sync primitive. |
| Cross-user shared state (everyone sees everyone) | ⚠️ Possible — write a custom `Authenticator` that returns a constant group ID — but you've now built a write-amplification machine that any visitor can poke. Production examples need rate limiting, read-only modes, or moderation. |

`AnonymousAuthenticator` keeps you on the easy side of every row: per-user groups bound the fan-out, and the multi-instance question only matters once you've outgrown a single Fly machine.

## What the wiring file actually does

The full handler in `handler.go` is just the constructor expressed as a function. It exists because this recipe is mounted by the docs site's `cmd/site` aggregator — there's no standalone `main()`. In your own app you'd write a `main()` that does the same thing inline (`livetemplate.Must(...)` → `tmpl.Handle(...)` → `http.ListenAndServe`) and call it a day. Exposing it as a `Handler()` constructor is just so it can be mounted inside another binary's HTTP server.

```go include="./_app/handler.go" lines="14-46"
```

The `embed.FS` + temp-file dance at the top is a workaround for `livetemplate.WithParseFiles` taking filesystem paths — when the template ships inside the binary, we extract it once at first use. If you're running the standard "ship a directory of templates next to the binary" shape, you skip all this and pass the relative path directly.

## What next?

- [Reference — Authentication](/reference/authentication) — the full `Authenticator` interface, beyond the anonymous default.
- [Reference — PubSub & Broadcasting](/reference/pubsub) — multi-instance broadcasting via Redis.
- [Reference — Server Actions](/reference/server-actions) — the action lifecycle, including `BroadcastAction` ordering rules and gotchas.
- [Sync, Broadcast & Multi-User Sessions](/recipes/sync-and-broadcast) — when `Sync()` is enough and when you need broadcast.
- [Your First App](/getting-started/your-first-app) — if you arrived here cold, the from-scratch walkthrough is the better starting point.
