package greetvalidate

import (
	"errors"
	"strings"

	"github.com/livetemplate/livetemplate"
)

// State is per-session state — pure data, cloned per session by livetemplate.
// Step 2 of the landing's progressive-narrative spine: the same greet app, now
// with validation. The only structural change from examples/greet is the
// guard in Greet plus the error slot in the template.
type State struct {
	Name string
}

// Controller exposes the action methods invoked by name from the template.
type Controller struct{}

// Greet validates the submitted name and, when it's empty, returns a
// field-scoped error instead of mutating state. NewFieldError binds the error
// to the "name" field; the template's {{.lvt.ErrorTag "name"}} renders it and
// {{.lvt.AriaInvalid "name"}} flips aria-invalid — no client-side validation,
// no second model. The validation rule lives in Go.
func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	name := strings.TrimSpace(ctx.GetString("name"))
	if name == "" {
		return s, livetemplate.NewFieldError("name", errors.New("Please enter a name"))
	}
	s.Name = name
	return s, nil
}
