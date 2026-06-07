package patterns

// >>> region:animations-state
type AnimationItem struct {
	ID   string
	Name string
	Time string
	Mode string
}

type AnimationsState struct {
	Title    string
	Category string
	Items    []AnimationItem
	Mode     string
}

// <<< region:animations-state

// >>> region:loading-states-state
type LoadingStatesState struct {
	Title    string
	Category string
	LastSave string
}

// <<< region:loading-states-state

// >>> region:highlight-state
type HighlightState struct {
	Title    string
	Category string
	Counter  int
}

// <<< region:highlight-state

// >>> region:flash-messages-state
type FlashMessagesState struct {
	Title    string
	Category string
}

// <<< region:flash-messages-state
