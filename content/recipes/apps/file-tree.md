---
title: "File Tree"
description: "A directory tree rendered by a template that invokes itself — the recursive shape template inlining cannot express, with updates that scope to a single leaf instead of re-sending the branch."
source_repo: "https://github.com/livetemplate/docs"
source_path: "content/recipes/apps/file-tree.md"
---

# File Tree — a template that calls itself

A directory contains directories. Rendering one means rendering its children
the same way, to whatever depth the data happens to go — and you don't know
that depth when you write the template.

Go's `html/template` handles this natively: a `{{define}}` block can invoke
itself. LiveTemplate could not, until v0.19.0, and the reason is worth
understanding because it explains what changed. The full source is
[`examples/file-tree/`](https://github.com/livetemplate/docs/tree/main/examples/file-tree).

## Try it

```embed-lvt path="/apps/file-tree/" upstream="http://localhost:9091" height="520px"
```

Expand a folder or star a file. Every control above is a plain
`<button name="...">` inside a form — no custom attributes, no hand-written
JavaScript.

## The template

The whole recipe is one self-referential block. `node` renders an entry, and
for a directory it loops over the children invoking `node` again:

```html
{{define "node"}}
<li data-key="{{.Path}}">
    {{if .IsDir}}
        <form method="POST" style="display:inline">
            <button name="toggle" value="{{.Path}}">{{if .Expanded}}▾{{else}}▸{{end}} {{.Name}}/</button>
        </form>
        {{if .Expanded}}
            <ul>
                {{range .Children}}{{template "node" .}}{{end}}
            </ul>
        {{end}}
    {{else}}
        <span class="file">
            <form method="POST" style="display:inline">
                <button name="star" value="{{.Path}}">{{if .Starred}}★{{else}}☆{{end}}</button>
            </form>
            {{.Name}}
        </span>
    {{end}}
</li>
{{end}}
```

## Why this used to fail

LiveTemplate normally **inlines** `{{template}}` calls when it parses: it
splices the invoked body into the caller, producing one flat template it can
analyze for statics and dynamics. A self-referential call has no fixed point
to inline toward — expanding `node` yields another `node` to expand, forever.
So recursion was rejected at parse time.

From v0.19.0, the parser first finds templates reachable from themselves and
leaves *those* calls un-inlined, evaluating them at build time instead. The
recursive region becomes nested tree nodes rather than flattened markup, which
is what keeps it inside the reactive tree instead of degrading to opaque HTML.

## Why the update stays small

Because the recursion produces real nested structure, a change deep in the
tree diffs against that structure. Starring `query.go` four levels down sends
an update addressed to that one node — not a re-render of `/internal`, and not
a re-send of the branch containing it. The sibling `migrate.go` isn't just
visually unchanged; its DOM node is never replaced.

That property is what makes deep trees practical. Before per-leaf diffing, a
change anywhere under a branch meant re-sending the branch, which on a
five-level tree is roughly a hundred times more bytes than the change itself.

## Keys matter here

Each `<li>` carries `data-key="{{.Path}}"`, and path — not name — is the right
choice. Sibling names repeat across directories: two `README.md` files in
different folders are different nodes. Keying on name would let the diff
engine confuse them and move the wrong row; a full path cannot collide.

## Depth is capped

Recursion runs until the data stops nesting, so data that refers to itself
would recurse forever. LiveTemplate caps invocation depth at **128** by
default, surfacing an error rather than overflowing the stack. Raise it with
`WithMaxTemplateDepth(n)` or `LVT_MAX_TEMPLATE_DEPTH` only when your data is
legitimately deeper — see the
[template support matrix](/references/template-support-matrix#recursion-depth)
for the full behavior, including why a too-low cap can go unnoticed on first
render.
