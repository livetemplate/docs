---
title: "Live Framework Releases"
description: "A live REST-backed docs page that fetches recent LiveTemplate releases from GitHub through tinkerdown sources."
---

# Live Framework Releases

The latest LiveTemplate releases, fetched from the GitHub Releases API on page load. This page proves the docs site can bind to **any** REST endpoint, not just the patterns API — same `lvt-source` block type, different upstream.

```lvt
<div lvt-source="releases" class="releases-feed">
    {{if .Error}}
    <p><mark>Could not fetch releases: {{.Error}}</mark></p>
    <p>The GitHub API rate-limits unauthenticated requests to 60/hour. If you're seeing this and the docs site has been busy, try again in a few minutes — or visit <a href="https://github.com/livetemplate/livetemplate/releases">the releases page directly</a>.</p>
    {{else if not .Data}}
    <p>No releases found.</p>
    {{else}}
    <p data-test="releases-summary">Showing the latest <strong>{{len .Data}}</strong> releases of <code>livetemplate/livetemplate</code>.</p>
    <table>
        <thead>
            <tr><th>Tag</th><th>Published</th><th>Highlights</th></tr>
        </thead>
        <tbody>
            {{range .Data}}
            <tr data-test="release-row" data-tag="{{.TagName}}">
                <td><a href="{{.HtmlUrl}}"><strong>{{.TagName}}</strong></a></td>
                <td>{{.PublishedAt}}</td>
                <td>
                    {{if .Name}}<em>{{.Name}}</em>{{end}}
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
    {{end}}
</div>
```

## How this page works

This page binds to a `releases` source declared in the docs site's `tinkerdown.yaml`:

```yaml
sources:
  releases:
    type: rest
    from: https://api.github.com/repos/livetemplate/livetemplate/releases
    cache:
      ttl: 1h          # GitHub allows 60 unauth requests/hour; 1h cache keeps us well under
```

Two things to notice:

- **Cache TTL of 1 hour.** GitHub's unauthenticated rate limit is 60 requests/hour from a given IP. Without caching, every page view would hit the API directly. The 1h cache means even at hundreds of visits per hour, the API is only called once per hour per docs-site machine.
- **No authentication.** `releases` is a public endpoint, so no `headers:` block is needed. For private endpoints (a private repo, an internal API behind a token), tinkerdown's `headers:` block supports `${ENV_VAR}` expansion so secrets never live in the YAML.

## Adding more REST sources

The pattern generalises. Want to surface a status page, a build dashboard, or your own API? Add another source:

```yaml
sources:
  status:
    type: rest
    from: https://your-status.example.com/api/incidents
    cache:
      ttl: 5m
    headers:
      Authorization: "Bearer ${STATUS_TOKEN}"
    result_path: data.incidents
```

Then any markdown page can render it with `<table lvt-source="status" lvt-columns="title,impact,started_at">` — no Go code, no template registration, just a YAML entry and an attribute.

For more REST source options (`result_path`, `query_params`, `retry`, `auto_bind: false`), see the [tinkerdown source reference](https://github.com/livetemplate/tinkerdown).
