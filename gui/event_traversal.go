package gui

// ShapeCallback is the type for shape event callbacks.
type ShapeCallback = func(*Layout, *Event, *Window)

// executeFocusCallback executes a callback if the layout has
// focus. Returns true if handled.
func isFocusedTarget(layout *Layout, w *Window) bool {
	if layout.Shape.IDFocus == 0 {
		return false
	}
	if !w.IsFocus(layout.Shape.IDFocus) &&
		layout.Shape.ID != reservedDialogID {
		return false
	}
	return true
}

func executeFocusCallback(
	layout *Layout, e *Event, w *Window,
	callback ShapeCallback, name string,
) bool {
	if !isFocusedTarget(layout, w) {
		return false
	}
	if callback == nil {
		return false
	}
	callback(layout, e, w)
	return e.IsHandled
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
	saved := *e
	e.MouseX = saved.MouseX - layout.Shape.X
	e.MouseY = saved.MouseY - layout.Shape.Y
	callback(layout, e, w)
	handled := e.IsHandled
	*e = saved
	if handled {
		e.IsHandled = true
		return true
	}
	return false
}

// isChildEnabled checks if a child layout should receive events.
func isChildEnabled(child *Layout) bool {
	return !child.Shape.Disabled
}
