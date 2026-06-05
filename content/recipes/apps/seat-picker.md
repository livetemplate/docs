---
title: "Seat Picker"
description: "A cross-user, real-time seat-booking hall built entirely from standard HTML — different people see each other's selections live over a shared pub/sub topic."
source_repo: "https://github.com/livetemplate/docs"
source_path: "content/recipes/apps/seat-picker.md"
---

# Seat Picker — cross-user real time, standard HTML

Most reactive demos show *your* clicks updating *your* screen. This one
shows a different person's clicks updating *your* screen — a live seat map
where everyone booking the same event sees every selection in real time.

It is the recipe that exercises all three of LiveTemplate's claims at once,
and it does so without a single custom attribute on the markup or a line of
hand-written JavaScript. The full source is
[`examples/seat-picker/`](https://github.com/livetemplate/docs/tree/main/examples/seat-picker).

## Try it

```embed-lvt path="/apps/seat-picker/" upstream="http://localhost:9091" height="560px"
```

This is the real app, served by the docs recipes binary. The seat hall is
**shared across everyone viewing this page** — open it in a second window
(or send the link to a friend), join under a different name, and watch your
selections and bookings appear in each other's halls in real time. Every
click above is a plain `<button name="...">` submit; there is no client-side
code driving it.

## The whole interaction vocabulary

```html
<!-- pick a seat -->
<button class="seat available" name="selectSeat" value="A5">A5</button>

<!-- a seat someone else is holding: disabled, greyed, no action -->
<button class="seat held" type="button" value="B3" disabled>B3</button>

<!-- confirm your held seats -->
<button name="confirm">Book 2 seats</button>
```

That is it. A seat is a `<button name="selectSeat">`; its id rides along as
the button's `value`, read on the server with `ctx.GetString("value")`. No
`hx-*`, no `x-*`, no `phx-*`, no client code.

## Why it's more than the chat recipe

The [chat](/recipes/apps/chat) and [todos](/recipes/todos) recipes are
real-time too, but they sync **one user's own tabs** via `ctx.SelfTopic()`.
Seat-picker broadcasts **across different users** on a developer-defined
topic, `event/main`, admitted past the deny-all default:

```go
opts = append(opts, livetemplate.WithTopicACL(
    func(topic, _ string, _ *http.Request) (bool, error) {
        return topic == "event/main", nil
    },
))
```

Every connection subscribes to that topic in `Mount`; every mutation
publishes a `Refresh` to it:

```go
func (c *Controller) Mount(state State, ctx *livetemplate.Context) (State, error) {
    if err := ctx.Subscribe("event/main"); err != nil { // shared, cross-session
        return state, err
    }
    c.mu.Lock(); c.project(&state); c.mu.Unlock()
    return state, nil
}

func (c *Controller) SelectSeat(state State, ctx *livetemplate.Context) (State, error) {
    id := ctx.GetString("value")
    owner := ctx.GroupID()   // server-assigned session id — the ownership key
    c.mu.Lock()
    c.expire()
    _, state.Message = c.tryHold(owner, id) // conflict rule lives here
    c.project(&state, owner)  // re-project for the *clicking* session…
    c.mu.Unlock()
    ctx.Publish("event/main", "Refresh", nil) // …and fan out to everyone else
    return state, nil
}
```

Both the clicking user and every peer end up running the same
`project`-and-diff path — the [one model, every surface](/getting-started/mental-model#one-model-every-surface)
pipeline. The publishing connection is excluded from its own fan-out, which
is why `SelectSeat` re-projects its own state *and* publishes.

## Ownership is your session, not your name

A seat belongs to the visitor's **server-assigned session id**
(`ctx.GroupID()`, from the anonymous-session cookie) — never to the name
they type. That distinction is a security one, not a stylistic one: a typed
name is forgeable, so if ownership keyed on it, anyone could enter "Alice"
and release Alice's seats. The name is only a display label; the unguessable
session id, read fresh from the request on every action, is the authority —
which is why it is never stored in the client-round-tripped state where it
could be tampered. In a real app you would key on your authenticated user id
exactly the same way. Two different people (two browsers, two sessions) are
two owners; the shared topic carries the broadcast between them.

## Conflicts and holds

`tryHold` is the single rule that makes double-booking impossible: a seat
held or booked by anyone other than you cannot be re-held. A race between
two users resolves server-side — one wins, the other is told the seat was
just taken — with no client-side locking or merge logic. Holds expire
lazily: an abandoned seat is reclaimed the next time anyone touches the
event.

## Run it

```bash
cd examples/seat-picker
GOWORK=off go run ./cmd
```

Open it in **two different browsers**, join under two names, and watch each
other's seats fill in live. The two-browser
[end-to-end test](https://github.com/livetemplate/docs/blob/main/examples/seat-picker/seat_picker_test.go)
drives exactly that — two separate Chrome sessions — and asserts the
standard "no inline handlers, no framework attributes" UI bar.
