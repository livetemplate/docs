package main

import "github.com/livetemplate/livetemplate"

// CounterState is per-session state — pure data, cloned per session by
// livetemplate. The tutorial app deliberately broadcasts every action
// (see BroadcastAction below) and uses a shared-group authenticator
// (main.go) so visible counts across embeds stay in lockstep.
type CounterState struct {
	Counter int
}

// CounterController holds shared dependencies (none in this demo) and
// exposes action methods invoked by name from the template.
type CounterController struct{}

// Increment is invoked when the user clicks the "+1" button. The
// runtime calls it with a clone of the current state and stores
// whatever you return. The BroadcastAction call tells the runtime
// to apply this same action on every other connected client, so
// multiple embeds and tabs stay in lockstep.
func (c *CounterController) Increment(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
	s.Counter++
	ctx.BroadcastAction("Increment", nil)
	return s, nil
}

// Decrement follows the same pattern.
func (c *CounterController) Decrement(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
	s.Counter--
	ctx.BroadcastAction("Decrement", nil)
	return s, nil
}
