package gui

import "testing"

// helper: build a layout with one focused child that has events.
func focusedChild(idFocus uint32, eh *EventHandlers) *Layout {
	return &Layout{
		Shape: &Shape{},
		Children: []Layout{
			{Shape: &Shape{
				IDFocus: idFocus,
				Events:  eh,
				ShapeClip: DrawClip{
					X: 0, Y: 0, Width: 100, Height: 100,
				},
			}},
		},
	}
}

// --- charHandler ---

func TestCharHandlerDelivers(t *testing.T) {
	called := false
	root := focusedChild(1, &EventHandlers{
		OnChar: func(_ *Layout, e *Event, _ *Window) {
			called = true
			e.IsHandled = true
		},
	})
	w := &Window{}
	w.SetIDFocus(1)
	e := &Event{CharCode: 'a'}
	charHandler(root, e, w)
	if !called {
		t.Error("OnChar not called")
	}
	if !e.IsHandled {
		t.Error("event not handled")
	}
}

func TestCharHandlerSkipsDisabled(t *testing.T) {
	called := false
	root := &Layout{
		Shape: &Shape{},
		Children: []Layout{
			{Shape: &Shape{
				IDFocus:  1,
				Disabled: true,
				Events: &EventHandlers{
					OnChar: func(_ *Layout, e *Event, _ *Window) {
						called = true
						e.IsHandled = true
					},
				},
			}},
		},
	}
	w := &Window{}
	w.SetIDFocus(1)
	e := &Event{CharCode: 'a'}
	charHandler(root, e, w)
	if called {
		t.Error("should skip disabled")
	}
}

// --- keydownHandler ---

func TestKeydownHandlerDelivers(t *testing.T) {
	called := false
	root := focusedChild(1, &EventHandlers{
		OnKeyDown: func(_ *Layout, e *Event, _ *Window) {
			called = true
			e.IsHandled = true
		},
	})
	w := &Window{}
	w.SetIDFocus(1)
	e := &Event{KeyCode: KeyEnter}
	keydownHandler(root, e, w)
	if !called {
		t.Error("OnKeyDown not called")
	}
}

func TestKeydownHandlerFallbackScroll(t *testing.T) {
	// Focused element with IDScroll but no OnKeyDown handler.
	root := &Layout{
		Shape: &Shape{},
		Children: []Layout{
			{Shape: &Shape{
				IDFocus:  1,
				IDScroll: 1,
				Width:    100,
				Height:   100,
				Padding:  Padding{},
			}},
		},
	}
	w := &Window{}
	w.SetIDFocus(1)
	guiTheme.ScrollDeltaLine = 20
	e := &Event{KeyCode: KeyDown, Modifiers: ModNone}
	keydownHandler(root, e, w)
	if !e.IsHandled {
		t.Error("scroll fallback should handle KeyDown")
	}
}

// --- keyDownScrollHandler ---

func TestKeyDownScrollHandlerArrows(t *testing.T) {
	guiTheme.ScrollDeltaLine = 20
	guiTheme.ScrollDeltaPage = 100
	guiTheme.ScrollMultiplier = 1

	layout := &Layout{Shape: &Shape{
		IDScroll: 1, Width: 100, Height: 100,
	}}
	w := &Window{}

	tests := []struct {
		name string
		key  KeyCode
		mod  Modifier
	}{
		{"up", KeyUp, ModNone},
		{"down", KeyDown, ModNone},
		{"page_up", KeyPageUp, ModNone},
		{"page_down", KeyPageDown, ModNone},
		{"home", KeyHome, ModNone},
		{"end", KeyEnd, ModNone},
		{"shift+left", KeyLeft, ModShift},
		{"shift+right", KeyRight, ModShift},
	}
	for _, tc := range tests {
		e := &Event{KeyCode: tc.key, Modifiers: tc.mod}
		keyDownScrollHandler(layout, e, w)
		if !e.IsHandled {
			t.Errorf("%s not handled", tc.name)
		}
	}
}

// --- mouseDownHandler ---

func TestMouseDownHandlerDelivers(t *testing.T) {
	clicked := false
	root := &Layout{
		Shape: &Shape{},
		Children: []Layout{
			{Shape: &Shape{
				ShapeClip: DrawClip{X: 0, Y: 0,
					Width: 100, Height: 100},
				Events: &EventHandlers{
					OnClick: func(_ *Layout, e *Event, _ *Window) {
						clicked = true
						e.IsHandled = true
					},
				},
			}},
		},
	}
	w := &Window{windowWidth: 800, windowHeight: 600}
	e := &Event{MouseX: 50, MouseY: 50, Type: EventMouseDown}
	mouseDownHandler(root, false, e, w)
	if !clicked {
		t.Error("OnClick not called")
	}
}

func TestMouseDownHandlerSetsFocus(t *testing.T) {
	root := &Layout{
		Shape: &Shape{},
		Children: []Layout{
			{Shape: &Shape{
				IDFocus: 42,
				ShapeClip: DrawClip{X: 0, Y: 0,
					Width: 100, Height: 100},
			}},
		},
	}
	w := &Window{windowWidth: 800, windowHeight: 600}
	e := &Event{MouseX: 50, MouseY: 50}
	mouseDownHandler(root, false, e, w)
	if w.IDFocus() != 42 {
		t.Errorf("focus: got %d, want 42", w.IDFocus())
	}
}

func TestMouseDownHandlerRespectsMouseLock(t *testing.T) {
	lockCalled := false
	w := &Window{windowWidth: 800, windowHeight: 600}
	w.MouseLock(MouseLockCfg{
		MouseDown: func(_ *Layout, e *Event, _ *Window) {
			lockCalled = true
			e.IsHandled = true
		},
	})
	root := &Layout{Shape: &Shape{}}
	e := &Event{MouseX: 50, MouseY: 50}
	mouseDownHandler(root, false, e, w)
	if !lockCalled {
		t.Error("mouse lock should intercept")
	}
}

func TestMouseDownHandlerReverseOrder(t *testing.T) {
	// Last child (topmost) should receive the event first.
	var hitID string
	mkChild := func(id string, x float32) Layout {
		return Layout{Shape: &Shape{
			ID: id,
			ShapeClip: DrawClip{X: x, Y: 0,
				Width: 100, Height: 100},
			Events: &EventHandlers{
				OnClick: func(l *Layout, e *Event, _ *Window) {
					hitID = l.Shape.ID
					e.IsHandled = true
				},
			},
		}}
	}
	// Both children overlap at x=50.
	root := &Layout{
		Shape: &Shape{},
		Children: []Layout{
			mkChild("first", 0),
			mkChild("second", 0),
		},
	}
	w := &Window{windowWidth: 800, windowHeight: 600}
	e := &Event{MouseX: 50, MouseY: 50}
	mouseDownHandler(root, false, e, w)
	if hitID != "second" {
		t.Errorf("hit: got %q, want second", hitID)
	}
}

// --- mouseMoveHandler ---

func TestMouseMoveHandlerRespectsMouseLock(t *testing.T) {
	lockCalled := false
	w := &Window{windowWidth: 800, windowHeight: 600}
	w.MouseLock(MouseLockCfg{
		MouseMove: func(_ *Layout, e *Event, _ *Window) {
			lockCalled = true
		},
	})
	root := &Layout{Shape: &Shape{}}
	e := &Event{MouseX: 50, MouseY: 50}
	mouseMoveHandler(root, e, w)
	if !lockCalled {
		t.Error("mouse lock should intercept move")
	}
}

func TestMouseMoveHandlerSkipsOutOfWindow(t *testing.T) {
	called := false
	root := &Layout{
		Shape: &Shape{
			ShapeClip: DrawClip{X: 0, Y: 0,
				Width: 100, Height: 100},
			Events: &EventHandlers{
				OnMouseMove: func(_ *Layout, e *Event, _ *Window) {
					called = true
					e.IsHandled = true
				},
			},
		},
	}
	w := &Window{windowWidth: 800, windowHeight: 600}
	e := &Event{MouseX: -10, MouseY: 50}
	mouseMoveHandler(root, e, w)
	if called {
		t.Error("should skip out-of-window move")
	}
}

// --- mouseUpHandler ---

func TestMouseUpHandlerRespectsMouseLock(t *testing.T) {
	lockCalled := false
	w := &Window{windowWidth: 800, windowHeight: 600}
	w.MouseLock(MouseLockCfg{
		MouseUp: func(_ *Layout, e *Event, _ *Window) {
			lockCalled = true
		},
	})
	root := &Layout{Shape: &Shape{}}
	e := &Event{MouseX: 50, MouseY: 50}
	mouseUpHandler(root, e, w)
	if !lockCalled {
		t.Error("mouse lock should intercept up")
	}
}

// --- mouseScrollHandler ---

func TestMouseScrollHandlerVertical(t *testing.T) {
	guiTheme.ScrollMultiplier = 1
	root := &Layout{Shape: &Shape{
		IDScroll: 1,
		Width:    100,
		Height:   50,
		ShapeClip: DrawClip{X: 0, Y: 0,
			Width: 100, Height: 50},
	}, Children: []Layout{
		{Shape: &Shape{Height: 200}}, // tall content
	}}
	w := &Window{windowWidth: 800, windowHeight: 600}
	e := &Event{
		MouseX:    50,
		MouseY:    25,
		ScrollY:   -10,
		Modifiers: ModNone,
	}
	mouseScrollHandler(root, e, w)
	if !e.IsHandled {
		t.Error("vertical scroll not handled")
	}
}

func TestMouseScrollHandlerHorizontalShift(t *testing.T) {
	guiTheme.ScrollMultiplier = 1
	root := &Layout{Shape: &Shape{
		IDScroll: 1,
		Width:    50,
		Height:   100,
		ShapeClip: DrawClip{X: 0, Y: 0,
			Width: 50, Height: 100},
		Axis: AxisLeftToRight,
	}, Children: []Layout{
		{Shape: &Shape{Width: 200}}, // wide content
	}}
	w := &Window{windowWidth: 800, windowHeight: 600}
	e := &Event{
		MouseX:    25,
		MouseY:    50,
		ScrollX:   -10,
		Modifiers: ModShift,
	}
	mouseScrollHandler(root, e, w)
	if !e.IsHandled {
		t.Error("horizontal scroll not handled")
	}
}

func TestMouseScrollHandlerFocusedOnMouseScroll(t *testing.T) {
	called := false
	root := &Layout{
		Shape: &Shape{},
		Children: []Layout{
			{Shape: &Shape{
				IDFocus: 5,
				Events: &EventHandlers{
					OnMouseScroll: func(_ *Layout, e *Event, _ *Window) {
						called = true
						e.IsHandled = true
					},
				},
			}},
		},
	}
	w := &Window{windowWidth: 800, windowHeight: 600}
	w.SetIDFocus(5)
	e := &Event{MouseX: 50, MouseY: 50, ScrollY: -10}
	mouseScrollHandler(root, e, w)
	if !called {
		t.Error("focused OnMouseScroll should be called")
	}
}
