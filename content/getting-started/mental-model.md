---
title: "Mental Model"
description: "How LiveTemplate fits together: templates, controller methods, per-session state, transport fallback, DOM patches, and optional pub/sub."
source_repo: https://github.com/livetemplate/docs
source_path: content/getting-started/mental-model.md
---

# Mental Model

LiveTemplate is server-driven UI for Go applications. You write standard `html/template` markup and a Go controller. The browser sends ordinary form data. The server runs a controller method, re-renders the template, diffs the result, and updates the browser.

The important constraint is also the useful part: start with HTML that works as a normal form POST, then opt into richer behavior only where the workflow needs it.

## One model, every surface

This is the idea the rest of the page elaborates, so it's worth stating once up front.

Every reactive thing LiveTemplate does is the same four-step pipeline:

```
state changes  →  re-render the template  →  diff against the last render  →  patch the browser
```

A user clicking a button runs it. A second tab reacting to that click runs it. A *different user* seeing a live update runs it. The server pushing an update on its own — a timer, a job finishing — runs it. There is no separate "real-time engine," no client-side store, no merge logic to reconcile. The only things that differ between these cases are **when** the action is enqueued and **which** connections receive the resulting patch:

| What happened | When the action runs | Who gets the patch |
|---|---|---|
| This user clicked | Immediately, on their connection | This connection |
| Another tab should follow | After the action, via `ctx.Publish` | Same user's other tabs (or any subscribed connections) |
| Another user should follow | After the action, via `ctx.Publish` to a shared topic | Everyone subscribed to that topic |
| The server decided | Whenever the server calls `session.TriggerAction` | The target connection(s) |

Because it is one pipeline, the program does not get a new shape as it grows. The counter you build in [Your First App](/getting-started/your-first-app) and a multi-user booking system are the same program with more controller methods and a `Subscribe`/`Publish` call — not a single-page app bolted onto a server. You never learn a second model.

## The three files

A small LiveTemplate app usually starts with three files:

- `main.go` wires the template, controller, initial state, and HTTP handler.
- A controller/state Go file defines the state struct and action methods.
- A `.tmpl` file contains ordinary HTML and Go template expressions.

The [Your First App](/getting-started/your-first-app) tutorial uses this shape directly: `main.go`, `counter.go`, and `counter.tmpl`.

## What happens on click

Buttons and forms provide the routing key. A button like this:

```html
<button name="increment">+1</button>
```

dispatches to this controller method:

```go
func (c *CounterController) Increment(state CounterState, ctx *livetemplate.Context) (CounterState, error) {
    state.Count++
    return state, nil
}
```

The method receives the current state and returns the next state. If it returns an error, the state is not committed and the error can be rendered back into the template.

## Where state lives

The controller is where dependencies live: databases, loggers, clients, and other long-lived services. State is per session group and should contain serializable UI state.

By default, anonymous visitors get a stable session group through a cookie. Tabs from the same browser share that group; different browsers get isolated groups. Authenticated apps can define their own grouping through an authenticator.

## Transport fallback

The same controller action can run in three modes:

| Browser capability | What happens |
|---|---|
| JavaScript disabled | The form submits normally and the browser navigates to the server-rendered response. |
| JavaScript enabled, WebSocket disabled | The client intercepts the form, sends HTTP, and patches the DOM in place. |
| JavaScript and WebSocket enabled | The client sends the action over WebSocket and receives DOM patches on the same connection. |

The app code does not need separate handlers for those modes. Progressive enhancement is a transport concern, not a different application model.

## DOM updates

After an action changes state, LiveTemplate renders the template on the server and compares it to the previous render. The browser receives the changed parts and patches the current DOM instead of replacing the whole page.

That means templates remain the source of truth. The browser client is there to preserve focus, submit actions, apply patches, and handle optional client attributes; it is not a second application.

## When to use lvt-* attributes

Use plain HTML first:

- Use `<form method="POST">` for submits.
- Use button `name` attributes for actions.
- Use normal inputs for form data.
- Use links for navigation when a full page navigation is correct.

Reach for `lvt-*` attributes when HTML cannot express the interaction cleanly: debounced input, explicit loading states, client-side DOM effects, click-away behavior, or SPA-style navigation that should keep the current LiveTemplate session.

## When to use pub/sub

Pub/sub is not needed for a single form updating a single tab. Add it when another connection needs to react to an action.

The smallest common case is same-user multi-tab sync:

```go
func (c *Controller) Mount(state State, ctx *livetemplate.Context) (State, error) {
    _ = ctx.Subscribe(ctx.SelfTopic())
    return state, nil
}

func (c *Controller) Save(state State, ctx *livetemplate.Context) (State, error) {
    // mutate state or durable storage
    ctx.Publish(ctx.SelfTopic(), "Refresh", nil)
    return state, nil
}
```

`Subscribe` opts the current connection into a topic. `Publish` sends an action to subscribed peers after the current action succeeds. Without both parts, no peer update happens.

## Next

- [Your First App](/getting-started/your-first-app) builds the counter from scratch.
- [Progressive Complexity](/guides/progressive-complexity) explains when to stay with plain HTML and when to add `lvt-*`.
- [Sync & Server Push](/recipes/sync-and-broadcast) covers `Subscribe`, `Publish`, and server-initiated actions in more detail.
- [Session Reference](/reference/session) documents state persistence and session groups.
