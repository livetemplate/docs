---
title: "Chat"
description: "A runnable realtime chat app showing LiveTemplate actions, shared state, server push, and scroll behavior."
source_repo: "https://github.com/livetemplate/docs"
source_path: "content/recipes/apps/chat.md"
---

# Building a real-time chat app

A complete tutorial for building a real-time chat application using LiveTemplate's simple kit. This demonstrates **automatic multi-tab syncing**, session management, and reactive UI updates with just **2 files**.

## What you'll build

- Real-time messaging with automatic tab syncing
- User login and presence tracking
- Instant UI updates across all tabs in the same browser
- Browser session isolation (each browser has its own chat room)
- Message history and timestamps

**All in just 2 files: `main.go` and `chat.tmpl`**

## Quick start

```bash
cd examples/chat
GOWORK=off go run main.go
```

Then open <http://localhost:8090> in **multiple browser tabs** to see automatic syncing in action:
- Messages sent in one tab appear instantly in all other tabs
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

### Step 2: define the chat state

Open `main.go` and replace the counter example with chat state:

```go
package main

import (
    "log"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/livetemplate/livetemplate"
)

type ChatState struct {
    Messages      []Message
    Users         map[string]*User
    CurrentUser   string
    OnlineCount   int
    TotalMessages int
    mu            sync.RWMutex  // Thread-safe access
}

type Message struct {
    ID        int
    Username  string
    Text      string
    Timestamp string
}

type User struct {
    Username string
    JoinedAt time.Time
    IsOnline bool
}
```

**Key concepts:**

- Single `ChatState` struct holds all app state
- `sync.RWMutex` for thread-safe concurrent access
- Simple Go structs - no database, no ORM, no complexity

### Step 3: implement actions

Add the `Change` method to handle user actions:

```go
func (s *ChatState) Change(ctx *livetemplate.ActionContext) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    switch ctx.Action {
    case "send":
        var data struct {
            Message string `json:"message"`
        }

        if err := ctx.Bind(&data); err != nil {
            return nil
        }

        if data.Message == "" {
            return nil
        }

        s.TotalMessages++
        msg := Message{
            ID:        s.TotalMessages,
            Username:  s.CurrentUser,
            Text:      data.Message,
            Timestamp: time.Now().Format("15:04:05"),
        }

        s.Messages = append(s.Messages, msg)
        return nil  // Auto-syncs to all tabs in same browser!

    case "join":
        var data struct {
            Username string `json:"username"`
        }

        if err := ctx.Bind(&data); err != nil {
            return nil
        }

        s.CurrentUser = data.Username

        if _, exists := s.Users[data.Username]; !exists {
            s.Users[data.Username] = &User{
                Username: data.Username,
                JoinedAt: time.Now(),
                IsOnline: true,
            }
            s.updateOnlineCount()
        }

        return nil
    }

    return nil
}

func (s *ChatState) updateOnlineCount() {
    count := 0
    for _, user := range s.Users {
        if user.IsOnline {
            count++
        }
    }
    s.OnlineCount = count
}
```

**Key concepts:**

- Actions route via `<form name="join">` and `<form name="send">` (button/form `name` routing)
- `ctx.GetString("field")` extracts form data
- Mutating state is not enough on its own — `Subscribe` opts a connection in,
  and `Publish` is what reaches the peers
- You don't write WebSocket code, but you do write both of those

### Step 4: initialize and run

Add initialization and main function:

```go
func (s *ChatState) Init() error {
    if s.Users == nil {
        s.Users = make(map[string]*User)
    }
    if s.Messages == nil {
        s.Messages = []Message{}
    }
    return nil
}

func main() {
    log.Println("chat starting...")

    state := &ChatState{
        Users:    make(map[string]*User),
        Messages: []Message{},
    }

    tmpl := livetemplate.Must(livetemplate.New("chat", livetemplate.WithDevMode(true)))
    http.Handle("/", tmpl.Handle(controller, livetemplate.AsState(state)))

    port := os.Getenv("PORT")
    if port == "" {
        port = "8090"
    }

    log.Printf("🚀 Chat server starting on http://localhost:%s", port)
    log.Println("📝 Open multiple browser tabs to test multi-user chat")
    log.Println("💬 Messages are broadcast to all connected users")

    http.ListenAndServe(":"+port, nil)
}
```

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
    |         [Auto-broadcast to group]       |
    |<---- update ------|------- update ----->|
    |                   |                     |
```

What happens, in order:

1. Each browser gets a session group ID in a cookie. Tabs in the same browser
   share it.
2. `Mount` subscribes the connection to `ctx.SelfTopic()` — the topic scoped to
   that session group.
3. `SendMessage` calls `ctx.Publish(ctx.SelfTopic(), "NewMessage", nil)`, which
   runs `NewMessage` on the other subscribed tabs.
4. Each tab re-renders, and the server sends only the parts that changed.

Both halves are explicit. `main.go` has one `Subscribe` and three `Publish`
calls, and they are the entire sync mechanism — nothing fans out on its own.

### What you don't write

You don't manage the WebSocket, and there are no API endpoints to define. The
server re-renders the template and sends the diff, so there's no client-side copy
of the message list to keep in step.

What you do write is all on this page: a state struct, four methods, and the
`Subscribe`/`Publish` pair above.

## Things to add

### Add persistence

Store messages in a slice that survives restarts:

```go
var persistedMessages []Message

func (s *ChatState) Init() error {
    s.Messages = persistedMessages  // Load from memory
    // Or load from file: loadFromJSON("messages.json")
    return nil
}

func (s *ChatState) Change(ctx *livetemplate.ActionContext) error {
    // ... after adding message
    persistedMessages = s.Messages  // Save to memory
    // Or save to file: saveToJSON("messages.json", s.Messages)
}
```

### Add typing indicators

```go
type ChatState struct {
    // ... existing fields
    TypingUsers map[string]bool
}

// In Change()
case "typing":
    var data struct {
        Username string `json:"username"`
    }
    ctx.Bind(&data)
    s.TypingUsers[data.Username] = true
    // Auto-broadcast!
```

### Add message reactions

```go
type Message struct {
    // ... existing fields
    Reactions map[string]int  // emoji -> count
}

case "react":
    var data struct {
        MessageID int    `json:"messageId"`
        Emoji     string `json:"emoji"`
    }
    ctx.Bind(&data)
    s.Messages[data.MessageID].Reactions[data.Emoji]++
```

### Add chat rooms

```go
type ChatState struct {
    Rooms       map[string]*Room
    CurrentRoom string
}

type Room struct {
    Name     string
    Messages []Message
}
```

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

```go
case "send":
    if time.Since(s.LastMessageTime) < time.Second {
        return nil  // Too fast, ignore
    }
    // ... process message
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
