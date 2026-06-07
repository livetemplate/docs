package greetwall

import (
	"testing"
	"time"
)

// TestServerHeartbeatNeverGrowsWall is the regression lock for the Step-7
// redesign: the server's heartbeat must update a single in-place slot and must
// NEVER append a row to the shared wall. The old design appended a "the server"
// greeting every tick, which over time filled the capped wall and crowded out
// real people — the bug this change fixes.
func TestServerHeartbeatNeverGrowsWall(t *testing.T) {
	c := newController()

	// Two genuine human greetings on the wall.
	c.mu.Lock()
	c.appendWall(Greeting{Name: "Ada", At: "10:00:00"})
	c.appendWall(Greeting{Name: "Bo", At: "10:00:01"})
	c.mu.Unlock()

	// Many server heartbeats — far more than the wall's cap, so an accidental
	// append would be obvious (and would have evicted the humans entirely).
	base := time.Date(2026, 6, 7, 10, 0, 0, 0, time.UTC)
	for i := range maxWall * 3 {
		c.markServerHeartbeat(base.Add(time.Duration(i) * time.Second))
	}

	wall, _ := c.WallRefresh(State{}, nil)
	if len(wall.Wall) != 2 {
		t.Fatalf("wall has %d rows, want 2 — the server heartbeat must not append to the wall", len(wall.Wall))
	}
	for _, g := range wall.Wall {
		if g.Name == "the server" {
			t.Errorf("a server line leaked into the wall (%+v); the wall is for people", g)
		}
	}

	srv, _ := c.ServerRefresh(State{}, nil)
	if srv.ServerAt == "" {
		t.Error("ServerAt is empty after heartbeats — the in-place slot was never stamped")
	}
}

// TestServerHeartbeatReplacesInPlace confirms the slot is REPLACED, not
// accumulated: the latest heartbeat time wins and earlier ones leave no trace.
func TestServerHeartbeatReplacesInPlace(t *testing.T) {
	c := newController()

	first := time.Date(2026, 6, 7, 10, 0, 0, 0, time.UTC)
	last := time.Date(2026, 6, 7, 10, 5, 30, 0, time.UTC)
	c.markServerHeartbeat(first)
	c.markServerHeartbeat(last)

	srv, _ := c.ServerRefresh(State{}, nil)
	if want := last.Format("15:04:05"); srv.ServerAt != want {
		t.Errorf("ServerAt = %q, want the most recent heartbeat %q (it must replace, not append)", srv.ServerAt, want)
	}
}

// TestServerHeartbeatFromEnv verifies the interval override the e2e harness
// relies on to drive a fast tick.
func TestServerHeartbeatFromEnv(t *testing.T) {
	t.Setenv("GREET_WALL_SERVER_INTERVAL", "1500ms")
	if got := serverHeartbeat(); got != 1500*time.Millisecond {
		t.Errorf("serverHeartbeat() = %v, want 1.5s from the env override", got)
	}

	t.Setenv("GREET_WALL_SERVER_INTERVAL", "garbage")
	if got := serverHeartbeat(); got != defaultServerHeartbeat {
		t.Errorf("serverHeartbeat() = %v, want the default %v when the override is invalid", got, defaultServerHeartbeat)
	}
}
