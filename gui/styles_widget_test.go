package gui

import "testing"

func TestDefaultInputStyleColors(t *testing.T) {
	s := DefaultInputStyle
	if s.Color.Eq(Color{}) {
		t.Error("input color should not be zero")
	}
	if s.Radius != RadiusMedium {
		t.Errorf("radius = %f, want %f", s.Radius, RadiusMedium)
	}
}

func TestDefaultScrollbarStyle(t *testing.T) {
	s := DefaultScrollbarStyle
	if s.Size != 7 {
		t.Errorf("scrollbar size = %f", s.Size)
	}
	if s.MinThumbSize != 20 {
		t.Errorf("min thumb = %f", s.MinThumbSize)
	}
}

func TestDefaultWidgetStylesNonZero(t *testing.T) {
	// Verify all widget default styles have non-zero radius
	// (except scrollbar which uses RadiusSmall).
	styles := []struct {
		name   string
		radius float32
	}{
		{"radio", DefaultRadioStyle.SizeBorder},
		{"switch", DefaultSwitchStyle.SizeBorder},
		{"toggle", DefaultToggleStyle.SizeBorder},
		{"select", DefaultSelectStyle.SizeBorder},
		{"listbox", DefaultListBoxStyle.SizeBorder},
	}
	for _, s := range styles {
		if s.radius == 0 {
			t.Errorf("%s has zero size_border", s.name)
		}
	}
}
