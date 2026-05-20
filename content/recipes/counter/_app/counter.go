package counter

import "github.com/livetemplate/livetemplate"

// CounterState is per-session state — pure data, cloned per session by
// livetemplate. AnonymousAuthenticator (handler.go) keeps state private per
// browser; Publish to SelfTopic() (below) keeps a single user's tabs in
// sync without leaking state to other visitors.
type CounterState struct {
	Counter int
}

// CounterController holds shared dependencies (none in this demo) and
// exposes action methods invoked by name from the template.
type CounterController struct{}

// Mount subscribes the self-topic so peer tabs of the same session receive
// the Increment / Decrement dispatches Publish'd from the actions below.
func (c *CounterController) Mount(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
	if err := ctx.Subscribe(ctx.SelfTopic()); err != nil {
		return s, err
	}
	return s, nil
}

// Increment is invoked when the user clicks the "+1" button. The runtime
// calls it with a clone of the current state and stores whatever you return.
// The Publish call tells peer tabs subscribed to the same SelfTopic() to run
// Increment too, keeping multiple embeds and tabs in lockstep.
func (c *CounterController) Increment(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
	s.Counter++
	if err := ctx.Publish(ctx.SelfTopic(), "Increment", nil); err != nil {
		return s, err
	}
	return s, nil
}

// Decrement follows the same pattern.
func (c *CounterController) Decrement(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
	s.Counter--
	if err := ctx.Publish(ctx.SelfTopic(), "Decrement", nil); err != nil {
		return s, err
	}
	return s, nil
}
