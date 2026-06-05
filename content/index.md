---
title: "LiveTemplate"
description: "A Go library for reactive web UIs. Write a Go template and a controller struct; when state changes, only the diff is sent to the browser. The same code runs over a plain form POST, fetch, or WebSocket."
source_repo: https://github.com/livetemplate/docs
source_path: content/index.md
---

# Reactive web UIs in standard HTML and Go

LiveTemplate is a Go library for building reactive web UIs from standard `html/template` templates. You write a template and a controller struct; when state changes, the template re-renders on the server and only the diff is sent to the browser. The same code runs in three browser modes: JavaScript disabled, JavaScript enabled with WebSocket disabled over HTTP, and JavaScript plus WebSocket enabled.

What sets it apart from htmx, Alpine, Phoenix LiveView, or templ + htmx: **you don't annotate HTML to make it reactive, and you don't learn a new templating language.** A `<button name="increment">` *is* the action — standard HTML is the interface. You reach for a framework attribute (`lvt-*`) only when the behavior is something HTML itself cannot define — debounced input, a keyboard shortcut, a reactive class toggle — never as boilerplate to make ordinary HTML work. And the same render-and-diff pipeline that updates one tab also drives multi-tab sync, cross-user real-time, and server push: the program you write for a counter is the program you write for a live booking system. **You never learn a second model.**

> **Alpha** — core features work and are tested, but the API may change before v1.0.

## Try it

```embed-lvt path="/apps/counter-basic/" upstream="http://localhost:9091" height="140px"
```

Click the buttons. Each click POSTs the action to the Go server; the server runs `Increment`, re-renders the template, diffs against the previous render, and sends only the changed text node back. The form, the buttons, and the count display are never re-created — only the count's text changes, patched into the page over a WebSocket with no full reload.

The widget above is a real, deployed LiveTemplate app — the same code as Steps 1–5 of the [Your First App](/getting-started/your-first-app) tutorial, embedded inline through tinkerdown's auto-proxy.

## The code that runs the demo above

Start with the markup — it is just standard HTML, `counter.tmpl`:

```html include="/examples/counter-basic/counter.tmpl"
```

No `hx-*`, no directives, no client code. A button's `name` attribute IS the routing key — `<button name="increment">` posts `increment` and LiveTemplate dispatches to the `Increment` method on the controller.

And the Go behind it — `counter.go`:

```go include="/examples/counter-basic/counter.go" lines="10-31"
```

The protocol between HTML and Go is just the form data the browser already sends.

[Read the full walkthrough →](/getting-started/your-first-app)

## Next level: real-time multi-tab sync

The counter above reacts within a single tab. The same app becomes *real-time across tabs* by adding two server-side calls — no client-side code, no extra dependencies:

```embed-lvt path="/apps/counter/" upstream="http://localhost:9091" height="140px"
```

Open this page in a second tab and click `+1` in either one — the count stays in sync across both tabs in real time. Two additions to `counter.go` make that happen: a `Mount` that opts the connection in with `ctx.Subscribe(ctx.SelfTopic())`, and a `ctx.Publish(ctx.SelfTopic(), ...)` at the end of each handler that fans the action out to every other tab in the same session (highlighted below):

```go include="/examples/counter/counter.go" lines="17-45" highlight="20,32,41"
```

[Counter, deeper](/recipes/counter) unpacks the session-group routing, why `AnonymousAuthenticator` is the right default for public demos, and where peer fan-out stops scaling.

## And the same code goes cross-user

Multi-tab sync and *cross-user* real time are the same mechanism — only the topic changes. Swap the per-session topic for a shared one and a selection by one person appears live in everyone else's browser. The [Seat Picker](/recipes/apps/seat-picker) recipe is a full booking hall built this way: every interaction is a plain `<button name="selectSeat" value="A5">`, there is no client code, and two different users watch each other's seats fill in real time. It is the proof that the program does not change shape as it grows from a counter to a multi-user app.

## What happens between a click and a DOM update

```mermaid
sequenceDiagram
    participant Browser
    participant Server

    Browser->>Server: User clicks button<br/>{action: "add", form: {title: "Buy milk"}}
    Note over Server: Add() returns new state<br/>(Items: [...] → [..., new])
    Note over Server: Tree diff calculated<br/>Only changed values sent
    Server->>Browser: {patches: [...]}
    Note over Browser: DOM patched in place<br/>(no full re-render)
```

When a user clicks a button, LiveTemplate calls a method on your Go struct, diffs the template output against the previous render, and sends only what changed.

[See the full architecture walkthrough →](/recipes/architecture-flow)

## Get started

1. **[Install](/getting-started/install)** — `go get`, ~30 seconds
2. **[Your First App](/getting-started/your-first-app)** — counter app from scratch in 10 minutes
3. **[Progressive Complexity](/guides/progressive-complexity)** — when to reach for `lvt-*` attributes (and when not to)
4. **[Recipes](/recipes/)** — basics, UI patterns, runnable apps, and deep dives

## How to place it

| If you're using... | LiveTemplate gives you... |
|---|---|
| htmx | Standard HTML actions — no `hx-*` attributes — with server-owned state and DOM diffing |
| templ + htmx | Go's own `html/template` instead of a new DSL, and reactivity built in instead of `hx-*` wiring |
| Alpine.js | Reactive DOM behavior without `x-*` directives or a separate client-side state model |
| Phoenix LiveView | Stateful server-driven UI without leaving Go — and it works over plain HTTP too |
| React SPA | Reactive workflows without a client build step for common app screens |

The pattern across that table: other tools make HTML reactive by adding a layer to it — attributes (`hx-*`, `x-*`, `phx-*`) or a DSL (templ). LiveTemplate keeps the HTML standard and moves the reactivity to the server, where one render-and-diff pipeline already lives. The composability comes from that simplified model, not from a new language. See [Standard HTML Reactivity](/guides/standard-html-reactivity) for the side-by-side.

LiveTemplate is not trying to replace every client app. It is a better fit when the server should own state, HTML should remain the primary interface, and progressive enhancement matters. For current constraints, see [Current Limitations](/reference/limitations).

## Or browse

- **[Mental Model](/getting-started/mental-model)** — how templates, controllers, sessions, transports, and pub/sub fit together
- **[Guides](/guides/progressive-complexity)** — conceptual walkthroughs, scaling, observability
- **[Reference](/reference/api)** — types, attributes, configuration, controller pattern
- **[CLI (`lvt`)](/cli)** — code generator, dev server, kit system
- **[TypeScript Client](/client)** — `@livetemplate/client` npm package
- **[Recipes](/recipes/)** — basics, UI patterns, runnable apps, and deep dives
- **[Changelog](/changelog)** — releases across all four repos

## How this site is built

This is a [tinkerdown](https://github.com/livetemplate/tinkerdown) site. Most reference and package pages are mirrored from canonical files in the source repos ([livetemplate](https://github.com/livetemplate/livetemplate), [client](https://github.com/livetemplate/client), [lvt](https://github.com/livetemplate/lvt), [examples](https://github.com/livetemplate/examples)) and re-published on each release. Recipe apps and UI pattern recipes are served by the docs-site recipes binary so the examples stay interactive inside the docs. The "Edit this page on GitHub" link in every footer points to the canonical source — that's where corrections should land. See [How This Docs Site Works](/recipes/how-this-site-works) for the full dogfood loop.
