---
title: "Your First App"
description: "Build a counter from scratch in 10 minutes. Walks through Tier 1 (no JavaScript) → JS client (in-place DOM patching) → multi-tab broadcast."
---

# Your First App

You're going to build a counter. The plain version takes about 5 minutes. The fully reactive multi-tab version takes another 5. By the end you'll have seen every layer of the LiveTemplate model.

> Prerequisite: Go 1.22 or later, and you've already run [`go get github.com/livetemplate/livetemplate`](/getting-started/install) in some directory.

## Step 1 — Set up the project

```bash
mkdir counter && cd counter
go mod init counter
go get github.com/livetemplate/livetemplate
```

You'll have a `go.mod` and an empty directory. We'll add two files: `main.go` and `counter.tmpl`.

## Step 2 — Define a controller and state

Create `main.go`:

```go
package main

import (
    "net/http"
    "github.com/livetemplate/livetemplate"
)

// State is pure data, cloned per session.
type CounterState struct {
    Counter int
}

// Controller holds shared dependencies (here, none) and the action methods.
type CounterController struct{}

func (c *CounterController) Increment(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
    s.Counter++
    return s, nil
}

func (c *CounterController) Decrement(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
    s.Counter--
    return s, nil
}

func main() {
    ctrl := &CounterController{}
    initial := &CounterState{Counter: 0}

    tmpl := livetemplate.Must(livetemplate.New("counter"))
    http.Handle("/", tmpl.Handle(ctrl, livetemplate.AsState(initial)))

    http.ListenAndServe(":8080", nil)
}
```

Two patterns to notice. First, **state is a value type, not a pointer** — controllers receive a copy and return a (possibly modified) copy. The framework manages the swap. Second, **action methods are exported on the controller**, and their names ARE the action names — `Increment` and `Decrement` are what the template will reference.

## Step 3 — Write the template

Create `counter.tmpl`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Counter</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
</head>
<body>
    <main class="container">
        <h1>Counter: {{.Counter}}</h1>
        <form method="POST" style="display:inline">
            <button name="increment">+1</button>
            <button name="decrement" class="secondary">-1</button>
        </form>
    </main>
</body>
</html>
```

`livetemplate.New("counter")` auto-discovers `counter.tmpl` in the current directory. The `<button name="...">` attribute is the routing trigger — clicking `<button name="increment">` posts the form and the framework calls `Increment()`.

## Step 4 — Run it

```bash
go run main.go
```

Open `http://localhost:8080` in your browser. Click `+1`. The page reloads (you'll see a brief flash) and the counter updates. Click `-1`, same flash. **This works without any JavaScript.** Forms POST. The framework re-renders. The browser navigates.

If your back button is enabled and you've never written a server-rendered app before, this might already feel surprising — there's no React, no client framework, no build step, and yet clicking buttons mutates server state and refreshes the view.

## Step 5 — Add the JS client (no more page reloads)

Add two lines inside `<head>` of `counter.tmpl`:

```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@livetemplate/client@latest/livetemplate.css">
<script defer src="https://cdn.jsdelivr.net/npm/@livetemplate/client@latest/dist/livetemplate-client.browser.js"></script>
```

Reload the page in your browser. Click `+1` again. **No flash this time** — the DOM is patched in place. The framework opens a WebSocket, your click sends a frame instead of a form POST, the server diffs the new template render against the previous, and only the changed text node (`Counter: 1` → `Counter: 2`) is sent back as a patch.

Same Go code. Same template. Two lines of HTML promote the experience from server-rendered-with-reload to in-place reactive.

## Step 6 — Multi-tab sync (broadcast)

Open the same page in TWO browser tabs side by side. Click `+1` in the left tab. The right tab doesn't update — the broadcast hasn't been wired yet.

Add ONE line to your `Increment` and `Decrement` methods:

```go
func (c *CounterController) Increment(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
    s.Counter++
    ctx.BroadcastAction("Refresh", nil) // ← add this line
    return s, nil
}

func (c *CounterController) Decrement(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
    s.Counter--
    ctx.BroadcastAction("Refresh", nil) // ← and this one
    return s, nil
}
```

Restart the server (`Ctrl-C` then `go run main.go` again), reload both tabs. Click in one tab. Watch the other one update in real time.

`BroadcastAction("Refresh", nil)` re-runs the page render in every other connected session. The framework's diff engine sends only the changed bytes — even at scale, this stays efficient.

## What you just built

You wrote a counter that:

- works without JavaScript (Tier 1)
- patches the DOM in place when the JS client is loaded
- syncs across browser tabs in real time

…in about 50 lines of Go and HTML, with no build step, no client-side framework, no custom template language.

## What next?

- [Progressive Complexity](/guides/progressive-complexity) — when to reach for `lvt-*` attributes (Tier 2) and when to stay in Tier 1. The decision matters more than you'd guess.
- [How a LiveTemplate Update Flows](/recipes/architecture-flow) — the sequence diagram of what happened between your click and the DOM patch.
- [Patterns catalog](/patterns/) — 33 live, reactive UI idioms you can copy. Forms, lists, search, real-time, navigation, feedback.
- [Server API reference](/reference/api) — `New`, `Handle`, `Context`, action method dispatch.
- [Sync, Broadcast & Multi-User Sessions](/recipes/sync-and-broadcast) — when `Sync()` vs `BroadcastAction()`, and how sessions are scoped.
- [Examples](/examples/) — runnable apps including chat, todos, file uploads, auth.
