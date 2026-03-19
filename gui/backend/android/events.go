//go:build android

package android

import "github.com/mike-ward/go-gui/gui"

// Touch-to-mouse event mapping. Single-touch maps to mouse
// events for widget compatibility. Multi-touch: only first
// touch drives mouse events.

func touchDown(x, y float32) {
	if androidWindow == nil {
		return
	}
	evt := gui.Event{
		Type:        gui.EventMouseDown,
		MouseX:      x,
		MouseY:      y,
		MouseButton: gui.MouseLeft,
	}
	androidWindow.EventFn(&evt)
}

func touchMoved(x, y float32) {
	if androidWindow == nil {
		return
	}
	evt := gui.Event{
		Type:   gui.EventMouseMove,
		MouseX: x,
		MouseY: y,
	}
	androidWindow.EventFn(&evt)
}

func touchUp(x, y float32) {
	if androidWindow == nil {
		return
	}
	evt := gui.Event{
		Type:        gui.EventMouseUp,
		MouseX:      x,
		MouseY:      y,
		MouseButton: gui.MouseLeft,
	}
	androidWindow.EventFn(&evt)
}
