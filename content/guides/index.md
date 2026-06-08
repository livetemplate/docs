---
title: "Concepts"
description: "The one LiveTemplate model, explained — standard-HTML reactivity, progressive complexity, the controller & state pattern, error handling, transports, and how an update flows."
source_repo: https://github.com/livetemplate/docs
source_path: content/guides/index.md
---

# Concepts

One model, explained from a few angles — plus the how-it-works deep dives. Start with the first two if you've just finished [Learn](/getting-started/); reach for the rest when you need the detail.

## The model

- [Standard HTML Reactivity](/guides/standard-html-reactivity) — why plain HTML is the baseline and reactivity layers on top.
- [Progressive Complexity](/guides/progressive-complexity) — start simple, add capability only when an interaction needs it.
- [The lvt-* Decision Tree](/recipes/progressive-complexity-tree) — choose plain HTML first and add `lvt-*` attributes only when required.

## Writing apps

- [Controller & State Pattern](/reference/controller-pattern) — how controllers and per-session state fit together.
- [Error Handling](/reference/error-handling) — server errors, validation failures, form lifecycle, template error state, and flash messages.
- [Transports & Progressive Enhancement](/recipes/progressive-enhancement/) — staying functional across full-live, HTTP-fetch, and raw form-POST from one controller and one template.

## How it works

- [The Update Flow](/recipes/architecture-flow) — a single action, step by step: browser event → controller method → render → tree diff → DOM patch.
- [Pubsub](/recipes/pubsub) — `Subscribe`/`Publish` peer fan-out across a session's connections.
- [Server push](/recipes/server-push) — server-initiated `TriggerAction` from goroutines, timers, and background jobs.
