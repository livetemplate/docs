---
title: "Todos: a real LiveTemplate app"
description: "Past the +1: how a recipe with auth, persistent data, and reusable components actually fits together. BasicAuthenticator scopes data per user, sqlc-generated queries handle storage, and lvt/components supply the modal + toast that make it feel finished."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/todos/index.md
---

# Todos: a real LiveTemplate app

The counter recipe ends at "click +1, see number." That's enough to prove the framework runs. It's not enough to know what a real LiveTemplate app *looks like* — what auth wires up, where state lives, how reusable components compose with the framework's render lifecycle.

This recipe is the smallest app that touches all of those: a multi-user todo list with BasicAuth scoping, SQLite persistence, and the `lvt/components` modal + toast for the finishing touches.

The live demo can't embed inline — it's gated by HTTP BasicAuth, which is part of the teaching point. Open it in a new tab and the browser will prompt for credentials:

<a href="/apps/todos/" target="_blank" rel="noopener"><strong>→ Launch the live demo</strong> &middot; <code>alice / password</code> or <code>bob / password</code></a>

Add a few todos, then open the same link in an incognito window and log in as the other user — your two browsers see two separate lists, even though they're talking to the same Go process. That separation is the recipe's central teaching point.

## Anatomy of the wiring

The handler ties together five things: BasicAuth, the in-memory database, the controller, the components, and the template. The whole orchestration is one function:

```go include="/examples/todos/handler.go" lines="87-138"
```

Three of those `livetemplate.With*` options carry teaching weight. Origins (`opts...`) are deployment plumbing.

## How `BasicAuthenticator` scopes data per user

Every action handler in the controller filters database queries by `ctx.UserID()`. That `UserID` comes from the authenticator: `BasicAuthenticator` returns the username that authenticated the request as both the user identity *and* the session group ID. The framework guarantees:

- **Same user, multiple tabs** → same group → an opt-in `ctx.Subscribe(ctx.SelfTopic())` in Mount plus `ctx.Publish(ctx.SelfTopic(), "RefreshTodos", nil)` from each mutating action keeps tabs in sync within one logged-in user
- **Different users on the same machine** → different groups → no leakage
- **Same user across devices** → still the same group (the username is the group), so a logged-in alice on phone + laptop sees the same list

The query layer enforces this at the SQL boundary too — every query takes a `user_id` parameter:

```sql include="/examples/todos/db/queries.sql" lines="1-15"
```

A bug in the controller that forgot to pass `ctx.UserID()` would silently return all users' todos. The compile-time signature on the generated `GetAllTodos(ctx, userID)` makes that mistake hard to write — and a real production app would also enforce it at the DB layer with a row-level security policy. Here `:memory:` SQLite skips that, but the column is laid down so the upgrade path is one-line.

## Why components live outside `lvt:"persist"`

The trick most LiveTemplate apps hit on day three is that **state objects must round-trip through JSON serialization on reconnect** — but rich UI primitives (modal stacks, toast queues) carry mutable state and aren't serializable. Components solve this by being *re-initialized* on every state-restoring lifecycle method.

Look at the state struct:

```go include="/examples/todos/state.go" lines="63-94"
```

`Toasts` and `DeleteConfirm` are pointer types from `lvt/components`. They're missing the `lvt:"persist"` tag deliberately — when a connection reconnects mid-conversation, the framework rehydrates everything else (the search query, the page number, the pending delete ID) but leaves these `nil`. The controller re-creates them in three places — every entry point where state may have just been hydrated:

```go include="/examples/todos/controller.go" lines="19-34"
```

And the re-init function:

```go include="/examples/todos/controller.go" lines="239-260"
```

The pattern: **persistable plain data with `lvt:"persist"`; non-serializable runtime objects re-built in `Mount` / `OnConnect` / `Sync`.** A toast queue that was serialized would be a re-render hazard (the same notifications would repaint after every reconnect); explicit re-init at lifecycle entry points is the right shape.

## Modal + toast in two action handlers

The delete-with-confirm flow is two action handlers and the modal component handles the open/close state for you:

```go include="/examples/todos/controller.go" lines="93-127"
```

`ConfirmDelete` is fired when the user clicks Delete on a row — the modal is opened, no DB work yet. `ConfirmDeleteConfirm` runs only if the user clicks the destructive button inside the modal — by then `state.DeleteID` is whatever the original click captured. `CancelDeleteConfirm` clears the modal without touching the DB. The component never round-trips to the server for its own UI state changes; it's just `state.DeleteConfirm.Show()` / `.Hide()`.

Toasts are even simpler — fire-and-forget from any action:

```go include="/examples/todos/controller.go" lines="36-59" highlight="56"
```

The `state.Toasts.AddSuccess(...)` call queues a notification; the rendered template walks `state.Toasts` and emits the toast container. The toast disappears client-side on its own dismiss timer; you don't write any of that.

## Where the recipe stops, and what production needs

This is the smallest app that exercises the full LiveTemplate idiom. A real production app would extend it on three axes:

| Concern | This recipe | Production shape |
|---|---|---|
| Persistence | `:memory:` SQLite, lost on restart | File-backed SQLite or Postgres; daily backup |
| Auth | Hardcoded alice/bob with plaintext passwords | OAuth/SSO + an `Authenticator` impl that validates session tokens |
| Multi-instance fan-out | Single Fly machine | `WithPubSubBroadcaster` (Redis) so `Publish` reaches peer instances |
| User registration | None | Companion endpoint + `lvt/components/form` validation |
| Audit trail | None | Append-only log table; query layer logs writes |

None of those changes the recipe's *shape* — the same controller methods, the same state struct, the same components. They swap implementations, not the surface. That's the architectural payoff for the upfront ceremony of `Authenticator` + `Mount`/`OnConnect` + explicit peer refresh actions + components: the apps that grow out of this recipe inherit a clean separation between deployment plumbing and the actual interaction surface.

## What next?

- [Counter, deeper](/recipes/counter) — the same `Subscribe(SelfTopic())` + `Publish` peer-fan-out pattern this app uses for multi-tab refresh, in isolation.
- [Reference — Authentication](/reference/authentication) — the full `Authenticator` interface and the contracts `BasicAuthenticator` implements.
- [Reference — Components](/reference/components) — the modal + toast APIs, plus the rest of `lvt/components`.
- [Pubsub](/recipes/pubsub) / [Server push](/recipes/server-push) — explicit `Publish` peer fan-out vs. `TriggerAction` server push.
