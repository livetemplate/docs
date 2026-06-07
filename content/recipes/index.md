---
title: "Recipes"
description: "Practical LiveTemplate recipes: basics, UI patterns, runnable apps, and deep dives."
source_repo: https://github.com/livetemplate/docs
source_path: content/recipes/index.md
---

# Recipes

Everything practical in these docs lives here. A recipe can be a small core concept, a focused UI behavior, a runnable app, or a deeper explanation of the tradeoffs behind an implementation.

## Basics

Small end-to-end recipes for the core model:

- [Counter](/recipes/apps/counter) — action routing, state updates, and minimal wiring.
- [Progressive enhancement](/recipes/apps/progressive-enhancement) — standard form behavior first, LiveTemplate enhancement on top.
- [HTTP only](/recipes/apps/ws-disabled) — LiveTemplate without WebSocket upgrades.

## UI Patterns

Focused interaction recipes for forms, lists, loading, feedback, navigation, realtime, and search:

- [UI pattern recipes](/recipes/ui-patterns/)

## Apps

Larger runnable recipes with more production shape:

- [Todos](/recipes/todos) — persistence, auth, components, and multi-step workflows.
- [Chat](/recipes/apps/chat) — realtime multi-user messaging.
- [Avatar upload](/recipes/apps/avatar-upload) — file upload flow.
- [Flash messages](/recipes/apps/flash-messages) — page-level notifications.

## Deep Dives

Narrative recipes for internals, tradeoffs, and operational shape:

- [Counter, deeper](/recipes/counter)
- [Pubsub](/recipes/pubsub) — `Subscribe`/`Publish` peer fan-out
- [Server push](/recipes/server-push) — server-initiated `TriggerAction`
- [How a LiveTemplate update flows](/recipes/architecture-flow)
- [How this docs site works](/recipes/how-this-site-works)
