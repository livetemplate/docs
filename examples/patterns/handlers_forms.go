package patterns

import (
	"fmt"
	"net/http"

	"github.com/livetemplate/livetemplate"
)

// --- Pattern #1: Click To Edit ---

// >>> region:click-to-edit
type ClickToEditController struct{}

func (c *ClickToEditController) Edit(state ClickToEditState, ctx *livetemplate.Context) (ClickToEditState, error) {
	state.Editing = true
	return state, nil
}

func (c *ClickToEditController) Save(state ClickToEditState, ctx *livetemplate.Context) (ClickToEditState, error) {
	state.FirstName = ctx.GetString("firstName")
	state.LastName = ctx.GetString("lastName")
	state.Email = ctx.GetString("email")
	state.Editing = false
	return state, nil
}

func (c *ClickToEditController) Cancel(state ClickToEditState, ctx *livetemplate.Context) (ClickToEditState, error) {
	state.Editing = false
	return state, nil
}

func clickToEditHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/forms/click-to-edit.tmpl")
	return tmpl.Handle(&ClickToEditController{}, livetemplate.AsState(&ClickToEditState{
		Title:     "Click To Edit",
		Category:  "Forms & Editing",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}))
}

// <<< region:click-to-edit

// --- Pattern #2: Edit Row ---

// >>> region:edit-row
type EditRowController struct{}

func (c *EditRowController) Edit(state EditRowState, ctx *livetemplate.Context) (EditRowState, error) {
	// Edit/Save buttons send their ID via `value` attribute — see
	// docs/references/progressive-complexity-reference.md.
	state.EditingID = ctx.GetString("value")
	return state, nil
}

func (c *EditRowController) Save(state EditRowState, ctx *livetemplate.Context) (EditRowState, error) {
	id := ctx.GetString("value")
	for i, contact := range state.Contacts {
		if contact.ID == id {
			state.Contacts[i].Name = ctx.GetString("name")
			state.Contacts[i].Email = ctx.GetString("email")
			break
		}
	}
	state.EditingID = ""
	return state, nil
}

func (c *EditRowController) Cancel(state EditRowState, ctx *livetemplate.Context) (EditRowState, error) {
	state.EditingID = ""
	return state, nil
}

func editRowHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/forms/edit-row.tmpl")
	return tmpl.Handle(&EditRowController{}, livetemplate.AsState(&EditRowState{
		Title:    "Edit Row",
		Category: "Forms & Editing",
		Contacts: sampleContacts(),
	}))
}

// <<< region:edit-row

// --- Pattern #3: Inline Validation ---

// >>> region:inline-validation
type InlineValidationController struct{}

func (c *InlineValidationController) Change(state InlineValidationState, ctx *livetemplate.Context) (InlineValidationState, error) {
	if ctx.Has("email") {
		state.Email = ctx.GetString("email")
	}
	if ctx.Has("username") {
		state.Username = ctx.GetString("username")
	}
	_ = ctx.ValidateForm()
	return state, nil
}

func (c *InlineValidationController) Submit(state InlineValidationState, ctx *livetemplate.Context) (InlineValidationState, error) {
	if err := ctx.ValidateForm(); err != nil {
		return state, err
	}
	state.Saved = true
	return state, nil
}

func inlineValidationHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/forms/inline-validation.tmpl")
	return tmpl.Handle(&InlineValidationController{}, livetemplate.AsState(&InlineValidationState{
		Title:    "Inline Validation",
		Category: "Forms & Editing",
	}))
}

// <<< region:inline-validation

// --- Pattern #4: Bulk Update ---

// >>> region:bulk-update
type BulkUpdateController struct{}

func (c *BulkUpdateController) BulkUpdate(state BulkUpdateState, ctx *livetemplate.Context) (BulkUpdateState, error) {
	changed := 0
	for i, user := range state.Users {
		newActive := ctx.GetBool("active-" + user.ID)
		if newActive != user.Active {
			changed++
		}
		state.Users[i].Active = newActive
	}
	if changed == 0 {
		ctx.ClearFlash("success")
		ctx.SetFlash("info", "No changes")
	} else {
		ctx.ClearFlash("info")
		ctx.SetFlash("success", fmt.Sprintf("Updated %d user(s)", changed))
	}
	return state, nil
}

func bulkUpdateHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/forms/bulk-update.tmpl")
	return tmpl.Handle(&BulkUpdateController{}, livetemplate.AsState(&BulkUpdateState{
		Title:    "Bulk Update",
		Category: "Forms & Editing",
		Users:    sampleUsers(),
	}))
}

// <<< region:bulk-update

// --- Pattern #5: Reset User Input ---

// >>> region:reset-input
type ResetInputController struct{}

func (c *ResetInputController) Submit(state ResetInputState, ctx *livetemplate.Context) (ResetInputState, error) {
	msg := ctx.GetString("message")
	if msg != "" {
		state.Messages = append(state.Messages, msg)
	}
	return state, nil
}

func resetInputHandler() http.Handler {
	tmpl := newLayoutTmpl("templates/layout.tmpl", "templates/forms/reset-input.tmpl")
	return tmpl.Handle(&ResetInputController{}, livetemplate.AsState(&ResetInputState{
		Title:    "Reset User Input",
		Category: "Forms & Editing",
	}))
}

// <<< region:reset-input

// --- Pattern #6: File Upload ---

// >>> region:file-upload
type FileUploadController struct{}

func (c *FileUploadController) Upload(state FileUploadState, ctx *livetemplate.Context) (FileUploadState, error) {
	for _, name := range []string{"document", "chunked-doc"} {
		if ctx.HasUploads(name) {
			entries := ctx.GetCompletedUploads(name)
			if len(entries) > 0 {
				ctx.SetFlash("success", "Uploaded: "+entries[0].ClientName, livetemplate.FlashExpiry(flashSuccessExpiry))
				nudgeFlashExpiry(ctx, flashSuccessExpiry)
				return state, nil
			}
		}
	}
	ctx.SetFlash("error", "No file selected")
	return state, nil
}

func (c *FileUploadController) Refresh(state FileUploadState, ctx *livetemplate.Context) (FileUploadState, error) {
	return state, nil
}

func fileUploadHandler() http.Handler {
	tmpl := newLayoutTmplWithOpts(
		[]string{"templates/layout.tmpl", "templates/forms/file-upload.tmpl"},
		livetemplate.WithUpload("document", livetemplate.UploadConfig{
			MaxFileSize: 10 << 20, // 10 MB
			MaxEntries:  1,
		}),
		livetemplate.WithUpload("chunked-doc", livetemplate.UploadConfig{
			MaxFileSize: 10 << 20, // 10 MB
			MaxEntries:  1,
			ChunkSize:   1024, // 1KB chunks — small so progress is visible for demo files
		}),
	)
	return tmpl.Handle(&FileUploadController{}, livetemplate.AsState(&FileUploadState{
		Title:    "File Upload",
		Category: "Forms & Editing",
	}))
}

// <<< region:file-upload

// --- Pattern #7: Preserving File Inputs ---

// >>> region:preserve-inputs
type PreserveInputsController struct{}

func (c *PreserveInputsController) Submit(state PreserveInputsState, ctx *livetemplate.Context) (PreserveInputsState, error) {
	state.Name = ctx.GetString("name")
	state.Description = ctx.GetString("description")
	if err := ctx.ValidateForm(); err != nil {
		return state, err
	}
	ctx.SetFlash("success", "Saved: "+state.Name, livetemplate.FlashExpiry(flashSuccessExpiry))
	nudgeFlashExpiry(ctx, flashSuccessExpiry)
	return state, nil
}

func (c *PreserveInputsController) Refresh(state PreserveInputsState, ctx *livetemplate.Context) (PreserveInputsState, error) {
	return state, nil
}

func preserveInputsHandler() http.Handler {
	tmpl := newLayoutTmplWithOpts(
		[]string{"templates/layout.tmpl", "templates/forms/preserve-inputs.tmpl"},
		livetemplate.WithUpload("attachment", livetemplate.UploadConfig{
			MaxFileSize: 10 << 20, // 10 MB
			MaxEntries:  1,
		}),
	)
	return tmpl.Handle(&PreserveInputsController{}, livetemplate.AsState(&PreserveInputsState{
		Title:    "Preserving File Inputs",
		Category: "Forms & Editing",
	}))
}

// <<< region:preserve-inputs
