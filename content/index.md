---
title: "LiveTemplate — Reactive web apps in Go, made with just HTML"
description: "Build reactive, real-time web apps in Go with standard HTML. No client framework, no build step."
layout: landing
---

<header class="nav"><div class="wrap nav-in">
  <div class="brand"><span class="glyph">◇</span> LiveTemplate</div>
  <nav class="nav-links"><a href="/getting-started/introduction">Docs</a><a href="/recipes/">Recipes</a><a href="/reference/api">Reference</a><a href="https://github.com/livetemplate/livetemplate">GitHub</a><a class="btn btn-primary" href="/getting-started/install">Get started →</a></nav>
</div></header>

<!-- HERO -->
<section class="hero"><div class="wrap">
  <span class="eyebrow">Server-driven UI for Go · Alpha</span>
  <h1 class="head">Reactive web apps in Go, <span class="g">made with just HTML.</span></h1>
  <p class="sub">Real-time apps from <b>standard HTML and a Go controller</b> — no client framework, no build step, no JavaScript to write.</p>
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
    <p class="hero-cap" style="margin:14px 0 8px">↑ a real, running app. Type a name, hit <b>Say hi</b>. Below is <b>every line</b> that makes it — both files, complete:</p>
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
    <p class="hero-cap">That's the whole app — ~20 lines of Go and standard HTML. No build step, no generated code.</p>
  </div>
</div></section>

<!-- UNDER THE HOOD: animated request/response over the wire -->
<section><div class="wrap">
  <div class="sec-tag">Step 1 · Render</div>
  <h2>That's the whole program — not a snippet.</h2>
  <p class="lead">These are the <b>real frames</b> on the wire. Your action goes up, the server runs your method, and <b>only the changed value comes back</b> — no reload, no fetch or route you wrote.</p>
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
  <p class="wiring-foot">All of it from the <b>two files above</b> plus one <code>&lt;script&gt;</code>. The framework moves the bytes — you never wrote a fetch, a route, or a diff.</p>
</div></section>

<!-- SPINE INTRO -->
<section class="alt"><div class="wrap spine-intro">
  <div class="sec-tag">One app, seven steps</div>
  <h2>Now watch that same app grow.</h2>
  <p class="lead">Everything below is the <b>same greeting app</b>, gaining one capability at a time — validation, loading, no-JS, then real-time. Each step is a <b>small diff</b> and the real bytes on the wire. No second model to learn — only more of the one you just saw.</p>
</div></section>

<!-- STEP 2 · VALIDATE -->
<section><div class="wrap two code-right">
  <div>
    <div class="sec-tag">Step 2 · Validation</div>
    <h2>Validate once, on both sides.</h2>
    <p class="lead">Write each rule once as a standard HTML attribute — <code>required</code>, <code>type="email"</code>. <code>ctx.ValidateForm()</code> re-checks the <b>same</b> rules server-side. For what HTML can't express, return a <code>FieldError</code>. Submit empty, or type <b>admin</b>:</p>
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

<!-- STEP 3 · LOADING -->
<section class="alt"><div class="wrap two code-right">
  <div>
    <div class="sec-tag">Step 3 · Loading state</div>
    <h2>A loading button, declared in HTML.</h2>
    <p class="lead">While the action is in flight, the client toggles a class on the button — a little CSS spins an icon in its place, no JavaScript. Two attributes mark the start and the end; the server code is unchanged. Click <b>Say hi</b>:</p>
    <div class="live-card" style="margin-top:24px">
      <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">greet-loading · live spinner</span></div>
      <div class="live-body">

```embed-lvt path="/apps/greet-loading/" upstream="http://localhost:9091" height="200px"
```

</div>
    </div>
  </div>
  <div>
    <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.tmpl · the loading button</span></div>
<pre><span class="tag">&lt;button</span> <span class="attr">name</span>=<span class="str">"greet"</span>
  <span class="attr">lvt-el:addClass:on:pending</span>=<span class="str">"is-loading"</span>
  <span class="attr">lvt-el:removeClass:on:done</span>=<span class="str">"is-loading"</span><span class="tag">&gt;</span>Say hi<span class="tag">&lt;/button&gt;</span></pre></div>
    <p class="demo-cap" style="margin-top:18px"><code>lvt-*</code> attributes are the escape hatch for <b>behavior HTML itself can't express</b> — a pending state, a debounce, a keyboard shortcut.</p>
    <div class="wire"><span class="wlabel">on the wire · HTTP fetch</span>
      <span class="wf up">▲ {"action":"greet","data":{"name":"Ada"}}</span>
      <span class="wf dn">▼ {"tree":{"0":"Ada"}}</span>
    </div>
  </div>
</div></section>

<!-- STEP 4 · WORKS WITHOUT JS -->
<section><div class="wrap">
  <div class="sec-tag">Step 4 · Works without JavaScript</div>
  <h2>The same program. Two transports, side by side — try them.</h2>
  <p class="lead">Both cards run the <b>identical app</b>, WebSocket off. <b>Left, JS on:</b> the client sends a fetch and patches the headline in place. <b>Right, JS disabled:</b> the same <code>&lt;form&gt;</code> does a plain POST and the server renders the page — watch it reload. Progressive enhancement is a <b>transport flag, not a different app</b>. Type a name in each.</p>
  <div class="two" style="margin-top:28px">
    <div class="live-card">
      <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">JavaScript on · fetch + DOM patch</span></div>
      <div class="live-body">

```embed-lvt path="/apps/greet-nojs/" upstream="http://localhost:9091" height="200px"
```

</div>
    </div>
    <div class="live-card">
      <div class="live-bar"><span class="live-badge nojs">○ no JS</span><span class="live-meta">JavaScript off · form POST → full render</span></div>
      <div class="live-body">
        <iframe class="nojs-frame" src="/apps/greet-nojs/" sandbox="allow-forms allow-same-origin" title="The greeting app with JavaScript disabled" loading="lazy"></iframe>
      </div>
    </div>
  </div>
  <div class="code" style="max-width:680px;margin:28px auto 0"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.tmpl · one form, either transport</span></div>
<pre><span class="com">&lt;!-- the only line that flips the transport: --&gt;</span>
<span class="tag">&lt;script</span> <span class="attr">defer src</span>=<span class="str">"…@livetemplate/client"</span><span class="tag">&gt;&lt;/script&gt;</span>

<span class="tag">&lt;form</span> <span class="attr">method</span>=<span class="str">"POST"</span><span class="tag">&gt;</span>   <span class="com">&lt;!-- JS on → fetch + patch · JS off → native POST --&gt;</span>
  <span class="tag">&lt;input</span> <span class="attr">name</span>=<span class="str">"name"</span><span class="tag">&gt;</span>
  <span class="tag">&lt;button</span> <span class="attr">name</span>=<span class="str">"greet"</span><span class="tag">&gt;</span>Say hi<span class="tag">&lt;/button&gt;</span>
<span class="tag">&lt;/form&gt;</span></pre></div>
  <p class="demo-cap" style="margin-top:14px">Same <code>&lt;form&gt;</code> and the same <code>Greet</code> action as Step 1 — no <code>if jsEnabled</code> branch anywhere. The client enhances the submit when the <code>&lt;script&gt;</code> loads, and the browser falls back to a native POST when it doesn't.</p>
  <p class="demo-cap" style="margin-top:18px">The third transport — <b>real-time WebSocket push</b>, where the server updates the page with no request at all — is live in the next three steps. ↓</p>
</div></section>

<!-- STEP 5 · YOUR TABS -->
<section class="alt"><div class="wrap">
  <div class="sec-tag">Step 5 · Sync your own tabs</div>
  <h2>Turn on real-time. Your tabs move together.</h2>
  <p class="lead">Subscribe a connection to its own topic and publish after a handler runs — <b>two calls</b> — and your greeting syncs across every open tab. Every reactive thing LiveTemplate does is this same four-step pipeline.</p>
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
  <p class="demo-cap"><b>Open this page in a second tab</b>, greet in either, and your headline updates in <b>both</b> — live, no reload. Your tabs share one self-topic.</p>
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
  <p class="demo-cap" style="margin-top:14px">No magic: <code>Publish(ctx.SelfTopic(), "Refresh", nil)</code> just <b>runs your <code>Refresh</code> method on your other tabs</b>. It re-reads the shared name and returns new state; the framework diffs and patches. A name lives in a tiny per-tab-group store because each tab has its own <code>State</code>.</p>
  <div class="wire" style="max-width:820px;margin:14px auto 0"><span class="wlabel">on the wire · WebSocket</span>
    <span class="wf up">▲ this tab · {"action":"greet","data":{"name":"Ada"}}</span>
    <span class="wf dn">▼ your other tab · {"tree":{"0":"Ada"}}</span>
  </div>
</div></section>

<!-- STEP 6 · EVERYONE -->
<section><div class="wrap">
  <div class="sec-tag">Step 6 · A wall everyone shares</div>
  <h2>One more topic, and it's cross-user.</h2>
  <p class="lead">Swap the self-topic for a <b>shared</b> topic — admitted by a tiny ACL — and the same publish fans out to <b>every visitor</b>. The two cards below are <b>separate sessions, like two different people</b>. Greet in one and your line lands on the other's wall, live.</p>
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
  <p class="demo-cap">Two independent sessions, one shared wall — type in either card and watch the <b>list</b> appear in both. Every greeting here is real, typed by someone else reading this page.</p>
  <div class="two code-right" style="margin-top:26px">
    <div>
      <p class="lead">Headlines stay independent (each card is its own session), but the wall is global — so a greeting crosses from one session to the other. That's the whole cross-user story: the same two pub/sub calls as step 5, with a different topic.</p>
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

<!-- STEP 7 · SERVER SPEAKS -->
<section class="alt"><div class="wrap two code-right">
  <div>
    <div class="sec-tag">Step 7 · The server speaks first</div>
    <h2>No click required.</h2>
    <p class="lead">Every update so far began with a user. But a connection runs both ways: hold a <code>Session</code> handle and the server pushes on its own. A heartbeat <b>refreshes one value in place</b> — a downstream patch with nothing going up.</p>
    <div class="wire"><span class="wlabel">on the wire · WebSocket</span>
      <span class="wf dn">▼ {"tree":{"3":{"0":"15:04:08"}}}</span>
      <span class="wf note">(no ▲ — the server started it; just the one changed value)</span>
    </div>
  </div>
  <div>
    <div class="code delta"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go · a handle, and a heartbeat</span></div>
<pre class="language-go"><code class="language-go">func (a *App) OnConnect(s State, ctx *lvt.Context) (State, error) {
    a.keep(ctx.GroupID(), ctx.Session())   // remember who's connected
    return s, nil
}
func (a *App) heartbeat() {
    for range time.Tick(30 * time.Second) {
        a.serverAt = now()                       // replace ONE slot, in place
        for _, sess := range a.sessions {
            sess.TriggerAction("ServerRefresh", nil)   // push, unprompted
        }
    }
}</code></pre></div>
    <p class="demo-cap">That's the full arc: one greeting app grew into a live, shared, self-updating wall — and you never left Go or wrote a line of client JavaScript.</p>
  </div>
</div></section>

<!-- DIFF -->
<section><div class="wrap two">
  <div>
    <div class="sec-tag">Only the diff goes over the wire</div>
    <h2>Send what changed, not the whole page.</h2>
    <p class="lead">Templates split into <b>static structure (cached)</b> and <b>dynamic values</b>. On change, LiveTemplate sends only the changed values — typically <b>85%+ less bandwidth</b> than shipping HTML. A greeting comes back as <code>{"tree":{"0":"Ada"}}</code>, not a page.</p>
  </div>
  <div class="bars">
    <div class="bar-row"><span class="lab">full HTML</span><span class="bar full"><span></span></span><span class="val" style="color:var(--slate-2)">2.4 KB</span></div>
    <div class="bar-row"><span class="lab">lvt diff</span><span class="bar diff"><span></span></span><span class="val" style="color:var(--sig-d)">340 B</span></div>
    <div style="font:500 13px 'JetBrains Mono';color:var(--slate);margin-top:8px">↓ 86% smaller per update</div>
  </div>
</div></section>

<!-- FEATURES -->
<section class="alt"><div class="wrap">
  <div class="sec-tag">And so much more</div>
  <h2>Batteries for real apps.</h2>
  <div class="grid">
    <a class="feat" href="/reference/uploads"><div class="ico"><svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg></div><h4>File uploads</h4><p>Add <code>lvt-upload</code> to a file input — chunked over the WebSocket with live progress, no extra route.</p></a>
    <a class="feat" href="/reference/pubsub"><div class="ico"><svg viewBox="0 0 24 24"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><line x1="12" y1="20" x2="12.01" y2="20"/></svg></div><h4>Pub/Sub</h4><p><code>Subscribe</code>/<code>Publish</code> for multi-tab and cross-user fan-out — the engine behind steps 5–7.</p></a>
    <a class="feat" href="/reference/session"><div class="ico"><svg viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg></div><h4>Sessions &amp; state</h4><p>Server-owned, per-session state. No cross-user leaks.</p></a>
    <a class="feat" href="/reference/error-handling"><div class="ico"><svg viewBox="0 0 24 24"><path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg></div><h4>Error handling</h4><p>Actions return <code>(State, error)</code>; field errors flow to the template, as in step 2.</p></a>
    <a class="feat" href="/cli/"><div class="ico"><svg viewBox="0 0 24 24"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg></div><h4>CLI (lvt)</h4><p>Scaffolds, dev server, component kits.</p></a>
    <a class="feat" href="/client/"><div class="ico"><svg viewBox="0 0 24 24"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg></div><h4>TypeScript client</h4><p><code>@livetemplate/client</code> on npm, ~75% smaller updates.</p></a>
    <a class="feat" href="/guides/observability"><div class="ico"><svg viewBox="0 0 24 24"><line x1="6" y1="20" x2="6" y2="14"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="18" y1="20" x2="18" y2="10"/></svg></div><h4>Observability</h4><p>Structured hooks for metrics and tracing.</p></a>
    <a class="feat" href="/guides/scaling"><div class="ico"><svg viewBox="0 0 24 24"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg></div><h4>Scaling</h4><p>Session groups, fan-out limits, deploy guidance.</p></a>
  </div>
</div></section>

<!-- COMPARE -->
<section><div class="wrap">
  <div class="sec-tag">How it compares</div>
  <h2>Others annotate HTML to make it reactive. LiveTemplate moves that to the server.</h2>
  <p class="lead">Other tools carry behavior in the markup — <code>hx-*</code>, <code>x-*</code>, <code>phx-*</code>, or a DSL. Here a plain <code>&lt;button name="greet"&gt;</code> is already the action, and state lives on the server. <code>lvt-*</code> attributes exist only as an escape hatch for what HTML can't express.</p>
  <table class="cmp">
    <thead><tr><th>If you’re using…</th><th>LiveTemplate gives you…</th></tr></thead>
    <tbody>
      <tr><td>htmx</td><td class="give">Standard HTML actions <span class="badge">server-owned state</span> with DOM diffing — no <code>hx-*</code> request wiring in the markup.</td></tr>
      <tr><td>templ + htmx</td><td class="give">Go’s own <code>html/template</code> instead of a new DSL, reactivity built in instead of wiring.</td></tr>
      <tr><td>Alpine.js</td><td class="give">Reactive UI <span class="badge">no client state model</span> — state lives on the server, not in <code>x-*</code> on the element.</td></tr>
      <tr><td>Phoenix LiveView</td><td class="give">Stateful server-driven UI without leaving Go — and it works over plain HTTP too.</td></tr>
      <tr><td>React SPA</td><td class="give">Reactive workflows without a client build step for common app screens.</td></tr>
    </tbody>
  </table>
</div></section>

<!-- DOGFOOD -->
<section class="alt"><div class="wrap">
  <div class="dogfood">
    <div><svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="#047857" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg></div>
    <p><b>Built in Go. This site proves it.</b> Every step above is a real LiveTemplate app, embedded live through this docs site — which itself runs on LiveTemplate + tinkerdown. <a href="/recipes/how-this-site-works">See how this site works →</a></p>
  </div>
</div></section>

<!-- FINAL -->
<section class="final"><div class="wrap">
  <h2>HTML apps in Go. Start in 30 seconds.</h2>
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
