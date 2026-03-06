package gui

import (
	"strconv"
	"time"
)

// DatePickerWeekdays identifies days of the week (1=Monday..7=Sunday).
type DatePickerWeekdays uint8

const (
	DatePickerMonday    DatePickerWeekdays = 1
	DatePickerTuesday   DatePickerWeekdays = 2
	DatePickerWednesday DatePickerWeekdays = 3
	DatePickerThursday  DatePickerWeekdays = 4
	DatePickerFriday    DatePickerWeekdays = 5
	DatePickerSaturday  DatePickerWeekdays = 6
	DatePickerSunday    DatePickerWeekdays = 7
)

// DatePickerMonths identifies months (1=January..12=December).
type DatePickerMonths uint16

const (
	DatePickerJanuary   DatePickerMonths = 1
	DatePickerFebruary  DatePickerMonths = 2
	DatePickerMarch     DatePickerMonths = 3
	DatePickerApril     DatePickerMonths = 4
	DatePickerMay       DatePickerMonths = 5
	DatePickerJune      DatePickerMonths = 6
	DatePickerJuly      DatePickerMonths = 7
	DatePickerAugust    DatePickerMonths = 8
	DatePickerSeptember DatePickerMonths = 9
	DatePickerOctober   DatePickerMonths = 10
	DatePickerNovember  DatePickerMonths = 11
	DatePickerDecember  DatePickerMonths = 12
)

// DatePickerWeekdayLen controls weekday header label length.
type DatePickerWeekdayLen uint8

const (
	WeekdayOneLetter   DatePickerWeekdayLen = iota // "S"
	WeekdayThreeLetter                             // "Sun"
	WeekdayFull                                    // "Sunday"
)

// datePickerState holds per-instance state for the date picker.
type datePickerState struct {
	ShowYearMonthPicker bool
	ViewMonth           int
	ViewYear            int
}

// DatePickerCfg configures a date picker calendar view.
type DatePickerCfg struct {
	ID                   string
	Dates                []time.Time
	AllowedWeekdays      []DatePickerWeekdays
	AllowedMonths        []DatePickerMonths
	AllowedYears         []int
	AllowedDates         []time.Time
	OnSelect             func([]time.Time, *Event, *Window)
	WeekdaysLen          DatePickerWeekdayLen
	TextStyle            TextStyle
	Color                Color
	ColorHover           Color
	ColorFocus           Color
	ColorClick           Color
	ColorBorder          Color
	ColorBorderFocus     Color
	ColorSelect          Color
	Padding              Opt[Padding]
	SizeBorder           Opt[float32]
	CellSpacing          Opt[float32]
	Radius               Opt[float32]
	RadiusBorder         Opt[float32]
	IDFocus              uint32
	Disabled             bool
	Invisible            bool
	SelectMultiple       bool
	HideTodayIndicator   bool
	MondayFirstDayOfWeek bool
	ShowAdjacentMonths   bool
}

type datePickerView struct {
	cfg DatePickerCfg
}

// DatePicker creates a calendar date picker view.
func DatePicker(cfg DatePickerCfg) View {
	applyDatePickerDefaults(&cfg)
	return &datePickerView{cfg: cfg}
}

func (dv *datePickerView) Content() []View { return nil }

func (dv *datePickerView) GenerateLayout(w *Window) Layout {
	cfg := &dv.cfg
	dn := &DefaultDatePickerStyle
	sizeBorder := cfg.SizeBorder.Get(dn.SizeBorder)
	cellSpacing := cfg.CellSpacing.Get(dn.CellSpacing)
	radiusBorder := cfg.RadiusBorder.Get(dn.RadiusBorder)

	// Get/init state.
	state := datePickerGetState(w, cfg)

	// Build view tree: controls + body.
	content := make([]View, 0, 2)
	content = append(content, datePickerControls(cfg, state, w))
	if state.ShowYearMonthPicker {
		content = append(content, datePickerYearMonthPicker(cfg, state))
	} else {
		content = append(content, datePickerCalendar(cfg, state, w))
	}

	col := &containerView{
		cfg: ContainerCfg{
			ID:          cfg.ID,
			IDFocus:     cfg.IDFocus,
			Color:       cfg.Color,
			ColorBorder: cfg.ColorBorder,
			SizeBorder:  Some(sizeBorder),
			Radius:      Some(radiusBorder),
			Padding:     cfg.Padding,
			Spacing:     Some(cellSpacing),
			Disabled:    cfg.Disabled,
			Invisible:   cfg.Invisible,
			Content:     content,
			axis:        AxisTopToBottom,
			OnKeyDown: func(_ *Layout, e *Event, w *Window) {
				datePickerOnKeyDown(cfg, e, w)
			},
		},
		content:   content,
		shapeType: ShapeRectangle,
	}
	return GenerateViewLayout(col, w)
}

// datePickerGetState retrieves or initializes per-instance state.
func datePickerGetState(w *Window, cfg *DatePickerCfg) datePickerState {
	sm := StateMap[string, datePickerState](w, nsDatePicker, capModerate)
	s, ok := sm.Get(cfg.ID)
	if !ok {
		now := time.Now()
		if len(cfg.Dates) > 0 {
			now = cfg.Dates[0]
		}
		s = datePickerState{
			ViewMonth: int(now.Month()),
			ViewYear:  now.Year(),
		}
		sm.Set(cfg.ID, s)
	}
	return s
}

// DatePickerReset clears the state for a date picker instance.
func (w *Window) DatePickerReset(id string) {
	sm := StateMap[string, datePickerState](w, nsDatePicker, capModerate)
	sm.Delete(id)
	w.UpdateWindow()
}

// datePickerControls builds the header row: month/year + prev/next.
func datePickerControls(
	cfg *DatePickerCfg, state datePickerState, w *Window,
) View {
	dn := &DefaultDatePickerStyle
	cellSpacing := cfg.CellSpacing.Get(dn.CellSpacing)
	cfgID := cfg.ID
	monthLabel := LocaleFormatDate(
		datePickerViewTime(state),
		guiLocale.Date.MonthYear,
	)

	onToggle := func(_ *Layout, e *Event, w *Window) {
		sm := StateMap[string, datePickerState](w, nsDatePicker, capModerate)
		s, _ := sm.Get(cfgID)
		s.ShowYearMonthPicker = !s.ShowYearMonthPicker
		sm.Set(cfgID, s)
		w.UpdateWindow()
		e.IsHandled = true
	}

	onPrev := func(_ *Layout, e *Event, w *Window) {
		datePickerNavMonth(cfgID, -1, w)
		e.IsHandled = true
	}

	onNext := func(_ *Layout, e *Event, w *Window) {
		datePickerNavMonth(cfgID, 1, w)
		e.IsHandled = true
	}

	return Row(ContainerCfg{
		Padding: Some(PaddingSmall),
		Spacing: Some(cellSpacing),
		Content: []View{
			Button(ButtonCfg{
				Sizing:  FillFit,
				OnClick: onToggle,
				Content: []View{Text(TextCfg{
					Text: monthLabel, TextStyle: cfg.TextStyle,
				})},
			}),
			Button(ButtonCfg{
				OnClick: onPrev,
				Content: []View{Text(TextCfg{Text: "◀"})},
			}),
			Button(ButtonCfg{
				OnClick: onNext,
				Content: []View{Text(TextCfg{Text: "▶"})},
			}),
		},
	})
}

// datePickerCalendar builds weekday headers + day grid.
func datePickerCalendar(
	cfg *DatePickerCfg, state datePickerState, w *Window,
) View {
	dn := &DefaultDatePickerStyle
	cellSpacing := cfg.CellSpacing.Get(dn.CellSpacing)
	content := make([]View, 0, 7)
	content = append(content, datePickerWeekdays(cfg))
	content = append(content, datePickerMonth(cfg, state, w)...)
	return Column(ContainerCfg{
		Spacing: Some(cellSpacing),
		Padding: Some(PaddingSmall),
		Content: content,
	})
}

// datePickerWeekdays builds the weekday header row.
func datePickerWeekdays(cfg *DatePickerCfg) View {
	dn := &DefaultDatePickerStyle
	cellSpacing := cfg.CellSpacing.Get(dn.CellSpacing)
	labels := make([]View, 0, 7)
	for i := range 7 {
		dow := datePickerWeekdayIndex(i, cfg.MondayFirstDayOfWeek)
		label := datePickerWeekdayLabel(dow, cfg.WeekdaysLen)
		centeredTS := cfg.TextStyle
		centeredTS.Align = TextAlignCenter
		labels = append(labels, Row(ContainerCfg{
			Width:   40,
			HAlign:  HAlignCenter,
			Padding: Some(PaddingNone),
			Content: []View{
				Text(TextCfg{
					Text:      label,
					TextStyle: centeredTS,
					Sizing:    FillFit,
				}),
			},
		}))
	}
	return Row(ContainerCfg{
		Spacing: Some(cellSpacing),
		Padding: Some(PaddingNone),
		Content: labels,
	})
}

// datePickerMonth builds 6 rows of 7 day cells.
func datePickerMonth(
	cfg *DatePickerCfg, state datePickerState, w *Window,
) []View {
	dn := &DefaultDatePickerStyle
	sizeBorder := cfg.SizeBorder.Get(dn.SizeBorder)
	radius := cfg.Radius.Get(dn.Radius)
	cellSpacing := cfg.CellSpacing.Get(dn.CellSpacing)
	viewTime := datePickerViewTime(state)
	year, month := viewTime.Year(), viewTime.Month()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	daysInMonth := datePickerDaysInMonth(int(month), year)
	startDOW := datePickerGoWeekday(firstDay.Weekday())
	if cfg.MondayFirstDayOfWeek {
		startDOW = (startDOW + 6) % 7 // shift Sunday from 0 to 6
	}

	today := time.Now()
	onSelect := cfg.OnSelect
	selectMultiple := cfg.SelectMultiple

	rows := make([]View, 0, 6)
	day := 1 - startDOW
	for row := range 6 {
		cells := make([]View, 0, 7)
		for col := range 7 {
			d := day + row*7 + col
			if d < 1 || d > daysInMonth {
				if cfg.ShowAdjacentMonths {
					cells = append(cells, datePickerAdjacentCell(
						cfg, state, d, daysInMonth))
				} else {
					cells = append(cells, Row(ContainerCfg{
						Width:   40,
						Height:  40,
						Padding: Some(PaddingNone),
					}))
				}
				continue
			}
			cellDate := time.Date(year, month, d, 0, 0, 0, 0, time.Local)
			isToday := isSameDay(cellDate, today)
			selected := datePickerIsSelected(cellDate, cfg.Dates)
			disabled := datePickerIsDisabled(cellDate, cfg)
			dayStr := strconv.Itoa(d)

			cellColor := cfg.Color
			if selected {
				cellColor = cfg.ColorSelect
			}
			borderColor := cfg.ColorBorder
			if isToday && !cfg.HideTodayIndicator {
				borderColor = cfg.TextStyle.Color
			}

			ts := cfg.TextStyle
			if disabled {
				ts.Color = RGBA(ts.Color.R, ts.Color.G, ts.Color.B, 100)
			}

			dayVal := d
			cells = append(cells, Button(ButtonCfg{
				Width:       40,
				Height:      40,
				Color:       cellColor,
				ColorBorder: borderColor,
				SizeBorder:  Some(sizeBorder),
				Radius:      Some(radius),
				Disabled:    disabled,
				Content: []View{Text(TextCfg{
					Text: dayStr, TextStyle: ts,
				})},
				OnClick: func(_ *Layout, e *Event, w *Window) {
					dates := datePickerUpdateSelections(
						dayVal, state, cfg.Dates,
						selectMultiple)
					if onSelect != nil {
						onSelect(dates, e, w)
					}
					e.IsHandled = true
				},
			}))
		}
		rows = append(rows, Row(ContainerCfg{
			Spacing: Some(cellSpacing),
			Padding: Some(PaddingNone),
			Content: cells,
		}))
		// Stop generating rows if all days rendered.
		if day+row*7+6 >= daysInMonth {
			break
		}
	}
	return rows
}

// datePickerAdjacentCell builds a faded cell for prev/next month.
func datePickerAdjacentCell(
	cfg *DatePickerCfg, state datePickerState,
	day, daysInMonth int,
) View {
	var adjDay int
	if day < 1 {
		// Previous month.
		prevMonth := state.ViewMonth - 1
		prevYear := state.ViewYear
		if prevMonth < 1 {
			prevMonth = 12
			prevYear--
		}
		adjDay = datePickerDaysInMonth(prevMonth, prevYear) + day
	} else {
		// Next month.
		adjDay = day - daysInMonth
	}
	ts := cfg.TextStyle
	ts.Color = RGBA(ts.Color.R, ts.Color.G, ts.Color.B, 80)
	ts.Align = TextAlignCenter
	return Row(ContainerCfg{
		Width:   40,
		Height:  40,
		HAlign:  HAlignCenter,
		VAlign:  VAlignMiddle,
		Padding: Some(PaddingNone),
		Content: []View{
			Text(TextCfg{
				Text:      strconv.Itoa(adjDay),
				TextStyle: ts,
			}),
		},
	})
}

// datePickerYearMonthPicker builds a roller picker.
func datePickerYearMonthPicker(
	cfg *DatePickerCfg, state datePickerState,
) View {
	cfgID := cfg.ID
	return DatePickerRoller(DatePickerRollerCfg{
		ID:           cfg.ID + ".roller",
		SelectedDate: datePickerViewTime(state),
		DisplayMode:  RollerMonthYear,
		OnChange: func(t time.Time, w *Window) {
			sm := StateMap[string, datePickerState](
				w, nsDatePicker, capModerate)
			s, _ := sm.Get(cfgID)
			s.ViewMonth = int(t.Month())
			s.ViewYear = t.Year()
			sm.Set(cfgID, s)
			w.UpdateWindow()
		},
	})
}

// datePickerOnKeyDown handles arrow key navigation.
func datePickerOnKeyDown(cfg *DatePickerCfg, e *Event, w *Window) {
	switch e.KeyCode {
	case KeyLeft:
		datePickerNavMonth(cfg.ID, -1, w)
		e.IsHandled = true
	case KeyRight:
		datePickerNavMonth(cfg.ID, 1, w)
		e.IsHandled = true
	}
}

// datePickerNavMonth shifts the view month by delta.
func datePickerNavMonth(id string, delta int, w *Window) {
	sm := StateMap[string, datePickerState](w, nsDatePicker, capModerate)
	s, _ := sm.Get(id)
	s.ViewMonth += delta
	if s.ViewMonth > 12 {
		s.ViewMonth = 1
		s.ViewYear++
	} else if s.ViewMonth < 1 {
		s.ViewMonth = 12
		s.ViewYear--
	}
	sm.Set(id, s)
	w.UpdateWindow()
}

// datePickerUpdateSelections toggles the selected day.
func datePickerUpdateSelections(
	day int, state datePickerState,
	current []time.Time, multi bool,
) []time.Time {
	sel := time.Date(state.ViewYear, time.Month(state.ViewMonth),
		day, 0, 0, 0, 0, time.Local)
	if !multi {
		return []time.Time{sel}
	}
	// Toggle in multi-select mode.
	for i, d := range current {
		if isSameDay(d, sel) {
			return append(current[:i], current[i+1:]...)
		}
	}
	return append(current, sel)
}

// datePickerIsSelected checks if a date is in the selection.
func datePickerIsSelected(d time.Time, dates []time.Time) bool {
	for _, sel := range dates {
		if isSameDay(sel, d) {
			return true
		}
	}
	return false
}

// datePickerIsDisabled checks if a date is disallowed.
func datePickerIsDisabled(d time.Time, cfg *DatePickerCfg) bool {
	if len(cfg.AllowedDates) > 0 {
		found := false
		for _, ad := range cfg.AllowedDates {
			if isSameDay(ad, d) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	if len(cfg.AllowedWeekdays) > 0 {
		dow := datePickerGoWeekday(d.Weekday())
		// Convert to 1-based Mon=1..Sun=7.
		dpDOW := DatePickerWeekdays((dow+6)%7 + 1)
		found := false
		for _, aw := range cfg.AllowedWeekdays {
			if aw == dpDOW {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	if len(cfg.AllowedMonths) > 0 {
		m := DatePickerMonths(d.Month())
		found := false
		for _, am := range cfg.AllowedMonths {
			if am == m {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	if len(cfg.AllowedYears) > 0 {
		y := d.Year()
		found := false
		for _, ay := range cfg.AllowedYears {
			if ay == y {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	return false
}

// isSameDay compares two times ignoring time-of-day.
func isSameDay(a, b time.Time) bool {
	return a.Year() == b.Year() &&
		a.Month() == b.Month() &&
		a.Day() == b.Day()
}

// datePickerViewTime returns a time for the current view month/year.
func datePickerViewTime(state datePickerState) time.Time {
	return time.Date(state.ViewYear, time.Month(state.ViewMonth),
		1, 0, 0, 0, 0, time.Local)
}

// datePickerGoWeekday converts Go's time.Weekday (Sun=0) to
// 0=Sunday..6=Saturday.
func datePickerGoWeekday(wd time.Weekday) int {
	return int(wd)
}

// datePickerWeekdayIndex returns the weekday index for column i.
func datePickerWeekdayIndex(i int, mondayFirst bool) int {
	if mondayFirst {
		return (i + 1) % 7 // Mon=1,Tue=2,...,Sun=0
	}
	return i // Sun=0,Mon=1,...,Sat=6
}

// datePickerWeekdayLabel returns the locale weekday label.
func datePickerWeekdayLabel(dow int, wdLen DatePickerWeekdayLen) string {
	switch wdLen {
	case WeekdayThreeLetter:
		return guiLocale.WeekdaysMed[dow]
	case WeekdayFull:
		return guiLocale.WeekdaysFull[dow]
	default:
		return guiLocale.WeekdaysShort[dow]
	}
}

// datePickerDaysInMonth returns the number of days in a month.
func datePickerDaysInMonth(month, year int) int {
	t := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local)
	return t.Day()
}

func applyDatePickerDefaults(cfg *DatePickerCfg) {
	d := &DefaultDatePickerStyle
	if !cfg.Color.IsSet() {
		cfg.Color = d.Color
	}
	if !cfg.ColorHover.IsSet() {
		cfg.ColorHover = d.ColorHover
	}
	if !cfg.ColorFocus.IsSet() {
		cfg.ColorFocus = d.ColorFocus
	}
	if !cfg.ColorClick.IsSet() {
		cfg.ColorClick = d.ColorClick
	}
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = d.ColorBorder
	}
	if !cfg.ColorBorderFocus.IsSet() {
		cfg.ColorBorderFocus = d.ColorBorderFocus
	}
	if !cfg.ColorSelect.IsSet() {
		cfg.ColorSelect = d.ColorSelect
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(d.Padding)
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = d.TextStyle
	}
	if !cfg.CellSpacing.IsSet() {
		cfg.CellSpacing = Some(d.CellSpacing)
	}
	if !cfg.Radius.IsSet() {
		cfg.Radius = Some(d.Radius)
	}
}
