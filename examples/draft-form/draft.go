package draftform

import (
	"strings"

	"github.com/livetemplate/livetemplate"
)

// State is per-session state — pure data, cloned per session by livetemplate.
// It models a tiny post editor: a title plus the last save outcome. The whole
// point of the example is that the SAME ctx.ValidateForm() call behaves
// differently depending on which button submitted the form.
type State struct {
	Title  string
	Status string // "", "draft", or "published"
}

// Controller exposes the action methods invoked by name from the template.
type Controller struct{}

// Publish enforces validation: the title is required, so an empty submit
// returns a field error and nothing is stored. ctx.ValidateForm() infers the
// rule from the `required` attribute in draft.tmpl — the authoritative check a
// malicious or no-JS client can't skip.
func (c *Controller) Publish(s State, ctx *livetemplate.Context) (State, error) {
	if err := ctx.ValidateForm(); err != nil {
		return s, err
	}
	s.Title = strings.TrimSpace(ctx.GetString("title"))
	s.Status = "published"
	return s, nil
}

// SaveDraft saves whatever is typed, even an empty title. The "Save draft"
// button carries formnovalidate, so the IDENTICAL ctx.ValidateForm() call here
// is skipped server-side — on every tier (WebSocket, HTTP-fetch, and no-JS
// native POST), with no client code. The framework records the button's name
// from the template (FormSchema.NoValidateSubmitters) and matches it against
// the submission's submitter; the kebab-case name save-draft also routes to
// this SaveDraft method verbatim on the no-JS tier.
func (c *Controller) SaveDraft(s State, ctx *livetemplate.Context) (State, error) {
	if err := ctx.ValidateForm(); err != nil {
		return s, err
	}
	s.Title = strings.TrimSpace(ctx.GetString("title"))
	s.Status = "draft"
	return s, nil
}
