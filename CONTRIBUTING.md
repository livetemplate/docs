# Contributing to livetemplate/docs

This repo is the source for the LiveTemplate documentation website. Most content lives in other repos and is synced here automatically — read this guide before editing anything under `content/`.

## Where content lives

- **`content/_meta/source-of-truth.md`** (created in Phase 2) — for every page, names the canonical source file and repo. **Edit content in the canonical source repo, not here**, unless this repo is the canonical source.
- **`content/recipes/`** — recipes are authored directly in this repo; this is the canonical home for them.
- **`content/patterns/index.md`** — catalog page; pattern detail pages are proxied to the deployed `examples/patterns` app via tinkerdown's external-app router (Phase 4).
- **Everything else under `content/`** — mirrored from a source repo. Don't edit here; edit the source. The Phase 3 sync action will overwrite local edits.

## Workflow

1. Determine where the content is canonical (check `content/_meta/source-of-truth.md`).
2. If canonical here → edit, open PR.
3. If canonical elsewhere → open a PR in the source repo. The next release will sync the change here automatically (Phase 3 onwards).

## Local dev

```bash
tinkerdown serve content/ --watch
```

## Validation & tests

```bash
tinkerdown validate content/
```

Browser e2e tests live in `e2e/` (added in Phase 1) and run via chromedp.

## Conventions

- No inline `<style>` or `<script>` (CSP-clean).
- Frontmatter required on every page (`title:` minimum).
- Use site-relative links (`/reference/...`), not full GitHub URLs.
- Pages mirrored from source repos must keep their `source_repo` and `source_path` frontmatter — these power the "Edit on GitHub" link.
