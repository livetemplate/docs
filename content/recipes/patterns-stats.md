---
title: "Anatomy of the Patterns Catalog"
---

# Anatomy of the Patterns Catalog

The same `patterns` data source that powers the [catalog page](/patterns/) drives this stats overview — a different view of identical data. This is the "one source, many views" pattern: a docs site can present REST data from many angles without re-fetching, re-modelling, or duplicating the source.

```lvt
<div lvt-source="patterns" class="patterns-stats">
    {{if .Error}}
    <p><mark>Failed to load: {{.Error}}</mark></p>
    {{else}}

    <p data-test="stats-summary">
        <strong>{{len .Data}}</strong> categories of reactive UI patterns. Click any to jump to its section in the catalog.
    </p>

    <table>
        <thead><tr><th>Category</th><th>Patterns</th><th>Sample</th></tr></thead>
        <tbody>
        {{range .Data}}
            <tr data-test="cat-row" data-category="{{.Slug}}">
                <td>
                    <a href="/patterns/#cat-{{.Slug}}"><strong>{{.Name}}</strong></a>
                </td>
                <td>{{len .Patterns}}</td>
                <td>
                    {{if .Patterns}}
                        <em>{{(index .Patterns 0).Name}}</em>
                        {{if gt (len .Patterns) 1}}, {{(index .Patterns 1).Name}}{{end}}
                        {{if gt (len .Patterns) 2}}, …{{end}}
                    {{end}}
                </td>
            </tr>
        {{end}}
        </tbody>
    </table>

    <h2 id="status-mix">Pattern status</h2>
    <p>
        Every pattern declares a <code>status</code> field — <em>stable</em> (production-ready) or <em>soon</em> (planned).
        The catalog filters on this so newcomers see what they can actually use today.
    </p>

    {{end}}
</div>
```

## How this page works

This page binds to the **same `patterns` source as `/patterns/`** — declared once in `tinkerdown.yaml` and reused by any number of pages. The catalog renders a per-category listing; this page renders a tabular summary with sample patterns per category.

The pattern that matters: **the source is the contract**. Adding a per-status view, a "patterns added this month" view, or an embedded auto-table on the home page — all of them read from `lvt-source="patterns"` without re-fetching the API or re-implementing the catalog's data model. If the upstream API ever adds a new field, every page that wants to surface it picks it up by re-rendering its template.

For richer aggregation (group_by, sum, avg) the right tool is a **computed source** declared in `tinkerdown.yaml`:

```yaml
sources:
  patterns:
    type: rest
    from: https://lt-patterns.fly.dev/api/index.json
    result_path: categories
  patterns_by_status:
    type: computed
    from: patterns
    group_by: status
    aggregate:
      count: count()
```

Computed sources move the work into Go and out of every render path — useful when the dataset is large enough that template-side aggregation gets noticeable.
