package progressiveenhancement

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/livetemplate/livetemplate"
)

// TodoController is a singleton that holds dependencies (the validator).
type TodoController struct {
	validate *validator.Validate
}

// TodoState is pure data, cloned per session.
type TodoState struct {
	Title string `json:"title"`
	Items []Todo `json:"items" lvt:"persist"`
	// InputTitle preserves the form value when validation fails so the
	// user doesn't have to retype on the second attempt.
	InputTitle string `json:"input_title" lvt:"persist"`
}

// Todo represents a single todo item.
type Todo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	CreatedAt string `json:"created_at"`
}

// AddInput is the input struct for the Add action. The validator tags
// drive both server-side validation and the inline error display via
// .lvt.ErrorTag in the template.
type AddInput struct {
	Title string `json:"title" validate:"required,min=3,max=100"`
}

// Mount runs once per session and seeds the in-memory store with three
// sample todos. Flash messages are carried by the framework's lvt-flash
// cookie across PRG redirects — no URL-param bridging needed, which
// also avoids letting strangers spoof banners with ?success=...
func (c *TodoController) Mount(state TodoState, ctx *livetemplate.Context) (TodoState, error) {
	state.Title = "Progressive Enhancement Todo List"

	if len(state.Items) == 0 {
		state.Items = []Todo{
			{ID: "1", Title: "Learn about progressive enhancement", Completed: true, CreatedAt: formatTime()},
			{ID: "2", Title: "Try the app without JavaScript", Completed: false, CreatedAt: formatTime()},
			{ID: "3", Title: "Enable JavaScript and see the difference", Completed: false, CreatedAt: formatTime()},
		}
	}

	return state, nil
}

// Add handles adding a new todo item. On validation failure, InputTitle
// is preserved on state so the form re-renders with the user's typed
// value rather than blanking it.
func (c *TodoController) Add(state TodoState, ctx *livetemplate.Context) (TodoState, error) {
	var input AddInput
	if err := ctx.BindAndValidate(&input, c.validate); err != nil {
		state.InputTitle = ctx.GetString("title")
		return state, err
	}

	title := strings.TrimSpace(input.Title)
	newID := fmt.Sprintf("%d", time.Now().UnixNano())
	state.Items = append(state.Items, Todo{
		ID:        newID,
		Title:     title,
		Completed: false,
		CreatedAt: formatTime(),
	})

	state.InputTitle = ""
	ctx.SetFlash("success", fmt.Sprintf("Added: %s", title))

	return state, nil
}

// Toggle flips a todo's completed status by ID.
func (c *TodoController) Toggle(state TodoState, ctx *livetemplate.Context) (TodoState, error) {
	id := ctx.GetString("id")
	found := false
	for i := range state.Items {
		if state.Items[i].ID == id {
			state.Items[i].Completed = !state.Items[i].Completed
			found = true
			ctx.SetFlash("success", "Item updated")
			break
		}
	}
	if !found {
		ctx.SetFlash("error", "Item not found")
	}
	return state, nil
}

// Delete removes a todo by ID.
func (c *TodoController) Delete(state TodoState, ctx *livetemplate.Context) (TodoState, error) {
	id := ctx.GetString("id")

	deleteIndex := -1
	for i, item := range state.Items {
		if item.ID == id {
			deleteIndex = i
			break
		}
	}

	if deleteIndex >= 0 {
		state.Items = append(state.Items[:deleteIndex], state.Items[deleteIndex+1:]...)
		ctx.SetFlash("success", "Item deleted")
	} else {
		ctx.SetFlash("error", "Item not found")
	}
	return state, nil
}

func formatTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
