---
title: "LiveTemplate — Build interactive web apps in Go with standard HTML templates"
description: "Write html/template and Go handlers, and the page updates itself. No SPA, no JSON API, no build step."
layout: landing
---

<!-- NOTE FOR EDITORS: never put a blank line inside a <pre> block on this page.
     Goldmark ends an HTML block at the first blank line, so everything after it
     is re-parsed as markdown: indentation is stripped and <p> tags appear
     inside the <code>. Keep snippets compact instead. -->

<header class="site"><div class="hdr-in">
  <a class="logo" href="#top">
    <svg width="19" height="19" viewBox="0 0 20 20" fill="none" aria-hidden="true"><rect x="1.5" y="1.5" width="17" height="17" rx="4" fill="none" stroke="currentColor" stroke-width="1.6"/><path d="M11.2 4.2 L6.6 10.6 H9.6 L8.8 15.8 L13.4 9.4 H10.4 Z" fill="var(--accent)"/></svg>
    LiveTemplate
  </a>
  <nav class="hdr-nav">
    <a href="/getting-started/introduction">Docs</a>
    <a href="/recipes/">Recipes</a>
    <a href="/reference/api">Reference</a>
    <a href="https://github.com/livetemplate/livetemplate">GitHub</a>
    <a class="btn btn-primary" href="/getting-started/install">Get started</a>
  </nav>
</div></header>

<div id="top" class="shell">

<aside class="rail">
  <div class="rail-label">On this page</div>
  <a href="#whole-app">The whole app</a>
  <a href="#steps">Five more steps</a>
  <a class="sub" href="#step-2">2 · No JavaScript</a>
  <a class="sub" href="#step-3">3 · Validation</a>
  <a class="sub" href="#step-4">4 · Loading state</a>
  <a class="sub" href="#step-5">5 · Sync tabs</a>
  <a class="sub" href="#step-6">6 · Shared wall</a>
  <a href="#compare">How it compares</a>
  <a href="#more">Everything else</a>
</aside>

<main class="doc">

<section class="hero">
  <div class="eyebrow">Alpha · a Go library for server-rendered app screens</div>
  <h1>Build interactive web apps in Go with standard HTML templates.</h1>
  <p class="sub">Write <code>html/template</code> and Go handlers, and the page updates itself. The goal is app-like screens without an SPA, a JSON API or a build step, so there's no JavaScript here that you have to write.</p>
  <div class="cta-row">
    <a class="btn btn-primary" href="/getting-started/install">Get started →</a>
    <a class="btn btn-ghost" href="/getting-started/introduction">Read the docs</a>
    <code class="get">go get github.com/livetemplate/livetemplate</code>
  </div>
</section>

<section id="whole-app">
  <div class="eyebrow">Step 1 · Render</div>
  <h2>This is the whole app, running on this page.</h2>
  <p class="lead">Type a name and hit Say hi. The submit calls a Go method, the server re-renders the template, and only the changed HTML comes back.</p>

  <div class="demo">
    <div class="demo-bar"><span class="dot"></span> greet · running in this page</div>
    <div class="demo-body">

```embed-lvt path="/apps/greet/" upstream="http://localhost:9091" height="130px"
```

</div>
  </div>

  <div class="pair">
    <div class="snip">
      <div class="snip-label">app.tmpl — the entire template</div>
<pre class="language-html"><code class="language-html">&lt;!DOCTYPE html&gt;
&lt;html&gt;&lt;head&gt;
  &lt;script defer src="{{lvtClientScriptURL}}"&gt;&lt;/script&gt;
&lt;/head&gt;&lt;body&gt;
  &lt;h1&gt;Hello, {{.Name}}&lt;/h1&gt;
  &lt;form method="POST"&gt;
    &lt;input name="name" placeholder="Your name"&gt;
    &lt;button name="greet"&gt;Say hi&lt;/button&gt;
  &lt;/form&gt;
&lt;/body&gt;&lt;/html&gt;</code></pre>
    </div>
    <div class="snip">
      <div class="snip-label">app.go — the entire program</div>
<pre class="language-go"><code class="language-go">package main
import (
    "net/http"
    lvt "github.com/livetemplate/livetemplate"
)
type State struct{ Name string }
type App struct{}
func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    s.Name = ctx.GetString("name")
    return s, nil
}
func main() {
    app := lvt.Must(lvt.New("app", lvt.WithParseFiles("app.tmpl")))
    http.ListenAndServe(":8080",
        app.Handle(&amp;App{}, lvt.AsState(&amp;State{Name: "there"})))
}</code></pre>
    </div>
  </div>

  <div class="wire">
    <div class="wire-label">on the wire · WebSocket</div>
    <div>▲ action · 40 B <span class="payload">{"action":"greet","data":{"name":"Ada"}}</span></div>
    <div>▼ diff &nbsp;· 20 B <span class="payload">{"tree":{"0":"Ada"}}</span></div>
  </div>

  <p class="close">That's the whole app, about 20 lines of Go and some standard HTML. Everything else on this page is a small diff on it.</p>
</section>

<section id="steps" class="intro">
  <div class="eyebrow">One app, five more steps</div>
  <h2>Adding the things an app usually ends up needing.</h2>
  <p class="lead">Everything below is the same greeting app. It picks up a plain POST fallback, then validation, then a pending state, then live updates over a WebSocket. It stays one Go codebase, and application logic never ends up in two places.</p>
</section>

<section id="step-2" class="step">
  <div class="eyebrow">Step 2 · Works without JavaScript</div>
  <h2>The same app works with JavaScript disabled.</h2>
  <p class="lead">When the script loads, the client enhances the submit and patches the headline. When it doesn't, the same <code>&lt;form&gt;</code> does a native POST and the server renders the page. There's no <code>if jsEnabled</code> branch to write. Both cards below run the same app — the right one has scripting switched off.</p>

  <div class="pair">
    <div class="demo">
      <div class="demo-bar"><span class="dot"></span> JavaScript on · fetch + patch</div>
      <div class="demo-body">
        <iframe class="nojs-frame" src="/apps/greet-nojs/" sandbox="allow-forms allow-same-origin allow-scripts" title="The greeting app with JavaScript enabled"></iframe>
      </div>
    </div>
    <div class="demo">
      <div class="demo-bar"><span class="dot off"></span> JavaScript off · form POST → full render</div>
      <div class="demo-body">
        <iframe class="nojs-frame" src="/apps/greet-nojs/" sandbox="allow-forms allow-same-origin" title="The greeting app with JavaScript disabled"></iframe>
      </div>
    </div>
  </div>

  <div class="snip">
    <div class="snip-label">app.tmpl — one form, either transport</div>
<pre class="language-html"><code class="language-html">&lt;!-- the only line that flips the transport --&gt;
&lt;script defer src="{{lvtClientScriptURL}}"&gt;&lt;/script&gt;
&lt;form method="POST"&gt;   &lt;!-- JS on → fetch + patch · JS off → native POST --&gt;
  &lt;input name="name"&gt;
  &lt;button name="greet"&gt;Say hi&lt;/button&gt;
&lt;/form&gt;</code></pre>
  </div>
</section>

<section id="step-3" class="step">
  <div class="eyebrow">Step 3 · Validation</div>
  <h2>Validation rules written in HTML, re-checked in Go.</h2>
  <p class="lead">Standard attributes like <code>required</code> run again server-side via <code>ctx.ValidateForm()</code>, then you add the rules HTML cannot express. Try an empty submit, or type <em>admin</em>.</p>

  <div class="demo">
    <div class="demo-bar"><span class="dot"></span> greet-validate · server-checked</div>
    <div class="demo-body">

```embed-lvt path="/apps/greet-validate/" upstream="http://localhost:9091" height="160px"
```

</div>
  </div>

  <div class="pair">
    <div class="snip">
      <div class="snip-label">app.tmpl · the rule, written once</div>
<pre class="language-html"><code class="language-html">&lt;input name="name" required {{.lvt.AriaInvalid "name"}}&gt;
{{.lvt.ErrorTag "name"}}</code></pre>
    </div>
    <div class="snip">
      <div class="snip-label">app.go · re-check, then add your own rule</div>
<pre class="language-go"><code class="language-go">func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    if err := ctx.ValidateForm(); err != nil {
        return s, err                 // re-runs the HTML rules
    }
    name := strings.TrimSpace(ctx.GetString("name"))
    if strings.EqualFold(name, "admin") {
        return s, lvt.NewFieldError("name",
            errors.New(`"admin" is reserved`))
    }
    s.Name = name
    return s, nil
}</code></pre>
    </div>
  </div>

  <div class="wire">
    <div class="wire-label">on the wire · HTTP fetch</div>
    <div>▲ <span class="payload">{"action":"greet","data":{"name":"admin"}}</span></div>
    <div>▼ <span class="payload">{"meta":{"errors":{"name":"\"admin\" is reserved"}}}</span></div>
  </div>
</section>

<section id="step-4" class="step">
  <div class="eyebrow">Step 4 · Loading state</div>
  <h2>Two ways to show a pending state.</h2>
  <p class="lead">The slow work runs on the server, so you can render its pending state with ordinary template conditionals. If you'd rather not touch the Go code at all, there's a button-level attribute for that.</p>

  <div class="pair">
    <div>
      <div class="snip-label">A · server-owned, template variables only</div>
      <div class="demo">
        <div class="demo-bar"><span class="dot"></span> greet-async · server-owned pending</div>
        <div class="demo-body">

```embed-lvt path="/apps/greet-async/" upstream="http://localhost:9091" height="130px"
```

</div>
      </div>
<pre class="language-html"><code class="language-html">&lt;button {{if .lvt.Pending}}type="button" aria-busy="true"
  disabled{{else}}name="greet"{{end}}&gt;Say hi&lt;/button&gt;</code></pre>
<pre class="language-go"><code class="language-go">lvt.Async(ctx,
    func(context.Context) (string, error) {
        time.Sleep(700 * time.Millisecond)
        return name, nil
    },
    func(s State, name string, _ error) (State, error) {
        s.Name = name
        return s, nil
    },
)</code></pre>
      <p class="note">No second action to wire up and no <code>Loading</code> field in state, though it does need a live session for the completion render.</p>
    </div>
    <div>
      <div class="snip-label">B · button-level escape hatch</div>
      <div class="demo">
        <div class="demo-bar"><span class="dot"></span> greet-loading · attribute version</div>
        <div class="demo-body">

```embed-lvt path="/apps/greet-loading/" upstream="http://localhost:9091" height="130px"
```

</div>
      </div>
<pre class="language-html"><code class="language-html">&lt;button name="greet"
  lvt-el:addClass:on:pending="is-loading"
  lvt-el:removeClass:on:done="is-loading"&gt;Say hi&lt;/button&gt;</code></pre>
<pre class="language-go"><code class="language-go">func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    time.Sleep(700 * time.Millisecond)
    if name := strings.TrimSpace(
        ctx.GetString("name")); name != "" {
        s.Name = name
    }
    return s, nil
}</code></pre>
      <p class="note">This one keeps pending UI out of server state and works as a single request/response. It's the right one when the spinner is just button chrome rather than something the app cares about.</p>
    </div>
  </div>

  <div class="wire">
    <div class="wire-label">on the wire · A, server-side</div>
    <div>▲ <span class="payload">{"action":"greet","data":{"name":"Ada"}}</span></div>
    <div>▼ <span class="payload">{"tree":{"1":{"aria-busy":"true","disabled":true,"type":"button"}}}</span></div>
    <div>▼ <span class="payload">{"tree":{"0":"Ada","1":{"name":"greet"}}}</span></div>
    <div class="wire-label second">on the wire · B, attribute version</div>
    <div>▲ <span class="payload">{"action":"greet","data":{"name":"Ada"}}</span> &nbsp; ▼ <span class="payload">{"tree":{"0":"Ada"}}</span></div>
  </div>
</section>

<section id="step-5" class="step">
  <div class="eyebrow">Step 5 · Sync your own tabs</div>
  <h2>Keeping your own tabs in sync with two calls.</h2>
  <p class="lead">Subscribe the session to its own topic and publish after the handler runs. The same live session also lets the server push first, without waiting for a click.</p>

  <div class="pipe">
    <span>1 state changes</span><span class="arrow">→</span>
    <span>2 re-render</span><span class="arrow">→</span>
    <span>3 diff vs last render</span><span class="arrow">→</span>
    <span>4 patch the browser</span>
  </div>

  <div class="demo">
    <div class="demo-bar"><span class="dot"></span> greet-wall · WebSocket on</div>
    <div class="demo-body">

```embed-lvt path="/apps/greet-wall/" upstream="http://localhost:9091" height="200px"
```

</div>
  </div>
  <p class="note">Open this page in a second tab, greet in either, and the headline updates in both.</p>

<pre class="language-go"><code class="language-go">func (a *App) Mount(s State, ctx *lvt.Context) (State, error) {
    ctx.Subscribe(ctx.SelfTopic())                 // your tabs share a topic
    s.Name = a.name(ctx.GroupID())
    return s, nil
}
func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    a.setName(ctx.GroupID(), sanitize(ctx.GetString("name")))
    ctx.Publish(ctx.SelfTopic(), "Refresh", nil)   // run Refresh on your other tabs
    return s, nil
}
// Refresh is an ordinary action — the publish above runs it on each peer tab.
func (a *App) Refresh(s State, ctx *lvt.Context) (State, error) {
    s.Name = a.name(ctx.GroupID())
    return s, nil
}</code></pre>

  <p class="lead">There's no magic here. The publish just runs your <code>Refresh</code> method on your other tabs. It re-reads the shared data and returns new state, and the framework does the diffing and patching. The server can start the same cycle itself with <code>sess.TriggerAction("ServerRefresh", nil)</code>.</p>

  <div class="wire">
    <div class="wire-label">on the wire · WebSocket</div>
    <div>▲ this tab &nbsp;&nbsp;<span class="payload">{"action":"greet","data":{"name":"Ada"}}</span></div>
    <div>▼ other tab &nbsp;<span class="payload">{"tree":{"0":"Ada"}}</span></div>
    <div>▼ server push <span class="payload">{"tree":{"3":{"0":"15:04:08"}}}</span> — no ▲; just the changed value goes down</div>
  </div>
</section>

<section id="step-6" class="step">
  <div class="eyebrow">Step 6 · A wall everyone shares</div>
  <h2>Changing the topic makes it cross-user.</h2>
  <p class="lead">Swap the self-topic for a shared one, admitted by a small ACL, and the same publish fans out to every visitor. The two cards below are separate sessions, like two different people. Greet in one and the line shows up on both walls.</p>

  <div class="pair">
    <div class="demo">
      <div class="demo-bar"><span class="dot"></span> visitor 1 · WebSocket on</div>
      <div class="demo-body">

```embed-lvt path="/apps/greet-wall/" upstream="http://localhost:9091" height="200px"
```

</div>
    </div>
    <div class="demo">
      <div class="demo-bar"><span class="dot"></span> visitor 2 · WebSocket on</div>
      <div class="demo-body">

```embed-lvt path="/apps/greet-wall/" upstream="http://localhost:9091" height="200px"
```

</div>
    </div>
  </div>

  <div class="pair">
    <div class="snip">
      <div class="snip-label">app.go · the topic is the only difference</div>
<pre class="language-go"><code class="language-go">func (a *App) Mount(s State, ctx *lvt.Context) (State, error) {
    ctx.Subscribe("wall")           // shared, cross-user topic
    return s, nil
}
func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    a.append(sanitize(ctx.GetString("name")))
    ctx.Publish("wall", "WallRefresh", nil)
    return s, nil
}</code></pre>
    </div>
    <div class="snip">
      <div class="snip-label">app.go · admit the shared topic</div>
<pre class="language-go"><code class="language-go">lvt.WithTopicACL(
    func(topic, _ string, _ *http.Request) (bool, error) {
        return topic == "wall", nil   // deny-all by default
    })</code></pre>
      <p class="note">Each card is its own session, so the headlines stay independent, but the wall is global. It's the same two pubsub calls as step 5, with a different topic.</p>
    </div>
  </div>

  <div class="wire">
    <div class="wire-label">on the wire · WebSocket</div>
    <div>▲ visitor 1 <span class="payload">{"action":"greet","data":{"name":"Ada"}}</span></div>
    <div>▼ visitor 2 <span class="payload">{"tree":{"3":[["a",[{"0":"Ada","1":"15:04"}]]]}}</span></div>
  </div>
</section>

<section class="step">
  <div class="diff-split">
    <div>
      <div class="eyebrow">Only the diff goes over the wire</div>
      <h2>Only the changed values go over the wire.</h2>
      <p class="lead">Templates get split into static structure, which is cached, and dynamic values, so a greeting comes back as <code>{"tree":{"0":"Ada"}}</code> instead of a page.</p>
    </div>
    <div class="bars">
      <div class="bar-row"><span>full HTML</span><span>2.4 KB</span></div>
      <div class="bar full"><span></span></div>
      <div class="bar-row"><span>lvt diff</span><span>340 B</span></div>
      <div class="bar diff"><span></span></div>
      <div class="bar-note">86% smaller per update</div>
    </div>
  </div>
</section>

<section id="compare" class="step">
  <div class="eyebrow">How it compares</div>
  <h2>How this sits next to htmx, templ and LiveView.</h2>
  <p class="lead">Here a plain <code>&lt;button name="greet"&gt;</code> is already the action. <code>lvt-*</code> attributes are an escape hatch for what HTML cannot express, not the main interface.</p>
  <div class="rows">
    <div class="row-k">htmx</div>
    <div class="row-v">A similar HTML-first feel, with server-owned state and diffing built in, so there is less request wiring in the markup.</div>
    <div class="row-k">templ + htmx</div>
    <div class="row-v">Use Go's built-in <code>html/template</code> and keep live behavior in one app model instead of stacking layers.</div>
    <div class="row-k">Alpine.js</div>
    <div class="row-v">Richer behavior without keeping a second copy of state in the browser.</div>
    <div class="row-k">Phoenix LiveView</div>
    <div class="row-v">The same server-driven idea, in Go, and it still falls back to plain HTTP forms.</div>
    <div class="row-k">React SPA</div>
    <div class="row-v">Forms, CRUD, dashboards and shared views without splitting the product into an API and a frontend.</div>
  </div>
</section>

<section id="more" class="step">
  <div class="eyebrow">Everything else</div>
  <h2>Other things in here.</h2>
  <p class="lead">This is aimed at what Go teams actually ship: admin screens, internal tools, CRUD, dashboards, approvals, uploads, auth, and the occasional shared view.</p>
  <div class="links">
    <a href="/reference/uploads"><span class="link-t">File uploads</span><span class="link-g">live progress, same app</span></a>
    <a href="/reference/pubsub"><span class="link-t">Shared views</span><span class="link-g">Subscribe &amp; Publish</span></a>
    <a href="/reference/session"><span class="link-t">Sessions &amp; state</span><span class="link-g">scoped per browser or user</span></a>
    <a href="/reference/error-handling"><span class="link-t">Forms &amp; errors</span><span class="link-g">Go errors, same template</span></a>
    <a href="/cli/"><span class="link-t">Scaffolding</span><span class="link-g">generate common app shapes</span></a>
    <a href="/client/"><span class="link-t">Browser client</span><span class="link-g">transport &amp; DOM patching</span></a>
    <a href="/guides/observability"><span class="link-t">Observability</span><span class="link-g">metrics &amp; tracing hooks</span></a>
    <a href="/guides/scaling"><span class="link-t">Scaling</span><span class="link-g">session groups &amp; fan-out</span></a>
  </div>
  <p class="lead">The <a href="/recipes/ui-patterns/">UI patterns catalog</a> has focused examples: loading states, inline validation, SPA-style navigation, sortable tables, pubsub, presence, server push. This site runs on LiveTemplate itself. <a href="/recipes/how-this-site-works">See how it works</a>.</p>
</section>

<section class="cta">
  <div>
    <h2>Getting started.</h2>
<pre class="install">$ go get github.com/livetemplate/livetemplate</pre>
    <div class="cta-row">
      <a class="btn btn-primary" href="/getting-started/install">Get started →</a>
      <a class="btn btn-ghost" href="/recipes/">Browse recipes</a>
    </div>
    <p class="note">This is alpha: the core works and is tested, but the API may still change before v1.0.</p>
  </div>
  <img class="gopher" src="/assets/gopher-front.svg" alt="The Go gopher" width="132">
</section>

</main>
</div>

<footer class="site"><div class="foot-in">
  <div class="foot-brand">
    <svg width="17" height="17" viewBox="0 0 20 20" fill="none" aria-hidden="true"><rect x="1.5" y="1.5" width="17" height="17" rx="4" fill="none" stroke="#6B6862" stroke-width="1.6"/><path d="M11.2 4.2 L6.6 10.6 H9.6 L8.8 15.8 L13.4 9.4 H10.4 Z" fill="#6B6862"/></svg>
    LiveTemplate
  </div>
  <div class="foot-links">
    <a href="/getting-started/introduction">Docs</a>
    <a href="/recipes/">Recipes</a>
    <a href="/reference/api">Reference</a>
    <a href="https://github.com/livetemplate/livetemplate">GitHub</a>
    <a href="/changelog">Changelog</a>
    <a href="https://github.com/livetemplate/livetemplate/blob/main/LICENSE">License</a>
  </div>
  <div class="attrib">
    <span>© the LiveTemplate authors</span>
    <span>Gopher by <a href="http://reneefrench.blogspot.com/">Renée French</a>, vector by <a href="https://github.com/golang-samples/gopher-vector">Takuya Ueda</a> · <a href="https://creativecommons.org/licenses/by/3.0/">CC BY 3.0</a></span>
  </div>
</div></footer>
