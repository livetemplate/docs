---
title: "Source-of-Truth Matrix"
sidebar: false
---

# Source-of-Truth Matrix

This file is the canonical answer to "where does this concept's documentation live, and where is it served from on this docs site?"

It is consumed by humans deciding where to edit content, and (Phase 3 onwards) by the automation that mirrors source-repo markdown into this site on each release.

> **Rule of thumb:** content that explains the framework to **users** is mirrored here from its source repo. Content that explains the **codebase** to **contributors** stays in the source repo. Edit the canonical source — never edit a mirrored page directly.

## How to use this matrix

- **Editing content?** Find the concept, open the canonical source file in its origin repo, edit there. The Phase 3 sync workflow will mirror it on next release. Until then, manual ports require a corresponding edit here.
- **Adding a new concept?** Decide whether it's user-facing (mirror) or contributor-facing (source-only). Add a new row here AND a corresponding entry to `source-of-truth.yaml` (machine-readable; Phase 3 reads this).
- **Removing a concept?** Drop the row, drop the .md file under `content/`, drop the YAML entry.

## Domain glossary

- **Source repo** — origin GitHub repo of the canonical content.
- **Source path** — file path within the source repo (e.g. `docs/guides/progressive-complexity.md`).
- **Site URL** — canonical URL on `https://livetemplate.fly.dev/` (or `livetemplate-docs-staging.fly.dev` while pre-launch).
- **Mirror?** — `yes` (sync from source), `no` (stays in source repo), `recipe` (Phase 5 authors directly here).

## Top-level decisions

| Decision | Choice | Reason |
|---|---|---|
| Source repo CHANGELOGs | Concatenated into a single `/changelog` page on the docs site, sorted by tag date | One reader-facing release log; per-repo CHANGELOGs remain canonical in source |
| `*/CLAUDE.md` files | NEVER mirrored | They're agent instructions, not user docs |
| `docs/proposals/`, `docs/plans/`, `docs/archive/`, `design/`, `WORKFLOWS.md`, `ROADMAP.md`, `AGENT_*.md` | NEVER mirrored | Internal RFCs / planning / agent setup |
| `docs/performance/` | NOT mirrored in v1 | May graduate to public if stabilized |
| `examples/<app>/README.md` | All mirrored under `/examples/<app>` | These are user-facing showcases |
| `examples/patterns/<...>` | NOT mirrored | Served live via reverse proxy at `/patterns/*` (Phase 1 PR-D) |
| `*/CONTRIBUTING.md` | All mirrored under `/contributing/<repo>` | Centralizes contributor onboarding |
| Per-repo `README.md` | Mirrored conditionally; some split (see below) | The four repos each have a different role |

---

## Concept-by-concept matrix

Order: getting-started → guides → reference → CLI → client → patterns/recipes → examples → contributing → changelog.

### Getting Started (mirror)

| Concept | Source repo | Source path | Site URL | Mirror? | Notes |
|---|---|---|---|---|---|
| Install | `livetemplate` | `README.md` (Quick Start section, lines 24–60) | `/getting-started/install` | manual port | Extract just the install steps; the full README also documents `lvt` and the philosophy |
| Your First App | _new content_ | _none_ | `/getting-started/your-first-app` | new (Phase 2 author) | Audit flagged this as missing — write a focused 30-min walkthrough |
| Hello, Recipes | this repo | `content/recipes/hello-world.md` (Phase 5) | `/getting-started/hello` | recipe | Tinkerdown-native interactive intro |

### Guides (mirror)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| Progressive Complexity | `livetemplate` | `docs/guides/progressive-complexity.md` | `/guides/progressive-complexity` | yes |
| Standard HTML Reactivity | `livetemplate` | `docs/guides/standard-html-reactivity.md` | `/guides/standard-html-reactivity` | yes |
| Ephemeral Components | `livetemplate` | `docs/guides/ephemeral-components.md` | `/guides/ephemeral-components` | yes |
| Observability | `livetemplate` | `docs/guides/OBSERVABILITY.md` | `/guides/observability` | yes |
| Scaling | `livetemplate` | `docs/guides/SCALING.md` | `/guides/scaling` | yes |
| New Contributor Walkthrough | `livetemplate` | `docs/guides/new-contributor-walkthrough.md` | `/contributing/livetemplate/walkthrough` | yes (under contributing) |

### Reference (mirror)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| Server API (Go) | `livetemplate` | `docs/references/api-reference.md` | `/reference/api` | yes |
| Client Attributes (`lvt-*`) | `livetemplate` | `docs/references/client-attributes.md` | `/reference/client-attributes` | yes |
| Configuration | `livetemplate` | `docs/references/CONFIGURATION.md` | `/reference/configuration` | yes |
| Session API | `livetemplate` | `docs/references/session.md` | `/reference/session` | yes |
| Server Actions | `livetemplate` | `docs/references/server-actions.md` | `/reference/server-actions` | yes |
| Authentication | `livetemplate` | `docs/references/authentication.md` | `/reference/authentication` | yes |
| File Uploads | `livetemplate` | `docs/references/uploads.md` | `/reference/uploads` | yes |
| Pub/Sub & Broadcasting | `livetemplate` | `docs/references/pubsub.md` | `/reference/pubsub` | yes |
| Error Handling | `livetemplate` | `docs/references/error-handling.md` | `/reference/error-handling` | yes |
| Controller + State Pattern | `livetemplate` | `docs/references/controller-pattern.md` | `/reference/controller-pattern` | yes |
| Navigate (multi-page SPA) | `livetemplate` | `docs/references/navigate.md` | `/reference/navigate` | yes |
| Template Support Matrix | `livetemplate` | `docs/references/template-support-matrix.md` | `/reference/template-support-matrix` | yes |
| Current Limitations | `livetemplate` | `docs/references/current-limitations.md` | `/reference/limitations` | yes |
| Progressive Complexity (quick ref) | `livetemplate` | `docs/references/progressive-complexity-reference.md` | `/reference/progressive-complexity` | yes |

### CLI (`lvt`) (mirror)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| CLI overview | `lvt` | `docs/guides/lvt-cli-guide.md` | `/cli` | yes |
| Auth customization | `lvt` | `docs/guides/auth-customization.md` | `/cli/auth-customization` | yes |
| Component library | `lvt` | `components/README.md` + `components/examples/EXAMPLES.md` | `/cli/components` | yes (merged) |
| Kit system | `lvt` | `internal/kits/system/single/README.md` + `.../multi/README.md` | `/cli/kits` | yes (merged) |
| Testing helpers | `lvt` | `testing/README.md` | `/cli/testing` | yes |
| AI assistant integration | `lvt` | `.github/copilot-instructions.md` | `/cli/ai-assistants` | yes |

### TypeScript Client (mirror, split)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| Client overview + install | `client` | `README.md` (Overview / Install sections) | `/client` | yes (split) |
| Configuration options | `client` | `README.md` (Config section) | `/client/configuration` | yes (split) |
| API surface | `client` | `README.md` (API section) | `/client/api` | yes (split) |
| CSS utilities | `client` | `README.md` (CSS section) | `/client/css` | yes (split) |

### Patterns (Phase 4 — proxied)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| Pattern catalog | this repo | `content/patterns/index.md` | `/patterns/` | manual (this repo) |
| Pattern detail (each of 31) | `examples` | `patterns/handlers_*.go` + templates | `/patterns/<category>/<slug>` | NO — proxied via PR-D router |

### Recipes (Phase 5 — interactive)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| Recipes (8–10 of them) | this repo | `content/recipes/*.md` | `/recipes/<slug>` | recipe (Phase 5 authors directly) |

### Examples (mirror, indexed)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| Examples index | `examples` | `README.md` | `/examples/` | yes |
| counter | `examples` | `counter/README.md` | `/examples/counter` | yes |
| todos | `examples` | `todos/README.md` | `/examples/todos` | yes |
| chat | `examples` | `chat/README.md` | `/examples/chat` | yes |
| avatar-upload | `examples` | `avatar-upload/README.md` | `/examples/avatar-upload` | yes |
| flash-messages | `examples` | `flash-messages/README.md` | `/examples/flash-messages` | yes |
| login | `examples` | `login/README.md` (when present) | `/examples/login` | yes |
| dialog-patterns | `examples` | `dialog-patterns/README.md` (when present) | `/examples/dialog-patterns` | yes |
| live-preview | `examples` | `live-preview/README.md` (when present) | `/examples/live-preview` | yes |
| progressive-enhancement | `examples` | `progressive-enhancement/README.md` | `/examples/progressive-enhancement` | yes |
| shared-notepad | `examples` | `shared-notepad/README.md` (when present) | `/examples/shared-notepad` | yes |
| ws-disabled | `examples` | `ws-disabled/README.md` | `/examples/ws-disabled` | yes |

### Contributing (mirror, per-repo)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| livetemplate contrib | `livetemplate` | `CONTRIBUTING.md` | `/contributing/livetemplate` | yes |
| client contrib | `client` | `CONTRIBUTING.md` | `/contributing/client` | yes |
| lvt contrib | `lvt` | `CONTRIBUTING.md` | `/contributing/cli` | yes |
| examples contrib | `examples` | `CONTRIBUTING.md` | `/contributing/examples` | yes |

### Changelog (manually concatenated)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| Unified changelog | all 4 | `CHANGELOG.md` from each | `/changelog` | manual concat by tag date |

### Deployment (Phase 6, skeleton now)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| fly.io | this repo | `content/deployment/flyio.md` | `/deployment/flyio` | new (Phase 6) |
| Docker | this repo | `content/deployment/docker.md` | `/deployment/docker` | new (Phase 6) |

---

## What stays in source repos (NOT mirrored)

For audit-trail clarity, here is everything intentionally left behind:

- All `*/CLAUDE.md` (livetemplate, lvt, examples)
- `livetemplate/docs/proposals/*` (10+ RFCs)
- `livetemplate/docs/archive/*` (30+ historical docs)
- `livetemplate/docs/performance/*` (4 internal perf reports)
- `livetemplate/docs/specifications/*` (wire-protocol specs — internal in v1; may graduate later)
- `livetemplate/docs/design/*` (ARCHITECTURE, CODE_STRUCTURE, FIRST_PRINCIPLES, etc. — internal-leaning; revisit in Phase 6 whether to surface a subset)
- `livetemplate/docs/roadmap/*` (planning docs)
- `livetemplate/ROADMAP.md`, `livetemplate/.github/RELEASE.md`, `livetemplate/.github/COMMIT_CONVENTION.md`
- `lvt/docs/AGENT_USAGE_GUIDE.md`, `lvt/docs/AGENT_SETUP.md`, `lvt/docs/WORKFLOWS.md`, `lvt/docs/ROADMAP.md`
- `lvt/docs/plans/*`, `lvt/e2e/*`, `lvt/components/CONTRIBUTING.md`
- `examples/docs/plans/*`

---

## Concepts known to need NEW content (audit flagged)

These don't currently exist anywhere and must be written from scratch — likely Phase 2 follow-up sessions or Phase 5/6:

1. **Your First App** (`/getting-started/your-first-app`) — focused 30-min counter walkthrough
2. **Tutorial: Todo CRUD App** (`/tutorials/todo-crud`) — guided rebuild of `examples/todos/`
3. **Best Practices & Patterns** (`/guides/best-practices`) — currently scattered
4. **Migration & Upgrade Guide** (`/guides/migration`) — v0.x → v1.0 future-proofing
5. **Deployment Cookbook** — Phase 6 work (Docker, Fly.io, K8s as separate pages)
6. **Troubleshooting / FAQ** (`/troubleshooting`) — none exists
7. **Glossary** (`/glossary`) — none exists

---

## Cross-repo link rewrite rules

Any markdown content sourced from another repo must have these rewrites applied during port (or by the Phase 3 sync script):

| Pattern | Rewrite |
|---|---|
| `https://github.com/livetemplate/livetemplate/blob/main/docs/<path>.md` | `/<computed>` (look up site URL from this matrix) |
| `https://github.com/livetemplate/livetemplate` | `/` (site root) |
| `https://github.com/livetemplate/client/blob/main/<path>.md` | `/client/<computed>` |
| `https://github.com/livetemplate/client` | `/client` |
| `https://github.com/livetemplate/lvt/blob/main/<path>.md` | `/cli/<computed>` |
| `https://github.com/livetemplate/lvt` | `/cli` |
| `https://github.com/livetemplate/examples/blob/main/<path>.md` | `/examples/<computed>` |
| `https://github.com/livetemplate/examples` | `/examples` |
| `https://github.com/livetemplate/<repo>/(issues|pull|commit|releases)/<n>` | KEEP AS-IS — external GitHub references |
| `https://pkg.go.dev/...` | KEEP AS-IS — external API docs |

When the rewritten URL would 404 on the docs site (e.g. linked file isn't mirrored), the link should fall back to the original GitHub URL — never silently break.
