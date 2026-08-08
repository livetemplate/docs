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
  <a href="#the-app">The app</a>
  <a href="#inside">What's going on</a>
  <a class="sub" href="#actions">No attributes</a>
  <a class="sub" href="#nojs">No JavaScript</a>
  <a class="sub" href="#validation">Validation</a>
  <a class="sub" href="#pending">Pending state</a>
  <a class="sub" href="#realtime">Multi-user</a>
  <a href="#diff">Only the diff</a>
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

<section id="the-app" class="lead-in">
  <div class="eyebrow">The app</div>
  <h2>A shared greeting wall, running on this page.</h2>
  <p class="lead">Type a name. Your headline updates and your line joins the wall, along with everyone else's — including anyone else reading this page right now. <a href="/" target="_blank" rel="noopener">Open this page in a second tab</a> and watch it land there too, with no reload. The code for all of it is directly below.</p>

  <div class="demo">
    <div class="demo-bar"><span class="dot"></span> greet-wall · live, shared with every visitor</div>
    <div class="demo-body">

```embed-lvt path="/apps/greet-wall/" upstream="http://localhost:9091" height="230px"
```

</div>
  </div>

  <div class="stack">
    <div class="snip">
      <div class="snip-label">app.tmpl — the whole template, verbatim</div>
<pre class="language-html"><code class="language-html">&lt;script defer src="{{lvtClientScriptURL}}"&gt;&lt;/script&gt;
&lt;h1&gt;Hello, {{.Name}}&lt;/h1&gt;
&lt;form method="POST"&gt;
  &lt;input name="name" placeholder="Your name" required {{.lvt.AriaInvalid "name"}}&gt;
  {{.lvt.ErrorTag "name"}}
  &lt;button name="greet"&gt;Say hi&lt;/button&gt;
&lt;/form&gt;
&lt;ul&gt;
  {{range .Wall}}&lt;li&gt;&lt;b&gt;{{.Name}}&lt;/b&gt; said hi {{.At}}&lt;/li&gt;{{end}}
&lt;/ul&gt;</code></pre>
    </div>
    <div class="snip">
      <div class="snip-label">app.go — the whole controller, and the wiring</div>
<pre class="language-go"><code class="language-go">type State struct {
    Name string      // your headline, synced across your tabs
    Wall []Greeting  // the shared list, synced across everyone
}
func (a *App) Mount(s State, ctx *lvt.Context) (State, error) {
    ctx.Subscribe(ctx.SelfTopic())   // your own tabs
    ctx.Subscribe("wall")            // every visitor
    s.Name, s.Wall = a.nameFor(ctx.GroupID()), a.snapshot()
    return s, nil
}
func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    if err := ctx.ValidateForm(); err != nil {
        return s, err                        // re-runs the HTML rules
    }
    name := sanitize(ctx.GetString("name"))
    if name == "" {
        return s, lvt.NewFieldError("name", errors.New("Please enter a name"))
    }
    if strings.EqualFold(name, "admin") {    // a rule HTML can't express
        return s, lvt.NewFieldError("name", errors.New(`"admin" is reserved`))
    }
    a.saveName(ctx.GroupID(), name)   // so Refresh can re-read it on your tabs
    a.appendWall(name)                // the one shared list everyone sees
    // A publish skips the connection that called it, so this one renders
    // from the values returned here.
    s.Name, s.Wall = name, a.snapshot()
    ctx.Publish(ctx.SelfTopic(), "Refresh", nil)  // your other tabs
    ctx.Publish("wall", "WallRefresh", nil)       // everyone else
    return s, nil
}
// A publish just runs an ordinary action on the peers it reaches.
func (a *App) Refresh(s State, ctx *lvt.Context) (State, error) {
    s.Name = a.nameFor(ctx.GroupID()); return s, nil
}
func (a *App) WallRefresh(s State, ctx *lvt.Context) (State, error) {
    s.Wall = a.snapshot(); return s, nil
}
func main() {
    app := lvt.Must(lvt.New("wall",
        lvt.WithParseFiles("app.tmpl"),
        // Developer topics are deny-all; admit just this one. Each user's
        // own SelfTopic() is always permitted.
        lvt.WithTopicACL(func(topic, _ string, _ *http.Request) (bool, error) {
            return topic == "wall", nil
        })))
    http.ListenAndServe(":8080", app.Handle(&amp;App{}, lvt.AsState(&amp;State{Name: "there"})))
}</code></pre>
    </div>
  </div>

  <p class="close">That is the whole interface: a template, four methods and a <code>main</code>, with no JavaScript you had to write. The running demo adds about forty more lines of ordinary Go — <code>sanitize</code>, the map writes behind <code>saveName</code>, a twenty-line cap in <code>appendWall</code> and a per-session throttle — none of which is framework API. <a href="https://github.com/livetemplate/docs/blob/main/examples/greet-wall/wall.go">Read the real file</a>.</p>
</section>

<section id="inside" class="intro">
  <div class="eyebrow">What's going on</div>
  <h2>The parts of that worth a second look.</h2>
  <p class="lead">Every section below points at lines you have just read — except the pending state, which the wall has no slow work to demonstrate, and which says so. Each one also runs here as its own app, so you can check the claim rather than take it.</p>
</section>

<section id="actions" class="step">
  <div class="eyebrow">No attributes</div>
  <h2>The button's name is the action.</h2>
  <p class="lead"><code>&lt;button name="greet"&gt;</code> calls <code>Greet</code>. That's the whole binding — no <code>hx-post</code>, no <code>onClick</code>, no route to register. Strip the wall away and the same idea is a complete app in twenty lines, this time with nothing elided at all.</p>

  <div class="pair">
    <div class="demo">
      <div class="demo-bar"><span class="dot"></span> greet · the smallest version</div>
      <div class="demo-body">

```embed-lvt path="/apps/greet/" upstream="http://localhost:9091" height="130px"
```

</div>
    </div>
    <div class="snip">
      <div class="snip-label">app.go · complete, nothing elided</div>
<pre class="language-go"><code class="language-go">type State struct{ Name string }
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

  <p class="note">There <em>are</em> <code>lvt-*</code> attributes, but only for behavior HTML cannot express — a debounce, a keyboard shortcut, a class toggle. They're an escape hatch, not the interface.</p>
</section>

<section id="nojs" class="step">
  <div class="eyebrow">No JavaScript</div>
  <h2>The same app works with scripting switched off.</h2>
  <p class="lead">One <code>&lt;script&gt;</code> tag is the only difference between the two cards below. With it, the client enhances the submit and patches the headline. Without it, the same <code>&lt;form&gt;</code> does a native POST and the server renders the page. There's no <code>if jsEnabled</code> branch anywhere in the Go.</p>

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
    <div class="snip-label">app.tmpl · the line that flips the transport</div>
<pre class="language-html"><code class="language-html">&lt;script defer src="{{lvtClientScriptURL}}"&gt;&lt;/script&gt;</code></pre>
  </div>
</section>

<section id="validation" class="step">
  <div class="eyebrow">Validation</div>
  <h2>The HTML rule runs again in Go.</h2>
  <p class="lead">Both halves are in the app above. The input carries <code>required</code>, and <code>ctx.ValidateForm()</code> re-runs exactly that rule on the server — a client that skipped it, scripting off or a direct POST, gets the same answer. Then <code>strings.EqualFold(name, "admin")</code> adds the rule HTML has no way to state.</p>

  <p class="lead">The template side is the other two lines you read: <code>{{.lvt.AriaInvalid "name"}}</code> marks the field, <code>{{.lvt.ErrorTag "name"}}</code> is where the message lands. Returning an error from <code>Greet</code> is the whole mechanism — there's nothing to route. Scroll up and type <em>admin</em>, or try the smaller app here.</p>

  <div class="demo">
    <div class="demo-bar"><span class="dot"></span> greet-validate · the same pair, on its own</div>
    <div class="demo-body">

```embed-lvt path="/apps/greet-validate/" upstream="http://localhost:9091" height="160px"
```

</div>
  </div>

  <div class="wire">
    <div class="wire-label">on the wire · the rejected submit</div>
    <div>▲ <span class="payload">{"action":"greet","data":{"name":"admin"}}</span></div>
    <div>▼ <span class="payload">{"meta":{"errors":{"name":"\"admin\" is reserved"}}}</span></div>
  </div>
</section>

<section id="pending" class="step">
  <div class="eyebrow">Pending state</div>
  <h2>Slow work has a pending state you can render.</h2>
  <p class="lead">This is the one thing the app above cannot show you: the wall answers instantly, so it has no pending state to render. Both apps below do have slow work. The first is the way to reach for — the pending flag is a template variable, so the spinner is ordinary Go and ordinary HTML, with no new attribute to learn.</p>

  <div class="pair">
    <div>
      <div class="snip-label">A · server-owned — template variables, no attributes</div>
      <div class="demo">
        <div class="demo-bar"><span class="dot"></span> greet-async</div>
        <div class="demo-body">

```embed-lvt path="/apps/greet-async/" upstream="http://localhost:9091" height="130px"
```

</div>
      </div>
<pre class="language-html"><code class="language-html">&lt;button {{if .lvt.Pending}}type="button" aria-busy="true"
  disabled{{else}}name="greet"{{end}}&gt;Say hi&lt;/button&gt;</code></pre>
<pre class="language-go"><code class="language-go">lvt.Async(ctx,
    func(context.Context) (string, error) { return slowWork() },
    func(s State, name string, _ error) (State, error) {
        s.Name = name
        return s, nil
    },
)</code></pre>
      <p class="note">No second action to wire up and no <code>Loading</code> field in state, though it does need a live session for the completion render.</p>
    </div>
    <div>
      <div class="snip-label">B · the escape hatch, when the Go should not change</div>
      <div class="demo">
        <div class="demo-bar"><span class="dot"></span> greet-loading</div>
        <div class="demo-body">

```embed-lvt path="/apps/greet-loading/" upstream="http://localhost:9091" height="130px"
```

</div>
      </div>
<pre class="language-html"><code class="language-html">&lt;button name="greet"
  lvt-el:addClass:on:pending="is-loading"
  lvt-el:removeClass:on:done="is-loading"&gt;Say hi&lt;/button&gt;</code></pre>
      <p class="note">Two <code>lvt-*</code> attributes, and the Go is untouched. This is what the escape hatch is for: the spinner is button chrome, not something the app models. It also works as a single request/response, where A needs a live session for its completion render.</p>
    </div>
  </div>
</section>

<section id="realtime" class="step">
  <div class="eyebrow">Multi-user</div>
  <h2>Two calls sync your tabs. Changing the topic syncs everyone.</h2>
  <p class="lead">This is the part of <code>Greet</code> worth re-reading. <code>ctx.SelfTopic()</code> reaches your own tabs; <code>"wall"</code> reaches every visitor. Same two calls, different topic — that's the entire difference between "keeps my tabs in sync" and "multiplayer".</p>

  <div class="pipe">
    <span>1 state changes</span><span class="arrow">→</span>
    <span>2 re-render</span><span class="arrow">→</span>
    <span>3 diff vs last render</span><span class="arrow">→</span>
    <span>4 patch the browser</span>
  </div>

  <p class="lead">The two cards below are separate sessions, like two different people. Greet in one and the line shows up on both walls — while the headlines stay independent.</p>

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

  <div class="snip">
    <div class="snip-label">app.go · the server can start the same cycle</div>
<pre class="language-go"><code class="language-go">sess.TriggerAction("ServerRefresh", nil)</code></pre>
    <p class="note">You already read the <code>WithTopicACL</code> in <code>main</code> that admits <code>"wall"</code> — developer topics are deny-all until one is named. This is the same publish path with no user action behind it: the "the server said hi at …" line in the cards above, pushed on a timer.</p>
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
