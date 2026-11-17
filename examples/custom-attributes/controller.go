package customattributes

import (
	"fmt"

	"github.com/livetemplate/livetemplate"
)

// ShareBase is the page a copied link points at. The rating rides along as a
// query parameter, so the copied text changes whenever the rating does.
const ShareBase = "https://livetemplate.fly.dev/recipes/custom-attributes"

// State is per-session data. Nothing in it knows the lvt-x:* attributes exist:
// the template renders them as plain HTML and the server never sees them.
type State struct {
	Rating    int
	Shortcuts bool
}

// Stars is the five positions the template renders, so it can loop over them.
func (s State) Stars() []int { return []int{1, 2, 3, 4, 5} }

// ShareURL is what lvt-x:copy copies.
func (s State) ShareURL() string {
	if s.Rating == 0 {
		return ShareBase
	}
	return fmt.Sprintf("%s?stars=%d", ShareBase, s.Rating)
}

// Controller has no dependencies; it exists to hold the action methods.
type Controller struct{}

// SetRating is the action lvt-x:rating dispatches. It is an ordinary action:
// nothing here can tell it came from a custom attribute rather than a button.
func (c *Controller) SetRating(s State, ctx *livetemplate.Context) (State, error) {
	stars := ctx.GetInt("stars")
	if stars < 1 || stars > 5 {
		return s, fmt.Errorf("stars must be 1 to 5, got %d", stars)
	}
	s.Rating = stars
	return s, nil
}

// ToggleShortcuts adds or removes lvt-x:hotkey from the stars, which is how
// the page shows a low-level handler noticing that the server took its
// attribute away.
func (c *Controller) ToggleShortcuts(s State, _ *livetemplate.Context) (State, error) {
	s.Shortcuts = !s.Shortcuts
	return s, nil
}
