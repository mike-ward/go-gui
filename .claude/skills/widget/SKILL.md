---
name: widget
description: Create a new go-gui widget with proper Cfg struct and factory function
disable-model-invocation: true
---

# New Widget

Create a new widget in the `gui/` package following established conventions.

## Arguments
- `name` (required): widget name (e.g., "Slider", "ColorPicker")

## Widget Structure

Every widget consists of:
1. A `*Cfg` struct (zero-initializable, exported fields)
2. A factory function returning `*Layout`
3. Event callbacks using `func(*Layout, *Event, *Window)` signature

## Template

```go
package gui

// <Name>Cfg configures the <Name> widget.
type <Name>Cfg struct {
    // IDFocus opts into tab-order focus when > 0.
    IDFocus uint32

    // Widget-specific fields...

    // Event callbacks
    OnClick func(*Layout, *Event, *Window)
}

// <Name> creates a <Name> widget.
func <Name>(cfg <Name>Cfg) *Layout {
    // Build layout tree
    // Wire event handlers (set e.IsHandled = true when consumed)
    // Return root *Layout
}
```

## Rules
- File name: `view_<lowercase_name>.go` in `gui/`
- Cfg struct must be zero-initializable (sensible defaults)
- Event callbacks: `func(*Layout, *Event, *Window)`, set `e.IsHandled = true`
- `IDFocus uint32` field for tab-order focus support
- No variable shadowing (use `=` not `:=` for outer-scope vars)
- Read existing widgets (e.g., `view_button.go`, `view_slider.go`) for patterns
- Must pass `golangci-lint run ./gui/...`
