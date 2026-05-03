---
title: "LiveTemplate Components Library"
source_repo: "https://github.com/livetemplate/lvt"
source_path: "components/README.md"
source_commit: "91211b7f0565d357db29bf390d39ea484bf96eb8"
---

# LiveTemplate Components Library

![Components Independence](https://github.com/livetemplate/lvt/actions/workflows/components-independence.yml/badge.svg)

A comprehensive collection of reusable UI components for the [LiveTemplate](https://github.com/livetemplate/livetemplate) framework.

## Features

- **Zero Boilerplate** - Use components directly via library API with functional options
- **Tailwind CSS** - Styled variants with Tailwind classes, plus unstyled semantic HTML options
- **Server-Side State** - All component state managed on the server via LiveTemplate
- **lvt-* Attributes** - Uses existing generic attributes, no client library changes needed
- **Full Customization** - Override templates or eject for complete control

## Installation

```bash
go get github.com/livetemplate/lvt/components
```

## Quick Start

### 1. Register Templates (once in main.go)

```go
import "github.com/livetemplate/lvt/components"

tmpl := livetemplate.NewTemplates(
    livetemplate.WithComponentTemplates(components.All()),
    livetemplate.ParseGlob("internal/app/**/*.tmpl"),
)
```

### 2. Use Components in Your Page

```go
import "github.com/livetemplate/lvt/components/dropdown"

type State struct {
    CountrySelect *dropdown.Searchable
}

// In init
CountrySelect: dropdown.NewSearchable("country", countries,
    dropdown.WithPlaceholder("Select country"),
    dropdown.WithSelected(user.CountryCode),
)
```

### 3. Render in Templates

```html
<div class="form-group">
    <label>Country</label>
    {{template "lvt:dropdown:searchable:v1" .CountrySelect}}
</div>
```

## Available Components

### Form Controls
| Component | Package | Templates | Description |
|-----------|---------|-----------|-------------|
| Dropdown | `dropdown` | default, searchable, multi | Single and multi-select dropdowns |
| Autocomplete | `autocomplete` | default | Search with suggestions |
| Date Picker | `datepicker` | single, range, inline | Date selection |
| Time Picker | `timepicker` | default | Time selection |
| Tags Input | `tagsinput` | default | Tag/chip input |
| Toggle | `toggle` | default, checkbox | Toggle switches |
| Rating | `rating` | default | Star ratings |

### Layout
| Component | Package | Templates | Description |
|-----------|---------|-----------|-------------|
| Tabs | `tabs` | horizontal, vertical, pills | Tab navigation |
| Accordion | `accordion` | default, single | Collapsible sections |
| Modal | `modal` | default, confirm, sheet | Modal dialogs |
| Drawer | `drawer` | default | Slide-out panels |

### Feedback
| Component | Package | Templates | Description |
|-----------|---------|-----------|-------------|
| Toast | `toast` | default, container | Toast notifications |
| Tooltip | `tooltip` | default | Tooltips |
| Popover | `popover` | default | Rich content popovers |
| Progress | `progress` | default, circular, spinner | Progress indicators |
| Skeleton | `skeleton` | default, avatar, card | Loading placeholders |

### Data Display
| Component | Package | Templates | Description |
|-----------|---------|-----------|-------------|
| Data Table | `datatable` | default | Tables with sorting/pagination |
| Timeline | `timeline` | default | Event timelines |
| Breadcrumbs | `breadcrumbs` | default | Navigation breadcrumbs |

### Navigation
| Component | Package | Templates | Description |
|-----------|---------|-----------|-------------|
| Menu | `menu` | default, nested | Navigation menus |

## Template Naming Convention

All templates follow the pattern: `lvt:<component>:<variant>:v<version>`

```html
{{template "lvt:dropdown:default:v1" .MyDropdown}}
{{template "lvt:dropdown:searchable:v1" .SearchDropdown}}
{{template "lvt:tabs:horizontal:v1" .MyTabs}}
{{template "lvt:modal:confirm:v1" .ConfirmModal}}
```

## Styling Options

### Styled (Default)
Components come with Tailwind CSS classes baked in:

```go
dropdown.New("id", options)
```

### Unstyled
For custom CSS or other frameworks, use unstyled mode:

```go
dropdown.New("id", options, dropdown.WithStyled(false))
```

This renders semantic HTML without any classes.

## Customization

### Option 1: Functional Options
Most customization can be achieved through functional options:

```go
dropdown.NewSearchable("country", countries,
    dropdown.WithPlaceholder("Select country"),
    dropdown.WithSelected("US"),
    dropdown.WithMinChars(2),
    dropdown.WithDebounce(300),
)
```

### Option 2: Override Template
Define the same template name in your project - it takes precedence:

```html
{{/* internal/app/templates/dropdown-override.tmpl */}}
{{define "lvt:dropdown:searchable:v1"}}
<div class="my-custom-dropdown">
    {{/* Your custom markup */}}
</div>
{{end}}
```

### Option 3: Eject Template Only
Extract just the HTML template while keeping the Go logic:

```bash
lvt component eject-template dropdown searchable
```

### Option 4: Full Eject
Get complete source code for total control:

```bash
lvt component eject dropdown
```

## CLI Commands

### List Available Components
```bash
lvt component list
```

### Eject Full Component
```bash
lvt component eject <component>
lvt component eject dropdown --dest internal/ui/dropdown
```

### Eject Template Only
```bash
lvt component eject-template <component> <template>
lvt component eject-template dropdown searchable
```

### Scaffold New Component
```bash
lvt new component <name>
lvt new component my-widget
```

## Component Development

### Creating a New Component

```bash
lvt new component rating
```

This creates:
```
components/rating/
  rating.go           # Component struct and constructor
  options.go          # Functional options
  templates.go        # Template embedding
  rating_test.go      # Tests
  templates/
    default.tmpl      # HTML template
```

### Component Structure

```go
// rating.go
package rating

import "github.com/livetemplate/lvt/components/base"

type Rating struct {
    base.Base
    Value    int
    MaxStars int
}

func New(id string, opts ...Option) *Rating {
    r := &Rating{
        Base:     base.NewBase(id, "rating"),
        MaxStars: 5,
    }
    for _, opt := range opts {
        opt(r)
    }
    return r
}
```

### Template Structure

```html
{{define "lvt:rating:default:v1"}}
{{if .IsStyled}}
<div class="flex gap-1" data-component="rating" data-id="{{.ID}}">
  {{/* Tailwind styled version */}}
</div>
{{else}}
<div data-component="rating" data-id="{{.ID}}">
  {{/* Unstyled semantic HTML */}}
</div>
{{end}}
{{end}}
```

## Independence Guarantee

The components module is fully independent from the parent `lvt` CLI tool.
This means external projects can import and use components without pulling
in the entire `lvt` dependency tree.

**What CI checks:**
- No imports of `github.com/livetemplate/lvt` (parent module)
- No imports of `github.com/livetemplate/lvt/internal` or `github.com/livetemplate/lvt/commands` (fully-qualified import paths)
- `components/go.mod` does not require the parent module
- `go build ./...` and `go test ./...` pass with `GOWORK=off`
- An external test module can import and build against components

**Using components in your project:**

```go
import "github.com/livetemplate/lvt/components/modal"

m := modal.New("my-modal", modal.WithTitle("Confirm"))
```

No `lvt` CLI installation required — components are a standalone Go library.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Component naming conventions
- Template requirements
- Test requirements
- PR process

## License

MIT License - see LICENSE file for details.
