---
title: "Progressive enhancement: graceful degradation"
description: "How a LiveTemplate app stays functional across three transports — full live, HTTP fetch fallback, and raw form POST — from a single controller and a single template. The only code-level switch between modes is one option flag."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/progressive-enhancement/index.md
---

# Progressive enhancement: graceful degradation

Most "live" frameworks have a load-bearing assumption: JavaScript is on, the WebSocket connects, and the user agent cooperates. LiveTemplate is built so that those assumptions can fail one at a time without the app breaking. The same controller, the same template, and the same form markup degrade through three modes:

- **Tier A** — JS on, WS on (default). Actions travel over the WebSocket; UI updates as diff patches.
- **Tier B** — JS on, WS off (`WithWebSocketDisabled()`). The client falls back to plain HTTP `fetch()`; same diff patches over a different transport.
- **Tier C** — JS off entirely. Browser submits the form natively; server responds with `303 See Other` (POST-Redirect-GET).

The interesting bit is that **the template doesn't change between tiers** — every form is `<form method="POST" name="add">`, every button is `<button name="add">`. Tier C falls out of native browser behavior, the framework just has to handle the POST. The only code-level distinction across the three modes is one option flag for Tier B.

## Tier A — full live (the default)

The live demo below uses LiveTemplate's normal transport: a WebSocket per session, diff patches over the wire, no whole-page reload between actions. Add a todo, toggle one, delete one — all without a navigation event.

```embed-lvt path="/apps/progressive-enhancement/" upstream="http://localhost:9091" height="420px"
```

The handler that produces this is the smallest possible recipe shape — no auth, no DB:

```go include="./_app/handler.go" lines="77-92"
```

That's it. `WithParseFiles(extractTemplate())` ships the embedded `.tmpl` into the framework; everything else (`opts...`) is what the caller — `cmd/site` in production, the e2e harness in tests — supplies for origin policy and dev-mode static assets.

## Tier B — WebSocket disabled (HTTP fetch fallback)

Same app, same template, same controller. The only difference is one option appended at construction time:

```go
livetemplate.WithWebSocketDisabled()
```

When the server rejects the WebSocket upgrade, the client library detects it, falls back to plain HTTP `fetch()` for action delivery, and applies the same diff patches it would have applied over WS. The user-visible behavior is identical — instant updates, no page reload — but the network path is request/response.

```embed-lvt path="/apps/progressive-enhancement/no-ws/" upstream="http://localhost:9091" height="420px"
```

A WebSocket upgrade against this mount is rejected before negotiation:

```text
GET /no-ws/ HTTP/1.1
Upgrade: websocket
Connection: Upgrade

→ HTTP/1.1 400 Bad Request   (or similar non-101)
```

The e2e suite asserts this directly with `TestPE_TierB_WebSocketRejected`. When you see a 101 from this mount, something is wrong.

## Tier C — JavaScript disabled (POST-Redirect-GET)

Tier C is the one that surprises people new to LiveTemplate, because there's no toggle for it on the server side — it's just what happens when the JS client isn't there to intercept the form submit. The browser sends `POST` with `Accept: text/html` to the form's action (the page's own URL), the framework handles the action, and the response is a `303 See Other` to the same URL with the flash message stashed in a cookie:

```text
POST / HTTP/1.1
Content-Type: application/x-www-form-urlencoded
Accept: text/html

add=&title=Pick+up+milk

→ HTTP/1.1 303 See Other
   Location: /
   Set-Cookie: lvt-flash=success=Added%3A+Pick+up+milk; ...
```

The browser follows the redirect, the next `GET` re-renders with the new state, and the flash cookie is consumed and cleared. POST-Redirect-GET is a well-known pattern; LiveTemplate just speaks it natively when the request shape says "no JS interception."

The template carries one piece of UX scaffolding for this mode — a `<noscript>` banner that's only visible when scripts are disabled:

```html include="./_app/progressive-enhancement.tmpl" lines="27-33"
```

To try Tier C live: open the [Tier A demo](/apps/progressive-enhancement/) in a new tab, then in DevTools (Cmd-Option-I / F12) → Settings → Debugger, check "Disable JavaScript" and refresh. The banner appears, every action causes a full page navigation, but the app remains fully functional.

## Why `InputTitle` is on the state struct

Forms reset on submit. If the user types `ab` (too short), submits, and gets a validation error, the framework re-renders — and on Tier A/B the framework's diff doesn't *reset* the input field, but on Tier C the page is fully reloaded after a 303 round-trip and the input is gone. Without explicit handling, the user retypes from scratch.

The fix is one persisted field on state and one template binding:

```go include="./_app/controller.go" lines="69-89"
```

On validation failure, `state.InputTitle = ctx.GetString("title")` captures whatever the user typed; the template binds it back via `value="{{.InputTitle}}"`. After a successful submit, `state.InputTitle = ""` clears it. The same persistence works across all three tiers because state round-trips through the framework regardless of transport.

## Action resolution: form name vs button name

All three forms in the template use the same shape:

```html include="./_app/progressive-enhancement.tmpl" lines="41-54"
```

The form has `name="add"` and a button with `name="add"`. Both naming the same action is intentional belt-and-suspenders:

- **Browser-native POST** (Tier C): the browser only includes form fields in the body. The clicked button's `name=value` pair is included, so the body is `add=&title=...` — the framework reads `add=` from the body to route the action.
- **JS-intercepted submit** (Tier A/B): the client library reads the form's `name` attribute as the action.

Either path resolves to the controller's `Add` method. The toggle and delete forms follow the same shape with hidden `id` inputs to carry the row identity.

## Where this stops being free

Three tiers from one controller is a lot of mileage from one option flag, but there are real limits:

- **Tier C requires `<form method="POST">` for every action.** Pure-button no-form interactions (e.g., `<button onclick="...">`) skip the browser's form-encoding step and have no Tier C path.
- **`Change()` is JS-only.** Live debounced input updates on every keystroke can't happen without JS — Tier C readers won't see incremental feedback. The `submit`-on-blur fallback is what they get.
- **Multi-user peer fan-out is WS-only by default.** `Publish` to a subscribed topic reaches other connections via the WebSocket transport; HTTP fetch is request-scoped. Tier B users see *their own* updates, not peers'.
- **Streamed responses require WS.** Anything range-streamed (SSE-style) needs the persistent connection.

These are the cliffs. For the 80% of CRUD forms that don't need any of them, the recipe shape — one controller, three transports, one option flag — covers all three tiers without conditionals.

## What next?

- [Counter, deeper](/recipes/counter) — the broadcast story Tier B doesn't get for free.
- [Todos: a real LiveTemplate app](/recipes/todos) — the same form-action shape with auth, persistence, and components on top.
- [Reference — Server Actions](/reference/server-actions) — the action resolution rules in detail (form name, button name, action priority).
- [Reference — POST-Redirect-GET](/reference/prg) — the framework's PRG implementation, including flash cookies and redirect target rules.
