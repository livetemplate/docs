package greetloading

import (
	"strings"
	"time"

	"github.com/livetemplate/livetemplate"
)

// State is per-session state — pure data, cloned per session by livetemplate.
// Step 3 of the landing's progressive-narrative spine: the same greet app,
// now with a loading button. The only change from Step 2 is the
// lvt-form:disable-with attribute on the button in the template — the server
// code is unchanged except for an artificial delay so the pending state is
// visible in the demo.
type State struct {
	Name string
}

// Controller exposes the action methods invoked by name from the template.
type Controller struct{}

// Greet stores the submitted name. The short sleep simulates real work (a DB
// write, an API call) so the button's "Saying hi…" pending state — declared
// purely in HTML via lvt-form:disable-with — is perceptible. There is no
// client-side loading state machine to write.
func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	time.Sleep(700 * time.Millisecond)
	if name := strings.TrimSpace(ctx.GetString("name")); name != "" {
		s.Name = name
	}
	return s, nil
}
