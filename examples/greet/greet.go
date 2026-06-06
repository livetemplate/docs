package greet

import "github.com/livetemplate/livetemplate"

// State is per-session state — pure data, cloned per session by livetemplate.
// This is the tiny "hello" app shown live in the docs homepage hero, matching
// the app.tmpl / app.go snippets right beside it.
type State struct {
	Name string
}

// Controller holds shared dependencies (none here) and exposes action methods
// invoked by name from the template's button.
type Controller struct{}

// Greet is invoked when the user clicks the "Say hi" button (button name="greet").
// It reads the submitted "name" field and stores it; the new render is diffed
// and only the changed text node ("Hello, …") is patched into the browser.
func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	if name := ctx.GetString("name"); name != "" {
		s.Name = name
	}
	return s, nil
}
