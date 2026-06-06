---
title: "LiveTemplate — HTML app development, made easy for Go"
description: "Build reactive, real-time web apps in Go with standard HTML. No client framework, no build step."
layout: landing
---

<header class="nav"><div class="wrap nav-in">
  <div class="brand"><span class="glyph">◇</span> LiveTemplate</div>
  <nav class="nav-links"><a href="/getting-started/your-first-app">Docs</a><a href="/recipes/">Recipes</a><a href="/reference/api">Reference</a><a href="https://github.com/livetemplate/livetemplate">GitHub</a><a class="btn btn-primary" href="/getting-started/install">Get started →</a></nav>
</div></header>

<!-- HERO -->
<section class="hero"><div class="wrap">
  <span class="eyebrow">Server-driven UI for Go · Alpha</span>
  <h1 class="head">HTML app development, <span class="g">made easy</span> for Go.</h1>
  <p class="sub">Build reactive, real-time web apps with <b>standard HTML and a Go controller</b> — no client framework, no build step, no JavaScript to write. Forms, buttons, and live updates just work.</p>
  <div class="cta-row">
    <a class="btn btn-primary btn-lg" href="/getting-started/install">Get started →</a>
    <a class="btn btn-ghost btn-lg" href="/getting-started/your-first-app">Read the docs</a>
  </div>
  <div class="hero-snip">
    <div class="live-card">
      <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">greet · running in this page</span></div>
      <div class="live-body">

```embed-lvt path="/apps/greet/" upstream="http://localhost:9091" height="200px"
```

</div>
    </div>
    <p class="hero-cap" style="margin:14px 0 8px">↑ a real, running app. Type a name and hit <b>Say hi</b> — and below is <b>every line</b> that makes it. Both files, complete:</p>
    <div class="code"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.tmpl &nbsp;— the entire template</span></div>
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
    <p class="hero-cap">That's it — a full reactive web app: <b>~20 lines of Go + standard HTML</b>. No client framework, no build step, no generated code.</p>
  </div>
</div></section>

<!-- UNDER THE HOOD: animated request/response over the wire -->
<section><div class="wrap">
  <div class="sec-tag">No hidden code</div>
  <h2>That's the whole program — not a snippet.</h2>
  <p class="lead">These are the <b>real frames</b> on the wire. Your action goes <b>up</b>, the server runs your method, and <b>only the changed value comes back</b> — the static HTML stays cached. No page reload, and no fetch, route, or JSON you ever wrote.</p>
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

<!-- FORMS -->
<section class="alt"><div class="wrap two code-right">
  <div>
    <div class="sec-tag">Forms that just work</div>
    <h2>Server-side forms, with validation built in.</h2>
    <p class="lead">Handle submissions with <b>standard <code>&lt;form&gt;</code> markup</b> and the form data Go already parses. A button’s <code>name</code> routes to a method; return an <code>error</code> and it renders inline. <b>No client-side state, no serialization, no JSON plumbing</b> — and progressive enhancement comes free.</p>
    <div class="live-card" style="margin-top:24px;max-width:420px">
      <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">signup · validates on submit</span></div>
      <div class="live-body"><form class="form-demo" onsubmit="return false">
        <label>Email</label><input class="bad" value="not-an-email">
        <p class="err">Enter a valid email address.</p>
        <button>Sign up</button>
      </form></div>
    </div>
  </div>
  <div>
    <div class="code"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">signup.tmpl</span></div>
<pre><span class="tag">&lt;form</span> <span class="attr">method</span>=<span class="str">"POST"</span><span class="tag">&gt;</span>
  <span class="tag">&lt;input</span> <span class="attr">name</span>=<span class="str">"email"</span> <span class="attr">required</span><span class="tag">&gt;</span>
  <span class="tag">&lt;button</span> <span class="attr">name</span>=<span class="str">"submit"</span><span class="tag">&gt;</span>Sign up<span class="tag">&lt;/button&gt;</span>
  {{<span class="kw">if</span> <span class="fn">.Error</span>}}<span class="tag">&lt;p</span> <span class="attr">class</span>=<span class="str">"err"</span><span class="tag">&gt;</span>{{<span class="fn">.Error</span>}}<span class="tag">&lt;/p&gt;</span>{{<span class="kw">end</span>}}
<span class="tag">&lt;/form&gt;</span></pre></div>
    <div class="code"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">signup.go</span></div>
<pre class="language-go"><code class="language-go">func (c *Signup) Submit(s State, ctx *lvt.Context) (State, error) {
    if !valid(ctx.GetString("email")) {
        return s, errors.New("Enter a valid email address.")
    }
    return s, nil
}</code></pre></div>
  </div>
</div></section>

<!-- REACTIVITY / STANDARD HTML -->
<section><div class="wrap">
  <div class="sec-tag">Reactivity, without the wiring</div>
  <h2>The button’s name is the action.</h2>
  <p class="lead"><code>&lt;button name="increment"&gt;</code> dispatches to the <code>Increment</code> method — that’s the whole protocol. You reach for an <code>lvt-*</code> attribute <b>only</b> when the behavior is something HTML itself cannot define — a debounce, a keyboard shortcut, a reactive class toggle — never as boilerplate to make ordinary HTML work.</p>
  <div class="two code-right" style="margin-top:28px">
    <div class="live-card">
      <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">counter · running in this page</span></div>
      <div class="live-body">

```embed-lvt path="/apps/counter-basic/" upstream="http://localhost:9091" height="150px"
```

</div>
    </div>
    <div>
      <div class="code"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">counter.tmpl</span></div>
<pre><span class="tag">&lt;h1&gt;</span>Counter: {{<span class="fn">.Counter</span>}}<span class="tag">&lt;/h1&gt;</span>
<span class="tag">&lt;button</span> <span class="attr">name</span>=<span class="str">"increment"</span><span class="tag">&gt;</span>+1<span class="tag">&lt;/button&gt;</span></pre></div>
      <div class="code"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">counter.go</span></div>
<pre class="language-go"><code class="language-go">func (c *Counter) Increment(s State, ctx *lvt.Context) (State, error) {
    s.Counter++
    return s, nil
}</code></pre></div>
    </div>
  </div>
</div></section>

<!-- TIER 2: lvt-* -->
<section class="alt"><div class="wrap">
  <div class="sec-tag">When HTML isn’t enough: lvt-*</div>
  <h2>One escape hatch — for what HTML can’t express.</h2>
  <p class="lead">Standard HTML can’t say “add a class <em>while the request is in flight</em>,” “debounce this input,” or “close when I click away.” That — and only that — is when you reach for an <code>lvt-*</code> attribute. Here a button gets a <code>loading</code> class for the duration of its action — no JavaScript:</p>
  <div class="two code-right" style="margin-top:26px">
    <div class="live-card">
      <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">reactive class · lvt-el</span></div>
      <div class="live-body">
        <button class="ld-btn is-loading"><span class="spin"></span> Saving…</button>
        <p class="demo-cap" style="margin-top:16px">the <code>loading</code> class is added on <code>pending</code>, removed on <code>done</code></p>
      </div>
    </div>
    <div>
      <div class="code"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">save.tmpl</span></div>
<pre><span class="tag">&lt;button</span> <span class="attr">name</span>=<span class="str">"save"</span>
  <span class="attr">lvt-el:addClass:on:pending</span>=<span class="str">"loading"</span>
  <span class="attr">lvt-el:removeClass:on:done</span>=<span class="str">"loading"</span><span class="tag">&gt;</span>Save<span class="tag">&lt;/button&gt;</span>

<span class="com">&lt;!-- the rest of the lvt-* family, only where HTML falls short --&gt;</span>
<span class="tag">&lt;input</span> <span class="attr">lvt-on:input</span>=<span class="str">"search"</span> <span class="attr">lvt-mod:debounce</span>=<span class="str">"300"</span><span class="tag">&gt;</span>
<span class="tag">&lt;div</span> <span class="attr">lvt-el:toggleClass:on:click</span>=<span class="str">"open"</span>
     <span class="attr">lvt-el:removeClass:on:click-away</span>=<span class="str">"open"</span><span class="tag">&gt;</span>…<span class="tag">&lt;/div&gt;</span></pre></div>
    </div>
  </div>
</div></section>

<!-- ONE MODEL -->
<section><div class="wrap">
  <div class="sec-tag">One model, every surface</div>
  <h2>The counter you build is the booking system you build.</h2>
  <p class="lead">Every reactive thing LiveTemplate does is the same four-step pipeline. A click runs it. A second tab runs it. A different user runs it. The server pushing on its own runs it. <b>You never learn a second model.</b></p>
  <div class="pipe" style="margin-top:28px">
    <div class="step"><div class="k">1 · state</div><div class="v">state changes</div></div><div class="arrow">→</div>
    <div class="step"><div class="k">2 · render</div><div class="v">re-render template</div></div><div class="arrow">→</div>
    <div class="step"><div class="k">3 · diff</div><div class="v">diff vs last render</div></div><div class="arrow">→</div>
    <div class="step"><div class="k">4 · patch</div><div class="v">patch the browser</div></div>
  </div>
  <div class="two" style="margin-top:34px">
    <div class="live-card"><div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">tab A</span></div><div class="live-body">

```embed-lvt path="/apps/counter/" upstream="http://localhost:9091" session="landing-sync" height="150px"
```

</div></div>
    <div class="live-card"><div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">tab B</span></div><div class="live-body">

```embed-lvt path="/apps/counter/" upstream="http://localhost:9091" session="landing-sync" height="150px"
```

</div></div>
  </div>
  <p class="demo-cap">Click <b>+1</b> in one tab — both stay in sync. Two server calls (<code>Subscribe</code> + <code>Publish</code>) make it real-time; no client code.</p>
</div></section>

<!-- PUB/SUB -->
<section class="alt"><div class="wrap">
  <div class="sec-tag">Pub/Sub</div>
  <h2>Real-time, in two calls.</h2>
  <p class="lead">Subscribe a connection to a topic; publish an action to that topic after a handler runs. The <b>same two calls</b> drive multi-tab sync and cross-user broadcast — only the topic changes.</p>
  <div class="code" style="max-width:780px;margin-top:22px"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">app.go</span></div>
<pre class="language-go"><code class="language-go">func (c *App) Mount(s State, ctx *lvt.Context) (State, error) {
    ctx.Subscribe(ctx.SelfTopic())                 // opt in to fan-out
    return s, nil
}
func (c *App) Add(s State, ctx *lvt.Context) (State, error) {
    s.Items = append(s.Items, newItem(ctx))
    ctx.Publish(ctx.SelfTopic(), "Refresh", nil)   // push to subscribed peers
    return s, nil
}</code></pre></div>
  <div class="callout-banner">Swap <code>SelfTopic()</code> for a shared topic and the same handler broadcasts to every user in the room — exactly what the seat picker below does. ↓</div>
</div></section>

<!-- CROSS USER -->
<section><div class="wrap">
  <div class="sec-tag">Cross-user, in real time</div>
  <h2>A different person’s clicks, updating your screen.</h2>
  <p class="lead">Most demos show <em>your</em> clicks updating <em>your</em> screen. This is a live seat map where <b>everyone booking the same event sees every selection in real time</b> — standard HTML, no hand-written JavaScript. Same pipeline; only the topic changes.</p>
  <div class="live-card" style="margin-top:26px">
    <div class="live-bar"><span class="live-badge"><span class="pulse"></span> cross-user · preview</span><span class="live-meta">seat-picker · everyone sees every pick</span></div>
    <div class="live-body" style="text-align:left;padding:26px">
      <div style="display:grid;grid-template-columns:repeat(8,1fr);gap:8px;max-width:520px;margin:0 auto">
        <div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#94a3b8"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#0f172a"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#e2e8f0"></div>
        <div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#047857"></div><div class="seat" style="--c:#047857"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#0f172a"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#e2e8f0"></div>
        <div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#047857"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#e2e8f0"></div><div class="seat" style="--c:#0f172a"></div><div class="seat" style="--c:#e2e8f0"></div>
      </div>
      <div style="display:flex;gap:18px;justify-content:center;margin-top:18px;font:500 13px 'JetBrains Mono';color:var(--slate);flex-wrap:wrap">
        <span style="display:inline-flex;align-items:center;gap:6px"><i style="width:12px;height:12px;border-radius:3px;background:#e2e8f0;display:block"></i> available</span>
        <span style="display:inline-flex;align-items:center;gap:6px"><i style="width:12px;height:12px;border-radius:3px;background:#047857;display:block"></i> you</span>
        <span style="display:inline-flex;align-items:center;gap:6px"><i style="width:12px;height:12px;border-radius:3px;background:#0f172a;display:block"></i> someone else (live)</span>
      </div>
    </div>
  </div>
  <p class="demo-cap"><code>&lt;button name="selectSeat" value="A5"&gt;</code> — the whole interaction vocabulary. <a href="/recipes/apps/seat-picker">Open the live seat picker →</a></p>
</div></section>
<style>.seat{aspect-ratio:1;border-radius:6px;background:var(--c)}</style>

<!-- UPLOADS -->
<section class="alt"><div class="wrap two">
  <div>
    <div class="sec-tag">File uploads</div>
    <h2>Uploads with live progress — no extra endpoints.</h2>
    <p class="lead">Add <code>lvt-upload</code> to a file input. Files stream over the WebSocket in chunks and progress renders straight into your template — <b>no multipart route, no client upload library</b>.</p>
    <div class="code" style="margin-top:18px"><div class="code-bar"><span class="dots"><i></i><i></i><i></i></span><span class="file">avatar.tmpl</span></div>
<pre><span class="tag">&lt;input</span> <span class="attr">type</span>=<span class="str">"file"</span> <span class="attr">lvt-upload</span>=<span class="str">"avatar"</span> <span class="attr">name</span>=<span class="str">"avatar"</span><span class="tag">&gt;</span>
<span class="tag">&lt;button</span> <span class="attr">name</span>=<span class="str">"save-profile"</span><span class="tag">&gt;</span>Save<span class="tag">&lt;/button&gt;</span>
{{<span class="kw">if</span> <span class="fn">.Uploading</span>}}<span class="tag">&lt;progress</span> <span class="attr">value</span>=<span class="str">"{{.Pct}}"</span> <span class="attr">max</span>=<span class="str">"100"</span><span class="tag">&gt;&lt;/progress&gt;</span>{{<span class="kw">end</span>}}</pre></div>
  </div>
  <div class="live-card">
    <div class="live-bar"><span class="live-badge"><span class="pulse"></span> live</span><span class="live-meta">avatar · chunked over WebSocket</span></div>
    <div class="live-body"><div class="up">
      <div class="row"><b>headshot.png</b><span>2.0 MB</span></div>
      <div class="track"><span></span></div>
      <div class="pct">64% · 1.3 MB / 2.0 MB · uploading…</div>
    </div></div>
  </div>
</div></section>

<!-- PROGRESSIVE ENHANCEMENT -->
<section><div class="wrap">
  <div class="sec-tag">Works everywhere — progressive enhancement</div>
  <h2>Same Go code. Same template. Three transports.</h2>
  <p class="lead">Progressive enhancement is a <b>transport concern, not a different application model.</b> Your app code never needs separate handlers for these modes.</p>
  <div class="tiers">
    <div class="tier"><div class="lvl">no javascript</div><h4>Form POST</h4><p>The form submits normally; the browser navigates to the server-rendered response. It just works.</p></div>
    <div class="tier"><div class="lvl">javascript</div><h4>fetch + DOM patch</h4><p>The client intercepts the form, sends HTTP, and patches the DOM in place — no full reload.</p></div>
    <div class="tier"><div class="lvl">+ websocket</div><h4>Real-time push</h4><p>Actions stream over a WebSocket; the server can push any time. Multi-tab and cross-user sync.</p></div>
  </div>
</div></section>

<!-- DIFF -->
<section class="alt"><div class="wrap two">
  <div>
    <div class="sec-tag">Only the diff goes over the wire</div>
    <h2>Send what changed, not the whole page.</h2>
    <p class="lead">Templates split into <b>static structure (cached)</b> and <b>dynamic values</b>. When state changes, LiveTemplate diffs the new render against the last one and sends only the changed values — typically <b>85%+ less bandwidth</b> than shipping HTML.</p>
  </div>
  <div class="bars">
    <div class="bar-row"><span class="lab">full HTML</span><span class="bar full"><span></span></span><span class="val" style="color:var(--slate-2)">2.4 KB</span></div>
    <div class="bar-row"><span class="lab">lvt diff</span><span class="bar diff"><span></span></span><span class="val" style="color:var(--sig-d)">340 B</span></div>
    <div style="font:500 13px 'JetBrains Mono';color:var(--slate);margin-top:8px">↓ 86% smaller per update</div>
  </div>
</div></section>

<!-- FEATURES -->
<section><div class="wrap">
  <div class="sec-tag">And so much more</div>
  <h2>Batteries for real apps.</h2>
  <div class="grid">
    <div class="feat"><div class="ico"><svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg></div><h4>File uploads</h4><p>WebSocket-backed chunked uploads with live progress.</p></div>
    <div class="feat"><div class="ico"><svg viewBox="0 0 24 24"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><line x1="12" y1="20" x2="12.01" y2="20"/></svg></div><h4>Pub/Sub</h4><p><code>Subscribe</code>/<code>Publish</code> for multi-tab and cross-user fan-out.</p></div>
    <div class="feat"><div class="ico"><svg viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg></div><h4>Sessions &amp; state</h4><p>Server-owned, per-session state. No cross-user leaks.</p></div>
    <div class="feat"><div class="ico"><svg viewBox="0 0 24 24"><path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg></div><h4>Error handling</h4><p>Actions return <code>(State, error)</code>; validation flows to the template.</p></div>
    <div class="feat"><div class="ico"><svg viewBox="0 0 24 24"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg></div><h4>CLI (lvt)</h4><p>Scaffolds, dev server, component kits.</p></div>
    <div class="feat"><div class="ico"><svg viewBox="0 0 24 24"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg></div><h4>TypeScript client</h4><p><code>@livetemplate/client</code> on npm, ~75% smaller updates.</p></div>
    <div class="feat"><div class="ico"><svg viewBox="0 0 24 24"><line x1="6" y1="20" x2="6" y2="14"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="18" y1="20" x2="18" y2="10"/></svg></div><h4>Observability</h4><p>Structured hooks for metrics and tracing.</p></div>
    <div class="feat"><div class="ico"><svg viewBox="0 0 24 24"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg></div><h4>Scaling</h4><p>Session groups, fan-out limits, deploy guidance.</p></div>
  </div>
</div></section>

<!-- COMPARE -->
<section class="alt"><div class="wrap">
  <div class="sec-tag">How it compares</div>
  <h2>Others add a layer. LiveTemplate keeps HTML standard.</h2>
  <p class="lead">Other tools make HTML reactive by adding attributes (<code>hx-*</code>, <code>x-*</code>, <code>phx-*</code>) or a DSL. LiveTemplate moves reactivity to the server, where one render-and-diff pipeline already lives.</p>
  <table class="cmp">
    <thead><tr><th>If you’re using…</th><th>LiveTemplate gives you…</th></tr></thead>
    <tbody>
      <tr><td>htmx</td><td class="give">Standard HTML actions <span class="badge">no hx-*</span> with server-owned state and DOM diffing.</td></tr>
      <tr><td>templ + htmx</td><td class="give">Go’s own <code>html/template</code> instead of a new DSL, reactivity built in instead of wiring.</td></tr>
      <tr><td>Alpine.js</td><td class="give">Reactive DOM behavior <span class="badge">no x-*</span> and no separate client-side state model.</td></tr>
      <tr><td>Phoenix LiveView</td><td class="give">Stateful server-driven UI without leaving Go — and it works over plain HTTP too.</td></tr>
      <tr><td>React SPA</td><td class="give">Reactive workflows without a client build step for common app screens.</td></tr>
    </tbody>
  </table>
</div></section>

<!-- DOGFOOD -->
<section><div class="wrap">
  <div class="dogfood">
    <div><svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="#047857" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg></div>
    <p><b>Built in Go. This site proves it.</b> Every demo above is a real LiveTemplate app, embedded live through this docs site — which itself runs on LiveTemplate + tinkerdown. <a href="/recipes/how-this-site-works">See how this site works →</a></p>
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
  <div class="foot-links"><a href="/getting-started/your-first-app">Docs</a><a href="/recipes/">Recipes</a><a href="/reference/api">Reference</a><a href="https://github.com/livetemplate/livetemplate">GitHub</a><a href="/changelog">Changelog</a><a href="https://github.com/livetemplate/livetemplate/blob/main/LICENSE">License</a></div>
  <div style="font-size:13px">© the LiveTemplate authors</div>
</div></footer>
