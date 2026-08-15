---
title: "Contributing to examples"
source_repo: "https://github.com/livetemplate/docs"
source_path: "content/contributing/examples.md"
---

# Contributing to examples

Examples used to live in `livetemplate/examples`. **GitHub has that repository
archived, so it is read-only.** Nothing there accepts contributions any more,
and the instructions this page used to carry described a workflow that no
longer exists.

Every runnable app now lives in this repository, under `examples/<slug>/`, next
to the page that documents it. You edit the two together, so a change to one
turns up in review beside the other.

## Where to start

[`CONTRIBUTING.md`](https://github.com/livetemplate/docs/blob/main/CONTRIBUTING.md)
in the docs repo has the current steps — the four files an example needs, how to
mount it in `cmd/site`, and how to run its chromedp test.

Two rules from it are worth repeating here, because getting either wrong has
already cost real work:

- **Include code, don't retype it.** Cite the app with
  `` ```go include="/examples/foo/foo.go" lines="5-15" ``. A pasted snippet
  drifts the moment either side changes and nothing catches it. The chat recipe
  documented a `Change(ctx *ActionContext)` API that never existed, for as long
  as it was hand-written.
- **Read [`VOICE.md`](https://github.com/livetemplate/docs/blob/main/VOICE.md)
  before writing the prose.** The ten pages under
  [Apps](/recipes/apps/) came from example READMEs and kept that register —
  Title Case headings, emoji, exclamation marks — until a later pass took it
  back out.

## Contributing to the libraries

This page is about example apps. For the framework itself, see
[LiveTemplate](/contributing/livetemplate), [the browser
client](/contributing/client), or [the CLI](/contributing/cli).
