# Working in this repo

## Before writing any prose, read `VOICE.md`

Every rewrite of the landing page so far has drifted back to the same flat
register (#65, #83, #135, #136). `VOICE.md` is the spec that exists to stop the
fifth. It covers the register, the banned vocabulary, and the two constructions
the repo owner has called out by name: **claudisms** and **rhetorical
reversals**. It applies to PR descriptions too.

`scripts/voice-check.sh` catches the countable tells. It will not catch a page
that is merely flat — that part is on you.

## Never hand-edit a mirrored page

A page with an upstream `source_repo` in its front matter is overwritten
wholesale by `cmd/sync` on the next release (`cmd/sync/sync.go`, `os.WriteFile`).
Editing it here is work that gets thrown away. Fix it in the source repo.

- Mirrored: `content/reference/*`, `content/changelog/*`, `content/guides/*`
  (except `index.md`), `content/cli/*`, `content/client/*`, `content/contributing/*`
- Docs-native: `content/index.md`, `content/getting-started/*`,
  `content/recipes/**` (including `recipes/apps/*`), the section `index.md` files

`content/_meta/source-of-truth.yaml` is the machine-readable mapping and the
authority when it and `source-of-truth.md` disagree.

## Builds and tests

```bash
GOWORK=off make test          # everything
make test-e2e                 # chromedp, needs the site serving
make sweep SWEEP_URL=...      # sitemap crawl for overflow
```

Verify UI in a real browser, never `curl`. `/recipes/ui-patterns/lists/large-table`
is a 4.9 MB page that times out the sweep on production too — that flag is
pre-existing.

## Editing `content/index.md`

Never put a blank line inside a `<pre>` block. Goldmark ends the HTML block
there and re-parses the rest as markdown: indentation is stripped and `<p>` tags
appear inside the `<code>`. The file carries a note saying so.
