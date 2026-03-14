package gui

import "testing"

func TestSliderGeneratesLayout(t *testing.T) {
	w := &Window{}
	v := Slider(SliderCfg{ID: "sl1", Max: 100})
	layout := GenerateViewLayout(v, w)
	if layout.Shape == nil {
		t.Fatal("expected shape")
	}
	if layout.Shape.ID != "sl1" {
		t.Errorf("ID: got %s", layout.Shape.ID)
	}
	if layout.Shape.A11YRole != AccessRoleSlider {
		t.Errorf("a11y role: got %d, want %d",
			layout.Shape.A11YRole, AccessRoleSlider)
	}
}

func TestSliderValueChange(t *testing.T) {
	var got float32
	w := &Window{}
	v := Slider(SliderCfg{
		ID:    "sl2",
		Value: 50,
		Max:   100,
		OnChange: func(val float32, _ *Event, _ *Window) {
			got = val
		},
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.Events == nil ||
		layout.Shape.Events.OnKeyDown == nil {
		t.Fatal("expected OnKeyDown for keyboard control")
	}
	// Simulate right arrow key.
	e := &Event{KeyCode: KeyRight}
	layout.Shape.Events.OnKeyDown(&layout, e, w)
	if !e.IsHandled {
		t.Error("expected key handled")
	}
	if got != 51 {
		t.Errorf("value: got %f, want 51", got)
	}
}

func TestSliderKeyboardHomeEnd(t *testing.T) {
	var got float32
	w := &Window{}
	v := Slider(SliderCfg{
		ID:    "sl3",
		Value: 50,
		Min:   0,
		Max:   100,
		OnChange: func(val float32, _ *Event, _ *Window) {
			got = val
		},
	})
	layout := GenerateViewLayout(v, w)

	// Home → min.
	e := &Event{KeyCode: KeyHome}
	layout.Shape.Events.OnKeyDown(&layout, e, w)
	if got != 0 {
		t.Errorf("home: got %f, want 0", got)
	}

	// End → max.
	e = &Event{KeyCode: KeyEnd}
	layout.Shape.Events.OnKeyDown(&layout, e, w)
	if got != 100 {
		t.Errorf("end: got %f, want 100", got)
	}
}

func TestSliderDisabled(t *testing.T) {
	w := &Window{}
	v := Slider(SliderCfg{
		ID:       "sl4",
		Max:      100,
		Disabled: true,
	})
	layout := GenerateViewLayout(v, w)
	if !layout.Shape.Disabled {
		t.Error("expected disabled")
	}
}

func TestSliderVertical(t *testing.T) {
	w := &Window{}
	v := Slider(SliderCfg{
		ID:       "sl5",
		Max:      100,
		Vertical: true,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape == nil {
		t.Fatal("expected shape")
	}
	// Vertical slider should have a top-to-bottom axis.
	if layout.Shape.Axis != AxisTopToBottom {
		t.Errorf("axis: got %d, want %d",
			layout.Shape.Axis, AxisTopToBottom)
	}
}

func TestSliderIDFocus(t *testing.T) {
	w := &Window{}
	v := Slider(SliderCfg{
		ID:      "sl6",
		Max:     100,
		IDFocus: 77,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.IDFocus != 77 {
		t.Errorf("IDFocus: got %d, want 77", layout.Shape.IDFocus)
	}
}

func TestSliderA11YValues(t *testing.T) {
	w := &Window{}
	v := Slider(SliderCfg{
		ID:    "sl7",
		Value: 25,
		Min:   10,
		Max:   90,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11Y == nil {
		t.Fatal("expected a11y info")
	}
	if !f32AreClose(layout.Shape.A11Y.ValueNum, 25) {
		t.Errorf("a11y value: got %f", layout.Shape.A11Y.ValueNum)
	}
	if !f32AreClose(layout.Shape.A11Y.ValueMin, 10) {
		t.Errorf("a11y min: got %f", layout.Shape.A11Y.ValueMin)
	}
	if !f32AreClose(layout.Shape.A11Y.ValueMax, 90) {
		t.Errorf("a11y max: got %f", layout.Shape.A11Y.ValueMax)
	}
}
