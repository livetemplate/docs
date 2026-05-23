---
title: "Login: form-based session auth"
description: "Form-based session login without leaving the framework: an HTTP POST that sets an HttpOnly cookie via ctx.SetCookie, an OnConnect lifecycle hook that pushes a welcome message back via session.TriggerAction, and a symmetric POST-driven logout. All three pieces are framework-native — no custom middleware, no JS auth code, no escape hatches."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/login/index.md
---

# Login: form-based session auth

Most LiveTemplate code is reactive — actions arrive over a WebSocket, state mutates, the framework patches the DOM. Login flow is the one place that doesn't fit that mould: the browser hasn't authenticated yet, so there's no WebSocket to ride. Either you reach for a separate auth middleware (and lose framework-native flash messages, validation tags, and lifecycle hooks), or you let LiveTemplate handle it the way it handles everything else — as a controller action.

This recipe shows the second path. A login form posts to the same handler that renders the dashboard. The controller's `Login` method validates the credentials, sets an HttpOnly cookie via `ctx.SetCookie`, and 303-redirects. When the browser follows the redirect and the WebSocket connects, an `OnConnect` lifecycle hook spawns a goroutine that pushes a welcome message back to the client via `session.TriggerAction`. Logout mirrors the login shape — a POST that deletes the cookie and redirects.

Three pieces of the framework that are easy to miss until you need them: **cookies as first-class context primitives**, **forms that opt out of WebSocket interception** for the auth round-trip, and **server-initiated state updates** for everything after the page loads.

Try it in a new tab — any username; the password is `secret`. After the dashboard loads, watch for the welcome message: it's pushed from the server ~500ms after the WebSocket connects, no client poll involved.

**[Launch the login demo →](/apps/login/)**

The demo opens at its own URL rather than embedding inline because `lvt-form:no-intercept` posts to the current URL — inside an inline embed that would be the docs page, not the recipe handler, so the `Login` action would never run. The same constraint applies to any recipe that needs real browser navigation semantics (cookies on a redirect, POST-Redirect-GET).

## The state struct

The state holds only what the template needs to render either the login form or the dashboard. `lvt:"persist"` tags survive WebSocket reconnects via the framework's client-side state checksum:

```go include="/examples/login/controller.go" region="state"
```

`ServerMessage` is intentionally **not** persisted — it's set every time the WebSocket connects and re-derived on reconnect, so persisting it would mean stale welcome text.

## Login: HTTP POST, cookie, redirect

The login form in the template opts out of WebSocket interception with `lvt-form:no-intercept`. When the user clicks Login, the browser submits a standard HTTP POST to the page URL with `Accept: text/html`, and the framework routes the request to the controller's `Login` method:

```go include="/examples/login/controller.go" region="login"
```

Three framework primitives carry the auth weight here:

- **`livetemplate.NewFieldError("field", err)`** — surfaces validation errors keyed to a form input. The template binds them via `{{.lvt.ErrorTag "username"}}` and `{{.lvt.AriaInvalid "username"}}` so the rendered form keeps the user's filled-in fields and decorates the bad one with `aria-invalid` plus an error message.

- **`ctx.SetCookie(&http.Cookie{...})`** — first-class cookie API on the action context. The framework writes the `Set-Cookie` header on the redirect response, so the browser stores it before the WebSocket connects. `HttpOnly`, `SameSite=Strict`, and a 1-hour `MaxAge` are sensible defaults for a session token; production would also set `Secure: true` under HTTPS.

- **`ctx.Redirect(c.mountPath, http.StatusSeeOther)`** — the action returns its modified state *and* a redirect. POST-Redirect-GET: the framework writes the 303, the browser follows it, and the next GET renders the dashboard branch of the template against the new state. The redirect target is the recipe's mount path because `http.StripPrefix` strips the mount before the handler sees the URL — the recipe can't reconstruct its own mount from the request, so the caller passes it in.

The flash message API works the same way it does for non-auth flows — `ctx.SetFlash("error", "Invalid credentials")` stashes the message in a cookie, and the next render reads it via `{{.lvt.FlashTag "error"}}`.

## The form: opting out of WebSocket interception

The login form looks like a normal HTML form with one extra attribute:

```html include="/examples/login/auth.html" region="loginform"
```

`lvt-form:no-intercept` tells the LiveTemplate JS client *not* to intercept the submit. The form posts the natural browser way — `application/x-www-form-urlencoded` body, button-name routing (`login`) — and the response is a 303 redirect. This is what makes the auth flow work identically with and without JavaScript, and what makes the `Set-Cookie` header land on a real navigation response rather than a WebSocket frame.

## Server-push welcome: `OnConnect` + `session.TriggerAction`

After the redirect, the dashboard page loads, the JS client connects the WebSocket, and the framework calls the controller's `OnConnect` lifecycle method:

```go include="/examples/login/controller.go" region="onconnect"
```

Two things to notice. First, `ctx.Session()` returns the live session handle — the same one that `ctx.Publish` peer fan-outs and `session.TriggerAction` server-pushes flow through. Storing it in the controller's sessions map is a stand-in for what a real app would do via `SessionStore` and a background pub/sub channel.

Second, the actual welcome push happens in a goroutine. `OnConnect` returns immediately so the framework can finish the WebSocket handshake; the goroutine sleeps long enough for the dashboard to paint, then fires:

```go include="/examples/login/controller.go" region="serverpush"
```

`session.TriggerAction("serverWelcome", data)` enqueues an action on the session as if the client had dispatched it. The framework routes it to the controller's `ServerWelcome` method, runs the standard state-mutation-and-diff pipeline, and patches the dashboard — specifically the `<ins id="server-welcome-message">` block — with the new message. The client never asked for it; the server decided to update.

The same pattern is how you push subscription updates, completed background-job results, or any "the server learned something new" event. For the deeper model of when to use `Publish` (peer connections in the same session group that opted in via `Subscribe`) versus `TriggerAction` (server-owned work on one session), see [Sync & Server Push](/recipes/sync-and-broadcast).

## Logout: symmetric to login

Logout mirrors login's shape — another `lvt-form:no-intercept` form, another action method, another redirect:

```go include="/examples/login/controller.go" region="logout"
```

`ctx.DeleteCookie` writes `Set-Cookie: session_token=; MaxAge=-1`, the browser drops the cookie, the redirect lands on the login form, and the cycle starts over.

## Where this recipe stops

Three things a production auth flow needs that this recipe deliberately doesn't have:

- **Server-side session validation on every request.** The session cookie here is opaque, but its contents are advisory — nothing verifies on the next request that `session_<username>_<timestamp>` corresponds to a real prior login. A real implementation would use the cookie as a key into a session store (Redis, SQLite, Postgres) and reject requests whose tokens don't match.
- **Password hashing.** The demo accepts any username with the hardcoded password `secret`. A real implementation would store bcrypt/argon2 hashes and compare via `subtle.ConstantTimeCompare`.
- **Authenticator integration.** This recipe runs without an `Authenticator` — every browser gets a fresh session group by default. A real app would implement `Authenticator` and have its `Authenticate` method consult the session store, so `ctx.UserID()` is populated for every action.

The framework primitives shown here — cookies, no-intercept forms, `OnConnect`, `TriggerAction` — compose cleanly with all three additions. The auth *shape* doesn't change as you harden it; the implementations of the cookie validator, password compare, and `Authenticate` method swap in.

For a header-driven alternative that does have an `Authenticator`, see the [Shared notepad recipe](../shared-notepad/) — BasicAuth instead of form login, with `ctx.UserID()` driving per-user state isolation.

## What next?

- [Reference — Authentication](/reference/authentication) — the full `Authenticator` interface contract.
- [Reference — Session](/reference/session) — `session.TriggerAction`, `SessionStore`, and the session group model.
- [Shared notepad](../shared-notepad/) — BasicAuth + per-user state + explicit peer refresh.
- [Sync & Server Push](/recipes/sync-and-broadcast) — when to use `Publish` peer fan-out vs. `TriggerAction` server-push.
