package gui

import "log"

// ShapeCallback is the type for shape event callbacks.
type ShapeCallback = func(*Layout, *Event, *Window)

// executeFocusCallback executes a callback if the layout has
// focus. Returns true if handled.
func executeFocusCallback(
	layout *Layout, e *Event, w *Window,
	callback ShapeCallback, name string,
) bool {
	if layout.Shape.IDFocus == 0 {
		return false
	}
	if !w.IsFocus(layout.Shape.IDFocus) &&
		layout.Shape.ID != reservedDialogID {
		return false
	}
	if callback == nil {
		return false
	}
	callback(layout, e, w)
	if e.IsHandled {
		log.Printf("debug: %s handled by %s", name, layout.Shape.ID)
		return true
	}
	return false
}

// executeMouseCallback executes a callback if the mouse is
// within shape bounds. Coordinates are made relative before
// calling. Returns true if handled.
func executeMouseCallback(
	layout *Layout, e *Event, w *Window,
	callback ShapeCallback, name string,
) bool {
	if !layout.Shape.PointInShape(e.MouseX, e.MouseY) {
		return false
	}
	if callback == nil {
		return false
	}
	ev := eventRelativeTo(layout.Shape, e)
	callback(layout, &ev, w)
	if ev.IsHandled {
		e.IsHandled = true
		log.Printf("debug: %s handled by %s", name, layout.Shape.ID)
		return true
	}
	return false
}

// isChildEnabled checks if a child layout should receive events.
func isChildEnabled(child *Layout) bool {
	return !child.Shape.Disabled
}
