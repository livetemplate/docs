package notepad

import (
	"sync"
	"time"
	"unicode/utf8"

	"github.com/livetemplate/livetemplate"
)

// >>> region:state
// NotepadController holds per-user notepad state. The map is keyed by
// ctx.UserID() (the username from BasicAuth). A real app would back
// this with a database; the map is fine for a recipe.
type NotepadController struct {
	mu    sync.RWMutex
	notes map[string]NotepadState // userID -> latest state
}

// NotepadState is pure data, cloned per session. lvt:"persist" tags
// keep the textarea content and metadata alive across page refreshes
// (the framework round-trips them through a client-side state
// checksum). Username is derived from ctx.UserID() on Mount and isn't
// persisted — it would be wrong to trust a client-supplied identity.
type NotepadState struct {
	Username  string `json:"username"`
	Content   string `json:"content" lvt:"persist"`
	SavedAt   string `json:"saved_at" lvt:"persist"`
	CharCount int    `json:"char_count" lvt:"persist"`
}

// <<< region:state

// >>> region:mount
// Mount runs on every fresh state (page load, reconnect with stale
// state). It binds Username to the authenticated user and rehydrates
// the textarea from the controller's per-user map.
func (c *NotepadController) Mount(state NotepadState, ctx *livetemplate.Context) (NotepadState, error) {
	state.Username = ctx.UserID()
	c.mu.RLock()
	if saved, ok := c.notes[ctx.UserID()]; ok {
		state.Content = saved.Content
		state.CharCount = saved.CharCount
		state.SavedAt = saved.SavedAt
	}
	c.mu.RUnlock()
	return state, nil
}

// <<< region:mount

// >>> region:save
// Save writes the textarea content into the per-user map and broadcasts
// a "Refresh" action to peer connections in the same session group
// (other tabs of the same user). The framework drains the broadcast
// queue after this action's response is sent.
func (c *NotepadController) Save(state NotepadState, ctx *livetemplate.Context) (NotepadState, error) {
	state.Content = ctx.GetString("content")
	state.CharCount = utf8.RuneCountInString(state.Content)
	state.SavedAt = time.Now().Format("15:04:05")

	c.mu.Lock()
	c.notes[ctx.UserID()] = state
	c.mu.Unlock()

	ctx.BroadcastAction("Refresh", nil)
	return state, nil
}

// <<< region:save

// Change is auto-wired to the textarea's input event (300ms debounce).
// It updates the in-memory state without persisting — that's what Save
// is for. Keeps the character count live as the user types.
func (c *NotepadController) Change(state NotepadState, ctx *livetemplate.Context) (NotepadState, error) {
	if ctx.Has("content") {
		state.Content = ctx.GetString("content")
		state.CharCount = utf8.RuneCountInString(state.Content)
	}
	return state, nil
}

// >>> region:refresh
// Refresh is the action peer tabs run when Save broadcasts. It re-reads
// the latest state from the per-user map. Note this is a regular
// controller action, not a framework-reserved name — pre-v0.9.0 the
// framework auto-dispatched a Sync() method; that was removed in
// livetemplate#406 in favour of explicit BroadcastAction("Refresh", nil)
// for clearer control over when peers actually refresh.
func (c *NotepadController) Refresh(state NotepadState, ctx *livetemplate.Context) (NotepadState, error) {
	c.mu.RLock()
	if saved, ok := c.notes[ctx.UserID()]; ok {
		state.Content = saved.Content
		state.CharCount = saved.CharCount
		state.SavedAt = saved.SavedAt
	}
	c.mu.RUnlock()
	return state, nil
}

// <<< region:refresh
