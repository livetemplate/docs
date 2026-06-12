---
title: "Avatar Upload"
description: "A profile form that uploads an avatar with live progress, validation, and an instant preview — using the default Volume upload mode."
source_repo: "https://github.com/livetemplate/docs"
source_path: "content/recipes/apps/avatar-upload.md"
---

# Avatar Upload — a profile form with a file field

A profile form with one extra field: an avatar. The file streams to the
server over the WebSocket with a live `<progress>` bar, is validated against
a type and size whitelist, and — once the form is saved — is moved to a
permanent location and shown back instantly. No page reload, no custom
JavaScript. The full source is
[`examples/avatar-upload/`](https://github.com/livetemplate/docs/tree/main/examples/avatar-upload).

## Which upload mode is this?

This recipe uses the **Volume** mode — LiveTemplate's default. The browser
sends the bytes to the server, which stages them on local disk; the app then
owns the file's lifecycle (here: move it into `uploads/`). It is the right
default when you want the server to see and keep the bytes.

Volume is one of [four upload modes](/reference/uploads#upload-modes); the
mode is chosen purely by server config on an otherwise identical
`<input lvt-upload>`. To stream bytes straight to remote storage with zero
local disk, or to let the browser upload directly to S3/GCS, see the
[Upload Modes recipe](/recipes/apps/upload-modes) and the
[Upload reference](/reference/uploads).

## Configure the upload field

`WithUpload` declares the named upload and its validation. With no `Mode`
set, the field is Volume (the default); with no `Dir`, bytes stage to a temp
directory and the app moves them where it wants on completion:

```go
lt := livetemplate.Must(livetemplate.New("avatar-upload",
    livetemplate.WithParseFiles("avatar-upload.tmpl"),
    livetemplate.WithUpload("avatar", livetemplate.UploadConfig{
        Accept:      []string{"image/jpeg", "image/png", "image/gif"},
        MaxFileSize: 5 * 1024 * 1024, // 5MB
        MaxEntries:  1,                // single file
    }),
))

handler := lt.Handle(&ProfileController{}, livetemplate.AsState(&ProfileState{
    Name:  "John Doe",
    Email: "john@example.com",
}))
```

> Set `Dir: "uploads"` (with `Mode: livetemplate.UploadModeVolume`) to have
> LiveTemplate **retain** the staged file in a directory you own, instead of
> the stage-then-move pattern below. See the
> [Volume mode reference](/reference/uploads#volume--staged-to-the-servers-disk).

## Handle the submission

The avatar rides along in the same `multipart/form-data` POST as the text
fields, so one action reads both. Text fields come from `ctx.GetString`;
completed files come from `ctx.GetCompletedUploads`. Each entry carries the
server-side staging path in `entry.TempPath` — move it to permanent storage
and record the URL in state:

```go
func (c *ProfileController) UpdateProfile(state ProfileState, ctx *livetemplate.Context) (ProfileState, error) {
    state.Name = ctx.GetString("name")
    state.Email = ctx.GetString("email")

    for _, entry := range ctx.GetCompletedUploads("avatar") {
        ext := filepath.Ext(entry.ClientName)
        dst := filepath.Join("uploads", fmt.Sprintf("avatar-%s%s", entry.ID, ext))
        if err := os.Rename(entry.TempPath, dst); err != nil {
            return state, fmt.Errorf("failed to save avatar: %w", err)
        }
        state.AvatarURL = "/" + dst
    }

    ctx.SetFlash("success", "Profile updated")
    return state, nil
}
```

## Template

The file input is a plain `<input type="file">` plus one attribute,
`lvt-upload="avatar"`. The `{{range .lvt.Uploads "avatar"}}` block renders
per-file progress as it streams; `.lvt.HasUploadError` / `.lvt.UploadError`
surface validation failures:

```html
<form method="POST" name="updateProfile" enctype="multipart/form-data" lvt-form:preserve>
    <input type="text" name="name" value="{{.Name}}" required>
    <input type="email" name="email" value="{{.Email}}" required>

    <input type="file" name="avatar" lvt-upload="avatar"
           accept="image/jpeg,image/png,image/gif">

    {{range .lvt.Uploads "avatar"}}
        <small><strong>{{.ClientName}}</strong> — {{.Progress}}%</small>
        <progress value="{{.Progress}}" max="100"></progress>
        {{if .Error}}<del>{{.Error}}</del>{{else if .Done}}<ins>Upload complete!</ins>{{end}}
    {{end}}

    {{if .lvt.HasUploadError "avatar"}}<del>{{.lvt.UploadError "avatar"}}</del>{{end}}

    <button type="submit">Save Profile</button>
</form>

{{if .AvatarURL}}<img src="{{.AvatarURL}}" alt="Avatar">{{end}}
```

`lvt-form:preserve` keeps the chosen file and typed text across the live
re-render so a validation error doesn't wipe the form.

## Validation

`UploadConfig` enforces the whitelist before your handler runs — a file
that fails is marked invalid, surfaced via `.lvt.UploadError`, and never
appears in `GetCompletedUploads`:

- **Wrong type** (e.g. a `.txt` or `.pdf`) — rejected by `Accept`.
- **Too large** (over 5MB) — rejected by `MaxFileSize`.
- **Too many files** — only the first is accepted (`MaxEntries: 1`).

MIME types can be spoofed, so for security-critical uploads also validate the
file's actual content in your handler — see
[Content validation](/reference/uploads#content-validation).

## Run it

```bash
cd examples/avatar-upload
GOWORK=off go run main.go
```

Open <http://localhost:8080>, choose an image, and click **Save Profile** to
watch the progress bar fill and the avatar appear. The
[end-to-end test](https://github.com/livetemplate/docs/blob/main/examples/avatar-upload/avatar-upload_test.go)
drives exactly that flow in a real browser.

## See also

- [Upload Modes recipe](/recipes/apps/upload-modes) — all four modes, one `lvt-upload`.
- [File Upload pattern](/recipes/ui-patterns/forms/file-upload) — the minimal two-tier form.
- [Upload reference](/reference/uploads) — full config, helpers, and the mode matrix.
