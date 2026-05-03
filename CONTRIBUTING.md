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
tinkerdown validate content/         # parses every page
cd cmd/sync && go test ./...         # sync tool unit tests
cd e2e && go test ./...              # chromedp browser e2e against staging
```

Browser e2e tests live in `e2e/` (added in Phase 1) and run via chromedp against the live staging site at `https://livetemplate-docs-staging.fly.dev/`.

## Sync workflow (Phase 3)

Most pages on this site are mirrored from canonical files in the source repos. The mapping lives in `content/_meta/source-of-truth.yaml` and is consumed by `cmd/sync/`, a Go program that:

1. Reads the YAML to find every entry whose `source_repo` matches the requested repo
2. Shallow-clones that repo at the requested ref into a temp dir
3. For each entry: copies the source file, replaces frontmatter with `source_repo` / `source_path` / `source_commit`, rewrites known cross-repo GitHub URLs to docs-site-relative paths
4. Writes the result into `content/`

The tool is wrapped by `.github/workflows/sync.yml`, which:

- Runs on **`workflow_dispatch`** — manual trigger from the Actions UI; pick a source repo and ref
- Runs on **`repository_dispatch` event `sync-source`** — fired by source-repo release workflows with `{source_repo, ref}` payload
- After running the tool, opens a PR titled `Sync from <repo>@<ref>` with only the content changes

### Manually triggering a sync

```bash
gh workflow run sync.yml \
  -f source_repo=https://github.com/livetemplate/livetemplate \
  -f ref=main
```

Or use the GitHub Actions UI at `https://github.com/livetemplate/docs/actions/workflows/sync.yml`.

### Resolving sync-vs-manual edit conflicts

Sometimes a docs page needs a one-off edit that the next sync would clobber. Two paths:

1. **The page should be docs-native, not mirrored.** Remove its entry from `content/_meta/source-of-truth.yaml`. The page now lives natively in this repo; future syncs leave it alone. (Example: `/getting-started/install` is a hybrid — Quick Start extracted from `livetemplate/README.md` plus docs-site-authored "What next?" section. Native.)
2. **The edit should land in the source repo.** Open a PR against the canonical source file in its origin repo. Mention the docs-site URL in the PR body so the source-repo reviewer knows where the content surfaces. The next sync after merge will mirror the change here.

If a sync PR conflicts with concurrent work in this repo, rebase the sync branch (`git rebase main`) and re-push — the sync action's commit message includes the source ref so the diff is replayable. If the conflict is in a page the operator believes should be docs-native, switch to path 1 above.

## Conventions

- No inline `<style>` or `<script>` (CSP-clean).
- Frontmatter required on every page (`title:` minimum).
- Use site-relative links (`/reference/...`), not full GitHub URLs.
- Pages mirrored from source repos must keep their `source_repo` and `source_path` frontmatter — these power the "Edit on GitHub" link.
