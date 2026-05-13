---
title: "Shared notepad: BasicAuth + per-user state + explicit peer refresh"
description: "BasicAuth turns the Authorization header into ctx.UserID() and a stable session group; a controller-owned map keys per-user state by that ID; and ctx.BroadcastAction(\"Refresh\", nil) keeps every tab of the same user in sync after a Save. The recipe sits between the [login recipe](../login/) (form-based session auth) and the [sync-and-broadcast deep dive](../sync-and-broadcast) (the general broadcast vs. trigger model)."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/shared-notepad/index.md
---

# Shared notepad: BasicAuth + per-user state + explicit peer refresh

The smallest authenticated multi-user app: a textarea, per-user persistence, and multi-tab sync. The whole thing fits in a controller with four action methods, a handler that wires `BasicAuthenticator` as a default option, and ten lines of template — the rest is framework machinery you don't write.

What makes this recipe worth a page is the *seam* between three independently-useful primitives:

- **`BasicAuthenticator`** turns the HTTP `Authorization` header into a stable identity. Username becomes both `ctx.UserID()` and the session-group ID.
- **A controller-owned map keyed by `ctx.UserID()`** is enough to isolate per-user state without a database. Alice's notes never leak to Bob.
- **`ctx.BroadcastAction("Refresh", nil)`** is the explicit peer-refresh primitive. After a Save, every other tab of the same user runs `Refresh` and re-reads from the map.

That last one is the v0.9.0 shape. Earlier versions of LiveTemplate auto-dispatched a controller method named `Sync` on every peer after every action; [livetemplate#406](https://github.com/livetemplate/livetemplate/pull/406) removed that auto-dispatch in favour of explicit broadcasts so authors control *when* peers refresh, not the framework.

Try it right here. Type some text and click Save:

```embed-lvt path="/apps/shared-notepad/" upstream="http://localhost:9091" height="520px"
```

To see the peer-refresh, open this page in a second browser tab — the embed there shares the same cookie-bound identity — and click Save in one. The other tab's content updates without you doing anything on it. For the cross-identity isolation story, open the page in a private window: different cookie, different `ctx.UserID()`, different session group, completely separate state.

> **A note on the embed's authenticator.** The recipe text below shows `NewDemoBasicAuth` — that's what `examples/shared-notepad/` and the e2e suite use, and it's the production-shaped wiring where the username from the `Authorization` header becomes `ctx.UserID()`. The embed above uses `AnonymousAuthenticator` instead — a cookie-bound session ID with no credential prompt — because tinkerdown's embed-lvt does a server-side prefetch to extract the LiveTemplate wrapper, and that prefetch can't forward `Authorization` headers. The controller code is identical either way; only the source of `ctx.UserID()` changes.

## The state struct

State is pure data, cloned per session. The three `lvt:"persist"` tags keep the textarea content and metadata alive across reconnects via the framework's client-side state checksum:

```go include="./_app/controller.go" region="state"
```

`Username` is intentionally **not** persisted — it's re-derived from `ctx.UserID()` on every `Mount`, and trusting a client-supplied username would be an authorization bug waiting to happen.

## The handler: pick the authenticator at mount time

The handler exposes two authenticator flavours and lets the caller pick. Production-shaped is BasicAuth:

```go include="./_app/handler.go" region="basicauth"
```

`BasicAuthenticator` answers two questions for every request:

- **`Authenticate(...)`** — does the credential pass? Here, any username with password `demo`. A real app would check against a user table.
- **`GroupID(...)`** — what session-group does this client belong to? `BasicAuthenticator` returns the username, so two tabs authenticated as alice land in the same group and broadcasts between them work; alice and bob land in different groups and nothing crosses.

`ctx.UserID()` in any action handler returns whatever the authenticator decided — username here.

## Mount: rehydrate from the per-user map

`Mount` runs on every fresh state — first page load, reconnect with stale state, or a state-restoring navigation. It binds `Username` to whoever just authenticated, then re-reads the textarea content from the per-user map:

```go include="./_app/controller.go" region="mount"
```

The `c.mu.RLock` is the only concurrency primitive in the recipe. Save takes the write lock; Mount, Refresh, and the implicit page-load reads take the read lock. For a production app this would be a database transaction, not a `map[string]NotepadState` — but the controller-shape is the same.

## Save: write through, then broadcast

The interesting line is the last one before the return:

```go include="./_app/controller.go" region="save"
```

`ctx.BroadcastAction("Refresh", nil)` doesn't run `Refresh` immediately on other connections — it *enqueues* the action for the framework's broadcast pipeline. After the current request's response is sent back to the originating tab, the framework drains the queue: for every peer connection in the same session group (other tabs of the same authenticated user), it dispatches `Refresh` against that connection's local state.

Two consequences worth knowing:

- **Each connection still has its own state copy.** Broadcast doesn't share state; it *replays an action*. A disconnected peer doesn't magically get a state update — it gets the missed actions applied in order when it reconnects.
- **The broadcast is fire-and-forget within a request.** The Save response goes to the originating tab immediately; peer tabs see the refresh milliseconds later as the queue drains.

For the deeper model — when to broadcast versus when to use `session.TriggerAction` for server-owned work — see [Broadcast & Server Push](/recipes/sync-and-broadcast).

## Refresh: a regular controller action

`Refresh` is the action peer tabs run when Save broadcasts. It's just a regular controller method — not a framework-reserved name:

```go include="./_app/controller.go" region="refresh"
```

The `Refresh` *name* is convention; the framework doesn't care. What matters is the `BroadcastAction("Refresh", nil)` call in Save naming this same string. Renaming the method also requires updating the broadcast call — and *every* peer connection that's running an older controller will see the broadcast and fail to route it.

## The form: persistence across re-render

The textarea form uses `lvt-form:preserve`, which tells the framework to retain the form's input values across a re-render:

```html include="./_app/notepad.tmpl" region="textarea"
```

Without `lvt-form:preserve`, a Save would re-render the article with the new `SavedAt` timestamp and the textarea would briefly flash empty before the framework re-binds `{{.Content}}`. Preserve keeps the user's typing intact during the round-trip.

For the deeper pattern (including how preserve composes with `Change()` for live preview), see [Patterns › Preserve inputs](/recipes/ui-patterns/forms/preserve-inputs).

## Where this recipe stops

Three production extensions that don't change the recipe's shape:

| Concern | This recipe | Production shape |
|---|---|---|
| Persistence | `map[string]NotepadState` in process memory | File-backed SQLite or Postgres; per-user rows; `Save` becomes an `UPDATE` |
| Auth | BasicAuth with hardcoded password check | `BasicAuthenticator` with bcrypt password compare, or a custom `Authenticator` that validates session tokens |
| Multi-instance broadcast | Single Fly machine, in-memory peer registry | `WithPubSubBroadcaster` (Redis) so `BroadcastAction("Refresh", ...)` reaches peer instances |
| Audit | No event log | Append-only `audit` table; `Save` writes the event before returning |

None of those changes the four action methods on `NotepadController`. The shape carries over.

## What next?

- [Login: form-based session auth](../login/) — cookies + `OnConnect` + `session.TriggerAction` instead of header auth.
- [Broadcast & Server Push](/recipes/sync-and-broadcast) — when to use explicit broadcast vs. server push.
- [Broadcasting, deeper](/recipes/broadcasting) — the broadcast pipeline in detail.
- [Patterns › Preserve inputs](/recipes/ui-patterns/forms/preserve-inputs) — the `lvt-form:preserve` story by itself.
- [Reference — Authentication](/reference/authentication) — the full `Authenticator` interface.
