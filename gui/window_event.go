package gui

// EventFn handles user events, dispatching to child views.
// Called by the backend event loop.
func (w *Window) EventFn(e *Event) {
	if e == nil {
		return
	}

	// Focus gate: block events when unfocused except right-click,
	// focused, and scroll.
	if !w.focused && e.Type == EventMouseDown &&
		e.MouseButton == MouseRight {
		// Allow right clicks without focus (browser behavior).
	} else if !w.focused &&
		e.Type != EventFocused &&
		e.Type != EventMouseScroll {
		return
	}

	// Top-level layout children represent z-axis layers.
	// Dialogs are modal: route events to last child (dialog layer).
	layout := &w.layout
	if w.dialogCfg.visible && len(w.layout.Children) > 0 {
		layout = &w.layout.Children[len(w.layout.Children)-1]
	}

	switch e.Type {
	case EventChar:
		charHandler(layout, e, w)

	case EventFocused:
		w.focused = true

	case EventUnfocused:
		w.focused = false

	case EventKeyDown:
		keydownHandler(layout, e, w)
		if !e.IsHandled && e.KeyCode == KeyTab &&
			e.Modifiers == ModShift {
			if shape, ok := layout.PreviousFocusable(w); ok {
				w.SetIDFocus(shape.IDFocus)
			}
		} else if !e.IsHandled && e.KeyCode == KeyTab {
			if shape, ok := layout.NextFocusable(w); ok {
				w.SetIDFocus(shape.IDFocus)
			}
		}

	case EventMouseDown:
		w.SetMouseCursor(CursorArrow)
		mouseDownHandler(layout, false, e, w)
		if !e.IsHandled {
			ss := StateMap[string, bool](w, nsSelect, capModerate)
			ss.Clear()
			cs := StateMap[string, bool](w, nsCombobox, capModerate)
			cs.Clear()
		}

	case EventMouseMove:
		w.SetMouseCursor(CursorArrow)
		w.viewState.menuKeyNav = false
		w.viewState.mousePosX = e.MouseX
		w.viewState.mousePosY = e.MouseY
		mouseMoveHandler(layout, e, w)

	case EventMouseUp:
		mouseUpHandler(layout, e, w)

	case EventMouseScroll:
		mouseScrollHandler(layout, e, w)

	case EventResized:
		w.windowWidth = e.WindowWidth
		w.windowHeight = e.WindowHeight

	default:
		// Unhandled event type.
	}

	if !e.IsHandled && w.OnEvent != nil {
		w.OnEvent(e, w)
	}
	w.UpdateWindow()
}
