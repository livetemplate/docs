package greetwall

import (
	"errors"
	"maps"
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
// cross-user list (synced to everyone via the "wall" topic).
type State struct {
	Name string
	Wall []Greeting
}

const (
	wallTopic = "wall" // shared cross-user topic (no ':' — developer-topic grammar)
	maxWall   = 20      // ring-buffer cap: the wall is ephemeral and bounded

	// throttle rate-limits greetings per session group so one visitor can't
	// flood the public wall by holding the button.
	throttle = 400 * time.Millisecond

	// serverGreetInterval is how often the server posts its own greeting
	// (Step 7, "the server can speak first"). Bounded and gentle — slow enough
	// that human greetings aren't drowned out on a quiet wall.
	serverGreetInterval = 30 * time.Second
)

// Controller holds the process-wide shared wall. Unlike the earlier greet
// steps (whose state is per-session), the wall is one list shared by every
// visitor — that shared state is the cross-user demo.
type Controller struct {
	mu        sync.Mutex
	wall      []Greeting                      // shared, capped, ephemeral
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
	c.mu.Unlock()
	return s, nil
}

// OnConnect registers this session's push handle and starts the single
// server-greeting loop (Step 7). The handle is per session group; storing it
// lets a background goroutine push wall updates with no user action.
func (c *Controller) OnConnect(s State, ctx *livetemplate.Context) (State, error) {
	if sess := ctx.Session(); sess != nil {
		c.mu.Lock()
		c.sessions[ctx.GroupID()] = sess
		c.mu.Unlock()
		c.tickOnce.Do(func() { go c.serverGreetLoop() })
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
// any user greets (Step 6) and when the server posts its own greeting (Step 7).
func (c *Controller) WallRefresh(s State, ctx *livetemplate.Context) (State, error) {
	c.mu.Lock()
	s.Wall = c.snapshot()
	c.mu.Unlock()
	return s, nil
}

// serverGreetLoop posts a greeting from "the server" on a fixed interval and
// pushes the new wall to every connected session — server-initiated, with no
// user action (Step 7). Sessions reported disconnected (ErrSessionDisconnected)
// are pruned, which self-heals the session map without an OnDisconnect hook
// (OnDisconnect carries no group identity). Pruning is scoped to that sentinel
// so a transient error never drops a still-connected group from server push.
func (c *Controller) serverGreetLoop() {
	for {
		time.Sleep(serverGreetInterval)

		c.mu.Lock()
		if len(c.sessions) == 0 {
			c.mu.Unlock()
			continue // nobody watching — don't fill the wall with server lines
		}
		c.appendWall(Greeting{Name: "the server", At: time.Now().Format("15:04:05")})
		pending := make(map[string]livetemplate.Session, len(c.sessions))
		maps.Copy(pending, c.sessions)
		c.mu.Unlock()

		for group, sess := range pending {
			if err := sess.TriggerAction("WallRefresh", nil); errors.Is(err, livetemplate.ErrSessionDisconnected) {
				c.mu.Lock()
				delete(c.sessions, group)
				c.mu.Unlock()
			}
		}
	}
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
