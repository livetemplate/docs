package patterns

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/livetemplate/livetemplate"
)

// ctx.ValidateForm doesn't surface a schema for forms inside <dialog>;
// use BindAndValidate (the dialog-patterns shape) instead.
var validateNav = validator.New()

// --- Pattern #17: Modal Dialog ---

// >>> region:modal-dialog
// On invalid submit, field errors must render inside the still-open dialog.
type ModalDialogController struct{}

type modalDialogInput struct {
	Name  string `json:"name"  validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
}

func (c *ModalDialogController) Save(state ModalDialogState, ctx *livetemplate.Context) (ModalDialogState, error) {
	var in modalDialogInput
	if err := ctx.BindAndValidate(&in, validateNav); err != nil {
		return state, err
	}
	state.Name = in.Name
	state.Email = in.Email
	state.SavedAt = time.Now().Format("15:04:05")
	ctx.SetFlash("success", "Profile saved", livetemplate.FlashExpiry(5*time.Second))
	return state, nil
}

func modalDialogHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/navigation/modal-dialog.tmpl")
	return tmpl.Handle(&ModalDialogController{}, livetemplate.AsState(&ModalDialogState{
		Title:    "Modal Dialog",
		Category: "Dialogs, Tabs & Navigation",
		Name:     "Ada Lovelace",
		Email:    "ada@analytical.engine",
	}))
}

// <<< region:modal-dialog

// --- Pattern #18: Confirm Dialog ---

// >>> region:confirm-dialog
// The Delete action reads the item id from the submit button's value attribute
// (the canonical Tier-1 row-action shape), not a hidden input.
type ConfirmDialogController struct{}

const confirmDialogItemCount = 5

func (c *ConfirmDialogController) Mount(state ConfirmDialogState, ctx *livetemplate.Context) (ConfirmDialogState, error) {
	if len(state.Items) == 0 && ctx.Action() == "" {
		state.Items = getItemPage(1, confirmDialogItemCount)
	}
	return state, nil
}

func (c *ConfirmDialogController) Delete(state ConfirmDialogState, ctx *livetemplate.Context) (ConfirmDialogState, error) {
	// "value" is the framework key for the clicked submit button's value
	// attribute (independent of the button's name). Same idiom as dialog-patterns.
	// id is a server-rendered Item.ID echoed back via the form, NOT free-form
	// user input — no escaping/allowlist check is needed.
	id := ctx.GetString("value")
	// Unknown ids are tolerated as a no-op — the next render reconciles any
	// drift between client and server item lists without surfacing a flash.
	state.Items = slices.DeleteFunc(state.Items, func(it Item) bool { return it.ID == id })
	return state, nil
}

func confirmDialogHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/navigation/confirm-dialog.tmpl")
	return tmpl.Handle(&ConfirmDialogController{}, livetemplate.AsState(&ConfirmDialogState{
		Title:    "Confirm Dialog",
		Category: "Dialogs, Tabs & Navigation",
	}))
}

// <<< region:confirm-dialog

// --- Pattern #19: Tabs (HATEOAS) ---

// >>> region:tabs
// Mount-only: tab links use the in-band __navigate__ action, which re-runs
// Mount with ctx.Action()=="" so the same guard covers initial GET.
type TabsController struct{}

var validTabs = map[string]bool{"overview": true, "settings": true, "activity": true}

func (c *TabsController) Mount(state TabsState, ctx *livetemplate.Context) (TabsState, error) {
	if ctx.Action() == "" {
		t := ctx.GetString("tab")
		switch {
		case t != "" && validTabs[t]:
			state.ActiveTab = t
		case t != "":
			// Explicit unknown tab → reset to overview (matches template promise).
			state.ActiveTab = "overview"
		case !validTabs[state.ActiveTab]:
			// First load (no param, empty state) → default to overview.
			state.ActiveTab = "overview"
		}
	}
	// Invariant: by here ActiveTab is always in validTabs. Belt-and-suspenders
	// in case a malformed message routes through Mount with ctx.Action() != "".
	if !validTabs[state.ActiveTab] {
		state.ActiveTab = "overview"
	}
	return state, nil
}

func tabsHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/navigation/tabs.tmpl")
	return tmpl.Handle(&TabsController{}, livetemplate.AsState(&TabsState{
		Title:    "Tabs (HATEOAS)",
		Category: "Dialogs, Tabs & Navigation",
	}))
}

// <<< region:tabs

// --- Pattern #20: SPA Navigation ---

// >>> region:spa-navigation
type SPANavController struct{}

const spaNavMaxStep = 3

func (c *SPANavController) Mount(state SPANavState, ctx *livetemplate.Context) (SPANavState, error) {
	if ctx.Action() == "" {
		// Out-of-range or non-integer step → fall through to the default below.
		if s := ctx.GetString("step"); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= spaNavMaxStep {
				state.Step = n
			}
		}
		if state.Step == 0 {
			state.Step = 1
		}
	}
	return state, nil
}

func spaNavigationHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/navigation/spa-navigation.tmpl")
	return tmpl.Handle(&SPANavController{}, livetemplate.AsState(&SPANavState{
		Title:    "SPA Navigation",
		Category: "Dialogs, Tabs & Navigation",
	}))
}

// <<< region:spa-navigation

// --- Pattern #21: Keyboard Shortcuts ---

// >>> region:keyboard-shortcuts
// Tier-2: lvt-on:window:keydown drives the panel; "/" opens, "Escape" closes
// (bound only while the panel is rendered).
type ShortcutsController struct{}

const shortcutsLogMax = 5

func (c *ShortcutsController) Open(state ShortcutsState, ctx *livetemplate.Context) (ShortcutsState, error) {
	if state.PanelOpen {
		return state, nil
	}
	state.PanelOpen = true
	state.Log = appendLog(state.Log, fmt.Sprintf("[%s] Opened panel", time.Now().Format("15:04:05")))
	return state, nil
}

func (c *ShortcutsController) Close(state ShortcutsState, ctx *livetemplate.Context) (ShortcutsState, error) {
	if !state.PanelOpen {
		return state, nil
	}
	state.PanelOpen = false
	state.Log = appendLog(state.Log, fmt.Sprintf("[%s] Closed panel", time.Now().Format("15:04:05")))
	return state, nil
}

func appendLog(log []string, entry string) []string {
	log = append(log, entry)
	if len(log) > shortcutsLogMax {
		log = log[len(log)-shortcutsLogMax:]
	}
	return log
}

func keyboardShortcutsHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/navigation/keyboard-shortcuts.tmpl")
	return tmpl.Handle(&ShortcutsController{}, livetemplate.AsState(&ShortcutsState{
		Title:    "Keyboard Shortcuts",
		Category: "Dialogs, Tabs & Navigation",
	}))
}

// <<< region:keyboard-shortcuts
