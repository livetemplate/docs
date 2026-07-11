package livedashboard

import (
	"sync"
	"time"

	"github.com/livetemplate/livetemplate"
)

// dashboardTopic is the shared developer topic every viewer joins. Because it
// is a developer topic (not ctx.SelfTopic()), subscribing to it is deny-all by
// default — Handler() authorizes it with WithTopicACL. That ACL is the only
// boundary on a cross-user topic, so making it explicit is the point of the
// recipe.
const dashboardTopic = "dashboard"

// >>> region:shared
// metrics is the process-wide shared source the background goroutine mutates
// and every session renders from. A real app would read these numbers from a
// database, a metrics endpoint, or a message queue; a mutex-guarded struct is
// enough to show the shape. It is owned by the controller (a singleton), not
// copied into per-session state.
type metrics struct {
	mu        sync.RWMutex
	ticks     int
	jobs      int
	updatedAt string
}

// advance mutates the shared metrics once per background tick.
func (m *metrics) advance() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ticks++
	// A cheap deterministic wobble so "active jobs" visibly changes without
	// needing a random source (Math.random()-free, like the rest of the demo).
	m.jobs = 3 + m.ticks%5
	m.updatedAt = time.Now().Format("15:04:05")
}

// snapshot copies the shared metrics into per-session state under a read lock.
func (m *metrics) snapshot(s DashboardState) DashboardState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s.Ticks = m.ticks
	s.Jobs = m.jobs
	s.UpdatedAt = m.updatedAt
	return s
}

// <<< region:shared

// DashboardController is a singleton holding the shared metrics. It is never
// cloned; only DashboardState is.
type DashboardController struct {
	metrics *metrics
}

// DashboardState is pure per-session data, cloned per connection and diffed on
// each render. It carries no dependencies — every field is a plain value read
// from the shared metrics.
type DashboardState struct {
	Ticks     int
	Jobs      int
	UpdatedAt string
}

// >>> region:mount
// Mount joins the shared "dashboard" topic and renders the current metrics.
// Subscribing in Mount (rather than an action) makes the join reconnect-durable:
// after a WebSocket blip the framework re-runs Mount and re-subscribes. Every
// connection — across every session group — that joins this topic receives the
// background goroutine's out-of-band Refresh.
func (c *DashboardController) Mount(state DashboardState, ctx *livetemplate.Context) (DashboardState, error) {
	if err := ctx.Subscribe(dashboardTopic); err != nil {
		return state, err
	}
	return c.metrics.snapshot(state), nil
}

// <<< region:mount

// >>> region:refresh
// Refresh is the action the background goroutine's handler.Publish dispatches.
// It re-reads the shared metrics into this session's state; the framework diffs
// and sends only what changed. It is a normal controller method — the fan-out
// picks it by name, exactly like a client-initiated action.
func (c *DashboardController) Refresh(state DashboardState, ctx *livetemplate.Context) (DashboardState, error) {
	return c.metrics.snapshot(state), nil
}

// <<< region:refresh
