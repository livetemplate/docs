---
title: "Patterns"
---

# Patterns

A catalog of **reactive UI patterns** built with LiveTemplate. Each pattern is a self-contained handler demonstrating a single idiom — forms, lists, navigation, real-time, and more.

Pattern detail pages open the **live demo**, served from a dedicated [`lt-patterns.fly.dev`](https://lt-patterns.fly.dev/) deployment of the [`livetemplate/examples/patterns/`](https://github.com/livetemplate/examples/tree/main/patterns) app. The docs site reverse-proxies the demo so you can interact with each pattern without leaving this site.

> The catalog below is **rendered live** from [`lt-patterns.fly.dev/api/index.json`](https://lt-patterns.fly.dev/api/index.json) via tinkerdown's `lvt-source` REST binding. New patterns added to the upstream show up here on the next page load — no docs-site PR needed.

```lvt
<div lvt-source="patterns" class="patterns-catalog">
    {{if .Error}}
    <p><mark>Failed to load catalog: {{.Error}}. Try the <a href="https://lt-patterns.fly.dev/">upstream catalog</a> directly.</mark></p>
    {{else}}
    <p data-test="catalog-summary">Loaded <strong>{{len .Data}}</strong> categories from the upstream API.</p>
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
    from: https://lt-patterns.fly.dev/api/index.json
    result_path: categories
    cache:
      ttl: 5m
```

On every visit, tinkerdown:

1. Renders a `<div class="loading">Connecting...</div>` placeholder server-side
2. Establishes a WebSocket to the docs server
3. Server fetches `/api/index.json` (cached 5 minutes), exposes the JSON as `.Data`
4. Re-renders the inner template with the fetched data and patches it into the page

Adding a new pattern to the upstream `examples/patterns` app — and re-deploying it — automatically updates this catalog on the next visit. There is no docs-site PR required to keep the list in sync.

If you spot a mismatch between this catalog and what's actually served at `/patterns/<category>/<slug>`, the patterns app is canonical — open an issue or PR against [livetemplate/examples](https://github.com/livetemplate/examples).
