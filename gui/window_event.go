package gui

// EventFn handles user events, dispatching to child views.
// Called by the backend event loop.
func (w *Window) EventFn(e *Event) {
	if e == nil {
		return
	}

	// Focus gate: block events when unfocused except right-click,
	// focused, and scroll.
	if !w.focused &&
		!(e.Type == EventMouseDown && e.MouseButton == MouseRight) &&
		e.Type != EventFocused &&
		e.Type != EventMouseScroll {
		return
	}

	if inspectorSupported && e.Type == EventKeyDown &&
		e.KeyCode == KeyF12 {
		inspectorToggle(w)
		e.IsHandled = true
		return
	}
	if inspectorSupported && w.inspectorEnabled &&
		e.Type == EventKeyDown &&
		e.Modifiers == ModCtrl {
		switch e.KeyCode {
		case KeyLeft:
			inspectorResize(inspectorResizeStep, w)
			e.IsHandled = true
			return
		case KeyRight:
			inspectorResize(-inspectorResizeStep, w)
			e.IsHandled = true
			return
		case KeyUp:
			inspectorToggleSide(w)
			e.IsHandled = true
			return
		}
	}

	// Top-level layout children represent z-axis layers.
	// Dialogs are modal: route events to last child (dialog layer).
	layout := &w.layout
	if w.dialogCfg.visible && len(w.layout.Children) > 0 {
		layout = &w.layout.Children[len(w.layout.Children)-1]
	}

	switch e.Type {
	case EventChar:
		w.imeClear()
		charHandler(layout, e, w)

	case EventIMEComposition:
		imeCompositionHandler(layout, e, w)

	case EventFocused:
		w.focused = true

	case EventUnfocused:
		w.focused = false
		w.imeClear()

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
		if inspectorSupported && w.inspectorEnabled {
			panelW := inspectorPanelWidth(w)
			left := inspectorIsLeft(w)
			var inApp bool
			if left {
				inApp = e.MouseX > panelW+inspectorMargin
			} else {
				inApp = e.MouseX < float32(w.windowWidth)-panelW-inspectorMargin
			}
			if inApp {
				if picked := inspectorPickPath(&w.layout, e.MouseX, e.MouseY); picked != "" {
					inspectorSelect(picked, w)
				}
				e.IsHandled = true
			}
		}
		// Dismiss open popups on any mouse down. Cleared
		// before dispatch so handlers can re-open. Focus
		// is only cleared when a popup was actually open
		// to avoid interfering with normal focus flow.
		popupOpen := false
		if cms := StateMapRead[string, contextMenuState](
			w, nsContextMenu); cms != nil && cms.Len() > 0 {
			cms.Clear()
			popupOpen = true
		}
		if rms := StateMapRead[string, rtfLinkMenuState](
			w, nsRtfLinkMenu); rms != nil && rms.Len() > 0 {
			rms.Clear()
			popupOpen = true
		}
		if ms := StateMapRead[uint32, string](
			w, nsMenu); ms != nil && ms.Len() > 0 {
			ms.Clear()
			popupOpen = true
		}
		if popupOpen {
			w.SetIDFocus(0)
		}
		if !e.IsHandled {
			mouseDownHandler(layout, false, e, w)
		}
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
