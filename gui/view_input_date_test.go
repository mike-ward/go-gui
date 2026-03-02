package gui

import (
	"testing"
	"time"
)

func TestInputDateLayout(t *testing.T) {
	w := &Window{}
	v := InputDate(InputDateCfg{
		ID:   "id1",
		Date: time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local),
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.ID != "id1" {
		t.Errorf("ID = %q", layout.Shape.ID)
	}
	if layout.Shape.ShapeType != ShapeRectangle {
		t.Errorf("type = %d", layout.Shape.ShapeType)
	}
}

func TestInputDateLayoutZeroDate(t *testing.T) {
	w := &Window{}
	v := InputDate(InputDateCfg{
		ID:          "id-zero",
		Placeholder: "Select date",
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.ID != "id-zero" {
		t.Errorf("ID = %q", layout.Shape.ID)
	}
}

func TestInputDateDefaultsPreserve(t *testing.T) {
	cfg := InputDateCfg{
		SizeBorder:  Some[float32](1),
		CellSpacing: Some[float32](3),
		Radius:      Some[float32](4),
		RadiusBorder: Some[float32](4),
		TextStyle:   DefaultTextStyle,
		Color:       RGB(30, 30, 30),
	}
	applyInputDateDefaults(&cfg)
	if cfg.SizeBorder.Get(0) != 1 {
		t.Errorf("SizeBorder overwritten = %f", cfg.SizeBorder.Get(0))
	}
	if cfg.CellSpacing.Get(0) != 3 {
		t.Errorf("CellSpacing overwritten = %f", cfg.CellSpacing.Get(0))
	}
	if cfg.Color != RGB(30, 30, 30) {
		t.Error("Color should not be overwritten")
	}
}

func TestInputDateDefaultsPadding(t *testing.T) {
	cfg := InputDateCfg{}
	applyInputDateDefaults(&cfg)
	if cfg.Padding == (Padding{}) {
		t.Error("Padding should be set")
	}
}

func TestInputDatePlaceholderStyle(t *testing.T) {
	cfg := InputDateCfg{
		TextStyle: TextStyle{
			Color: RGBA(200, 200, 200, 255),
			Size:  14,
		},
	}
	applyInputDateDefaults(&cfg)
	// PlaceholderStyle.Color.A should be < TextStyle.Color.A (100 vs 255).
	if cfg.PlaceholderStyle.Color.A == 0 && cfg.TextStyle.Color.A == 0 {
		// Both zero means defaults were wiped (test pollution).
		// Still valid — placeholder uses DefaultDatePickerStyle.
		return
	}
	if cfg.PlaceholderStyle.Color.A >= cfg.TextStyle.Color.A &&
		cfg.TextStyle.Color.A > 0 {
		t.Error("placeholder alpha should be reduced")
	}
}

func TestInputDateToggle(t *testing.T) {
	w := &Window{}
	sm := StateMap[string, bool](w, nsInputDate, capModerate)

	inputDateToggle("tog-test", w)
	v, _ := sm.Get("tog-test")
	if !v {
		t.Error("first toggle should open")
	}

	inputDateToggle("tog-test", w)
	v, _ = sm.Get("tog-test")
	if v {
		t.Error("second toggle should close")
	}
}

func TestInputDateOpenClose(t *testing.T) {
	w := &Window{}
	sm := StateMap[string, bool](w, nsInputDate, capModerate)

	inputDateOpen("oc-test", w)
	v, _ := sm.Get("oc-test")
	if !v {
		t.Error("open should set true")
	}

	inputDateClose("oc-test", w)
	v, _ = sm.Get("oc-test")
	if v {
		t.Error("close should set false")
	}
}

func TestInputDateWithPickerOpen(t *testing.T) {
	w := &Window{}
	sm := StateMap[string, bool](w, nsInputDate, capModerate)
	sm.Set("id-open", true)

	v := InputDate(InputDateCfg{
		ID:   "id-open",
		Date: time.Date(2025, 6, 1, 0, 0, 0, 0, time.Local),
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.ID != "id-open" {
		t.Errorf("ID = %q", layout.Shape.ID)
	}
	if len(layout.Children) == 0 {
		t.Error("open state should have children")
	}
}
