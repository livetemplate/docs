package greetwall

import (
	"errors"
	"maps"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/livetemplate/livetemplate"
)

// Greeting is one line on the shared wall. Pure data, rendered through
// html/template (which auto-escapes Name — the wall is public, so the
// submitted name is untrusted and must never be treated as trusted HTML).
type Greeting struct {
	Name string
	At   string
}

// State is per-session state. Name is THIS session's greeting headline (synced
// across the user's own tabs via SelfTopic); Wall is a snapshot of the shared,
// cross-user list (synced to everyone via the "wall" topic); ServerAt is the
// timestamp of the latest server-initiated heartbeat (Step 7) — a single slot
// the server REPLACES in place, never a row it appends to the wall.
type State struct {
	Name     string
	Wall     []Greeting
	ServerAt string
}

const (
	wallTopic = "wall" // shared cross-user topic (no ':' — developer-topic grammar)
	maxWall   = 20      // ring-buffer cap: the wall is ephemeral and bounded

	// throttle rate-limits greetings per session group so one visitor can't
	// flood the public wall by holding the button.
	throttle = 400 * time.Millisecond

	// defaultServerHeartbeat is how often the server posts its own heartbeat
	// (Step 7, "the server can speak first"). The heartbeat REPLACES a single
	// in-place slot rather than appending to the wall, so frequency no longer
	// risks crowding out human greetings. Override via GREET_WALL_SERVER_INTERVAL
	// (e.g. a short value so e2e can assert the push without a 30s wait).
	defaultServerHeartbeat = 30 * time.Second
)

// serverHeartbeat resolves the heartbeat interval, honoring the
// GREET_WALL_SERVER_INTERVAL override (any valid time.Duration > 0) so the
// e2e harness can drive a fast tick; it falls back to defaultServerHeartbeat.
func serverHeartbeat() time.Duration {
	if v := os.Getenv("GREET_WALL_SERVER_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return defaultServerHeartbeat
}

// Controller holds the process-wide shared wall. Unlike the earlier greet
// steps (whose state is per-session), the wall is one list shared by every
// visitor — that shared state is the cross-user demo.
type Controller struct {
	mu        sync.Mutex
	wall      []Greeting                      // shared, capped, ephemeral
	serverAt  string                          // latest server heartbeat time — replaced in place (Step 7)
	names     map[string]string               // groupID -> latest name (SelfTopic sync target)
	lastGreet map[string]time.Time            // groupID -> last greet time (throttle)
	sessions  map[string]livetemplate.Session // groupID -> push handle (server push, Step 7)
	tickOnce  sync.Once
}

func newController() *Controller {
	return &Controller{
		names:     map[string]string{},
		lastGreet: map[string]time.Time{},
		sessions:  map[string]livetemplate.Session{},
	}
}

// Mount subscribes each session to BOTH topics — its own (so the user's
// greeting headline syncs across their tabs, Step 5) and the shared "wall"
// (so other people's greetings appear, Step 6) — then projects current state.
// SelfTopic() is ACL-exempt; the "wall" topic is admitted by WithTopicACL.
func (c *Controller) Mount(s State, ctx *livetemplate.Context) (State, error) {
	if err := ctx.Subscribe(ctx.SelfTopic()); err != nil {
		return s, err
	}
	if err := ctx.Subscribe(wallTopic); err != nil {
		return s, err
	}
	c.mu.Lock()
	if n, ok := c.names[ctx.GroupID()]; ok {
		s.Name = n
	}
	s.Wall = c.snapshot()
	s.ServerAt = c.serverAt // project the latest heartbeat so a fresh visitor sees it immediately
	c.mu.Unlock()
	return s, nil
}

// OnConnect registers this session's push handle and starts the single
// server-heartbeat loop (Step 7). The handle is per session group; storing it
// lets a background goroutine push the heartbeat with no user action.
func (c *Controller) OnConnect(s State, ctx *livetemplate.Context) (State, error) {
	if sess := ctx.Session(); sess != nil {
		c.mu.Lock()
		c.sessions[ctx.GroupID()] = sess
		c.mu.Unlock()
		c.tickOnce.Do(func() { go c.serverHeartbeatLoop(serverHeartbeat()) })
	}
	return s, nil
}

// Greet validates the name, records it for the user's tabs, appends it to the
// shared wall, and fans both changes out: SelfTopic -> Refresh (the user's own
// headline, Step 5) and the wall topic -> WallRefresh (everyone's list, Step
// 6). The calling connection is excluded from both — it already has the result.
func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	name := sanitize(ctx.GetString("name"))
	if name == "" {
		return s, livetemplate.NewFieldError("name", errors.New("Please enter a name"))
	}

	group := ctx.GroupID()
	now := time.Now()

	c.mu.Lock()
	if last, ok := c.lastGreet[group]; ok && now.Sub(last) < throttle {
		c.mu.Unlock()
		return s, nil // drop rapid-fire greetings from the same session
	}
	c.lastGreet[group] = now
	c.names[group] = name
	c.appendWall(Greeting{Name: name, At: now.Format("15:04:05")})
	s.Name = name
	s.Wall = c.snapshot()
	c.mu.Unlock()

	_ = ctx.Publish(ctx.SelfTopic(), "Refresh", nil)  // Step 5: your other tabs
	_ = ctx.Publish(wallTopic, "WallRefresh", nil)    // Step 6: every other visitor
	return s, nil
}

// Refresh reloads only the user's own headline. It runs on the user's peer
// tabs when Greet publishes to SelfTopic — never touches the shared wall, so a
// SelfTopic broadcast can't leak one user's name into another's headline.
func (c *Controller) Refresh(s State, ctx *livetemplate.Context) (State, error) {
	c.mu.Lock()
	if n, ok := c.names[ctx.GroupID()]; ok {
		s.Name = n
	}
	c.mu.Unlock()
	return s, nil
}

// WallRefresh reloads only the shared wall. It runs on every subscriber when
// any user greets (Step 6) — the wall now holds only human greetings.
func (c *Controller) WallRefresh(s State, ctx *livetemplate.Context) (State, error) {
	c.mu.Lock()
	s.Wall = c.snapshot()
	c.mu.Unlock()
	return s, nil
}

// ServerRefresh reloads only the server-heartbeat slot. It runs on every
// connected session when the heartbeat loop pushes (Step 7) — replacing one
// in-place value, so the server can speak without ever growing the wall.
func (c *Controller) ServerRefresh(s State, ctx *livetemplate.Context) (State, error) {
	c.mu.Lock()
	s.ServerAt = c.serverAt
	c.mu.Unlock()
	return s, nil
}

// serverHeartbeatLoop is Step 7, "the server can speak first": on a fixed
// interval it stamps an in-place heartbeat slot (it does NOT append to the
// wall — that's the whole point of this redesign; a server line replaces one
// value instead of piling up rows) and pushes ServerRefresh to every connected
// session, server-initiated with no user action. Sessions reported disconnected
// (ErrSessionDisconnected) are pruned, which self-heals the session map without
// an OnDisconnect hook (OnDisconnect carries no group identity). Pruning is
// scoped to that sentinel so a transient error never drops a still-connected
// group from server push.
func (c *Controller) serverHeartbeatLoop(interval time.Duration) {
	for {
		time.Sleep(interval)
		for group, sess := range c.markServerHeartbeat(time.Now()) {
			if err := sess.TriggerAction("ServerRefresh", nil); errors.Is(err, livetemplate.ErrSessionDisconnected) {
				c.mu.Lock()
				delete(c.sessions, group)
				c.mu.Unlock()
			}
		}
	}
}

// markServerHeartbeat stamps the in-place heartbeat slot and returns a snapshot
// of the sessions to push ServerRefresh to (nil when nobody is connected). It
// NEVER appends to the wall — that invariant is the whole redesign: a server
// line replaces one value, so it can't crowd out human greetings no matter how
// often it ticks. Split out from the loop so a test can drive a single tick.
func (c *Controller) markServerHeartbeat(at time.Time) map[string]livetemplate.Session {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.serverAt = at.Format("15:04:05")
	if len(c.sessions) == 0 {
		return nil
	}
	pending := make(map[string]livetemplate.Session, len(c.sessions))
	maps.Copy(pending, c.sessions)
	return pending
}

// appendWall pushes one greeting onto the shared ring buffer, dropping the
// oldest beyond maxWall. Caller must hold c.mu. The re-slice copies into a
// fresh backing array so the trimmed prefix is released.
func (c *Controller) appendWall(g Greeting) {
	c.wall = append(c.wall, g)
	if len(c.wall) > maxWall {
		c.wall = append([]Greeting(nil), c.wall[len(c.wall)-maxWall:]...)
	}
}

// snapshot returns a defensive copy of the shared wall. Caller must hold c.mu.
func (c *Controller) snapshot() []Greeting {
	return append([]Greeting(nil), c.wall...)
}

// sanitize trims, strips control characters, and caps the name length. The
// wall is public and rendered to every visitor; html/template handles HTML
// escaping, but bounding length and dropping control runes keeps the list
// readable and resists layout-breaking input.
func sanitize(s string) string {
	s = strings.TrimSpace(s)
	out := make([]rune, 0, 24)
	for _, r := range s {
		if unicode.IsControl(r) {
			continue
		}
		out = append(out, r)
		if len(out) >= 24 {
			break
		}
	}
	return string(out)
}
