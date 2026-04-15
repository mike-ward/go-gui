package gui

import (
	"fmt"
)

// TimeTravelController drives the time-travel scrubber UI. Holds
// a reference to the app window being debugged and a cursor into
// its history ring. All motion methods auto-freeze the app
// window; ResumeLive releases the freeze.
//
// Safe to construct without history or with a nil app window —
// methods are no-ops in those cases so the UI renders gracefully
// before state exists.
type TimeTravelController struct {
	// App is the window being debugged. Nil is allowed; methods
	// become no-ops.
	App *Window
	// Cursor indexes into App.history (0-based, 0 = oldest).
	// Out-of-range values are clamped by Jump.
	Cursor int
}

// Len returns the current history length, or 0 when history is
// unset.
func (c *TimeTravelController) Len() int {
	if c == nil || c.App == nil || c.App.history == nil {
		return 0
	}
	return c.App.history.len()
}

// Bytes returns the current history byte usage, or 0 when
// history is unset.
func (c *TimeTravelController) Bytes() int {
	if c == nil || c.App == nil || c.App.history == nil {
		return 0
	}
	return c.App.history.bytes()
}

// Cause returns the cause label of the entry at the cursor, or
// "" when out of range.
func (c *TimeTravelController) Cause() string {
	if c == nil || c.App == nil || c.App.history == nil {
		return ""
	}
	e, ok := c.App.history.at(c.Cursor)
	if !ok {
		return ""
	}
	return e.cause
}

// Jump moves the cursor to idx, clamped to [0, len-1]. Auto-
// freezes the app window and posts a restore request. No-op
// when history is empty.
func (c *TimeTravelController) Jump(idx int) {
	if c == nil || c.App == nil || c.App.history == nil {
		return
	}
	n := c.App.history.len()
	if n == 0 {
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	c.Cursor = idx
	c.App.Freeze()
	c.App.PostRestore(idx)
}

// StepBack moves one entry toward the past. Clamps at 0.
func (c *TimeTravelController) StepBack() {
	if c == nil {
		return
	}
	c.Jump(c.Cursor - 1)
}

// StepForward moves one entry toward the present. Clamps at len-1.
func (c *TimeTravelController) StepForward() {
	if c == nil {
		return
	}
	c.Jump(c.Cursor + 1)
}

// First jumps to the oldest entry.
func (c *TimeTravelController) First() {
	c.Jump(0)
}

// Last jumps to the newest entry.
func (c *TimeTravelController) Last() {
	if c == nil || c.App == nil || c.App.history == nil {
		return
	}
	c.Jump(c.App.history.len() - 1)
}

// ResumeLive releases the app window's freeze and snaps the
// cursor to the newest entry. The app resumes accepting input
// and w.Now() returns live time.
func (c *TimeTravelController) ResumeLive() {
	if c == nil || c.App == nil {
		return
	}
	c.App.Resume()
	if c.App.history != nil {
		if n := c.App.history.len(); n > 0 {
			c.Cursor = n - 1
		}
	}
}

// ToggleFreeze flips the frozen state. Intended for a freeze
// button or Space-key shortcut.
func (c *TimeTravelController) ToggleFreeze() {
	if c == nil || c.App == nil {
		return
	}
	if c.App.IsFrozen() {
		c.ResumeLive()
	} else {
		c.App.Freeze()
	}
}

// View returns the debug scrubber UI composed from the
// controller's current state. Intended to be returned from a
// Window's view generator. The view is standalone — it can be
// dropped into any layout, not only a dedicated debug window.
func (c *TimeTravelController) View() View {
	n := c.Len()
	bytes := c.Bytes()
	displayIdx := c.Cursor + 1
	if n == 0 {
		displayIdx = 0
	}
	counter := fmt.Sprintf("%d / %d (%d KiB)", displayIdx, n, bytes>>10)
	cause := c.Cause()
	if cause == "" {
		cause = "(empty)"
	}

	sliderMax := float32(0)
	if n > 1 {
		sliderMax = float32(n - 1)
	}

	return Column(ContainerCfg{
		IDFocus:   ttDebugFocusID,
		OnKeyDown: c.handleKey,
		Content: []View{
			Text(TextCfg{Text: counter}),
			Text(TextCfg{Text: cause}),
			Slider(SliderCfg{
				Min:   0,
				Max:   sliderMax,
				Value: float32(c.Cursor),
				OnChange: func(v float32, _ *Event, _ *Window) {
					c.Jump(int(v))
				},
			}),
			Row(ContainerCfg{
				Content: []View{
					ttButton("<<", c.First),
					ttButton("<", c.StepBack),
					ttButton(">", c.StepForward),
					ttButton(">>", c.Last),
					ttButton(freezeLabel(c.App), c.ToggleFreeze),
					ttButton("Resume", c.ResumeLive),
				},
			}),
		},
	})
}

// ttDebugFocusID is the IDFocus for the debug window's root
// container. Fixed because there's only one focusable widget
// in the scrubber UI.
const ttDebugFocusID uint32 = 1

// handleKey maps scrubber keyboard shortcuts to controller
// actions. Called from the root container's OnKeyDown.
func (c *TimeTravelController) handleKey(_ *Layout, e *Event, _ *Window) {
	if c == nil || e == nil {
		return
	}
	switch e.KeyCode {
	case KeyLeft:
		c.StepBack()
	case KeyRight:
		c.StepForward()
	case KeyHome:
		c.First()
	case KeyEnd:
		c.Last()
	case KeySpace:
		c.ToggleFreeze()
	case KeyEscape:
		c.ResumeLive()
	default:
		return
	}
	e.IsHandled = true
}

// ttButton builds a labelled scrubber button whose click
// handler invokes the given zero-arg function.
func ttButton(label string, fn func()) View {
	return Button(ButtonCfg{
		Content: []View{Text(TextCfg{Text: label})},
		OnClick: func(_ *Layout, _ *Event, _ *Window) { fn() },
	})
}

// freezeLabel returns the text for the freeze-toggle button.
func freezeLabel(w *Window) string {
	if w != nil && w.IsFrozen() {
		return "Frozen"
	}
	return "Freeze"
}
