---
title: "Changelog"
description: "Release history for the LiveTemplate ecosystem — the Go library, the browser client, and the lvt CLI."
source_repo: https://github.com/livetemplate/docs
source_path: content/changelog.md
---

# Changelog

LiveTemplate ships as three released pieces, each with its own history. Every
page below is mirrored straight from that repository's `CHANGELOG.md` on
release, so it stays current by construction rather than by anyone remembering
to copy it across.

- [**Core library**](/changelog/livetemplate) — `github.com/livetemplate/livetemplate`.
  The Go server: the reactive tree, actions, sessions, uploads, pub/sub.
- [**Browser client**](/changelog/client) — `@livetemplate/client`.
  The TypeScript client: DOM morphing, transports, `lvt-*` attribute behaviour.
- [**CLI**](/changelog/cli) — `github.com/livetemplate/lvt`.
  Project scaffolding, kits, components, and the testing helpers.

## How the versions relate

The server and the client are **pinned to each other**. There is no runtime
handshake between them, so each server release names the exact client it is
wire-compatible with — `livetemplate.ClientVersion`, which is what
`{{lvtClientScriptURL}}` renders. Upgrading the Go module is what moves the
client, so the two stay in lockstep. See
[Install](/getting-started/install) for why that matters and how to
self-host at the pinned version.

Because of that pairing the client's major.minor tracks the core library's:
client 0.20.0 accompanies livetemplate v0.20.0, and a client release will jump
its minor version to re-sync after drifting. The CLI versions independently.
