---
title: "LiveTemplate"
description: "Build interactive web applications in Go using a simplified programming model. Write server-side code, get reactive UIs automatically."
---

# LiveTemplate

Build interactive web applications in Go using a simplified programming model. Write server-side code, get reactive UIs automatically.

## Get started

- [Install](/getting-started/install) — `go get`, a counter app, three steps to a running page
- [Progressive Complexity](/guides/progressive-complexity) — start with standard HTML; add `lvt-*` attributes only when needed
- [Server API reference](/reference/api) — `New`, `Handle`, `Context`, action method dispatch
- [Client Attributes reference](/reference/client-attributes) — the full `lvt-*` attribute catalog

## Browse

- [Guides](/guides/progressive-complexity) — conceptual walkthroughs, scaling, observability
- [Reference](/reference/api) — types, attributes, configuration
- [CLI (lvt)](/cli) — code generator, dev server, kit system
- [TypeScript Client](/client) — npm package, browser-side runtime
- [Patterns](/patterns/) — 31 interactive UI patterns served live by `examples/patterns`
- [Examples](/examples/) — runnable apps for every common pattern
- [Changelog](/changelog) — releases across all four repos

## How this site is built

This is a [tinkerdown](https://github.com/livetemplate/tinkerdown) site. Most pages are mirrored from canonical files in the source repos ([livetemplate](https://github.com/livetemplate/livetemplate), [client](https://github.com/livetemplate/client), [lvt](https://github.com/livetemplate/lvt), [examples](https://github.com/livetemplate/examples)) and re-published on each release. Pattern detail pages are reverse-proxied to a deployed [livetemplate/examples/patterns](https://github.com/livetemplate/examples/tree/main/patterns) showcase. The "Edit this page on GitHub" link in every footer points to the canonical source — that's where corrections should land.
