package patterns

// >>> region:modal-dialog-state
type ModalDialogState struct {
	Title    string
	Category string
	Name     string
	Email    string
	SavedAt  string
}

// <<< region:modal-dialog-state

// >>> region:confirm-dialog-state
type ConfirmDialogState struct {
	Title    string
	Category string
	Items    []Item
}

// <<< region:confirm-dialog-state

// >>> region:tabs-state
type TabsState struct {
	Title     string
	Category  string
	ActiveTab string
}

// <<< region:tabs-state

// >>> region:spa-navigation-state
type SPANavState struct {
	Title    string
	Category string
	Step     int
}

// <<< region:spa-navigation-state

// >>> region:keyboard-shortcuts-state
type ShortcutsState struct {
	Title     string
	Category  string
	PanelOpen bool
	Log       []string
}

// <<< region:keyboard-shortcuts-state
