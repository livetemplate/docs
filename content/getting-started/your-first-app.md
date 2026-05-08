---
title: "Your First App"
description: "Build a counter from scratch in 10 minutes. Walks through Tier 1 (no JavaScript) → JS client (in-place DOM patching) → multi-tab broadcast."
source_repo: https://github.com/livetemplate/docs
source_path: content/getting-started/your-first-app.md
---

# Your First App

You're going to build a counter. The plain version takes about 5 minutes. The fully reactive multi-tab version takes another 5. By the end you'll have seen every layer of the LiveTemplate model — and you'll have been clicking the same widget you wrote, embedded right in this page.

> Prerequisite: Go 1.22 or later, and you've already run [`go get github.com/livetemplate/livetemplate`](/getting-started/install) in some directory.

## Step 1 — Set up the project

```bash
mkdir counter && cd counter
go mod init counter
go get github.com/livetemplate/livetemplate
```

You'll have a `go.mod` and an empty directory. We'll add three files: `counter.go` (state and handlers), `main.go` (wiring), and `counter.tmpl` (the template).

## Step 2 — Define the state and handlers

Create `counter.go`. First the state:

```go include="./_app/counter/counter.go" lines="5-11"
```

State is a value type, not a pointer — controllers receive a copy and return a (possibly modified) copy. The framework manages the swap.

Then a controller and two action methods:

```go include="./_app/counter/counter.go" lines="13-33"
```

Action methods are exported on the controller, and their names ARE the action names — `Increment` and `Decrement` are what the template will reference. The `BroadcastAction` calls are how multi-tab sync works (Step 6).

Now wire it up in `main.go`:

```go include="./_app/counter/main.go" lines="25-52"
```

`livetemplate.New("counter")` parses `counter.tmpl` from the same directory. `tmpl.Handle(controller, AsState(initial))` is the standard wiring — controller for actions, initial state for new sessions.

The `WithAuthenticator(sharedAuth{})` option uses a constant-groupID authenticator so all connections share state — Step 6 has the why and the `sharedAuth` definition.

## Step 3 — Write the template

Create `counter.tmpl`:

```html include="./_app/counter/counter.tmpl"
```

The `<button name="increment">` attribute is the routing trigger — clicking that button posts the form and the framework calls `Increment()` on the controller.

The two `<link>` and `<script>` tags in `<head>` load the LiveTemplate JS client; we'll see what they do at Step 5.

## Step 4 — Run it

```bash
go run .
```

Open `http://localhost:9090` in your browser to see your local counter. Or click `+1` and `-1` right here — a hosted copy of the same source files (`lt-firstapp.fly.dev`) running below:

```embed-lvt path="/apps/counter/" upstream="https://lt-firstapp.fly.dev"
```

Click and the count changes — no full-page reload, just a DOM patch streamed over WebSocket. That's the JS client at work.

## Step 5 — Tier 1: it works without JavaScript

Remove these two lines from the template:

```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@livetemplate/client@latest/livetemplate.css">
<script defer src="https://cdn.jsdelivr.net/npm/@livetemplate/client@latest/dist/livetemplate-client.browser.js"></script>
```

…and the counter still works. Each click does a full form POST and page reload (you'll see a brief flash). The framework re-renders. The browser navigates. **No JavaScript needed.**

This is LiveTemplate's Tier 1: forms POST, server re-renders, browser navigates. Add the JS client back (the two CDN lines) and the framework opens a WebSocket — your click sends a frame instead of a form POST, the server diffs the new render against the previous, and only the changed text node (`Counter: 1` → `Counter: 2`) is sent back as a patch.

Same Go code. Same template. Two lines of HTML promote the experience from server-rendered-with-reload to in-place reactive.

## Step 6 — Multi-tab sync (broadcast)

Look at the handlers from Step 2 — note the highlighted lines:

```go include="./_app/counter/counter.go" lines="22-33" highlight="24,31"
```

`ctx.BroadcastAction("Increment", nil)` (and the matching `Decrement`) tells LiveTemplate to apply the same action on every other connected client — multiple tabs, multiple embeds, multiple users. Without it, each session has its own count; with it, they stay in lockstep.

To prove it, here are two embeds against the same counter, side by side:

<div class="firstapp-side-by-side" style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">

```embed-lvt path="/apps/counter/" upstream="https://lt-firstapp.fly.dev" session="counter-tour" height="200px"
```

```embed-lvt path="/apps/counter/" upstream="https://lt-firstapp.fly.dev" session="counter-tour" height="200px"
```

</div>

Click `+1` in one — watch the other update in real time. They're talking to the same upstream session, and `BroadcastAction` is what makes them stay synced. (On a narrow viewport the embeds stack vertically — the broadcast still works.)

> **Why constant-groupID auth?** Here's the `sharedAuth` referenced in `main.go`:
>
> ```go include="./_app/counter/main.go" lines="11-23"
> ```
>
> Every connection lands in the same session group, so `BroadcastAction` from any one client reaches all the others. A real app uses a per-user authenticator; for a tutorial counter served alongside the docs, putting everyone in one group is what makes the side-by-side demo visible to every reader.

## What you just built

You wrote a counter that:

- works without JavaScript (Tier 1)
- patches the DOM in place when the JS client is loaded
- syncs across browser tabs and embedded widgets in real time

…in about 50 lines of Go and HTML, with no build step, no client-side framework, no custom template language. The two embeds above? They're the same code rendered live. Every click you've done has gone through your handler, broadcast across, and patched the DOM.

## What next?

- [Progressive Complexity](/guides/progressive-complexity) — when to reach for `lvt-*` attributes (Tier 2) and when to stay in Tier 1.
- [How a LiveTemplate Update Flows](/recipes/architecture-flow) — the sequence diagram of what happened between your click and the DOM patch.
- [Patterns catalog](/patterns/) — 33 live, reactive UI idioms you can copy. Forms, lists, search, real-time, navigation, feedback.
- [Server API reference](/reference/api) — `New`, `Handle`, `Context`, action method dispatch.
- [Sync, Broadcast & Multi-User Sessions](/recipes/sync-and-broadcast) — when `Sync()` vs `BroadcastAction()`, and how sessions are scoped.
- [Examples](/examples/) — runnable apps including chat, todos, file uploads, auth.
