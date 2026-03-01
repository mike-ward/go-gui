package gui

import (
	"testing"
	"time"
)

func TestDatePickerLayout(t *testing.T) {
	w := &Window{}
	v := DatePicker(DatePickerCfg{
		ID:    "dp1",
		Dates: []time.Time{time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)},
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.ID != "dp1" {
		t.Errorf("ID = %q", layout.Shape.ID)
	}
	if layout.Shape.ShapeType != ShapeRectangle {
		t.Errorf("type = %d", layout.Shape.ShapeType)
	}
}

func TestDatePickerStateInit(t *testing.T) {
	w := &Window{}
	d := time.Date(2025, 6, 10, 0, 0, 0, 0, time.Local)
	cfg := DatePickerCfg{ID: "dp-state", Dates: []time.Time{d}}
	applyDatePickerDefaults(&cfg)
	state := datePickerGetState(w, &cfg)
	if state.ViewMonth != 6 {
		t.Errorf("month = %d, want 6", state.ViewMonth)
	}
	if state.ViewYear != 2025 {
		t.Errorf("year = %d, want 2025", state.ViewYear)
	}
}

func TestDatePickerNavMonth(t *testing.T) {
	w := &Window{}
	sm := StateMap[string, datePickerState](w, nsDatePicker, capModerate)
	sm.Set("nav-test", datePickerState{ViewMonth: 1, ViewYear: 2025})

	datePickerNavMonth("nav-test", -1, w)
	s, _ := sm.Get("nav-test")
	if s.ViewMonth != 12 || s.ViewYear != 2024 {
		t.Errorf("prev = %d/%d", s.ViewMonth, s.ViewYear)
	}

	datePickerNavMonth("nav-test", 1, w)
	s, _ = sm.Get("nav-test")
	if s.ViewMonth != 1 || s.ViewYear != 2025 {
		t.Errorf("next = %d/%d", s.ViewMonth, s.ViewYear)
	}
}

func TestDatePickerDaysInMonth(t *testing.T) {
	tests := []struct {
		month, year, want int
	}{
		{1, 2025, 31},
		{2, 2024, 29}, // leap year
		{2, 2025, 28},
		{4, 2025, 30},
		{12, 2025, 31},
	}
	for _, tt := range tests {
		got := datePickerDaysInMonth(tt.month, tt.year)
		if got != tt.want {
			t.Errorf("daysInMonth(%d, %d) = %d, want %d",
				tt.month, tt.year, got, tt.want)
		}
	}
}

func TestIsSameDay(t *testing.T) {
	a := time.Date(2025, 3, 15, 10, 30, 0, 0, time.Local)
	b := time.Date(2025, 3, 15, 23, 59, 0, 0, time.Local)
	c := time.Date(2025, 3, 16, 0, 0, 0, 0, time.Local)

	if !isSameDay(a, b) {
		t.Error("same day should match")
	}
	if isSameDay(a, c) {
		t.Error("different days should not match")
	}
}

func TestDatePickerIsSelected(t *testing.T) {
	d1 := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	d2 := time.Date(2025, 3, 16, 0, 0, 0, 0, time.Local)
	dates := []time.Time{d1}

	if !datePickerIsSelected(d1, dates) {
		t.Error("d1 should be selected")
	}
	if datePickerIsSelected(d2, dates) {
		t.Error("d2 should not be selected")
	}
}

func TestDatePickerIsDisabledWeekday(t *testing.T) {
	mon := time.Date(2025, 3, 17, 0, 0, 0, 0, time.Local) // Monday
	cfg := DatePickerCfg{
		AllowedWeekdays: []DatePickerWeekdays{DatePickerTuesday},
	}
	if !datePickerIsDisabled(mon, &cfg) {
		t.Error("Monday should be disabled")
	}
	tue := time.Date(2025, 3, 18, 0, 0, 0, 0, time.Local) // Tuesday
	if datePickerIsDisabled(tue, &cfg) {
		t.Error("Tuesday should not be disabled")
	}
}

func TestDatePickerIsDisabledMonth(t *testing.T) {
	mar := time.Date(2025, 3, 1, 0, 0, 0, 0, time.Local)
	cfg := DatePickerCfg{
		AllowedMonths: []DatePickerMonths{DatePickerJune},
	}
	if !datePickerIsDisabled(mar, &cfg) {
		t.Error("March should be disabled")
	}
}

func TestDatePickerIsDisabledYear(t *testing.T) {
	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	cfg := DatePickerCfg{
		AllowedYears: []int{2025, 2026},
	}
	if !datePickerIsDisabled(d, &cfg) {
		t.Error("2024 should be disabled")
	}
}

func TestDatePickerIsDisabledDates(t *testing.T) {
	allowed := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	other := time.Date(2025, 3, 16, 0, 0, 0, 0, time.Local)
	cfg := DatePickerCfg{AllowedDates: []time.Time{allowed}}
	if datePickerIsDisabled(allowed, &cfg) {
		t.Error("allowed date should not be disabled")
	}
	if !datePickerIsDisabled(other, &cfg) {
		t.Error("non-allowed date should be disabled")
	}
}

func TestDatePickerUpdateSelections(t *testing.T) {
	state := datePickerState{ViewMonth: 3, ViewYear: 2025}

	// Single select.
	dates := datePickerUpdateSelections(15, state, nil, false)
	if len(dates) != 1 || dates[0].Day() != 15 {
		t.Errorf("single = %v", dates)
	}

	// Multi select — add.
	dates = datePickerUpdateSelections(16, state, dates, true)
	if len(dates) != 2 {
		t.Errorf("multi add = %v", dates)
	}

	// Multi select — toggle off.
	dates = datePickerUpdateSelections(15, state, dates, true)
	if len(dates) != 1 || dates[0].Day() != 16 {
		t.Errorf("multi toggle = %v", dates)
	}
}

func TestDatePickerWeekdayIndex(t *testing.T) {
	// Sunday first: col 0 = Sunday.
	if datePickerWeekdayIndex(0, false) != 0 {
		t.Error("Sunday first col 0")
	}
	// Monday first: col 0 = Monday.
	if datePickerWeekdayIndex(0, true) != 1 {
		t.Error("Monday first col 0")
	}
	// Monday first: col 6 = Sunday.
	if datePickerWeekdayIndex(6, true) != 0 {
		t.Error("Monday first col 6")
	}
}

func TestDatePickerWeekdayLabel(t *testing.T) {
	lbl := datePickerWeekdayLabel(0, WeekdayOneLetter)
	if lbl != "S" {
		t.Errorf("one letter Sunday = %q", lbl)
	}
	lbl = datePickerWeekdayLabel(1, WeekdayThreeLetter)
	if lbl != "Mon" {
		t.Errorf("three letter Monday = %q", lbl)
	}
	lbl = datePickerWeekdayLabel(2, WeekdayFull)
	if lbl != "Tuesday" {
		t.Errorf("full Tuesday = %q", lbl)
	}
}

func TestDatePickerDefaults(t *testing.T) {
	cfg := DatePickerCfg{}
	applyDatePickerDefaults(&cfg)
	if cfg.CellSpacing != 3 {
		t.Errorf("spacing = %f", cfg.CellSpacing)
	}
	if cfg.Radius == 0 {
		t.Error("radius should be set")
	}
}

func TestDatePickerReset(t *testing.T) {
	w := &Window{}
	sm := StateMap[string, datePickerState](w, nsDatePicker, capModerate)
	sm.Set("reset-test", datePickerState{ViewMonth: 6, ViewYear: 2025})

	w.DatePickerReset("reset-test")
	_, ok := sm.Get("reset-test")
	if ok {
		t.Error("state should be cleared")
	}
}
