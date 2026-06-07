---
title: "Reconnection Recovery"
description: "Persist state across WebSocket drops, server restarts, and reloads with the lvt:\"persist\" tag."
source_repo: "https://github.com/livetemplate/docs"
source_path: "examples/patterns/templates/realtime/reconnection.tmpl"
---

# Reconnection Recovery

State fields tagged `lvt:"persist"` survive WebSocket disconnects, server restarts,
and full page reloads — the framework restores them from the session store via the
group cookie before the first render after reconnect. The counter and notes here are
both persisted; the notes textarea also uses `lvt-form:preserve` so in-progress
typing isn't lost when another action re-renders the page.

```embed-lvt path="/apps/ui-patterns/realtime/reconnection" upstream="http://localhost:9091" height="400px"
```

## Template

A persisted counter and a persisted notes textarea — reload the page and both come
back.

```html include="/examples/patterns/templates/realtime/reconnection.tmpl"
```

## Handler & state

The actions are ordinary mutations; persistence comes entirely from the `lvt:"persist"`
struct tags.

```go include="/examples/patterns/handlers_realtime.go" region="reconnection"
```

```go include="/examples/patterns/state_realtime.go" region="reconnection-state"
```

## When to use

- Long-lived state that must outlast a flaky connection or a deploy — drafts,
  in-progress forms, accumulated counts.
- A user should return to exactly where they left off after a reload.

For per-connection state that should *not* persist across tabs, see
[Pubsub](/recipes/ui-patterns/realtime/pubsub).
