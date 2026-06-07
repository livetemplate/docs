package patterns

// ClickToEditState holds the state for the Click To Edit pattern (#1).
// >>> region:click-to-edit-state
type ClickToEditState struct {
	Title     string
	Category  string
	FirstName string
	LastName  string
	Email     string
	Editing   bool
}

// <<< region:click-to-edit-state

// EditRowState holds the state for the Edit Row pattern (#2).
// >>> region:edit-row-state
type EditRowState struct {
	Title     string
	Category  string
	Contacts  []Contact
	EditingID string
}

// <<< region:edit-row-state

// InlineValidationState holds the state for the Inline Validation pattern (#3).
// >>> region:inline-validation-state
type InlineValidationState struct {
	Title    string
	Category string
	Email    string
	Username string
	Saved    bool
}

// <<< region:inline-validation-state

// BulkUpdateState holds the state for the Bulk Update pattern (#4).
// >>> region:bulk-update-state
type BulkUpdateState struct {
	Title    string
	Category string
	Users    []UserRow
}

// <<< region:bulk-update-state

// ResetInputState holds the state for the Reset User Input pattern (#5).
// >>> region:reset-input-state
type ResetInputState struct {
	Title    string
	Category string
	Messages []string
}

// <<< region:reset-input-state

// FileUploadState holds the state for the File Upload pattern (#6).
// >>> region:file-upload-state
type FileUploadState struct {
	Title    string
	Category string
}

// <<< region:file-upload-state

// PreserveInputsState holds the state for the Preserving File Inputs pattern (#7).
// >>> region:preserve-inputs-state
type PreserveInputsState struct {
	Title       string
	Category    string
	Name        string
	Description string
}

// <<< region:preserve-inputs-state
