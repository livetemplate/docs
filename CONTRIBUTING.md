# Contributing to livetemplate/docs

This repo is the source for the LiveTemplate documentation website. Most content lives in other repos and is synced here automatically — read this guide before editing anything under `content/`.

## Where content lives

- **`content/_meta/source-of-truth.md`** — for every page, names the canonical source file and repo. **Edit content in the canonical source repo, not here**, unless this repo is the canonical source.
- **`content/recipes/`** — recipes (the docs *about* runnable apps) are authored directly in this repo. Markdown only — no Go code lives under `content/`.
- **`examples/<slug>/`** — runnable apps. The single canonical home for every recipe's code, template, e2e test, and standalone runner. `cmd/site` imports these packages and mounts them at `/apps/<slug>/`; tinkerdown proxies inline `embed-lvt` blocks here.
- **Everything else under `content/`** — mirrored from a source repo. Don't edit here; edit the source. The sync action will overwrite local edits.

## Workflow

1. Determine where the content is canonical (check `content/_meta/source-of-truth.md`).
2. If canonical here → edit, open PR.
3. If canonical elsewhere → open a PR in the source repo. The next release will sync the change here automatically.

## Adding an example

Every runnable app — whether it has a recipe write-up or not — lives at `examples/<slug>/`. Pattern:

```
examples/foo/
├── handler.go              # exports Handler(opts ...livetemplate.Option) http.Handler
├── foo.tmpl                # //go:embed-ed by handler.go
├── foo_test.go             # chromedp e2e (package foo_test); spawns `go run ./cmd`
└── cmd/
    └── main.go             # standalone runner — supports PORT env + --dev flag
```

Steps:

1. Create the folder + four files (use `examples/counter/` as the template — it's the smallest runnable shape).
2. Optional: add `content/recipes/foo/index.md` if you want a recipe write-up. Cite source via site-rooted includes: `` ```go include="/examples/foo/foo.go" lines="5-15" `` ``.
3. Wire `cmd/site/main.go`: import `"github.com/livetemplate/docs/examples/foo"` and add a `mux.Handle("/apps/foo/", ...)` line.
4. `go build ./...` + `go test ./examples/foo` to confirm.

No `livetemplate/livetemplate` CI changes required — the cross-repo workflow runs `go test ./examples/...` and picks up the new folder automatically.

## Local dev

```bash
tinkerdown serve content/ --watch        # docs site
go run ./examples/counter/cmd --dev      # any example, standalone
make serve                               # both together (cmd/site + tinkerdown)
```

## Validation & tests

```bash
tinkerdown validate content/         # parses every page (includes resolve, etc.)
go test ./cmd/sync/...               # sync tool unit tests
go test ./examples/...               # all example e2e tests (chromedp)
go test ./e2e/...                    # site-level browser tests
```

The `examples/` tests need Docker Chrome (chromedp pulls it on first use); they pass under `-short` by compiling but skip the browser-driven parts.

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

### One-time PAT setup (so syncs auto-open PRs)

The `peter-evans/create-pull-request` step uses `GITHUB_TOKEN` by default, but org-level policy ("Allow GitHub Actions to create and approve pull requests") may block that token from creating PRs. The workflow falls through to a `DOCS_BOT_TOKEN` secret if it exists; configure that PAT once and every sync auto-opens a PR thereafter.

1. Create a fine-grained PAT at https://github.com/settings/personal-access-tokens with:
   - Repository access: `livetemplate/docs` only
   - Repository permissions: `Contents: Read and write`, `Pull requests: Read and write`
2. Add as a repo secret: `gh secret set DOCS_BOT_TOKEN --repo livetemplate/docs --body <token>`
3. Confirm: trigger the workflow manually and verify the "Open sync PR" step succeeds.

Until the PAT is set up, syncs still push their branches; open the PR manually with `gh pr create --head sync/<slug>-<ref> --base main`.

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
