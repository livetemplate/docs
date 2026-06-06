package greetloading

import (
	"strings"
	"time"

	"github.com/livetemplate/livetemplate"
)

// State is per-session state — pure data, cloned per session by livetemplate.
// Step 3 of the landing's progressive-narrative spine: the same greet app with
// an HTML-declared loading button (the spinner is driven by a class the client
// toggles on pending/done — see greet.tmpl).
type State struct {
	Name string
}

// Controller exposes the action methods invoked by name from the template.
type Controller struct{}

// Greet stores the submitted name after a short artificial delay so the
// button's pending spinner is visible. The server code is otherwise unchanged
// from Step 1 — the loading state is declared entirely in the template.
func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	time.Sleep(700 * time.Millisecond)
	if name := strings.TrimSpace(ctx.GetString("name")); name != "" {
		s.Name = name
	}
	return s, nil
}
