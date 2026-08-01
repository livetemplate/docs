# Avatar Upload Example

A simple example demonstrating LiveTemplate's file upload feature with avatar upload functionality.

## Features

- 📸 **Image Upload**: Upload JPEG, PNG, or GIF avatars
- 🧩 **Tier 1 markup**: a plain `<input type="file">` — no `lvt-*` upload attribute
- ✅ **Validation**: Automatic file type and size validation (5MB limit)
- 🔄 **Live Updates**: Profile updates instantly without page reload

## What This Example Demonstrates

### Upload Configuration

Uploads are registered on the template with `WithUpload`, keyed by the name the
file input carries:

```go
lt := livetemplate.Must(livetemplate.New("avatar-upload",
    livetemplate.WithParseFiles("avatar-upload.tmpl"),
    livetemplate.WithDevMode(true),
    livetemplate.WithUpload("avatar", livetemplate.UploadConfig{
        Accept:      []string{"image/jpeg", "image/png", "image/gif"},
        MaxFileSize: 5 * 1024 * 1024, // 5MB
        MaxEntries:  1,               // Single file
    }),
))
```

Mode is omitted, so this is **Volume** — the default. The bytes are staged on
the server and the app owns the file from there.

### Upload Processing

The file arrives with the ordinary form submit, so the same action that saves
the profile also consumes the upload:

```go
func (c *ProfileController) UpdateProfile(state ProfileState, ctx *livetemplate.Context) (ProfileState, error) {
    state.Name = ctx.GetString("name")
    state.Email = ctx.GetString("email")

    if ctx.HasUploads("avatar") {
        var err error
        state, err = c.processAvatarUpload(state, ctx) // moves entry.TempPath into uploads/
        if err != nil {
            return state, err
        }
    }

    ctx.SetFlash("success", "Profile updated")
    return state, nil
}
```

`processAvatarUpload` reads `ctx.GetCompletedUploads("avatar")` and renames each
`entry.TempPath` to a permanent path under `uploads/`.

### Template Helpers

The file input needs **no upload attribute** — `name="avatar"` is what pairs it
with the `WithUpload("avatar", …)` registration, and the file rides the form's
`multipart/form-data` submit:

```html
<input type="file" id="avatar" name="avatar" accept="image/jpeg,image/png,image/gif">

<!-- Render the resulting upload entries -->
{{range .lvt.Uploads "avatar"}}
    <div class="upload-entry">
        <span>{{.ClientName}} - {{.Progress}}%</span>
        <progress value="{{.Progress}}" max="100"></progress>
        {{if .Error}}<span class="error">{{.Error}}</span>{{end}}
    </div>
{{end}}
```

Because the whole file arrives in one multipart POST, the entry is already
complete when it first renders — the `<progress>` reports a finished upload
rather than animating toward one. Add `lvt-upload="avatar"` to switch to the
chunked WebSocket transport if you want byte-by-byte progress; see
`examples/upload-autoupload` and `examples/upload-modes`.

## Running the Example

### 1. Install Dependencies

```bash
cd examples/avatar-upload
go mod download
```

### 2. Run the Server

```bash
go run main.go
```

The server will start at http://localhost:8080

### 3. Try It Out

1. Open http://localhost:8080 in your browser
2. Click "Choose File" and select an image (JPEG, PNG, or GIF)
3. Click "Save Profile"
4. Watch the real-time progress bar as your file uploads
5. See your avatar appear instantly when upload completes!

## Upload Strategies

This example uses **WebSocket Chunked Upload**:
- ✅ Real-time progress tracking
- ✅ Handles large files efficiently (256KB chunks)
- ✅ Non-blocking uploads
- ✅ Works with LiveTemplate's reactive updates

## File Structure

```
avatar-upload/
├── main.go              # Server code with ProfileStore
├── avatar-upload.tmpl   # HTML template with upload UI
├── go.mod              # Dependencies (uses local livetemplate)
├── README.md           # This file
└── uploads/            # Created at runtime for uploaded avatars
```

## Testing Different Scenarios

### Valid Upload
- Upload a JPEG, PNG, or GIF under 5MB
- ✅ Should show progress and complete successfully

### File Too Large
- Upload an image over 5MB
- ❌ Should show validation error

### Invalid File Type
- Upload a non-image file (e.g., .txt, .pdf)
- ❌ Should show "file type not accepted" error

### Multiple Files
- Try selecting multiple images
- ℹ️ Only the first will be accepted (MaxEntries: 1)

## Code Quality

This example demonstrates:
- ✅ Clean separation of concerns (Store pattern)
- ✅ Proper error handling
- ✅ File validation and security
- ✅ Temp file cleanup
- ✅ LiveTemplate best practices

## Next Steps

Want to extend this example?

1. **Add S3 Upload**: Replace local storage with S3 presigner
2. **Multiple Avatars**: Change `MaxEntries` to allow multiple images
3. **Image Cropping**: Add client-side cropping before upload
4. **Drag & Drop**: Add drag-and-drop file selection
5. **Auto-Upload**: Set `AutoUpload: true` for instant uploads

## Learn More

- [Upload Reference](https://livetemplate.fly.dev/reference/uploads)
- [Avatar Upload recipe](https://livetemplate.fly.dev/recipes/apps/avatar-upload)
- [Upload Modes recipe](https://livetemplate.fly.dev/recipes/apps/upload-modes) — Volume, Direct, Proxied, Preview
- [LiveTemplate Documentation](https://github.com/livetemplate/livetemplate)
- [Other Examples](../)

## Troubleshooting

**Upload not working?**
- Check browser console for errors
- Ensure WebSocket connection is established (look for green indicator)
- Verify file meets validation criteria (type, size)

**Progress not updating?**
- Make sure you're using WebSocket (not HTTP fallback)
- Check that ChunkSize is set appropriately
- Verify client library is loaded

**Files not saving?**
- Check that `uploads/` directory exists (created automatically)
- Verify file permissions on the uploads directory
- Check server logs for errors

---

Built with ❤️ using [LiveTemplate](https://github.com/livetemplate/livetemplate)
