// Package greetnojs is Step 4 of the homepage spine — "works without
// JavaScript". It is the same greeting app as the hero, but built to prove the
// no-JS transport LIVE: with JavaScript disabled the browser submits the form
// as a plain POST, the server replies 303 (POST-Redirect-GET), and the followed
// GET must still greet you by name.
//
// That last part is why this is its own package rather than a reuse of
// examples/greet. Transient per-session State is re-cloned to its initial value
// on every fresh HTTP request, so over the stateless no-JS round-trip the name
// would be lost. Here the Controller keeps a small per-session name store keyed
// by the session group (the livetemplate-id cookie) and re-hydrates it in Mount
// on every render — including the GET after the redirect — so the greeting
// survives with no WebSocket and no client JavaScript. (cmd/site mounts this
// WS-disabled, and rewrites the 303 Location back under the mount prefix.)
package greetnojs

import (
	"sync"

	"github.com/livetemplate/livetemplate"
)

// State is per-session state. Name is re-hydrated from the Controller's store in
// Mount on every render, so it persists across the no-JS POST-redirect-GET.
type State struct {
	Name string
}

// Controller keeps the latest greeting per session group so the name survives a
// stateless no-JS round-trip. (The hero's examples/greet stays storeless — its
// source is shown verbatim on the page and must stay tiny.)
type Controller struct {
	mu    sync.Mutex
	names map[string]string // session group (livetemplate-id) -> latest name
}

func newController() *Controller {
	return &Controller{names: map[string]string{}}
}

// Mount re-hydrates Name from the per-session store. livetemplate runs Mount on
// every render — including the plain GET after a no-JS POST-redirect-GET — so
// the greeting persists without a WebSocket.
func (c *Controller) Mount(s State, ctx *livetemplate.Context) (State, error) {
	c.mu.Lock()
	if n, ok := c.names[ctx.GroupID()]; ok {
		s.Name = n
	}
	c.mu.Unlock()
	return s, nil
}

// Greet records the submitted name against the session group and reflects it in
// state. Storing it (not just setting s.Name) is what lets the next GET — the
// redirect target of a no-JS submit — render the greeting.
func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	if name := ctx.GetString("name"); name != "" {
		c.mu.Lock()
		c.names[ctx.GroupID()] = name
		c.mu.Unlock()
		s.Name = name
	}
	return s, nil
}
