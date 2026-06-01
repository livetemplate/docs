---
title: "UI Pattern Recipes"
description: "A catalog of focused LiveTemplate UI recipes for forms, lists, loading states, feedback, navigation, realtime, and search."
---

# UI Pattern Recipes

A catalog of focused **UI pattern recipes** built with LiveTemplate. Each recipe is a self-contained handler demonstrating a single idiom — forms, lists, navigation, real-time, and more.

Detail pages open the **live demo**, served by the docs-site recipes binary (`cmd/site`) from the [`docs/examples/patterns/`](https://github.com/livetemplate/docs/tree/main/examples/patterns) package. Tinkerdown reverse-proxies the in-process mount so you can interact with each pattern recipe without leaving this site.

> The catalog below is **rendered live** from the `/recipes/ui-patterns/api/index.json` endpoint that the same package exposes, via tinkerdown's `lvt-source` REST binding. New UI pattern recipes added to the package show up here on the next deploy — no separate sync needed.

```lvt
<div lvt-source="patterns" class="patterns-catalog">
    {{if .Error}}
    <p><mark>Failed to load catalog: {{.Error}}.</mark></p>
    {{else}}
    <p data-test="catalog-summary">Loaded <strong>{{len .Data}}</strong> categories from the in-process patterns endpoint.</p>
    {{range .Data}}
    <section data-category="{{.Slug}}">
        <h2 id="cat-{{.Slug}}">{{.Name}}</h2>
        <ul>
            {{range .Patterns}}
            <li data-test="pattern-row" data-pattern-slug="{{.Slug}}">
                <a href="{{.Path}}"><strong>{{.Name}}</strong></a> — {{.Description}}
            </li>
            {{end}}
        </ul>
    </section>
    {{end}}
    {{end}}
</div>
```

## How this page works

This page is itself a recipe. The catalog above is **not static markdown** — it is a single `lvt-source="patterns"` block bound to a REST data source declared in the site's `tinkerdown.yaml`:

```yaml
sources:
  patterns:
    type: rest
    from: http://localhost:9091/recipes/ui-patterns/api/index.json
    result_path: categories
    cache:
      ttl: 5m
```

On every visit, tinkerdown:

1. Renders a `<div class="loading">Connecting...</div>` placeholder server-side
2. Establishes a WebSocket to the docs server
3. Server fetches `/recipes/ui-patterns/api/index.json` from the in-process recipes binary (cached 5 minutes), exposes the JSON as `.Data`
4. Re-renders the inner template with the fetched data and patches it into the page

The recipes binary (`cmd/site`) compiles the UI pattern catalog metadata directly from `data.go`, so any pattern added to `docs/examples/patterns/` shows up in this catalog on the next deploy — no separate sync needed.

If you spot a mismatch between this catalog and what's actually served at `/recipes/ui-patterns/<category>/<slug>`, the patterns package is canonical — open an issue or PR against [livetemplate/docs](https://github.com/livetemplate/docs).
