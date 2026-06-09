---
title: "LiveTemplate — Build interactive web apps in Go with standard HTML templates"
description: "Use html/template and Go handlers to build rich app screens without writing JavaScript for the common cases."
layout: landing
---

<header class="nav"><div class="wrap nav-in">
  <div class="brand"><span class="glyph">◇</span> LiveTemplate</div>
  <nav class="nav-links"><a href="/getting-started/introduction">Docs</a><a href="/recipes/">Recipes</a><a href="/reference/api">Reference</a><a href="https://github.com/livetemplate/livetemplate">GitHub</a><a class="btn btn-primary" href="/getting-started/install">Get started →</a></nav>
</div></header>

<!-- HERO -->
<section class="hero"><div class="wrap">
  <span class="eyebrow">A simpler way to build interactive Go web apps · Alpha</span>
  <h1 class="head">Build interactive web apps in Go with <span class="g">standard HTML templates.</span></h1>
  <p class="sub">Use <b><code>html/template</code> and Go handlers</b> to build rich app screens without writing JavaScript for the common cases.</p>
  <div class="cta-row">
    <a class="btn btn-primary btn-lg" href="/getting-started/install">Get started →</a>
    <a class="btn btn-ghost btn-lg" href="/getting-started/introduction">Read the docs</a>
  </div>
  <div class="hero-snip">
    <div class="live-card">
      <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">greet · running in this page</span></div>
      <div class="live-body">

```embed-lvt path="/apps/greet/" upstream="http://localhost:9091" height="200px"
```

</div>
    </div>
    <p class="hero-cap" style="margin:14px 0 8px">↑ a real, running app. Type a name, hit <b>Say hi</b>. Below is <b>the whole thing</b> — the template and the Go code, complete:</p>
    <div class="code"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.tmpl &nbsp;— the entire template, just standard HTML</span></div>
<pre><span class="tag">&lt;!DOCTYPE html&gt;</span>
<span class="tag">&lt;html&gt;&lt;head&gt;</span>
  <span class="tag">&lt;script</span> <span class="attr">defer src</span>=<span class="str">"https://cdn.jsdelivr.net/npm/@livetemplate/client"</span><span class="tag">&gt;&lt;/script&gt;</span>
<span class="tag">&lt;/head&gt;&lt;body&gt;</span>
  <span class="tag">&lt;h1&gt;</span>Hello, {{<span class="fn">.Name</span>}}<span class="tag">&lt;/h1&gt;</span>
  <span class="tag">&lt;form</span> <span class="attr">method</span>=<span class="str">"POST"</span><span class="tag">&gt;</span>
    <span class="tag">&lt;input</span> <span class="attr">name</span>=<span class="str">"name"</span> <span class="attr">placeholder</span>=<span class="str">"Your name"</span><span class="tag">&gt;</span>
    <span class="tag">&lt;button</span> <span class="attr">name</span>=<span class="str">"greet"</span><span class="tag">&gt;</span>Say hi<span class="tag">&lt;/button&gt;</span>
  <span class="tag">&lt;/form&gt;</span>
<span class="tag">&lt;/body&gt;&lt;/html&gt;</span></pre></div>
    <div class="code"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go &nbsp;— the entire program</span></div>
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
}</code></pre></div>
    <p class="hero-cap">That's the whole app — ~20 lines of Go and standard HTML. No separate frontend, no build step, no generated code.</p>
  </div>
</div></section>

<!-- UNDER THE HOOD: animated request/response over the wire -->
<section><div class="wrap">
  <div class="sec-tag">Step 1 · Render</div>
  <h2>This is the whole app, not a toy example.</h2>
  <p class="lead">These are the <b>real frames</b> on the wire. A form submit calls your Go method, the server re-renders the template, and <b>only the changed HTML comes back</b> — no reload, no extra JSON API, no client route you had to build.</p>
  <div class="uh">
    <!-- browser -->
    <div class="uh-side uh-browser">
      <span class="uh-tag">browser</span>
      <div class="uh-app">
        <div class="uh-h"><span class="uh-there">Hello, there</span><span class="uh-ada">Hello, Ada</span></div>
        <button class="uh-btn">Say&nbsp;hi</button>
      </div>
    </div>
    <!-- the wire -->
    <div class="uh-wire">
      <span class="uh-wlabel">WebSocket</span>
      <div class="uh-lane">
        <span class="uh-arrow">▲ action · 40 B</span>
        <span class="uh-pkt uh-up">{"action":<b>"greet"</b>,"data":{"name":"Ada"}}</span>
      </div>
      <div class="uh-lane">
        <span class="uh-arrow">▼ diff · 20 B</span>
        <span class="uh-pkt uh-down">{"tree":{"0":<b>"Ada"</b>}}</span>
      </div>
    </div>
    <!-- go server -->
    <div class="uh-side uh-server">
      <span class="uh-tag">Go server</span>
      <div class="uh-go">Greet(state)</div>
      <div class="uh-steps">
        <span class="uh-step st1">re-render</span>
        <span class="uh-step st2">diff</span>
      </div>
    </div>
  </div>
  <p class="wiring-foot">All of it comes from the <b>two files above</b> plus one <code>&lt;script&gt;</code>. The framework handles transport and DOM patching, so you stay in Go handlers and HTML templates.</p>
</div></section>

<!-- SPINE INTRO -->
<section class="alt"><div class="wrap spine-intro">
  <div class="sec-tag">One app, seven steps</div>
  <h2>Start with a normal Go app. Then add the parts real apps need.</h2>
  <p class="lead">Everything below is the <b>same greeting app</b>. We add plain POST fallback, validation, pending state, then WebSocket updates. Each step is a <b>small diff</b>. You keep one Go codebase and one place for application logic.</p>
</div></section>

<!-- STEP 2 · WORKS WITHOUT JS -->
<section><div class="wrap two code-right">
  <div>
    <div class="sec-tag">Step 2 · Works without JavaScript</div>
    <h2>The same app. With and without JavaScript.</h2>
    <p class="lead">Both cards run the <b>identical app</b>, with WebSocket off. <b>Left, JS on:</b> the browser enhances the form submit and patches the headline in place. <b>Right, JS disabled:</b> the same <code>&lt;form&gt;</code> does a plain POST and the server renders the page. JavaScript changes the <b>browser behavior</b>, not the app you have to build. Type a name in each.</p>
    <div class="two" style="margin-top:28px">
      <div class="live-card">
        <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">JavaScript on · fetch + DOM patch</span></div>
        <div class="live-body">
        <iframe class="nojs-frame" src="/apps/greet-nojs/" sandbox="allow-forms allow-same-origin allow-scripts" title="The greeting app with JavaScript enabled"></iframe>
</div>
      </div>
      <div class="live-card">
        <div class="live-bar"><span class="live-badge nojs">○ no JS</span><span class="live-meta">JavaScript off · form POST → full render</span></div>
        <div class="live-body">
          <iframe class="nojs-frame" src="/apps/greet-nojs/" sandbox="allow-forms allow-same-origin" title="The greeting app with JavaScript disabled"></iframe>
        </div>
      </div>
    </div>
  </div>
  <div>
    <div class="code" style="max-width:680px;margin:0 auto"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.tmpl · one form, either transport</span></div>
<pre><span class="com">&lt;!-- the only line that flips the transport: --&gt;</span>
<span class="tag">&lt;script</span> <span class="attr">defer src</span>=<span class="str">"…@livetemplate/client"</span><span class="tag">&gt;&lt;/script&gt;</span>

<span class="tag">&lt;form</span> <span class="attr">method</span>=<span class="str">"POST"</span><span class="tag">&gt;</span>   <span class="com">&lt;!-- JS on → fetch + patch · JS off → native POST --&gt;</span>
  <span class="tag">&lt;input</span> <span class="attr">name</span>=<span class="str">"name"</span><span class="tag">&gt;</span>
  <span class="tag">&lt;button</span> <span class="attr">name</span>=<span class="str">"greet"</span><span class="tag">&gt;</span>Say hi<span class="tag">&lt;/button&gt;</span>
<span class="tag">&lt;/form&gt;</span></pre></div>
    <p class="demo-cap loading-cap" style="margin-top:14px">Same <code>&lt;form&gt;</code> and the same <code>Greet</code> handler as Step 1 — no <code>if jsEnabled</code> branch anywhere. When the <code>&lt;script&gt;</code> loads, the client enhances the submit; when it doesn't, the browser falls back to a native POST.</p>
  </div>
</div></section>

<!-- STEP 3 · VALIDATE -->
<section><div class="wrap two code-right">
  <div>
    <div class="sec-tag">Step 3 · Validation</div>
    <h2>Write the rule in HTML. Enforce it again on the server.</h2>
    <p class="lead">Use standard HTML attributes like <code>required</code> and <code>type="email"</code>. <code>ctx.ValidateForm()</code> re-runs the <b>same</b> rules server-side, then you can add Go-only checks for business rules. Submit empty, or type <b>admin</b>:</p>
    <div class="live-card" style="margin-top:24px">
      <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">greet-validate · server-checked</span></div>
      <div class="live-body">

```embed-lvt path="/apps/greet-validate/" upstream="http://localhost:9091" height="220px"
```

</div>
    </div>
  </div>
  <div>
    <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.tmpl · the rule, written once</span></div>
<pre><span class="tag">&lt;input</span> <span class="attr">name</span>=<span class="str">"name"</span> <span class="attr">required</span> {{<span class="fn">.lvt.AriaInvalid</span> <span class="str">"name"</span>}}<span class="tag">&gt;</span>
{{<span class="fn">.lvt.ErrorTag</span> <span class="str">"name"</span>}}</pre></div>
    <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go · the server re-checks, then adds its own rule</span></div>
<pre class="language-go"><code class="language-go">func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    if err := ctx.ValidateForm(); err != nil {   // re-runs the HTML rules server-side
        return s, err
    }
    name := strings.TrimSpace(ctx.GetString("name"))
    if strings.EqualFold(name, "admin") {         // a rule HTML can't express
        return s, lvt.NewFieldError("name", errors.New(`"admin" is reserved`))
    }
    s.Name = name
    return s, nil
}</code></pre></div>
    <div class="wire"><span class="wlabel">on the wire · HTTP fetch</span>
      <span class="wf up">▲ {"action":"greet","data":{"name":"admin"}}</span>
      <span class="wf dn">▼ {"meta":{"errors":{"name":"\"admin\" is reserved"}}}</span>
    </div>
  </div>
</div></section>

<!-- STEP 4 · LOADING -->
<section class="alt"><div class="wrap">
  <div class="sec-tag">Step 4 · Loading state</div>
  <h2>Two ways to show pending state, in HTTP and WebSocket mode.</h2>
  <p class="lead">LiveTemplate works in both <b>plain HTTP</b> and <b>live-session WebSocket</b> mode. You can model loading in <b>server state</b> with ordinary template conditionals, or use a small <b>button-level escape hatch</b> when the server code should stay unchanged. The server-state version below needs a <b>live session connection</b> for its follow-up push; the attribute version works as a single request/response.</p>
  <div class="two loading-cols" style="margin-top:28px">
    <div class="loading-col">
      <div class="live-card">
        <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">greet loading server owned</span></div>
        <div class="live-body">

```embed-lvt path="/apps/greet-loading-server/" upstream="http://localhost:9091" height="200px"
```

</div>
      </div>
      <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.tmpl · server-owned loading, only template variables</span></div>
<pre><span class="tag">&lt;button</span> <span class="attr">class</span>=<span class="str">"greet-btn"</span> {{<span class="kw">if</span> <span class="fn">.Loading</span>}}<span class="attr">type</span>=<span class="str">"button"</span> <span class="attr">aria-busy</span>=<span class="str">"true"</span> <span class="attr">disabled</span>{{<span class="kw">else</span>}}<span class="attr">name</span>=<span class="str">"greet"</span>{{<span class="kw">end</span>}}<span class="tag">&gt;</span>Say hi<span class="tag">&lt;/button&gt;</span></pre></div>
      <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go · set Loading, then finish via server push</span></div>
<pre class="language-go"><code class="language-go">func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    if s.Loading {
        return s, nil
    }
    session := ctx.Session()
    if session == nil {
        return s, nil
    }
    name := strings.TrimSpace(ctx.GetString("name"))
    s.Loading = true
    go func() {
        time.Sleep(700 * time.Millisecond)
        _ = session.TriggerAction("finishGreet", map[string]any{"name": name})
    }()
    return s, nil
}
func (a *App) FinishGreet(s State, ctx *lvt.Context) (State, error) {
    s.Name = ctx.GetString("name")
    s.Loading = false
    return s, nil
}</code></pre></div>
      <p class="demo-cap loading-cap" style="margin-top:18px">This version keeps loading entirely in <b>server state</b>, but it needs a second action over the live session to clear the spinner. It is a good fit when loading is part of the app's actual state machine.</p>
      <div class="wire"><span class="wlabel">on the wire · server-state version</span>
        <span class="wf up">▲ {"action":"greet","data":{"name":"Ada"}}</span>
        <span class="wf dn">▼ {"tree":{"1":{"aria-busy":"true","disabled":true,"type":"button"}}}</span>
        <span class="wf dn">▼ {"action":"finishGreet","data":{"name":"Ada"}} → {"tree":{"0":"Ada","1":{"name":"greet"}}}</span>
      </div>
    </div>
    <div class="loading-col">
      <div class="live-card">
        <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">greet loading attribute</span></div>
        <div class="live-body">

```embed-lvt path="/apps/greet-loading/" upstream="http://localhost:9091" height="200px"
```

</div>
      </div>
      <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.tmpl · button-level pending with two <code>lvt-*</code> attributes</span></div>
<pre><span class="tag">&lt;button</span> <span class="attr">name</span>=<span class="str">"greet"</span>
  <span class="attr">lvt-el:addClass:on:pending</span>=<span class="str">"is-loading"</span>
  <span class="attr">lvt-el:removeClass:on:done</span>=<span class="str">"is-loading"</span><span class="tag">&gt;</span>Say hi<span class="tag">&lt;/button&gt;</span></pre></div>
      <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go · no loading state machine needed</span></div>
<pre class="language-go"><code class="language-go">func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    time.Sleep(700 * time.Millisecond)
    if name := strings.TrimSpace(ctx.GetString("name")); name != "" {
        s.Name = name
    }
    return s, nil
}</code></pre></div>
      <p class="demo-cap loading-cap" style="margin-top:18px">This version keeps the <b>Go code simpler</b> by leaving pending UI out of server state. It works as a single request/response, so use it when the loading indicator is just button chrome rather than meaningful application state.</p>
      <div class="wire"><span class="wlabel">on the wire · attribute version</span>
        <span class="wf up">▲ {"action":"greet","data":{"name":"Ada"}}</span>
        <span class="wf dn">▼ {"tree":{"0":"Ada"}}</span>
      </div>
    </div>
  </div>
</div></section>

<!-- STEP 5 · YOUR TABS -->
<section class="alt"><div class="wrap">
  <div class="sec-tag">Step 5 · Sync your own tabs</div>
  <h2>Add WebSocket updates. Keep your tabs in sync.</h2>
  <p class="lead">Subscribe this browser session to its own topic and publish after a handler runs — <b>two calls</b> — and your greeting syncs across every open tab. The same live session also lets the <b>server push first</b> when it has something new to say.</p>
  <div class="pipe" style="margin-top:26px">
    <div class="step"><div class="k">1 · state</div><div class="v">state changes</div></div><div class="arrow">→</div>
    <div class="step"><div class="k">2 · render</div><div class="v">re-render template</div></div><div class="arrow">→</div>
    <div class="step"><div class="k">3 · diff</div><div class="v">diff vs last render</div></div><div class="arrow">→</div>
    <div class="step"><div class="k">4 · patch</div><div class="v">patch the browser</div></div>
  </div>
  <div class="live-card" style="margin:32px auto 0;max-width:540px">
    <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">greet wall · WebSocket on</span></div>
    <div class="live-body">

```embed-lvt path="/apps/greet-wall/" upstream="http://localhost:9091" height="260px"
```

</div>
  </div>
  <p class="demo-cap"><b>Open this page in a second tab</b>, greet in either, and your headline updates in <b>both</b> — live, no reload. The same connection also allows <b>server-initiated refreshes</b> in this app, without waiting for a user click. This is the kind of step from "single-page form" to "real workflow" that usually pushes teams toward a separate frontend.</p>
  <div class="code delta" style="max-width:820px;margin:26px auto 0"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go · subscribe, publish on greet, and the Refresh it runs</span></div>
<pre class="language-go"><code class="language-go">func (a *App) Mount(s State, ctx *lvt.Context) (State, error) {
    ctx.Subscribe(ctx.SelfTopic())                 // your tabs share a topic
    s.Name = a.name(ctx.GroupID())                 // load your latest name
    return s, nil
}
func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    a.setName(ctx.GroupID(), sanitize(ctx.GetString("name")))
    ctx.Publish(ctx.SelfTopic(), "Refresh", nil)   // run Refresh on your other tabs
    return s, nil
}
// Refresh is an ordinary action — the publish above runs it on each peer tab.
func (a *App) Refresh(s State, ctx *lvt.Context) (State, error) {
    s.Name = a.name(ctx.GroupID())                 // re-read state, then re-render
    return s, nil
}</code></pre></div>
  <p class="demo-cap" style="margin-top:14px">No magic: <code>Publish(ctx.SelfTopic(), "Refresh", nil)</code> just <b>runs your <code>Refresh</code> method on your other tabs</b>. It re-reads shared data and returns new state; the framework diffs and patches.</p>
  <div class="wire" style="max-width:820px;margin:14px auto 0"><span class="wlabel">on the wire · WebSocket</span>
    <span class="wf up">▲ this tab · {"action":"greet","data":{"name":"Ada"}}</span>
    <span class="wf dn">▼ your other tab · {"tree":{"0":"Ada"}}</span>
  </div>
  <div class="code delta" style="max-width:820px;margin:18px auto 0"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go · the same session can be pushed by the server</span></div>
<pre class="language-go"><code class="language-go">func (a *App) OnConnect(s State, ctx *lvt.Context) (State, error) {
    a.keep(ctx.GroupID(), ctx.Session())   // remember who's connected
    return s, nil
}
func (a *App) heartbeat() {
    for range time.Tick(30 * time.Second) {
        a.serverAt = now()                      // replace one slot in place
        for _, sess := range a.sessions {
            sess.TriggerAction("ServerRefresh", nil)
        }
    }
}</code></pre></div>
  <div class="wire" style="max-width:820px;margin:14px auto 0"><span class="wlabel">on the wire · server push</span>
    <span class="wf dn">▼ {"tree":{"3":{"0":"15:04:08"}}}</span>
    <span class="wf note">(no ▲ — the server started it; just the changed value goes down)</span>
  </div>
</div></section>

<!-- STEP 6 · EVERYONE -->
<section><div class="wrap">
  <div class="sec-tag">Step 6 · A wall everyone shares</div>
  <h2>Change the topic, and it becomes cross-user.</h2>
  <p class="lead">Swap the self-topic for a <b>shared</b> topic — admitted by a small ACL — and the same publish fans out to <b>every visitor</b>. The two cards below are <b>separate sessions, like two different people</b>. Greet in one and your line lands on the other's wall, live.</p>
  <div class="two" style="margin-top:28px">
    <div class="live-card"><div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">visitor 1 · WebSocket on</span></div><div class="live-body">

```embed-lvt path="/apps/greet-wall/" upstream="http://localhost:9091" height="260px"
```

</div></div>
    <div class="live-card"><div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">visitor 2 · WebSocket on</span></div><div class="live-body">

```embed-lvt path="/apps/greet-wall/" upstream="http://localhost:9091" height="260px"
```

</div></div>
  </div>
  <p class="demo-cap">Two independent sessions, one shared wall — type in either card and watch the <b>list</b> appear in both. This is the same pattern you would use for shared dashboards, approval queues, team status boards, or lightweight collaboration.</p>
  <div class="two code-right" style="margin-top:26px">
    <div>
      <p class="lead">Headlines stay independent (each card is its own session), but the wall is global — so a greeting crosses from one session to the other. That's the whole cross-user story: the same two pub/sub calls as step 5, with a different topic and an ACL around it.</p>
      <div class="wire"><span class="wlabel">on the wire · WebSocket</span>
        <span class="wf up">▲ visitor 1 · {"action":"greet","data":{"name":"Ada"}}</span>
        <span class="wf dn">▼ visitor 2 · {"tree":{"3":[["a",[{"0":"Ada","1":"15:04"}]]]}}</span>
      </div>
    </div>
    <div>
      <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go · the topic is the only difference</span></div>
<pre class="language-go"><code class="language-go">func (a *App) Mount(s State, ctx *lvt.Context) (State, error) {
    ctx.Subscribe("wall")                       // a shared, cross-user topic
    return s, nil
}
func (a *App) Greet(s State, ctx *lvt.Context) (State, error) {
    a.append(sanitize(ctx.GetString("name")))   // shared, capped, ephemeral
    ctx.Publish("wall", "WallRefresh", nil)     // fan out to every visitor
    return s, nil
}</code></pre></div>
      <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go · admit the shared topic</span></div>
<pre class="language-go"><code class="language-go">lvt.WithTopicACL(func(topic, _ string, _ *http.Request) (bool, error) {
    return topic == "wall", nil   // deny-all by default; admit just this one
})</code></pre></div>
    </div>
  </div>
</div></section>

<!-- DIFF -->
<section><div class="wrap two">
  <div>
    <div class="sec-tag">Only the diff goes over the wire</div>
    <h2>Send what changed, not the whole page.</h2>
    <p class="lead">Templates split into <b>static structure (cached)</b> and <b>dynamic values</b>. On change, LiveTemplate sends only the changed values — typically <b>85%+ less bandwidth</b> than re-sending full HTML. A greeting comes back as <code>{"tree":{"0":"Ada"}}</code>, not a page.</p>
  </div>
  <div class="bars">
    <div class="bar-row"><span class="lab">full HTML</span><span class="bar full"><span></span></span><span class="val" style="color:var(--slate-2)">2.4 KB</span></div>
    <div class="bar-row"><span class="lab">lvt diff</span><span class="bar diff"><span></span></span><span class="val" style="color:var(--sig-d)">340 B</span></div>
    <div style="font:500 13px 'JetBrains Mono';color:var(--slate);margin-top:8px">↓ 86% smaller per update</div>
  </div>
</div></section>

<section><div class="wrap">
  <div class="sec-tag">UI Patterns</div>
  <h2>Want deeper demos?</h2>
  <p class="lead">The <a href="/recipes/ui-patterns/">UI patterns catalog</a> breaks these ideas out into focused examples: loading states, inline validation, SPA-style navigation, sortable tables, pubsub, presence, server push, and more.</p>
</div></section>

<!-- FEATURES -->
<section class="alt"><div class="wrap">
  <div class="sec-tag">And so much more</div>
  <h2>The pieces real Go apps need.</h2>
  <p class="lead">This is aimed at the kinds of apps Go teams actually ship: admin screens, internal tools, CRUD flows, dashboards, approval systems, uploads, auth, and lightweight collaborative views.</p>
  <div class="grid">
    <a class="feat" href="/reference/uploads"><div class="ico"><svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg></div><h4>File uploads</h4><p>Handle uploads in the same app and show live progress, without bolting on a separate upload flow.</p></a>
    <a class="feat" href="/reference/pubsub"><div class="ico"><svg viewBox="0 0 24 24"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><line x1="12" y1="20" x2="12.01" y2="20"/></svg></div><h4>Shared views</h4><p>Keep tabs, dashboards, queues, and team screens in sync with <code>Subscribe</code> and <code>Publish</code>.</p></a>
    <a class="feat" href="/reference/session"><div class="ico"><svg viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg></div><h4>Sessions &amp; state</h4><p>Keep UI state on the server, scoped per browser or per user, without leaking data across sessions.</p></a>
    <a class="feat" href="/reference/error-handling"><div class="ico"><svg viewBox="0 0 24 24"><path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg></div><h4>Forms &amp; errors</h4><p>Return validation and business-rule errors from Go and render them back into the same template.</p></a>
    <a class="feat" href="/cli/"><div class="ico"><svg viewBox="0 0 24 24"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg></div><h4>Scaffolding</h4><p>Generate a starting point for common app shapes so teams can get to real screens faster.</p></a>
    <a class="feat" href="/client/"><div class="ico"><svg viewBox="0 0 24 24"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg></div><h4>Browser client</h4><p>The browser layer handles DOM patching and transport so your application logic stays in Go.</p></a>
    <a class="feat" href="/guides/observability"><div class="ico"><svg viewBox="0 0 24 24"><line x1="6" y1="20" x2="6" y2="14"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="18" y1="20" x2="18" y2="10"/></svg></div><h4>Observability</h4><p>Measure handler timings, update paths, and runtime behavior with hooks for metrics and tracing.</p></a>
    <a class="feat" href="/guides/scaling"><div class="ico"><svg viewBox="0 0 24 24"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg></div><h4>Scaling</h4><p>Run the same model in production with guidance for session groups, fan-out, and deployment shape.</p></a>
  </div>
</div></section>

<!-- COMPARE -->
<section><div class="wrap">
  <div class="sec-tag">How it compares</div>
  <h2>Others add a frontend layer. LiveTemplate keeps the app in Go.</h2>
  <p class="lead">Other tools carry more behavior in the markup — <code>hx-*</code>, <code>x-*</code>, <code>phx-*</code>, or a DSL. Here a plain <code>&lt;button name="greet"&gt;</code> is already the action, handlers stay in Go, and state lives on the server. <code>lvt-*</code> attributes exist only as an escape hatch for what HTML can't express.</p>
  <table class="cmp">
    <thead><tr><th>If you’re using…</th><th>LiveTemplate gives you…</th></tr></thead>
    <tbody>
      <tr><td>htmx</td><td class="give">A similar HTML-first feel, but with server-owned state and DOM diffing built in, so there is less request wiring in markup.</td></tr>
      <tr><td>templ + htmx</td><td class="give">Use Go's built-in <code>html/template</code> and keep live behavior in one app model instead of composing multiple layers.</td></tr>
      <tr><td>Alpine.js</td><td class="give">Handle richer app behavior without introducing a separate client-side state model for common server-rendered screens.</td></tr>
      <tr><td>Phoenix LiveView</td><td class="give">A comparable server-driven model, but staying in Go and still falling back cleanly to plain HTTP forms.</td></tr>
      <tr><td>React SPA</td><td class="give">Get modern app behavior for forms, CRUD, dashboards, and shared views without splitting the product into an API plus a frontend app.</td></tr>
    </tbody>
  </table>
</div></section>

<!-- DOGFOOD -->
<section class="alt"><div class="wrap">
  <div class="dogfood">
    <div><svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="#047857" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg></div>
    <p><b>Built in Go. This site proves the point.</b> Every step above is a real LiveTemplate app, embedded live through this docs site — which itself runs on LiveTemplate + tinkerdown. <a href="/recipes/how-this-site-works">See how this site works →</a></p>
  </div>
</div></section>

<!-- FINAL -->
<section class="final"><div class="wrap">
  <h2>Build a real Go web app. Start in 30 seconds.</h2>
  <div class="install"><span class="p">$</span> go get github.com/livetemplate/livetemplate</div>
  <div class="cta-row">
    <a class="btn btn-primary btn-lg" href="/getting-started/install">Get started →</a>
    <a class="btn btn-ghost btn-lg" href="/recipes/">Browse recipes</a>
  </div>
  <div class="alpha">⚠ Alpha — core features work and are tested; the API may change before v1.0</div>
</div></section>

<footer><div class="wrap foot-in">
  <div class="brand" style="color:#fff"><span class="glyph">◇</span> LiveTemplate</div>
  <div class="foot-links"><a href="/getting-started/introduction">Docs</a><a href="/recipes/">Recipes</a><a href="/reference/api">Reference</a><a href="https://github.com/livetemplate/livetemplate">GitHub</a><a href="/changelog">Changelog</a><a href="https://github.com/livetemplate/livetemplate/blob/main/LICENSE">License</a></div>
  <div style="font-size:13px">© the LiveTemplate authors</div>
</div></footer>
