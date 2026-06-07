package patterns

// DeleteRowState holds the state for the Delete Row pattern (#8).
// >>> region:delete-row-state
type DeleteRowState struct {
	Title    string
	Category string
	Items    []Item
}

// <<< region:delete-row-state

// ClickToLoadState holds the state for the Click To Load pattern (#9).
// >>> region:click-to-load-state
type ClickToLoadState struct {
	Title       string
	Category    string
	Items       []Item
	CurrentPage int
	HasMore     bool
}

// <<< region:click-to-load-state

// InfiniteScrollState holds the state for the Infinite Scroll pattern (#10).
// >>> region:infinite-scroll-state
type InfiniteScrollState struct {
	Title       string
	Category    string
	Items       []Item
	CurrentPage int
	HasMore     bool
}

// <<< region:infinite-scroll-state

// ValueSelectState holds the state for the Value Select pattern (#11).
// >>> region:value-select-state
type ValueSelectState struct {
	Title    string
	Category string
	Makes    []string
	Models   []string
	Make     string
	Model    string
}

// <<< region:value-select-state

// >>> region:sortable-state
type SortableState struct {
	Title    string
	Category string
	Items    []SortableItem
}

type SortableItem struct {
	Key  string
	Name string
}

// <<< region:sortable-state

// LargeTableState holds the per-session view of the Large Table demo.
// Filter/SortKey/SortDir are session-local; the underlying row dataset
// lives process-wide on the controller.
// >>> region:large-table-state
type LargeTableState struct {
	Title    string
	Category string
	Items    []LargeRow
	Filter   string
	SortKey  string
	SortDir  string
	Total    int
}

// <<< region:large-table-state
