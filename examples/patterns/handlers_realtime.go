package patterns

import (
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/livetemplate/livetemplate"
)

// --- Pattern #26: Multi-User Refresh ---

// >>> region:multi-user-sync
type MultiUserSyncController struct {
	mu      sync.RWMutex
	counter int
}

// Mount runs on every initial render. Subscribing the self-topic wires this
// connection to receive RefreshCounter from peer Publishes. Without the
// initial-render counter read, a tab that opens AFTER other tabs have
// incremented would render Counter:0 and only converge on the next peer
// publish. Same fix as PresenceController.
func (c *MultiUserSyncController) Mount(state MultiUserSyncState, ctx *livetemplate.Context) (MultiUserSyncState, error) {
	if err := ctx.Subscribe(ctx.SelfTopic()); err != nil {
		return state, err
	}
	c.mu.RLock()
	state.Counter = c.counter
	c.mu.RUnlock()
	return state, nil
}

// RefreshCounter is the action peer connections run when Increment publishes
// to SelfTopic(). The state arg is the peer's local state; we replace its
// Counter from the shared controller value so all tabs converge.
func (c *MultiUserSyncController) RefreshCounter(state MultiUserSyncState, ctx *livetemplate.Context) (MultiUserSyncState, error) {
	c.mu.RLock()
	state.Counter = c.counter
	c.mu.RUnlock()
	return state, nil
}

func (c *MultiUserSyncController) Increment(state MultiUserSyncState, ctx *livetemplate.Context) (MultiUserSyncState, error) {
	c.mu.Lock()
	c.counter++
	state.Counter = c.counter
	c.mu.Unlock()
	if err := ctx.Publish(ctx.SelfTopic(), "RefreshCounter", nil); err != nil {
		return state, err
	}
	return state, nil
}

func multiUserSyncHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/realtime/multi-user-sync.tmpl")
	return tmpl.Handle(&MultiUserSyncController{}, livetemplate.AsState(&MultiUserSyncState{
		Title:    "Multi-User Sync",
		Category: "Real-Time & Multi-User",
	}))
}

// <<< region:multi-user-sync

// --- Pattern #27: Pubsub ---

// >>> region:pubsub-controller
type PubSubController struct {
	mu       sync.RWMutex
	nextID   int
	messages []PubSubMessage
}

// snapshotLocked returns a copy of c.messages. The Locked suffix signals
// that the caller MUST hold c.mu (read or write) — without that, slices.Clone
// reads c.messages concurrently with Send's append and races.
func (c *PubSubController) snapshotLocked() []PubSubMessage {
	return slices.Clone(c.messages)
}

func (c *PubSubController) Mount(state PubSubState, ctx *livetemplate.Context) (PubSubState, error) {
	if err := ctx.Subscribe(ctx.SelfTopic()); err != nil {
		return state, err
	}
	c.mu.RLock()
	state.Messages = c.snapshotLocked()
	c.mu.RUnlock()
	return state, nil
}

// <<< region:pubsub-controller

func (c *PubSubController) Join(state PubSubState, ctx *livetemplate.Context) (PubSubState, error) {
	name := strings.TrimSpace(ctx.GetString("username"))
	if name == "" {
		return state, nil
	}
	state.Username = name
	return state, nil
}

// >>> region:pubsub-send
func (c *PubSubController) Send(state PubSubState, ctx *livetemplate.Context) (PubSubState, error) {
	if state.Username == "" {
		return state, nil
	}
	text := strings.TrimSpace(ctx.GetString("text"))
	if text == "" {
		return state, nil
	}
	c.mu.Lock()
	c.nextID++
	// No cap on c.messages: deliberately omitted to keep the demo focused
	// on the Publish-to-SelfTopic mechanism. Production apps would
	// ring-buffer, paginate, or persist to a store with TTL.
	c.messages = append(c.messages, PubSubMessage{ID: c.nextID, User: state.Username, Text: text})
	state.Messages = c.snapshotLocked()
	c.mu.Unlock()
	// Publish must come after the lock release — holding the connection
	// registry mutex while queuing peer dispatches can deadlock with peer
	// dispatches that take the same mutex from the other side. Peers
	// receive "NewMessage" and refresh their local copy.
	if err := ctx.Publish(ctx.SelfTopic(), "NewMessage", nil); err != nil {
		return state, err
	}
	return state, nil
}

// <<< region:pubsub-send

// >>> region:pubsub-newmessage
func (c *PubSubController) NewMessage(state PubSubState, ctx *livetemplate.Context) (PubSubState, error) {
	c.mu.RLock()
	state.Messages = c.snapshotLocked()
	c.mu.RUnlock()
	return state, nil
}

// <<< region:pubsub-newmessage

func pubsubHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/realtime/pubsub.tmpl")
	return tmpl.Handle(&PubSubController{}, livetemplate.AsState(&PubSubState{
		Title:    "Pubsub",
		Category: "Real-Time & Multi-User",
	}))
}

// --- Pattern #28: Presence Tracking ---

// >>> region:presence
type PresenceController struct {
	mu          sync.RWMutex
	onlineUsers map[string]bool
}

func newPresenceController() *PresenceController {
	return &PresenceController{onlineUsers: make(map[string]bool)}
}

// Mount runs on every initial render. Subscribing the self-topic wires this
// connection to receive PresenceChanged from peer Publishes. Without the
// initial-render OnlineCount read, a new visitor's state.OnlineCount would
// default to 0 even when other users are already in the shared map — they'd
// see "0 user(s) online" until the next Join/Leave publish updates them.
func (c *PresenceController) Mount(state PresenceState, ctx *livetemplate.Context) (PresenceState, error) {
	if err := ctx.Subscribe(ctx.SelfTopic()); err != nil {
		return state, err
	}
	c.mu.RLock()
	state.OnlineCount = len(c.onlineUsers)
	c.mu.RUnlock()
	return state, nil
}

func (c *PresenceController) Join(state PresenceState, ctx *livetemplate.Context) (PresenceState, error) {
	name := strings.TrimSpace(ctx.GetString("username"))
	if name == "" {
		return state, nil
	}
	c.mu.Lock()
	c.onlineUsers[name] = true
	state.Username = name
	state.Joined = true
	state.OnlineCount = len(c.onlineUsers)
	c.mu.Unlock()
	if err := ctx.Publish(ctx.SelfTopic(), "PresenceChanged", nil); err != nil {
		return state, err
	}
	return state, nil
}

func (c *PresenceController) Leave(state PresenceState, ctx *livetemplate.Context) (PresenceState, error) {
	if state.Username == "" {
		return state, nil
	}
	c.mu.Lock()
	delete(c.onlineUsers, state.Username)
	state.Username = ""
	state.Joined = false
	state.OnlineCount = len(c.onlineUsers)
	c.mu.Unlock()
	if err := ctx.Publish(ctx.SelfTopic(), "PresenceChanged", nil); err != nil {
		return state, err
	}
	return state, nil
}

// PresenceChanged refreshes only the shared OnlineCount. Username and
// Joined are per-connection identity and must NOT be overwritten from a
// peer publish — every connection's own Join/Leave is the only thing
// that mutates those fields locally.
func (c *PresenceController) PresenceChanged(state PresenceState, ctx *livetemplate.Context) (PresenceState, error) {
	c.mu.RLock()
	state.OnlineCount = len(c.onlineUsers)
	c.mu.RUnlock()
	return state, nil
}

func presenceHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/realtime/presence.tmpl")
	return tmpl.Handle(newPresenceController(), livetemplate.AsState(&PresenceState{
		Title:    "Presence Tracking",
		Category: "Real-Time & Multi-User",
	}))
}

// <<< region:presence

// --- Pattern #29: Reconnection Recovery ---

// >>> region:reconnection
type ReconnectionController struct{}

func (c *ReconnectionController) Increment(state ReconnectionState, ctx *livetemplate.Context) (ReconnectionState, error) {
	state.Counter++
	return state, nil
}

func (c *ReconnectionController) SaveNotes(state ReconnectionState, ctx *livetemplate.Context) (ReconnectionState, error) {
	// Notes is a free-form textarea — leading/trailing whitespace AND
	// internal newlines are deliberate user content. Unlike Send/Join
	// inputs (which use TrimSpace to reject all-whitespace submissions),
	// SaveNotes preserves whatever the user typed verbatim.
	state.Notes = ctx.GetString("notes")
	return state, nil
}

func reconnectionHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/realtime/reconnection.tmpl")
	return tmpl.Handle(&ReconnectionController{}, livetemplate.AsState(&ReconnectionState{
		Title:    "Reconnection Recovery",
		Category: "Real-Time & Multi-User",
	}))
}

// <<< region:reconnection

// --- Pattern #30: Live Preview ---

// >>> region:live-preview
type LivePreviewController struct{}

// Change is auto-bound by the framework when the controller exposes it.
// Reads the input's current value via ctx.GetString and updates state.Preview.
// Does NOT write back to state.Input — patching the input element's value
// attribute mid-typing would reset the cursor position. (See
// examples/live-preview/main.go:26-29 for the same constraint.) An explicit
// Submit action commits state.Input on form submission.
func (c *LivePreviewController) Change(state LivePreviewState, ctx *livetemplate.Context) (LivePreviewState, error) {
	if ctx.Has("input") {
		state.Preview = "Hello, " + ctx.GetString("input") + "!"
	}
	return state, nil
}

func (c *LivePreviewController) Submit(state LivePreviewState, ctx *livetemplate.Context) (LivePreviewState, error) {
	state.Input = ctx.GetString("input")
	state.Preview = "Saved: " + state.Input
	return state, nil
}

func livePreviewHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/realtime/live-preview.tmpl")
	return tmpl.Handle(&LivePreviewController{}, livetemplate.AsState(&LivePreviewState{
		Title:    "Live Preview",
		Category: "Real-Time & Multi-User",
		// Preview is intentionally empty initially — Change builds the
		// "Hello, …!" value as the user types. Mirrors live-preview/main.go's
		// initial state (preview("")  → empty until the first Change fires).
	}))
}

// <<< region:live-preview

// --- Pattern #31: Server Push ---

// >>> region:server-push
type ServerPushController struct{}

const serverPushTickInterval = 1 * time.Second
const serverPushTickCount = 10

// StartTimer flips state.Running and spawns a 10×1s ticker goroutine.
//
// Running is intentionally NOT lvt:"persist". If the connection drops
// mid-timer (browser refresh, network blip), the reconnected client
// will see Running=false in its initial render — the goroutine on the
// server keeps ticking and eventually fires TimerDone, so the client
// will pop directly to the "Last completed: Xs" view rather than the
// running view. Trade-off: a persisted Running flag could survive the
// reconnect, but a stale "Running=true" with no in-flight goroutine is
// a worse failure mode (the UI would be stuck forever waiting for ticks
// that aren't coming).
func (c *ServerPushController) StartTimer(state ServerPushState, ctx *livetemplate.Context) (ServerPushState, error) {
	if state.Running {
		return state, nil
	}
	// Check session BEFORE flipping Running. Framework guarantees a
	// session for WebSocket connections, but if it ever is nil we'd
	// render "Timer running" with no goroutine to ever clear it.
	session := ctx.Session()
	if session == nil {
		return state, nil
	}
	state.Running = true
	state.Elapsed = 0
	state.Total = serverPushTickCount
	go func() {
		ticker := time.NewTicker(serverPushTickInterval)
		defer ticker.Stop()
		for i := 0; i < serverPushTickCount; i++ {
			<-ticker.C
			// session.TriggerAction returns an error when the session group has
			// no live connections (livetemplate/session_impl.go:91-159). Bail
			// out cleanly so the goroutine exits when the user closes the tab.
			if err := session.TriggerAction("tick", map[string]any{
				"elapsed": i + 1,
			}); err != nil {
				return
			}
		}
		// timerDone fires after the loop completes. We discard the error
		// here (unlike the per-tick error check above): if the connection
		// is gone by now, the goroutine exits anyway — no recovery action
		// is meaningful, and propagating would force the caller of the
		// goroutine to handle a context that's already finished.
		_ = session.TriggerAction("timerDone", nil)
	}()
	return state, nil
}

func (c *ServerPushController) Tick(state ServerPushState, ctx *livetemplate.Context) (ServerPushState, error) {
	state.Elapsed = ctx.GetInt("elapsed")
	return state, nil
}

func (c *ServerPushController) TimerDone(state ServerPushState, ctx *livetemplate.Context) (ServerPushState, error) {
	state.Running = false
	return state, nil
}

func serverPushHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/realtime/server-push.tmpl")
	return tmpl.Handle(&ServerPushController{}, livetemplate.AsState(&ServerPushState{
		Title:    "Server Push",
		Category: "Real-Time & Multi-User",
	}))
}

// <<< region:server-push
