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
