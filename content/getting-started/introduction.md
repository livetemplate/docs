---
title: "Introduction"
description: "What LiveTemplate is, when to reach for it, and where to go next: reactive web UIs written in standard HTML and Go, with no client framework and no second model."
source_repo: https://github.com/livetemplate/docs
source_path: content/getting-started/introduction.md
---

# Introduction

Write an `html/template` and a small Go controller. The browser posts a form, the
server re-renders, and only the changed parts of the page get patched.

A client framework does this too, and for a canvas editor it should. But for a
settings screen it means keeping two copies of the same data in sync. That cost
is the one thing this is trying to avoid.

The defining idea is that you never leave HTML. A `<button name="increment">`
*is* the action — you don't annotate it to make it reactive. You reach for an
`lvt-*` attribute only for behavior HTML itself cannot express (a debounce, a
keyboard shortcut, a reactive class toggle), never as boilerplate.

## When LiveTemplate fits

Good fit: app screens in Go that need live behavior. Forms with inline
validation, multi-tab sync, dashboards that update themselves, views shared
across users. All of it without standing up a separate frontend.

The same program works as a plain form POST first, so it
[keeps working](/recipes/progressive-enhancement/) where JavaScript is off, and
picks up WebSockets where you want them.

Weaker fit: canvas editors, offline-first apps, animation-heavy UIs. The logic
really does belong in the browser there, and this would be fighting you.

## How it compares

If you've used other tools, the short version: LiveTemplate keeps HTML standard
and moves reactivity to the server, instead of layering a new vocabulary on top
of it.

- **htmx** — standard HTML actions with no `hx-*`, plus server-owned state and DOM diffing built in.
- **templ + htmx** — Go's own `html/template` instead of a new DSL, with reactivity built in rather than wired up.
- **Alpine.js** — reactive DOM behavior with no `x-*` and no separate client-side state model.
- **Phoenix LiveView** — stateful server-driven UI without leaving Go, and it works over plain HTTP too.
- **React SPA** — reactive workflows for common app screens without a client build step.

## Where to go next

- **[Install](/getting-started/install)** — add LiveTemplate to a Go module.
- **[Your First App](/getting-started/your-first-app)** — build a counter from scratch in about 10 minutes, from plain HTML up to multi-tab sync.
- **[Mental Model](/getting-started/mental-model)** — the one pipeline (state → re-render → diff → patch) that every reactive feature runs on.

Once the model clicks, the [Concepts](/guides/standard-html-reactivity) section
goes a level deeper, and the [Recipes](/recipes/) and
[Apps](/recipes/apps/) sections are copy-paste starting points.
