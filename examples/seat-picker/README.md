# Seat Picker — cross-user real-time with standard HTML

A live, multi-**user** seat-booking hall. Different people — in different
browsers — see each other's selections update in real time. Every
interaction is a plain `<button name="...">` inside a `<form>`: there is no
custom JavaScript and no framework attribute on the markup.

This is the recipe that proves LiveTemplate's whole pitch end to end:

1. **Standard HTML.** Selecting seat A5 is `<button name="selectSeat" value="A5">`. Booking is `<button name="confirm">`. That is the entire interaction vocabulary.
2. **One reactive model.** A click mutates server state, the template re-renders, and only the diff is patched into the page.
3. **Pub/sub is that same model, across users.** The mutation also publishes to a shared topic, so *every other viewer* re-runs the same render-and-diff against their own view. Single-user and multi-user are the same program.

## What makes it different from the other recipes

`counter`, `chat`, and `todos` broadcast to a single user's own tabs via
`ctx.SelfTopic()`. This recipe broadcasts **across session boundaries** on a
developer-defined topic, `event/main`, admitted past the deny-all default
with `WithTopicACL`. Two different people see each other live — not two tabs
of one person.

Seat ownership is keyed on the visitor's **server-assigned session id**
(`ctx.GroupID()`), never on the name they type. A typed name is forgeable —
if it owned seats, anyone could enter "Alice" and release Alice's holds — so
the name is only a display label and the unguessable session id is the
authority. Two different people (two browsers, two sessions) are two owners;
the shared topic carries the broadcast between them.

## Quick start

```bash
cd examples/seat-picker
GOWORK=off go run ./cmd
```

Open <http://localhost:8095> in **two different browsers** (or one normal +
one private window so the cookies differ), join under two names, and pick
seats. Each browser sees the other's holds and bookings instantly.

## How it works

| Piece | Role |
|---|---|
| `Controller.seats` (mutex-guarded) | The shared source of truth every viewer agrees on |
| `ctx.GroupID()` | The server-assigned session id that **owns** seats (the typed `State.Viewer` name is display-only) |
| `ctx.Subscribe("event/main")` in `Mount` | Opts each connection into cross-user fan-out |
| `ctx.Publish("event/main", "Refresh", nil)` | Fans every mutation out to all viewers |
| `Controller.tryHold` | The conflict rule that makes double-booking impossible |
| Lazy `expire()` | Reclaims abandoned holds on the next interaction |

The seat id travels as the clicked button's value, read with
`ctx.GetString("value")` — the same convention the `dialog-patterns` recipe
uses for delete-by-id.

For multiple server instances, add `WithPubSubBroadcaster(redis)`; the
recipe code does not change — the same publish relays between processes.

## Testing

```bash
GOWORK=off go test -short ./examples/seat-picker   # fast white-box logic tests
GOWORK=off go test ./examples/seat-picker          # + two-browser cross-user e2e (needs Docker)
```

The e2e drives **two separate Chrome containers** (Alice and Bob) so the
broadcast genuinely crosses session-group boundaries, and asserts the
standard "no inline event handlers / no framework attributes" UI bar.
