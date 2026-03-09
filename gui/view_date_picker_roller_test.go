package gui

import (
	"testing"
	"time"
)

func TestDatePickerRollerLayout(t *testing.T) {
	w := &Window{}
	v := DatePickerRoller(DatePickerRollerCfg{
		ID:           "roller1",
		SelectedDate: time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local),
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.ID != "roller1" {
		t.Errorf("ID = %q", layout.Shape.ID)
	}
}

func TestRollerDefaults(t *testing.T) {
	cfg := DatePickerRollerCfg{}
	applyRollerDefaults(&cfg)
	if cfg.MinYear != 1900 {
		t.Errorf("MinYear = %d", cfg.MinYear)
	}
	if cfg.MaxYear != 2100 {
		t.Errorf("MaxYear = %d", cfg.MaxYear)
	}
	if cfg.ItemHeight != 24 {
		t.Errorf("ItemHeight = %f", cfg.ItemHeight)
	}
	if cfg.VisibleItems != 3 {
		t.Errorf("VisibleItems = %d", cfg.VisibleItems)
	}
	if cfg.SelectedDate.IsZero() {
		t.Error("SelectedDate should default to now")
	}
}

func TestRollerDefaultsEvenVisible(t *testing.T) {
	cfg := DatePickerRollerCfg{VisibleItems: 4}
	applyRollerDefaults(&cfg)
	if cfg.VisibleItems != 5 {
		t.Errorf("VisibleItems = %d, want 5 (rounded up)", cfg.VisibleItems)
	}
}

func TestRollerDisplayModes(t *testing.T) {
	w := &Window{}
	modes := []DatePickerRollerDisplayMode{
		RollerDayMonthYear,
		RollerMonthDayYear,
		RollerMonthYear,
		RollerYearOnly,
	}
	for _, m := range modes {
		v := DatePickerRoller(DatePickerRollerCfg{
			ID:           "rm",
			SelectedDate: time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local),
			DisplayMode:  m,
		})
		layout := GenerateViewLayout(v, w)
		if layout.Shape.ID != "rm" {
			t.Errorf("mode %d: ID = %q", m, layout.Shape.ID)
		}
	}
}

func TestRollerDayFormat(t *testing.T) {
	if rollerDayFormat(1) != "01" {
		t.Errorf("day 1 = %q", rollerDayFormat(1))
	}
	if rollerDayFormat(15) != "15" {
		t.Errorf("day 15 = %q", rollerDayFormat(15))
	}
}

func TestRollerYearFormat(t *testing.T) {
	if rollerYearFormat(2025) != "2025" {
		t.Errorf("year = %q", rollerYearFormat(2025))
	}
}

func TestRollerMonthFormatShort(t *testing.T) {
	fn := rollerMonthFormat(false)
	if fn(1) != guiLocale.MonthsShort[0] {
		t.Errorf("short month 1 = %q", fn(1))
	}
	if fn(12) != guiLocale.MonthsShort[11] {
		t.Errorf("short month 12 = %q", fn(12))
	}
}

func TestRollerMonthFormatLong(t *testing.T) {
	fn := rollerMonthFormat(true)
	if fn(1) != guiLocale.MonthsFull[0] {
		t.Errorf("long month 1 = %q", fn(1))
	}
}

func TestRollerMonthFormatOutOfRange(t *testing.T) {
	fn := rollerMonthFormat(false)
	if fn(0) != "" {
		t.Errorf("month 0 = %q", fn(0))
	}
	if fn(13) != "" {
		t.Errorf("month 13 = %q", fn(13))
	}
}

func TestRollerAdjustDay(t *testing.T) {
	sel := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	var got time.Time
	onChange := func(d time.Time, _ *Window) { got = d }
	w := &Window{}

	rollerAdjustDay(1, sel, onChange, w)
	if got.Day() != 16 {
		t.Errorf("day+1 = %d", got.Day())
	}

	rollerAdjustDay(-1, sel, onChange, w)
	if got.Day() != 14 {
		t.Errorf("day-1 = %d", got.Day())
	}
}

func TestRollerAdjustDayNilOnChange(_ *testing.T) {
	sel := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	// Should not panic.
	rollerAdjustDay(1, sel, nil, &Window{})
}

func TestRollerAdjustMonth(t *testing.T) {
	sel := time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local)
	var got time.Time
	onChange := func(d time.Time, _ *Window) { got = d }
	w := &Window{}

	rollerAdjustMonth(1, sel, onChange, w)
	if got.Day() != 15 || got.Month() != 2 {
		t.Errorf("month+1 from Jan 15 = %v", got)
	}

	rollerAdjustMonth(-1, sel, onChange, w)
	if got.Day() != 15 || got.Month() != 12 || got.Year() != 2024 {
		t.Errorf("month-1 from Jan 15 = %v", got)
	}
}

func TestRollerAdjustYear(t *testing.T) {
	sel := time.Date(2024, 2, 29, 0, 0, 0, 0, time.Local) // leap day
	var got time.Time
	onChange := func(d time.Time, _ *Window) { got = d }
	w := &Window{}

	rollerAdjustYear(1, sel, 1900, 2100, onChange, w)
	// 2024 leap → 2025 non-leap, Feb 29 clamped to 28.
	if got.Day() != 28 || got.Month() != 2 || got.Year() != 2025 {
		t.Errorf("year+1 from leap = %v", got)
	}
}

func TestRollerAdjustYearBounds(t *testing.T) {
	sel := time.Date(2100, 6, 1, 0, 0, 0, 0, time.Local)
	called := false
	onChange := func(_ time.Time, _ *Window) { called = true }
	w := &Window{}

	rollerAdjustYear(1, sel, 1900, 2100, onChange, w)
	if called {
		t.Error("should not call onChange beyond maxYear")
	}

	sel = time.Date(1900, 6, 1, 0, 0, 0, 0, time.Local)
	called = false
	rollerAdjustYear(-1, sel, 1900, 2100, onChange, w)
	if called {
		t.Error("should not call onChange below minYear")
	}
}
