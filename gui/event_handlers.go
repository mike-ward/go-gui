package gui

// charHandler handles character input events (typing).
// Traverses forward (depth-first) and delivers to focused element.
func charHandler(layout *Layout, e *Event, w *Window) {
	for i := range layout.Children {
		if !isChildEnabled(&layout.Children[i]) {
			continue
		}
		charHandler(&layout.Children[i], e, w)
		if e.IsHandled {
			return
		}
	}
	if layout.Shape == nil {
		return
	}
	var onChar ShapeCallback
	if layout.Shape.HasEvents() {
		onChar = layout.Shape.Events.OnChar
	}
	executeFocusCallback(layout, e, w, onChar, "charHandler")
}

// imeCompositionHandler handles IME composition events.
// Updates the per-window IME state for the focused input.
func imeCompositionHandler(_ *Layout, e *Event, w *Window) {
	w.imeUpdate(e)
	e.IsHandled = true
}

// keydownHandler handles key down events (special keys, shortcuts).
// Traverses forward and delivers to focused element. Falls back to
// keyboard scroll if the focused scroll container has no handler.
func keydownHandler(layout *Layout, e *Event, w *Window) {
	for i := range layout.Children {
		if !isChildEnabled(&layout.Children[i]) {
			continue
		}
		keydownHandler(&layout.Children[i], e, w)
		if e.IsHandled {
			return
		}
	}
	if layout.Shape == nil || !isFocusedTarget(layout, w) {
		return
	}
	var onKeyDown ShapeCallback
	if layout.Shape.HasEvents() {
		onKeyDown = layout.Shape.Events.OnKeyDown
	}
	executeFocusCallback(layout, e, w, onKeyDown, "keydownHandler")
	if e.IsHandled {
		return
	}
	if layout.Shape.IDScroll > 0 {
		keyDownScrollHandler(layout, e, w)
	}
}

// keyDownScrollHandler handles keyboard-based scrolling.
// Supports arrow keys, page up/down, and home/end.
func keyDownScrollHandler(layout *Layout, e *Event, w *Window) {
	deltaLine := guiTheme.ScrollDeltaLine
	deltaPage := guiTheme.ScrollDeltaPage
	const deltaHome float32 = 10_000_000

	if e.Modifiers == ModNone {
		switch e.KeyCode {
		case KeyUp:
			e.IsHandled = scrollVertical(layout, deltaLine, w)
		case KeyDown:
			e.IsHandled = scrollVertical(layout, -deltaLine, w)
		case KeyHome:
			e.IsHandled = scrollVertical(layout, deltaHome, w)
		case KeyEnd:
			e.IsHandled = scrollVertical(layout, -deltaHome, w)
		case KeyPageUp:
			e.IsHandled = scrollVertical(layout, deltaPage, w)
		case KeyPageDown:
			e.IsHandled = scrollVertical(layout, -deltaPage, w)
		}
	} else if e.Modifiers == ModShift {
		switch e.KeyCode {
		case KeyLeft:
			e.IsHandled = scrollHorizontal(layout, deltaLine, w)
		case KeyRight:
			e.IsHandled = scrollHorizontal(layout, -deltaLine, w)
		}
	}
}

// mouseDownHandler handles mouse button press events.
// Traverses reverse (topmost first) and delivers to element under
// cursor. Also handles focus changes on click.
func mouseDownHandler(
	layout *Layout, inHandler bool, e *Event, w *Window,
) {
	// Check mouse lock (only at top level).
	if !inHandler {
		if w.viewState.mouseLock.MouseDown != nil {
			w.viewState.mouseLock.MouseDown(layout, e, w)
			return
		}
	}
	// Traverse children in reverse (topmost/last child first).
	ox, oy := rotateMouseInverse(layout.Shape, e)
	for i := len(layout.Children) - 1; i >= 0; i-- {
		if !isChildEnabled(&layout.Children[i]) {
			continue
		}
		mouseDownHandler(&layout.Children[i], true, e, w)
		if e.IsHandled {
			e.MouseX, e.MouseY = ox, oy
			return
		}
	}
	e.MouseX, e.MouseY = ox, oy
	if layout.Shape == nil {
		return
	}
	if layout.Shape.PointInShape(e.MouseX, e.MouseY) {
		if layout.Shape.IDFocus > 0 {
			w.SetIDFocus(layout.Shape.IDFocus)
		}
		var onClick ShapeCallback
		if layout.Shape.HasEvents() {
			onClick = layout.Shape.Events.OnClick
		}
		executeMouseCallback(layout, e, w, onClick,
			"mouseDownHandler")
	}
}

// mouseMoveHandler handles mouse movement events.
// Traverses reverse (topmost first).
func mouseMoveHandler(layout *Layout, e *Event, w *Window) {
	if w.viewState.mouseLock.MouseMove != nil {
		w.viewState.mouseLock.MouseMove(layout, e, w)
		return
	}
	if !w.PointerOverApp(e) {
		return
	}
	ox, oy := rotateMouseInverse(layout.Shape, e)
	for i := len(layout.Children) - 1; i >= 0; i-- {
		if !isChildEnabled(&layout.Children[i]) {
			continue
		}
		mouseMoveHandler(&layout.Children[i], e, w)
		if e.IsHandled {
			e.MouseX, e.MouseY = ox, oy
			return
		}
	}
	e.MouseX, e.MouseY = ox, oy
	if layout.Shape == nil {
		return
	}
	var onMouseMove ShapeCallback
	if layout.Shape.HasEvents() {
		onMouseMove = layout.Shape.Events.OnMouseMove
	}
	executeMouseCallback(layout, e, w, onMouseMove,
		"mouseMoveHandler")
}

// mouseUpHandler handles mouse button release events.
// Traverses reverse (topmost first).
func mouseUpHandler(layout *Layout, e *Event, w *Window) {
	if w.viewState.mouseLock.MouseUp != nil {
		w.viewState.mouseLock.MouseUp(layout, e, w)
		return
	}
	ox, oy := rotateMouseInverse(layout.Shape, e)
	for i := len(layout.Children) - 1; i >= 0; i-- {
		if !isChildEnabled(&layout.Children[i]) {
			continue
		}
		mouseUpHandler(&layout.Children[i], e, w)
		if e.IsHandled {
			e.MouseX, e.MouseY = ox, oy
			return
		}
	}
	e.MouseX, e.MouseY = ox, oy
	if layout.Shape == nil {
		return
	}
	var onMouseUp ShapeCallback
	if layout.Shape.HasEvents() {
		onMouseUp = layout.Shape.Events.OnMouseUp
	}
	executeMouseCallback(layout, e, w, onMouseUp,
		"mouseUpHandler")
}

func focusedScrollTarget(layout *Layout, w *Window) *Layout {
	if w == nil {
		return nil
	}
	idFocus := w.IDFocus()
	if idFocus == 0 {
		return nil
	}
	ly, ok := FindLayoutByIDFocus(layout, idFocus)
	if !ok || ly.Shape == nil || !ly.Shape.HasEvents() ||
		ly.Shape.Events.OnMouseScroll == nil {
		return nil
	}
	return ly
}

// mouseScrollHandler handles mouse wheel scroll events.
// Delivers to the focused element's OnMouseScroll handler first.
// If no focused handler exists, traverses reverse (topmost first)
// and falls back to the scroll container under cursor.
func mouseScrollHandler(layout *Layout, e *Event, w *Window) {
	if ly := focusedScrollTarget(layout, w); ly != nil {
		ly.Shape.Events.OnMouseScroll(ly, e, w)
		return
	}
	mouseScrollFallbackHandler(layout, e, w)
}

func mouseScrollFallbackHandler(layout *Layout, e *Event, w *Window) {
	ox, oy := rotateMouseInverse(layout.Shape, e)
	for i := len(layout.Children) - 1; i >= 0; i-- {
		if !isChildEnabled(&layout.Children[i]) {
			continue
		}
		mouseScrollFallbackHandler(&layout.Children[i], e, w)
		if e.IsHandled {
			e.MouseX, e.MouseY = ox, oy
			return
		}
	}
	e.MouseX, e.MouseY = ox, oy
	if layout.Shape == nil || layout.Shape.Disabled {
		return
	}
	// Deliver to OnMouseScroll handler under cursor.
	if layout.Shape.HasEvents() &&
		layout.Shape.Events.OnMouseScroll != nil {
		if layout.Shape.PointInShape(e.MouseX, e.MouseY) {
			layout.Shape.Events.OnMouseScroll(layout, e, w)
			return
		}
	}
	// Handle scroll on scroll container under cursor.
	if layout.Shape.IDScroll > 0 {
		if layout.Shape.PointInShape(e.MouseX, e.MouseY) {
			if e.Modifiers == ModShift {
				e.IsHandled = scrollHorizontal(
					layout, e.ScrollX, w)
			} else if e.Modifiers == ModNone {
				e.IsHandled = scrollVertical(
					layout, e.ScrollY, w)
			}
		}
	}
}
