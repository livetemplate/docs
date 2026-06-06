package greetvalidate

import (
	"errors"
	"strings"

	"github.com/livetemplate/livetemplate"
)

// State is per-session state — pure data, cloned per session by livetemplate.
// Step 2 of the landing's progressive-narrative spine: the same greet app, now
// validated on BOTH sides. The HTML attributes in greet.tmpl (required) drive
// the browser's native client-side check; the same rules are re-enforced on the
// server, because you never trust the client.
type State struct {
	Name string
}

// Controller exposes the action methods invoked by name from the template.
type Controller struct{}

// Greet validates in two layers, then stores the name:
//
//  1. ctx.ValidateForm() re-checks the HTML constraints (required, type, …)
//     server-side. The framework infers the schema from the template, so the
//     rule you wrote once as a standard HTML attribute holds on the server too
//     — the authoritative check a malicious or no-JS client can't skip.
//  2. NewFieldError handles a custom business rule HTML can't express.
//
// Both render inline via {{.lvt.ErrorTag}} / {{.lvt.AriaInvalid}}.
func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	if err := ctx.ValidateForm(); err != nil {
		return s, err
	}
	name := strings.TrimSpace(ctx.GetString("name"))
	if strings.EqualFold(name, "admin") {
		return s, livetemplate.NewFieldError("name", errors.New(`"admin" is reserved — pick another`))
	}
	s.Name = name
	return s, nil
}
