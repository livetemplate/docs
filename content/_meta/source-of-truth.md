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
| `examples/<app>/README.md` | All mirrored under `/recipes/apps/<app>` | These are user-facing showcases |
| `examples/patterns/<...>` | NOT mirrored | Served live via reverse proxy at `/recipes/ui-patterns/*` (Phase 1 PR-D) |
| `*/CONTRIBUTING.md` | All mirrored under `/contributing/<repo>` | Centralizes contributor onboarding |
| Per-repo `README.md` | Mirrored conditionally; some split (see below) | The four repos each have a different role |

---

## Concept-by-concept matrix

Order: getting-started -> guides -> reference -> CLI -> client -> recipes -> contributing -> changelog.

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

### UI Pattern Recipes (proxied)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| UI pattern catalog | this repo | `content/recipes/ui-patterns/index.md` | `/recipes/ui-patterns/` | manual (this repo) |
| UI pattern detail | `examples` | `patterns/handlers_*.go` + templates | `/recipes/ui-patterns/<category>/<slug>` | NO — proxied via PR-D router |

### Recipes (Phase 5 — interactive)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| Recipes (8–10 of them) | this repo | `content/recipes/*.md` | `/recipes/<slug>` | recipe (Phase 5 authors directly) |

### App Recipes (mirrored from examples)

| Concept | Source repo | Source path | Site URL | Mirror? |
|---|---|---|---|---|
| App recipes index | `examples` | `README.md` | `/recipes/apps/` | yes |
| counter | `examples` | `counter/README.md` | `/recipes/apps/counter` | yes |
| todos | `examples` | `todos/README.md` | `/recipes/apps/todos` | yes |
| chat | `examples` | `chat/README.md` | `/recipes/apps/chat` | yes |
| avatar-upload | `examples` | `avatar-upload/README.md` | `/recipes/apps/avatar-upload` | yes |
| flash-messages | `examples` | `flash-messages/README.md` | `/recipes/apps/flash-messages` | yes |
| login | `examples` | `login/README.md` (when present) | `/recipes/apps/login` | yes |
| dialog-patterns | `examples` | `dialog-patterns/README.md` (when present) | `/recipes/apps/dialog-patterns` | yes |
| live-preview | `examples` | `live-preview/README.md` (when present) | `/recipes/apps/live-preview` | yes |
| progressive-enhancement | `examples` | `progressive-enhancement/README.md` | `/recipes/apps/progressive-enhancement` | yes |
| shared-notepad | `examples` | `shared-notepad/README.md` (when present) | `/recipes/apps/shared-notepad` | yes |
| ws-disabled | `examples` | `ws-disabled/README.md` | `/recipes/apps/ws-disabled` | yes |

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

## Literate primitives in mirrored content

Mirrored upstream READMEs may use tinkerdown's literate authoring
primitives (since tinkerdown v0.2.0). The sync tool passes them through
byte-for-byte and mirrors a single adjacency convention so they resolve
correctly post-sync.

### Fence attributes and block kinds (passthrough)

The sync tool's link rewriter operates on full URL matches only. It
does **not** touch fence attributes or non-default block kinds, so
upstream content can use:

| Primitive | Example |
|---|---|
| `include="..."` fence attribute | `` ```go include="/examples/counter/counter.go" lines="5-15" highlight="7" `` |
| `embed-lvt` block | `` ```embed-lvt path="/apps/counter/" upstream="https://lt-firstapp.fly.dev" `` |
| `show-source` / `hide-source` flag | `` ```lvt show-source `` |

These survive sync without modification.

### Frontmatter contract

The sync tool **owns** five provenance keys; whatever the upstream sets
for these is overridden:

- `title` — extracted from upstream (frontmatter `title:`, first H1, or filename fallback)
- `source_repo` — from `source-of-truth.yaml`
- `source_path` — from `source-of-truth.yaml`
- `source_ref` — the `--ref` sync was invoked with (typically a release tag); drives tinkerdown's source-link footer URLs for `include=` blocks
- `source_commit` — `git rev-parse HEAD` at sync time (immutable provenance record)

The sync tool **preserves** an explicit allowlist of upstream keys
(passed through verbatim, with YAML types respected):

- `description` (string)
- `lvt_show_source` (bool — enables show-source default for the page)
- `sidebar` (bool / string)

Anything else upstream sets in frontmatter is **dropped**. This is
intentional — the docs site stays in control of its frontmatter
contract; new keys are added to the allowlist deliberately, in
`cmd/sync/sync.go`.

### Site-rooted includes for cross-tree literate authoring

Recipe pages cite code that lives outside the `content/` tree (under
`docs/examples/<slug>/`). Tinkerdown's include resolver supports a
site-rooted form for this:

```markdown
```go include="/examples/counter/counter.go" lines="9-33"
```

A leading `/` in the include attribute is interpreted as
project-root-relative — resolved against `filepath.Dir(siteRoot)`,
confined to the project root (not the content root). This replaces
the old `_app/` adjacency convention where the sync tool mirrored an
`_app/` folder next to each synced README. Examples now live at a
single canonical location (`docs/examples/<slug>/`); recipe markdown
references them by site-rooted path.

Page-relative includes (`./foo.go`, `../foo.go`, `foo.go`) keep their
v1 behavior: resolved against the markdown file's directory, confined
to the content root.

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
| `https://github.com/livetemplate/<repo>/(issues|pull|commit|releases)/<n>` | KEEP AS-IS — external GitHub references |
| `https://pkg.go.dev/...` | KEEP AS-IS — external API docs |

When the rewritten URL would 404 on the docs site (e.g. linked file isn't mirrored), the link should fall back to the original GitHub URL — never silently break.
