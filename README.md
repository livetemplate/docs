# livetemplate/docs

Source for the LiveTemplate documentation website, built with [tinkerdown](https://github.com/livetemplate/tinkerdown).

Production site: https://livetemplate.fly.dev (Phase 6 onwards)

## How content gets here

Most reference content is mirrored from canonical files in the source repos:

| Section | Source repo | Sync mechanism |
|---|---|---|
| Reference, Guides | [livetemplate/livetemplate](https://github.com/livetemplate/livetemplate) | GitHub Action on release tag |
| TypeScript client docs | [livetemplate/client](https://github.com/livetemplate/client) | GitHub Action on release tag |
| CLI docs | [livetemplate/lvt](https://github.com/livetemplate/lvt) | GitHub Action on release tag |
| Recipes (markdown) | this repo, under `content/recipes/` | Authored directly here |
| Runnable apps + their tests | this repo, under `examples/<slug>/` | Authored directly here |

The source-of-truth matrix lives at `content/_meta/source-of-truth.md`.

Every runnable demo cited by a recipe lives at `examples/<slug>/` —
one folder per app, containing the Go package, template, chromedp
e2e test, and a `cmd/main.go` standalone runner. `cmd/site` imports
these packages and mounts them at `/apps/<slug>/` so tinkerdown's
inline `embed-lvt` blocks can render live widgets on docs pages. See
`CONTRIBUTING.md` → "Adding an example" for the per-app shape.

## Local development

```bash
# Build tinkerdown from source if you haven't already
go install github.com/livetemplate/tinkerdown/cmd/tinkerdown@latest

# Serve the site
tinkerdown serve content/

# Open http://localhost:8080
```

Use `tinkerdown serve content/ --watch` for hot reload while editing.

## Validating content

```bash
tinkerdown validate content/
```

CI runs validate on every PR.

## Build & deploy

Phase 1 wires up the staging Dockerfile + fly.toml. Phase 6 promotes to production at `livetemplate.fly.dev`.

## Plan

The full multi-phase build plan lives at `/home/adnaan/.claude/plans/i-jaunty-boot.md` (in Claude's plan store, not in this repo). Per-phase progress and learnings are tracked at `/home/adnaan/.claude/plans/learnings/`.
