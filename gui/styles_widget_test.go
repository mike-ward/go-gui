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

func TestDefaultDialogStyle(t *testing.T) {
	s := DefaultDialogStyle
	if s.Color.Eq(Color{}) {
		t.Error("dialog color should not be zero")
	}
	if s.Radius != RadiusMedium {
		t.Errorf("radius = %f, want %f", s.Radius, RadiusMedium)
	}
	if s.AlignButtons != HAlignCenter {
		t.Errorf("align = %d, want center", s.AlignButtons)
	}
}

func TestDefaultToastStyle(t *testing.T) {
	s := DefaultToastStyle
	if s.MaxVisible != 5 {
		t.Errorf("max_visible = %d, want 5", s.MaxVisible)
	}
	if s.Width != 260 {
		t.Errorf("width = %f, want 260", s.Width)
	}
	if s.Anchor != ToastBottomRight {
		t.Errorf("anchor = %d, want bottom-right", s.Anchor)
	}
}

func TestDefaultTooltipStyle(t *testing.T) {
	s := DefaultTooltipStyle
	if s.Delay == 0 {
		t.Error("delay should not be zero")
	}
	if s.Radius != RadiusSmall {
		t.Errorf("radius = %f, want %f", s.Radius, RadiusSmall)
	}
}

func TestDefaultTreeStyle(t *testing.T) {
	s := DefaultTreeStyle
	if !s.ColorHover.IsSet() {
		t.Error("tree hover color should be set")
	}
	if s.Radius != RadiusMedium {
		t.Errorf("radius = %f, want %f", s.Radius, RadiusMedium)
	}
	if s.TextStyleIcon.Family != IconFontName {
		t.Errorf("icon family = %q, want %q", s.TextStyleIcon.Family, IconFontName)
	}
}
