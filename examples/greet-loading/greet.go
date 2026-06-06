package greetloading

import (
	"strings"
	"time"

	"github.com/livetemplate/livetemplate"
)

// State is per-session state — pure data, cloned per session by livetemplate.
// Step 3 of the landing's progressive-narrative spine: the same greet app with
// an HTML-declared loading button. Count makes every submit advance the state
// (see Greet) so the pending→done lifecycle always resolves.
type State struct {
	Name  string
	Count int
}

// Controller exposes the action methods invoked by name from the template.
type Controller struct{}

// Greet stores the submitted name after a short artificial delay so the
// button's pending spinner (declared in HTML via aria-busy) is visible.
//
// Count is incremented on every submit. This is deliberate: the client clears
// the button's aria-busy on the action's "done" lifecycle event, which only
// fires when the response carries a render diff. A repeat submit of the same
// name would otherwise produce an identical render (empty diff) and leave the
// spinner stuck — bumping Count guarantees a diff every time. (The session is
// WebSocket-backed so Count persists and keeps advancing across clicks.)
func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	time.Sleep(700 * time.Millisecond)
	if name := strings.TrimSpace(ctx.GetString("name")); name != "" {
		s.Name = name
	}
	s.Count++
	return s, nil
}
