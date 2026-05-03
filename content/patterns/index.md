---
title: "Patterns"
---

# Patterns

A catalog of **33 reactive UI patterns** built with LiveTemplate. Each pattern is a self-contained handler demonstrating a single idiom — forms, lists, navigation, real-time, and more.

Pattern detail pages open the **live demo**, served from a dedicated [`lt-patterns.fly.dev`](https://lt-patterns.fly.dev/) deployment of the [`livetemplate/examples/patterns/`](https://github.com/livetemplate/examples/tree/main/patterns) app. The docs site reverse-proxies the demo so you can interact with each pattern without leaving this site.

> Looking for the catalog as machine-readable data? Fetch [`lt-patterns.fly.dev/api/index.json`](https://lt-patterns.fly.dev/api/index.json) — versioned schema, CORS-enabled.

## Forms & Editing

- [Click To Edit](/patterns/forms/click-to-edit) — Toggle between view and edit mode
- [Edit Row](/patterns/forms/edit-row) — Inline editing of table rows
- [Inline Validation](/patterns/forms/inline-validation) — Server-side field validation as you type
- [Bulk Update](/patterns/forms/bulk-update) — Batch checkbox operations
- [Reset User Input](/patterns/forms/reset-input) — Auto-clear forms after submission
- [File Upload](/patterns/forms/file-upload) — Standard and chunked file uploads
- [Preserving File Inputs](/patterns/forms/preserve-inputs) — Retain form values across re-renders

## Lists & Data

- [Delete Row](/patterns/lists/delete-row) — Animated row removal
- [Click To Load](/patterns/lists/click-to-load) — Append-only pagination
- [Infinite Scroll](/patterns/lists/infinite-scroll) — Auto-load on scroll with IntersectionObserver
- [Value Select](/patterns/lists/value-select) — Cascading dependent selects
- [Sortable List](/patterns/lists/sortable) — Drag-and-drop reordering with native HTML5 drag events
- [Large Table](/patterns/lists/large-table) — 10k-row table with filter, sort, append, update, delete, reset (streaming range)

## Search & Filtering

- [Active Search](/patterns/search/active-search) — Debounced live search
- [URL-Preserved Filters](/patterns/search/url-filters) — Bookmarkable filter state via query params

## Loading & Progress

- [Lazy Loading](/patterns/loading/lazy-loading) — Load content after page render via server push
- [Progress Bar](/patterns/loading/progress-bar) — WebSocket-pushed progress updates
- [Async Operations](/patterns/loading/async-operations) — Loading / success / error state machine

## Dialogs, Tabs & Navigation

- [Modal Dialog](/patterns/navigation/modal-dialog) — Native `<dialog>` with `command` / `commandfor`
- [Confirm Dialog](/patterns/navigation/confirm-dialog) — CSP-compliant confirmation flow
- [Tabs (HATEOAS)](/patterns/navigation/tabs) — Server-driven tabs via SPA navigation
- [SPA Navigation](/patterns/navigation/spa-navigation) — Auto link interception with `pushState`
- [Keyboard Shortcuts](/patterns/navigation/keyboard-shortcuts) — Global keyboard event binding

## Visual Feedback

- [Animations](/patterns/feedback/animations) — Entry animations with `lvt-fx:animate`
- [Loading States](/patterns/feedback/loading-states) — Auto `aria-busy` and custom loading text
- [Highlight on Change](/patterns/feedback/highlight) — Visual flash on DOM updates
- [Flash Messages](/patterns/feedback/flash-messages) — Toast notifications via `ctx.SetFlash`

## Real-Time & Multi-User

- [Multi-User Sync](/patterns/realtime/multi-user-sync) — Auto-sync across tabs via `Sync()` handler
- [Broadcasting](/patterns/realtime/broadcasting) — Cross-connection updates via `BroadcastAction`
- [Presence Tracking](/patterns/realtime/presence) — Explicit join / leave with shared state
- [Reconnection Recovery](/patterns/realtime/reconnection) — State persistence across disconnects
- [Live Preview](/patterns/realtime/live-preview) — Real-time input preview via `Change()`
- [Server Push](/patterns/realtime/server-push) — Background goroutine pushing updates

## How this catalog stays in sync

The catalog above is hand-written markdown but the **pattern names + descriptions are sourced from the same `data.go` that drives the patterns app's own index page**. The `/api/index.json` endpoint exposes that data, and any future re-render of this page (or a Phase 5 tinkerdown auto-table recipe) can fetch from there to stay automatically in sync.

If you spot a mismatch between this catalog and what's actually served at `/patterns/<category>/<slug>`, the patterns app is canonical — open an issue or PR against [livetemplate/examples](https://github.com/livetemplate/examples).
