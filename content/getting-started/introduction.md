---
title: "Introduction"
description: "What LiveTemplate is, when to reach for it, and where to go next: reactive web UIs written in standard HTML and Go, with no client framework and no second model."
source_repo: https://github.com/livetemplate/docs
source_path: content/getting-started/introduction.md
---

# Introduction

LiveTemplate builds reactive web UIs in **standard HTML and Go**. You write an
ordinary `html/template` and a small Go controller; the browser sends ordinary
form data; the server re-renders, diffs, and patches the page. There is no
client-side framework to learn, no second state model to keep in sync, and no
build step for the common app screens.

The defining idea is that you never leave HTML. A `<button name="increment">`
*is* the action — you don't annotate it to make it reactive. You reach for an
`lvt-*` attribute only for behavior HTML itself cannot express (a debounce, a
keyboard shortcut, a reactive class toggle), never as boilerplate.

## When LiveTemplate fits

It's a good fit when you're building app screens in Go and want live behavior —
forms with inline validation, multi-tab sync, dashboards that update
themselves, cross-user views — without standing up a separate frontend. Because
the same program works as a plain form POST first, it
[degrades gracefully](/recipes/progressive-enhancement/) to environments with no
JavaScript and upgrades to WebSocket-driven real time where you want it.

It's a weaker fit for highly bespoke client-side interaction (canvas editors,
offline-first apps, animation-heavy UIs) where the logic genuinely belongs in
the browser.

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
