// Package seatpicker is a cross-user, real-time seat-booking recipe.
// There is no main() here: production runs via the docs single-binary
// container, mounted by cmd/site at /apps/seat-picker/; cmd/main.go wraps
// the same Handler in a standalone listener for local dev and e2e tests.
//
// It is the example that proves LiveTemplate's central claim end to end:
// you build a genuinely multi-user, real-time UI out of standard HTML —
// every interaction is a plain <button name="..."> inside a <form> — and
// the same render-and-diff pipeline that updates one tab fans a seat
// selection out to *every other user* watching the same event.
//
// What makes it different from the chat / todos / counter recipes: those
// broadcast to a single user's own tabs via ctx.SelfTopic(). This one
// broadcasts across session boundaries on a developer-defined topic
// ("event/main"), admitted past the deny-all default with WithTopicACL.
// Two different people — not two tabs of one person — see each other's
// actions live.
//
// Seat ownership is keyed on the visitor's server-assigned session id
// (ctx.GroupID(), from the anonymous-session cookie) — never on the name
// they type. A typed name is forgeable: if ownership keyed on it, anyone
// could enter "Alice" and release Alice's seats. So the name is only a
// display label; the unguessable session id, derived fresh from the
// request on every action, is the authority. Two different people (two
// browsers, two sessions) are two owners; the cross-user broadcast rides
// the shared topic between them. In a real app you would key on your
// authenticated user id the same way — the rule is "owner comes from the
// server's notion of who you are, not from client-supplied data."
package seatpicker

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/livetemplate/livetemplate"
	e2etest "github.com/livetemplate/lvt/testing"
)

//go:embed seat-picker.tmpl
var templateFS embed.FS

var (
	tmplPath string
	tmplOnce sync.Once
)

// extractTemplate writes the embedded template to a temp file so
// livetemplate's file-based loader can parse it at runtime (the same
// approach the counter recipe uses). Done once per process.
func extractTemplate() string {
	tmplOnce.Do(func() {
		dir, err := os.MkdirTemp("", "seat-picker-tmpl-*")
		if err != nil {
			log.Fatalf("seat-picker: mkdtemp: %v", err)
		}
		data, err := templateFS.ReadFile("seat-picker.tmpl")
		if err != nil {
			log.Fatalf("seat-picker: read embedded tmpl: %v", err)
		}
		tmplPath = filepath.Join(dir, "seat-picker.tmpl")
		if err := os.WriteFile(tmplPath, data, 0o644); err != nil {
			log.Fatalf("seat-picker: write tmpl: %v", err)
		}
	})
	return tmplPath
}

// eventTopic is the shared, cross-session fan-out channel. Every
// connection subscribes to it in Mount; every seat mutation publishes a
// "Refresh" to it. Because the connection registry spans all session
// groups in a process, a publish here reaches every viewer — that is the
// cross-user behaviour. (A multi-instance deployment would add
// WithPubSubBroadcaster(redis) to relay the same publish between
// processes; the recipe code would not change.)
const eventTopic = "event/main"

// holdDuration is how long a selected-but-unconfirmed seat stays yours.
// Kept short so the demo's expiry behaviour is observable in a sitting.
// Holds expire lazily — they are reclaimed the next time anyone touches
// the event (see (*Controller).expire). A background server-push sweep
// could free them on a timer too, but that is a separate concern from the
// standard-HTML thesis, so the recipe stays lazy.
const holdDuration = 45 * time.Second

// seat is the shared, authoritative state of one seat. It lives only on
// the Controller, guarded by the mutex — never in per-connection state,
// because every viewer must agree on it.
type seat struct {
	holder  string    // session id of the current holder/booker ("" = free)
	booked  bool      // true once confirmed; booked seats never expire
	expires time.Time // hold deadline (ignored when booked)
}

// Controller is the singleton. It owns the seat map (the single source of
// truth shared by all viewers) and the fixed seat layout. Per-connection
// state holds only this viewer's name and the projection they should see.
type Controller struct {
	mu    sync.Mutex
	seats map[string]*seat
	rows  []string // row labels, ordered
	cols  int      // seats per row
}

// newController builds a rows×cols hall of empty seats.
func newController(rows []string, cols int) *Controller {
	c := &Controller{seats: make(map[string]*seat), rows: rows, cols: cols}
	for _, r := range rows {
		for n := 1; n <= cols; n++ {
			c.seats[seatID(r, n)] = &seat{}
		}
	}
	return c
}

func seatID(row string, col int) string { return fmt.Sprintf("%s%d", row, col) }

// State is per-connection — each tab gets its own clone. Viewer is the
// display name (persisted across reconnects); Rows and the summary fields
// are a projection recomputed from the shared seat map on every action,
// from this session's perspective. Note there is deliberately no owner-id
// field here: the ownership key is never stored in client-round-tripped
// state (which could be tampered) — it is read fresh from ctx.GroupID() on
// the server for every action.
type State struct {
	Viewer   string    `json:"viewer" lvt:"persist"`
	Rows     []RowView `json:"rows"`
	HeldByMe int       `json:"held_by_me"`
	Message  string    `json:"message"`
}

// RowView / SeatView are the rendered projection. Everything the template
// needs is precomputed here so the markup stays plain HTML — the template
// never branches on holder identity, it just reads Status/Action/Disabled.
type RowView struct {
	Label string     `json:"label"`
	Seats []SeatView `json:"seats"`
}

// SeatView is one seat as a specific viewer sees it.
//   - Status drives the CSS class: available | mine | held | booked | booked-mine
//   - Action is the button name to route to ("selectSeat", "releaseSeat", or "" when disabled)
//   - Disabled is true for seats this viewer cannot act on (someone else's hold/booking)
type SeatView struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Action   string `json:"action"`
	Disabled bool   `json:"disabled"`
}

// Mount runs once per session group. Subscribing to the shared event
// topic is what opts this connection into cross-user fan-out; without it,
// a viewer would see only their own actions. The projection is computed
// from this session's perspective (ctx.GroupID()).
func (c *Controller) Mount(state State, ctx *livetemplate.Context) (State, error) {
	if err := ctx.Subscribe(eventTopic); err != nil {
		return state, err
	}
	c.mu.Lock()
	c.project(&state, ctx.GroupID())
	c.mu.Unlock()
	return state, nil
}

// Join records the visitor's display name on this connection only. It does
// not mutate the shared seat map, so it does not publish — there is
// nothing for peers to react to yet. The name is a label; the session id
// (ctx.GroupID()) is what actually owns seats.
func (c *Controller) Join(state State, ctx *livetemplate.Context) (State, error) {
	name := ctx.GetString("name")
	if name == "" {
		return state, nil
	}
	state.Viewer = name
	state.Message = ""
	c.mu.Lock()
	c.project(&state, ctx.GroupID())
	c.mu.Unlock()
	return state, nil
}

// SelectSeat holds a seat for this viewer. The clicked
// <button name="selectSeat" value="A5"> routes here by its name; the seat
// id arrives as the submitted button's value, which LiveTemplate exposes
// under the "value" key (the same convention the dialog-patterns recipe
// uses for delete-by-id).
//
// Note the two-audience update: we re-project into the caller's own state
// (so the clicking user sees their selection immediately) AND publish a
// Refresh to peers. The publishing connection is excluded from its own
// fan-out, so without the self-projection the clicker's view would not
// update over WebSocket.
func (c *Controller) SelectSeat(state State, ctx *livetemplate.Context) (State, error) {
	if state.Viewer == "" {
		return state, nil
	}
	id := ctx.GetString("value")
	owner := ctx.GroupID()

	c.mu.Lock()
	c.expire()
	_, state.Message = c.tryHold(owner, id)
	c.project(&state, owner)
	c.mu.Unlock()

	// Publish AFTER all mutations (and after any ctx.With*() calls, of
	// which there are none here) so the queued fan-out is not stranded on
	// a pre-copy Context.
	if err := ctx.Publish(eventTopic, "Refresh", nil); err != nil {
		return state, err
	}
	return state, nil
}

// ReleaseSeat gives up a seat this viewer is holding (booked seats cannot
// be released). Same two-audience pattern as SelectSeat.
func (c *Controller) ReleaseSeat(state State, ctx *livetemplate.Context) (State, error) {
	if state.Viewer == "" {
		return state, nil
	}
	id := ctx.GetString("value")
	owner := ctx.GroupID()

	c.mu.Lock()
	c.expire()
	if s := c.seats[id]; s != nil && !s.booked && s.holder == owner {
		s.holder = ""
	}
	state.Message = ""
	c.project(&state, owner)
	c.mu.Unlock()

	if err := ctx.Publish(eventTopic, "Refresh", nil); err != nil {
		return state, err
	}
	return state, nil
}

// Confirm books every seat this viewer is currently holding.
func (c *Controller) Confirm(state State, ctx *livetemplate.Context) (State, error) {
	if state.Viewer == "" {
		return state, nil
	}

	owner := ctx.GroupID()
	c.mu.Lock()
	c.expire()
	booked := c.bookHeld(owner)
	if booked > 0 {
		state.Message = fmt.Sprintf("Booked %d seat%s. Enjoy the show!", booked, plural(booked))
	} else {
		state.Message = "Select a seat before booking."
	}
	c.project(&state, owner)
	c.mu.Unlock()

	if err := ctx.Publish(eventTopic, "Refresh", nil); err != nil {
		return state, err
	}
	return state, nil
}

// Refresh is the peer-side handler: it runs on every *other* subscribed
// connection when someone publishes to eventTopic. It mutates nothing
// shared — it just re-projects the latest seat map into that peer's own
// state, from that peer's perspective. This is the same method the
// runtime would call for a server-initiated TriggerAction too; one
// handler, every fan-out path.
func (c *Controller) Refresh(state State, ctx *livetemplate.Context) (State, error) {
	c.mu.Lock()
	c.expire()
	c.project(&state, ctx.GroupID())
	c.mu.Unlock()
	return state, nil
}

// tryHold attempts to put a hold on seat id for owner (a session id).
// Caller must hold mu and should call expire() first. Returns ok=true,
// msg="" on success; ok=false and a human-readable reason when the seat is
// already taken. This is the conflict rule that makes double-booking
// impossible: a seat held or booked by any session other than owner cannot
// be re-held.
func (c *Controller) tryHold(owner, id string) (ok bool, msg string) {
	s := c.seats[id]
	switch {
	case s == nil:
		return false, "That seat does not exist."
	case s.booked:
		return false, fmt.Sprintf("Seat %s is already booked.", id)
	case s.holder != "" && s.holder != owner:
		return false, fmt.Sprintf("Seat %s was just taken by someone else.", id)
	default:
		s.holder = owner
		s.expires = time.Now().Add(holdDuration)
		return true, ""
	}
}

// bookHeld confirms every active hold owned by owner (a session id) and
// returns how many seats were booked. Caller must hold mu and should call
// expire() first.
func (c *Controller) bookHeld(owner string) int {
	now := time.Now()
	booked := 0
	for _, s := range c.seats {
		if !s.booked && s.holder == owner && now.Before(s.expires) {
			s.booked = true
			booked++
		}
	}
	return booked
}

// expire reclaims holds whose deadline has passed. Caller must hold mu.
// Booked seats are never reclaimed.
func (c *Controller) expire() {
	now := time.Now()
	for _, s := range c.seats {
		if !s.booked && s.holder != "" && now.After(s.expires) {
			s.holder = ""
		}
	}
}

// project rebuilds state.Rows and the summary fields from the shared seat
// map, as seen by owner (a session id). Caller must hold mu.
func (c *Controller) project(state *State, owner string) {
	now := time.Now()
	heldByMe := 0
	rows := make([]RowView, 0, len(c.rows))

	for _, r := range c.rows {
		seats := make([]SeatView, 0, c.cols)
		for n := 1; n <= c.cols; n++ {
			id := seatID(r, n)
			s := c.seats[id]
			sv := SeatView{ID: id}

			active := s.holder != "" && now.Before(s.expires)
			switch {
			case s.booked && s.holder == owner:
				sv.Status, sv.Disabled = "booked-mine", true
			case s.booked:
				sv.Status, sv.Disabled = "booked", true
			case active && s.holder == owner:
				sv.Status, sv.Action = "mine", "releaseSeat" // click to release
				heldByMe++
			case active:
				sv.Status, sv.Disabled = "held", true
			default:
				sv.Status, sv.Action = "available", "selectSeat"
			}
			seats = append(seats, sv)
		}
		rows = append(rows, RowView{Label: r, Seats: seats})
	}

	state.Rows = rows
	state.HeldByMe = heldByMe
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// Handler returns the seat-picker app as an http.Handler ready to mount.
// The Controller is a process-wide singleton, so the seat map is shared
// across every visitor — that shared state is what makes the demo
// genuinely cross-user (the same shape the realtime UI patterns use).
// AnonymousAuthenticator gives each browser its own session group; the
// cross-user fan-out rides the shared "event/main" topic instead, admitted
// past the deny-all default with WithTopicACL. Callers (cmd/site for
// production, cmd/main.go for dev/e2e) supply origin/dev options via opts.
func Handler(opts ...livetemplate.Option) http.Handler {
	baseOpts := []livetemplate.Option{
		livetemplate.WithParseFiles(extractTemplate()),
		livetemplate.WithAuthenticator(&livetemplate.AnonymousAuthenticator{}),
		// Admit the shared cross-session topic; every other topic stays
		// denied, so viewers can subscribe to the hall and nothing else.
		livetemplate.WithTopicACL(func(topic, _ string, _ *http.Request) (bool, error) {
			return topic == eventTopic, nil
		}),
	}
	tmpl := livetemplate.Must(livetemplate.New("seat-picker", append(baseOpts, opts...)...))

	mux := http.NewServeMux()
	mux.Handle("/", tmpl.Handle(newController([]string{"A", "B", "C", "D", "E"}, 8), livetemplate.AsState(&State{})))
	// Dev-mode static assets — the template loads these when .lvt.DevMode
	// is set (local dev and e2e); production renders the CDN URLs instead.
	mux.HandleFunc("/livetemplate-client.js", e2etest.ServeClientLibrary)
	mux.HandleFunc("/livetemplate.css", e2etest.ServeCSS)
	return mux
}
