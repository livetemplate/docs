---
title: "LiveTemplate — Build interactive web apps in Go with standard HTML templates"
description: "Write an html/template and a Go handler. The server re-renders on each action and patches what changed, so apps work without an SPA or a build step."
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
  <a href="#the-app">Live demo</a>
  <a href="#inside">How it works</a>
  <a class="sub" href="#actions">Actions</a>
  <a class="sub" href="#nojs">Without JavaScript</a>
  <a class="sub" href="#validation">Validation</a>
  <a class="sub" href="#pending">Loading states</a>
  <a class="sub" href="#realtime">Real-time updates</a>
  <a href="#compare">Comparison</a>
  <a href="#more">More features</a>
</aside>

<main class="doc">

<section class="hero">
  <div class="eyebrow">Alpha · a Go library for server-rendered apps</div>
  <h1>Build interactive web apps in Go with standard HTML templates.</h1>
  <p class="sub">Write <code>html/template</code> and Go handlers. The server re-renders on each action and patches only what changed. Every demo below is a real app running on this page, and you write no JavaScript for any of it.</p>
  <div class="cta-row">
    <a class="btn btn-primary" href="/getting-started/install">Get started →</a>
    <a class="btn btn-ghost" href="/getting-started/introduction">Read the docs</a>
    <code class="get">go get github.com/livetemplate/livetemplate</code>
  </div>
</section>

<section id="the-app" class="lead-in">
  <div class="eyebrow">Live demo</div>
  <h2>A shared guestbook in one template and one Go file</h2>
  <p class="lead">Type a name and press the button. The heading greets you, and your name joins the list below it for everyone viewing this page. <a href="/" target="_blank" rel="noopener">Open the page in a second tab</a> to see the list update without a reload. The full source is below the demo.</p>

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

  <p class="close">This app uses one template, four Go methods and a <code>main</code> function. The live demo adds about forty lines of plain Go for input cleanup, storage, a twenty-entry limit and rate limiting, none of which calls the library. <a href="https://github.com/livetemplate/docs/blob/main/examples/greet-wall/wall.go">Read the full file</a>.</p>
</section>

<section id="inside" class="intro">
  <div class="eyebrow">How it works</div>
  <h2>How the demo works</h2>
  <p class="lead">Each section below takes one part of the app above and runs it as a smaller app of its own. The loading-states section uses two other apps, because the guestbook responds instantly and has no slow work to show.</p>
</section>

<section id="actions" class="step">
  <div class="eyebrow">Actions</div>
  <h2>Buttons call Go methods</h2>
  <p class="lead">A button's <code>name</code> is the method it calls, so <code>&lt;button name="greet"&gt;</code> runs <code>Greet</code>, with no <code>hx-post</code>, <code>onClick</code> or route to register in between. Here is the smallest version of the app: twenty lines, with nothing left out.</p>

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

  <p class="note">Standard HTML is the interface. Add an <code>lvt-*</code> attribute only for behavior HTML can't express, such as a debounce, a keyboard shortcut or a class toggle.</p>
</section>

<section id="nojs" class="step">
  <div class="eyebrow">Without JavaScript</div>
  <h2>Works with JavaScript turned off</h2>
  <p class="lead">Both cards below run the same app. The left one loads the client script, so the form submits in the background and only the changed part of the page updates. The right one has JavaScript turned off, so the form does a normal POST and the server sends back the whole page. The Go code is the same for both, with no <code>if jsEnabled</code> check.</p>

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
    <div class="snip-label">app.tmpl · the one line that differs</div>
<pre class="language-html"><code class="language-html">&lt;script defer src="{{lvtClientScriptURL}}"&gt;&lt;/script&gt;</code></pre>
  </div>
</section>

<section id="validation" class="step">
  <div class="eyebrow">Validation</div>
  <h2>Form validation on the server</h2>
  <p class="lead">The input has the HTML <code>required</code> attribute, and <code>ctx.ValidateForm()</code> checks the same rule again on the server. So a request that skips the browser's check, with JavaScript off or sent directly, still gets the error. Rules HTML can't express are plain Go, like <code>strings.EqualFold(name, "admin")</code>.</p>

  <p class="lead">In the template, <code>{{.lvt.AriaInvalid "name"}}</code> marks the field as invalid and <code>{{.lvt.ErrorTag "name"}}</code> shows the message. To report a problem, return an <a href="/reference/error-handling">error</a> from <code>Greet</code>. Type <em>admin</em> in the guestbook above, or try this smaller app.</p>

  <div class="demo">
    <div class="demo-bar"><span class="dot"></span> greet-validate · both rules, on their own</div>
    <div class="demo-body">

```embed-lvt path="/apps/greet-validate/" upstream="http://localhost:9091" height="160px"
```

</div>
  </div>

  <div class="wire">
    <div class="wire-label">over the WebSocket · a rejected submit</div>
    <div>▲ <span class="payload">{"action":"greet","data":{"name":"admin"}}</span></div>
    <div>▼ <span class="payload">{"meta":{"errors":{"name":"\"admin\" is reserved"}}}</span></div>
  </div>
</section>

<section id="pending" class="step">
  <div class="eyebrow">Loading states</div>
  <h2>Loading states for slow actions</h2>
  <p class="lead">Both apps below do slow work. Use the first approach by default: the template reads <code>.lvt.Pending</code>, so the spinner is plain Go and HTML, with no new attribute to learn.</p>

  <div class="pair">
    <div>
      <div class="snip-label">A · in the template, no attributes</div>
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
      <p class="note">You don't add a second action or a <code>Loading</code> field to your state. The final render needs a live connection.</p>
    </div>
    <div>
      <div class="snip-label">B · with attributes, when the Go code shouldn't change</div>
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
      <p class="note">Two <code>lvt-*</code> attributes, and the Go code stays the same. Use this when the spinner is only button styling, not part of your app's state. It also works over a plain request and response, where A needs a live connection.</p>
    </div>
  </div>
</section>

<section id="realtime" class="step">
  <div class="eyebrow">Real-time updates</div>
  <h2>Sync across tabs and users</h2>
  <p class="lead"><code>Greet</code> publishes to two topics. <code>ctx.SelfTopic()</code> reaches your own open tabs, and <code>"wall"</code> reaches every visitor. Changing that one argument is the difference between syncing your tabs and syncing everyone.</p>

  <div class="pipe">
    <span>1 state changes</span><span class="arrow">→</span>
    <span>2 re-render</span><span class="arrow">→</span>
    <span>3 diff vs last render</span><span class="arrow">→</span>
    <span>4 patch the browser</span>
  </div>

  <p class="lead">The two cards below are separate sessions, like two different people. Say hi in one and your name appears in both lists, while each heading stays separate.</p>

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
    <div class="snip-label">app.go · the server can send updates too</div>
<pre class="language-go"><code class="language-go">sess.TriggerAction("ServerRefresh", nil)</code></pre>
    <p class="note"><code>TriggerAction</code> runs an action on a session with no user involved. A timer uses it to update the "the server said hi at …" line in the cards above. Topics you define are blocked until you allow them, which is what <a href="/reference/pubsub"><code>WithTopicACL</code></a> in <code>main</code> does for <code>"wall"</code>.</p>
  </div>

</section>

<section id="compare" class="step">
  <div class="eyebrow">Comparison</div>
  <h2>How it compares</h2>
  <p class="lead">The server renders HTML and sends updates over a WebSocket, like Phoenix LiveView. Your templates use ordinary form markup and the browser holds no app state, like htmx. Both are years more mature, and each row says where the other tool is the better choice.</p>
  <div class="rows">
    <div class="row-k">htmx</div>
    <div class="row-v">Works against any backend in any language, which this does not. Pick it when the server is already written. With LiveTemplate the server keeps the state and works out what changed, so your markup has no request attributes.</div>
    <div class="row-k">templ + htmx</div>
    <div class="row-v">templ type-checks your markup at compile time. <code>html/template</code> cannot, and that is a real thing to give up. The trade is a code generation step and one more library to track.</div>
    <div class="row-k">Alpine.js</div>
    <div class="row-v">Keeps real state in the browser: expressions, loops and computed values, which LiveTemplate doesn't have. An <code>lvt-el:</code> attribute toggles a class or an attribute on a DOM event and stops there. That covers a dropdown. It doesn't cover a widget with its own state.</div>
    <div class="row-k">Phoenix LiveView</div>
    <div class="row-v">The same model, with years of production use at scale and a much larger ecosystem. LiveTemplate brings the model to Go and is still in alpha.</div>
    <div class="row-k">React SPA</div>
    <div class="row-v">For a canvas editor or an offline-first app, a client framework is the right call. For a settings screen, it means keeping two copies of the same data in sync. LiveTemplate keeps one copy, on the server.</div>
  </div>
</section>

<section id="more" class="step">
  <div class="eyebrow">More features</div>
  <h2>Built-in features</h2>
  <p class="lead">Made for admin screens, internal tools, CRUD apps and dashboards.</p>
  <div class="links">
    <a href="/reference/uploads"><span class="link-t">File uploads</span><span class="link-g">live progress, same app</span></a>
    <a href="/recipes/login/"><span class="link-t">Auth &amp; login</span><span class="link-g">form-based sessions, no middleware</span></a>
    <a href="/reference/session"><span class="link-t">Sessions &amp; state</span><span class="link-g">scoped per browser or user</span></a>
    <a href="/reference/navigate"><span class="link-t">Navigation</span><span class="link-g">move between pages, no reload</span></a>
    <a href="/cli/"><span class="link-t">Scaffolding</span><span class="link-g">generate common app shapes</span></a>
    <a href="/client/"><span class="link-t">Browser client</span><span class="link-g">transport &amp; DOM patching</span></a>
    <a href="/guides/observability"><span class="link-t">Observability</span><span class="link-g">metrics &amp; tracing hooks</span></a>
    <a href="/guides/scaling"><span class="link-t">Scaling</span><span class="link-g">session groups &amp; fan-out</span></a>
  </div>
  <p class="lead">The <a href="/recipes/ui-patterns/">UI patterns catalog</a> has focused examples: loading states, inline validation, SPA-style navigation, sortable tables, pubsub, presence, server push. This site is a LiveTemplate app too, including this page. <a href="/recipes/how-this-site-works">See how it's built</a>.</p>
</section>

<section class="cta">
  <div>
    <h2>Get started</h2>
<pre class="install">$ go get github.com/livetemplate/livetemplate</pre>
    <div class="cta-row">
      <a class="btn btn-primary" href="/getting-started/install">Get started →</a>
      <a class="btn btn-ghost" href="/recipes/">Browse recipes</a>
    </div>
    <p class="note">This is alpha: the core works and has tests, but the API may still change before v1.0.</p>
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
