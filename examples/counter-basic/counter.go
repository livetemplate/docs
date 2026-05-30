package counterbasic

import "github.com/livetemplate/livetemplate"

// CounterState is per-session state — pure data, cloned per session by
// livetemplate. AnonymousAuthenticator (handler.go) keeps state private per
// browser, so each visitor gets their own count with nothing shared across
// users. This is the basic, single-session version of the counter; the
// pubsub variant in examples/counter adds cross-tab sync on top.
type CounterState struct {
	Counter int
}

// CounterController holds shared dependencies (none in this demo) and
// exposes action methods invoked by name from the template.
type CounterController struct{}

// Increment is invoked when the user clicks the "+1" button. The runtime
// calls it with a clone of the current state and stores whatever you return —
// the new render is diffed against the previous one and only the changed text
// node is sent to the browser.
func (c *CounterController) Increment(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
	s.Counter++
	return s, nil
}

// Decrement follows the same pattern.
func (c *CounterController) Decrement(s CounterState, ctx *livetemplate.Context) (CounterState, error) {
	s.Counter--
	return s, nil
}
