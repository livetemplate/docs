---
title: "Install"
---

# Install

LiveTemplate is a Go framework. The minimum to get a reactive page running is one `go get` and a few lines of code.

## Get the framework

```bash
go get github.com/livetemplate/livetemplate
```

That's enough to write a server-rendered reactive page. The TypeScript client (loaded from CDN by default — see step 2 below) wires up the browser side automatically.

## A counter in three steps

**1. Define controller and state** ([full source](/examples/counter))

```go
type CounterState struct {
    Counter int
}

type CounterController struct{}

func (c *CounterController) Increment(state CounterState, ctx *livetemplate.Context) (CounterState, error) {
    state.Counter++
    return state, nil
}

func (c *CounterController) Decrement(state CounterState, ctx *livetemplate.Context) (CounterState, error) {
    state.Counter--
    return state, nil
}

func main() {
    controller := &CounterController{}
    state := &CounterState{Counter: 0}
    tmpl := livetemplate.Must(livetemplate.New("counter"))
    http.Handle("/", tmpl.Handle(controller, livetemplate.AsState(state)))
    http.ListenAndServe(":8080", nil)
}
```

`New` auto-discovers `*.tmpl` files in the current directory — `counter.tmpl` is picked up automatically.

**2. Write the template** (`counter.tmpl`)

```html
<h1>Counter: {{.Counter}}</h1>
<form method="POST" style="display:inline">
    <button name="increment">+</button>
    <button name="decrement">-</button>
</form>

<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@livetemplate/client@latest/livetemplate.css">
<script defer src="https://cdn.jsdelivr.net/npm/@livetemplate/client@latest/dist/livetemplate-client.browser.js"></script>
```

**3. Run it**

```bash
go run main.go  # Open http://localhost:8080
```

## Optional: install the `lvt` CLI

The `lvt` CLI generates LiveTemplate apps with database scaffolding, auth, and a router pre-wired. Optional — you can write apps directly in Go without it.

```bash
go install github.com/livetemplate/lvt/cmd/lvt@latest
lvt new myapp
cd myapp && lvt serve
```

See the [CLI guide](/cli) for the full command reference.

## What next?

- [Progressive Complexity](/guides/progressive-complexity) — the framework's two-tier model (start with standard HTML, layer on `lvt-*` attributes only when needed).
- [Server API reference](/reference/api) — `New`, `Handle`, `Context`, action method dispatch.
- [Client Attributes reference](/reference/client-attributes) — the full `lvt-*` attribute catalog.
- [Examples](/examples/) — runnable apps for every common pattern.
