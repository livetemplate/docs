package greetloadingserver

import (
	"strings"
	"time"

	"github.com/livetemplate/livetemplate"
)

// State models loading in ordinary server state rather than client-only button
// chrome. The second action clears Loading after a server push.
type State struct {
	Name    string
	Loading bool
}

type Controller struct{}

func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	if s.Loading {
		return s, nil
	}
	session := ctx.Session()
	if session == nil {
		return s, nil
	}
	name := strings.TrimSpace(ctx.GetString("name"))
	s.Loading = true
	go func() {
		time.Sleep(700 * time.Millisecond)
		_ = session.TriggerAction("finishGreet", map[string]any{"name": name})
	}()
	return s, nil
}

func (c *Controller) FinishGreet(s State, ctx *livetemplate.Context) (State, error) {
	if name := strings.TrimSpace(ctx.GetString("name")); name != "" {
		s.Name = name
	}
	s.Loading = false
	return s, nil
}
