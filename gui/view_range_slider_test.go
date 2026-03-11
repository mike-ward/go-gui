package gui

import (
	"math"
	"testing"
)

func TestRangeSliderDefaultLayout(t *testing.T) {
	v := RangeSlider(RangeSliderCfg{
		ID:       "rs",
		Value:    50,
		OnChange: func(float32, *Event, *Window) {},
	})
	layout := GenerateViewLayout(v, &Window{})
	// Wrapper container with 1 child (track)
	if len(layout.Children) != 1 {
		t.Fatalf("children: got %d, want 1", len(layout.Children))
	}
	track := layout.Children[0]
	// Track has fill bar + thumb
	if len(track.Children) != 2 {
		t.Fatalf("track children: got %d, want 2",
			len(track.Children))
	}
}

func TestRangeSliderA11Y(t *testing.T) {
	v := RangeSlider(RangeSliderCfg{
		ID:       "rs",
		Value:    30,
		Min:      0,
		Max:      100,
		OnChange: func(float32, *Event, *Window) {},
	})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.A11YRole != AccessRoleSlider {
		t.Errorf("role = %d, want Slider", layout.Shape.A11YRole)
	}
	a := layout.Shape.A11Y
	if a == nil {
		t.Fatal("a11y should be set")
	}
	if a.ValueNum != 30 {
		t.Errorf("value_num = %f, want 30", a.ValueNum)
	}
	if a.ValueMin != 0 || a.ValueMax != 100 {
		t.Errorf("range = %f-%f, want 0-100",
			a.ValueMin, a.ValueMax)
	}
}

func TestRangeSliderMinMaxValidation(t *testing.T) {
	v := RangeSlider(RangeSliderCfg{
		ID:       "rs",
		Min:      50,
		Max:      50, // invalid: min >= max
		OnChange: func(float32, *Event, *Window) {},
	})
	layout := GenerateViewLayout(v, &Window{})
	// Should auto-adjust max to min+1
	if layout.Shape.A11Y.ValueMax != 51 {
		t.Errorf("adjusted max = %f, want 51",
			layout.Shape.A11Y.ValueMax)
	}
}

func TestRangeSliderKeyDown(t *testing.T) {
	var got float32
	onChange := func(v float32, _ *Event, _ *Window) { got = v }
	tests := []struct {
		key  KeyCode
		want float32
	}{
		{KeyHome, 0},
		{KeyEnd, 100},
		{KeyRight, 51},
		{KeyLeft, 49},
	}
	for _, tt := range tests {
		got = -1
		e := &Event{KeyCode: tt.key}
		rangeSliderOnKeyDown(nil, e, &Window{},
			onChange, 50, 0, 100, 1, false)
		if got != tt.want {
			t.Errorf("key %d: got %f, want %f",
				tt.key, got, tt.want)
		}
	}
}

func TestRangeSliderMouseScroll(t *testing.T) {
	var got float32
	onChange := func(v float32, _ *Event, _ *Window) { got = v }
	e := &Event{ScrollY: 5}
	rangeSliderOnMouseScroll(e, &Window{}, onChange,
		50, 0, 100, false)
	if got != 55 {
		t.Errorf("scroll: got %f, want 55", got)
	}
	if !e.IsHandled {
		t.Error("scroll should mark handled")
	}
}

func TestRangeSliderVertical(t *testing.T) {
	v := RangeSlider(RangeSliderCfg{
		ID:       "rs",
		Value:    50,
		Vertical: true,
		OnChange: func(float32, *Event, *Window) {},
	})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.Axis != AxisTopToBottom {
		t.Error("vertical slider should use top-to-bottom axis")
	}
}

func TestRangeSliderRoundValue(t *testing.T) {
	var got float32
	onChange := func(v float32, _ *Event, _ *Window) { got = v }
	e := &Event{ScrollY: 0.7}
	rangeSliderOnMouseScroll(e, &Window{}, onChange,
		50, 0, 100, true)
	if got != float32(math.Round(50.7)) {
		t.Errorf("rounded: got %f, want %f",
			got, float32(math.Round(50.7)))
	}
}

func TestRangeSliderNonZeroMin(t *testing.T) {
	// B1: percent calc with non-zero Min
	t.Run("percent", func(t *testing.T) {
		v := RangeSlider(RangeSliderCfg{
			ID:       "rs",
			Value:    60,
			Min:      10,
			Max:      110,
			OnChange: func(float32, *Event, *Window) {},
		})
		layout := GenerateViewLayout(v, &Window{})
		track := layout.Children[0]
		leftBar := track.Children[0]
		// value=60, min=10, max=110 → 50% of track width
		want := track.Shape.Width * 0.5
		if diff := leftBar.Shape.Width - want; diff > 1 || diff < -1 {
			t.Errorf("left bar width = %f, want ~%f",
				leftBar.Shape.Width, want)
		}
	})
	// B2: mouse value with non-zero Min
	t.Run("mouse_value", func(t *testing.T) {
		var got float32
		onChange := func(v float32, _ *Event, _ *Window) { got = v }
		v := RangeSlider(RangeSliderCfg{
			ID:       "rs-nz",
			Value:    10,
			Min:      10,
			Max:      110,
			Sizing:   FixedFit,
			Width:    100,
			OnChange: onChange,
		})
		w := &Window{}
		layout := GenerateViewLayout(v, w)
		// Simulate click at 50% along the slider
		e := &Event{
			MouseX: layout.Shape.X + layout.Shape.Width/2,
			MouseY: layout.Shape.Y + layout.Shape.Height/2,
		}
		rangeSliderMouseMove(&layout, e, w,
			"rs-nz", onChange, 10, 10, 110, false, false)
		// 50% of [10,110] → 60
		if got < 59 || got > 61 {
			t.Errorf("mouse value = %f, want ~60", got)
		}
	})
}

func TestRangeSliderKeyDownHandled(t *testing.T) {
	onChange := func(float32, *Event, *Window) {}
	// Recognized key sets IsHandled
	e := &Event{KeyCode: KeyRight}
	rangeSliderOnKeyDown(nil, e, &Window{},
		onChange, 50, 0, 100, 1, false)
	if !e.IsHandled {
		t.Error("arrow key should set IsHandled")
	}
	// Unrecognized key does not set IsHandled
	e2 := &Event{KeyCode: KeyA}
	rangeSliderOnKeyDown(nil, e2, &Window{},
		onChange, 50, 0, 100, 1, false)
	if e2.IsHandled {
		t.Error("unrecognized key should not set IsHandled")
	}
}

func TestRangeSliderVerticalMouseDedup(t *testing.T) {
	callCount := 0
	onChange := func(float32, *Event, *Window) { callCount++ }
	v := RangeSlider(RangeSliderCfg{
		ID:       "rs-vd",
		Value:    50,
		Vertical: true,
		OnChange: onChange,
	})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	// Mouse position that maps to curValue=50 (50% of 100)
	e := &Event{
		MouseX: layout.Shape.X,
		MouseY: layout.Shape.Y + layout.Shape.Height/2,
	}
	rangeSliderMouseMove(&layout, e, w,
		"rs-vd", onChange, 50, 0, 100, true, true)
	if callCount != 0 {
		t.Errorf("onChange called %d times, want 0 (value unchanged)",
			callCount)
	}
}
