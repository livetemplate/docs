---
title: "Chat"
description: "A runnable realtime chat app showing LiveTemplate actions, shared state, server push, and scroll behavior."
source_repo: "https://github.com/livetemplate/docs"
source_path: "content/recipes/apps/chat.md"
---

# Building a real-time chat app

A tutorial for building a real-time chat room on LiveTemplate's simple kit: multi-tab sync, session management and reactive UI updates, in two files.

## What you'll build

- Real-time messaging, synced across a browser's tabs
- User login and presence tracking
- Updates that reach the other tabs without a reload
- Browser session isolation (each browser has its own chat room)
- Message history and timestamps

Two files: `main.go` and `chat.tmpl`. Every snippet here comes straight out of
`examples/chat/main.go`, so it cannot drift from the app you are running.

## Quick start

```bash
cd examples/chat
GOWORK=off go run main.go
```

Then open <http://localhost:8090> in **two or more browser tabs**:
- A message sent in one tab shows up in the others
- Each browser gets its own isolated chat session

## Building it from scratch

### Step 1: create a new app

Start by creating a new LiveTemplate application with the `simple` kit:

```bash
lvt new chat --kit simple
cd chat
```

The `simple` kit generates a minimal structure:

- `main.go` - Application logic (single file)
- `chat.tmpl` - HTML template (single file)
- `go.mod` - Go module configuration
- `README.md` - Documentation

No `cmd/`, no `internal/`, no database. A larger app will grow some of those; a
chat room this size doesn't need them.

### Step 2: define the controller and the state

Two types, and the split between them is the thing to get right.

The **controller** is a singleton. It holds what every tab must agree on — the
message list, who is online — behind a mutex.

The **state** is per connection. Each tab gets its own copy, which is why
`CurrentUser` can differ between two tabs of the same browser.

```go include="/examples/chat/main.go" lines="16-37"
```

Put dependencies on the controller and serializable UI data in the state. Put a
mutex on the state struct and each connection gets its own copy, guarding
nothing.

### Step 3: subscribe, then publish

`Mount` runs once per session group. It opts this connection into its own topic
and seeds the state from the controller:

```go include="/examples/chat/main.go" lines="41-51"
```

`Send` appends to the shared list, then tells the peers:

```go include="/examples/chat/main.go" lines="101-125"
```

`Publish` does not push state. It runs a named action — here `NewMessage` — on
every other subscribed connection, and that action rebuilds its own tab's view:

```go include="/examples/chat/main.go" lines="129-135"
```

You need both. Without the `Subscribe` in `Mount` the publish reaches nobody;
without the `Publish` no peer ever runs. Neither happens on its own.

`OnConnect` fires per WebSocket rather than per session group, which is how a
second tab starts logged out instead of inheriting the first tab's user:

```go include="/examples/chat/main.go" lines="55-63"
```

### Step 4: wire it up

```go include="/examples/chat/main.go" lines="179-198"
```

`Handle` takes the controller and the initial state. Method names are the action
names, so `<button name="send">` reaches `Send` with nothing to register.

### Step 5: create the UI

Replace `chat.tmpl` with the chat interface. Key template concepts:

**Conditional Rendering:**

```html
{{if not .CurrentUser}}
    <!-- Show login form -->
{{else}}
    <!-- Show chat interface -->
{{end}}
```

**Message Loop:**

```html
{{range .Messages}}
<div class="message {{if eq .Username $.CurrentUser}}mine{{end}}">
    <div class="message-header">
        <span class="message-username">{{.Username}}</span>
        <span class="message-time">{{.Timestamp}}</span>
    </div>
    <div class="message-text">{{.Text}}</div>
</div>
{{end}}
```

**Form Actions:**

```html
<form method="POST" name="join">
    <input type="text" name="username" required autofocus>
    <button type="submit">Join Chat</button>
</form>

<form method="POST" name="send">
    <input type="text" name="message" autocomplete="off">
    <button type="submit">Send</button>
</form>
```

**Auto-scroll Script:**

```html
<script>
    {{if .CurrentUser}}
    function scrollToBottom() {
        const messages = document.getElementById('messages');
        if (messages) {
            messages.scrollTop = messages.scrollHeight;
        }
    }

    scrollToBottom();

    if (window.LiveTemplate) {
        const originalUpdate = window.LiveTemplate.prototype.updateDOM;
        window.LiveTemplate.prototype.updateDOM = function(...args) {
            originalUpdate.apply(this, args);
            setTimeout(scrollToBottom, 50);
        };
    }
    {{end}}
</script>
```

### Step 6: run and test

```bash
go run main.go
```

Open <http://localhost:8090> in multiple browser tabs:

**Test 1 - Same browser, multiple tabs:**

- Open 2+ tabs in Chrome
- Login with any username in tab 1
- Send a message in tab 1
- It shows up in tab 2
- Send from tab 2 and it shows up in tab 1

**Test 2 - Different browsers (isolated sessions):**

- Open Chrome and Firefox
- Each browser gets its own chat room
- Messages in Chrome don't appear in Firefox
- Each browser maintains separate state

## How it works

### How tabs stay in sync

```text
Chrome Tab 1       Server (Go)        Chrome Tab 2
    |                   |                     |
    |---- join -------->|                     |
    |          [groupID: session-abc]         |
    |                   |<------ join --------|
    |          [Same groupID: session-abc]    |
    |                   |                     |
    |--- send msg ----->|                     |
    |    [Send appends, then Publish's        |
    |     NewMessage runs on tab 2]           |
    |<---- update ------|------- update ----->|
    |                   |                     |
```

What happens, in order:

1. Each browser gets a session group ID in a cookie. Tabs in the same browser
   share it.
2. `Mount` subscribes the connection to `ctx.SelfTopic()` — the topic scoped to
   that session group.
3. `Send` calls `ctx.Publish(ctx.SelfTopic(), "NewMessage", nil)`, which
   runs `NewMessage` on the other subscribed tabs.
4. Each tab re-renders, and the server sends only the parts that changed.

Both halves are explicit. `main.go` has one `Subscribe` and three `Publish`
calls, and they are the entire sync mechanism — nothing fans out on its own.

### What you don't write

You don't manage the WebSocket, and there are no API endpoints to define. The
server re-renders the template and sends the diff, so there's no client-side copy
of the message list to keep in step.

What you do write is on this page: two structs, the action methods, and the
`Subscribe`/`Publish` pair above.

## Things to add

### Add persistence

Messages live on the controller and die with the process. Give the controller a
store and read it in `Mount`:

```go
type ChatController struct {
    mu       sync.RWMutex
    store    MessageStore   // your database, file, whatever
    users    map[string]bool
}

func (c *ChatController) Mount(state ChatState, ctx *livetemplate.Context) (ChatState, error) {
    if err := ctx.Subscribe(ctx.SelfTopic()); err != nil {
        return state, err
    }
    msgs, err := c.store.Recent(ctx.Request().Context(), 50)
    if err != nil {
        return state, err
    }
    state.Messages = msgs
    return state, nil
}
```

The store belongs on the controller, not the state — every connection clones the
state and round-trips it through JSON.

### Add typing indicators

A new action method and a publish, the same shape as `Send`:

```go
func (c *ChatController) Typing(state ChatState, ctx *livetemplate.Context) (ChatState, error) {
    c.mu.Lock()
    c.typing[state.CurrentUser] = time.Now()
    c.mu.Unlock()
    return state, ctx.Publish(ctx.SelfTopic(), "TypingChanged", nil)
}

func (c *ChatController) TypingChanged(state ChatState, ctx *livetemplate.Context) (ChatState, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    state.TypingUsers = c.activeTypers()
    return state, nil
}
```

### Add message reactions

The button carries the id and the emoji; the server reads them off the form:

```go
// <button name="react" value="{{.ID}}"> with a hidden emoji input
func (c *ChatController) React(state ChatState, ctx *livetemplate.Context) (ChatState, error) {
    id, emoji := ctx.GetInt("value"), ctx.GetString("emoji")
    c.mu.Lock()
    c.reactions[id][emoji]++
    c.mu.Unlock()
    return state, ctx.Publish(ctx.SelfTopic(), "NewMessage", nil)
}
```

### Add chat rooms

A room is a topic. Subscribe to the room instead of `SelfTopic()` and every
member of that room receives the publish — including other people, which
`SelfTopic()` deliberately will not do:

```go
func (c *ChatController) JoinRoom(state ChatState, ctx *livetemplate.Context) (ChatState, error) {
    state.CurrentRoom = ctx.GetString("room")
    return state, ctx.Subscribe("room:" + state.CurrentRoom)
}
```

Developer topics are deny-all until named, so a room app also needs
`WithTopicACL` in `main`. See [Pubsub](/recipes/pubsub).

## Before production

### 1. Load the client library from the CDN

In `chat.tmpl` — the framework function renders the pinned CDN URL for the
client this server release is wire-compatible with, so the two stay in
lockstep:

```html
<link rel="stylesheet" href="{{lvtClientStyleURL}}">
<script defer src="{{lvtClientScriptURL}}"></script>
```

### 2. Add rate limiting

Keep the clock on the controller, keyed by session group — a per-connection
field is trivially reset by opening a new tab:

```go
func (c *ChatController) Send(state ChatState, ctx *livetemplate.Context) (ChatState, error) {
    c.mu.Lock()
    last := c.lastSend[ctx.GroupID()]
    if time.Since(last) < time.Second {
        c.mu.Unlock()
        return state, nil
    }
    c.lastSend[ctx.GroupID()] = time.Now()
    c.mu.Unlock()
    // ... append and publish
}
```

### 3. Add message limits

```go
if len(s.Messages) > 100 {
    s.Messages = s.Messages[len(s.Messages)-100:]  // Keep last 100
}
```

### 4. Add authentication

For production, use real auth instead of just username:

```go
auth := livetemplate.NewBasicAuthenticator(func(username, password string) (bool, error) {
    return validateUser(username, password)
})

tmpl := livetemplate.New("chat",
    livetemplate.WithDevMode(false),
    livetemplate.WithAuthenticator(auth),
)
```

### 5. Create a global chat room (cross-browser)

By default, each browser has its own isolated chat. To make all users share the same chat room:

```go
// Custom authenticator that puts everyone in same session group
type GlobalChatAuthenticator struct{}

func (a *GlobalChatAuthenticator) Identify(r *http.Request) (string, error) {
    return "", nil // Anonymous
}

func (a *GlobalChatAuthenticator) GetSessionGroup(r *http.Request, userID string) (string, error) {
    return "global-chat-room", nil // Everyone shares same group!
}

// Use it:
tmpl := livetemplate.New("chat",
    livetemplate.WithDevMode(true),
    livetemplate.WithAuthenticator(&GlobalChatAuthenticator{}),
)
```

Now Chrome, Firefox and Safari share one room instead of getting one each.

## What this recipe showed

Two files: `main.go` and `chat.tmpl`. Messages are Go structs, so there's no JSON
marshaling between the server and the template. Tabs stay in step through the
`Subscribe`/`Publish` pair, and the server sends only the parts of the page that
changed.

What it doesn't show is persistence. Messages live in a slice on the controller
and are gone when the process restarts — see [Add Persistence](#add-persistence).

## Compared with the counter example

The counter is the smaller version of the same shape:

| Counter | Chat |
|---|---|
| `AppState{Counter int}` | `ChatState{Messages []Message}` |
| `increment/decrement` actions | `join/send` actions |
| Single user | Multi-user with broadcasting |
| Simple int update | List of messages |

Same shape, different data.

## Next

- Try the `counter` example for a simpler starting point
- Try the `todos` example for CRUD operations
- Use `lvt new myapp --kit multi` for apps needing databases

## Related

- [Server API reference](/reference/api)
- [PubSub reference](/reference/pubsub) — the fan-out this app is built on
- [Server push](/recipes/server-push)
- [Template Syntax](https://pkg.go.dev/html/template)
- [LiveTemplate API](https://pkg.go.dev/github.com/livetemplate/livetemplate)
