package gui

import "testing"

func BenchmarkEventFnMouseMove(b *testing.B) {
	w := newEventTestWindow()
	w.layout = Layout{
		Shape: &Shape{},
		Children: []Layout{
			{Shape: &Shape{
				ShapeClip: DrawClip{X: 0, Y: 0, Width: 200, Height: 200},
				Events: &EventHandlers{
					OnMouseMove: func(_ *Layout, e *Event, _ *Window) {
						e.IsHandled = true
					},
				},
			}},
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	e := &Event{
		Type:      EventMouseMove,
		MouseX:    25,
		MouseY:    25,
		Modifiers: ModNone,
	}
	for b.Loop() {
		e.IsHandled = false
		w.EventFn(e)
	}
}

func BenchmarkEventFnMouseScrollFocused(b *testing.B) {
	w := newEventTestWindow()
	w.layout = Layout{
		Shape: &Shape{},
		Children: []Layout{
			{Shape: &Shape{
				IDFocus: 77,
				Events: &EventHandlers{
					OnMouseScroll: func(_ *Layout, _ *Event, _ *Window) {},
				},
			}},
			{Shape: &Shape{
				IDScroll: 1,
				Width:    100,
				Height:   50,
				ShapeClip: DrawClip{
					X: 0, Y: 0, Width: 100, Height: 50,
				},
			}, Children: []Layout{
				{Shape: &Shape{Height: 200}},
			}},
		},
	}
	w.SetIDFocus(77)
	b.ReportAllocs()
	b.ResetTimer()
	e := &Event{
		Type:      EventMouseScroll,
		MouseX:    10,
		MouseY:    10,
		ScrollY:   -1,
		Modifiers: ModNone,
	}
	for b.Loop() {
		e.IsHandled = false
		w.EventFn(e)
	}
}

func BenchmarkExecuteMouseCallback(b *testing.B) {
	layout := &Layout{
		Shape: &Shape{
			ShapeClip: DrawClip{X: 10, Y: 10, Width: 100, Height: 100},
			X:         10,
			Y:         10,
		},
	}
	w := newEventTestWindow()
	cb := func(_ *Layout, e *Event, _ *Window) {
		if e.MouseX >= 0 && e.MouseY >= 0 {
			e.IsHandled = true
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	e := &Event{MouseX: 25, MouseY: 25}
	for b.Loop() {
		e.IsHandled = false
		e.MouseX = 25
		e.MouseY = 25
		executeMouseCallback(layout, e, w, cb, "bench")
	}
}
