package gui

import (
	"fmt"
	"time"
)

// DatePickerRollerDisplayMode controls which drums are shown.
type DatePickerRollerDisplayMode uint8

const (
	RollerDayMonthYear DatePickerRollerDisplayMode = iota // DD MMM YYYY
	RollerMonthDayYear                                    // MMM DD YYYY
	RollerMonthYear                                       // MMM YYYY
	RollerYearOnly                                        // YYYY
)

// DatePickerRollerCfg configures a roller-style date picker.
type DatePickerRollerCfg struct {
	ID           string
	IDFocus      uint32
	SelectedDate time.Time
	DisplayMode  DatePickerRollerDisplayMode
	MinYear      int
	MaxYear      int
	ItemHeight   float32
	VisibleItems int // must be odd
	MinWidth     float32
	MaxWidth     float32
	LongMonths   bool // true = "January", false = "Jan"
	Color        Color
	TextStyle    TextStyle
	OnChange     func(time.Time, *Window)
}

type datePickerRollerView struct {
	cfg DatePickerRollerCfg
}

// DatePickerRoller creates a roller-style date picker view.
func DatePickerRoller(cfg DatePickerRollerCfg) View {
	applyRollerDefaults(&cfg)
	return &datePickerRollerView{cfg: cfg}
}

func (rv *datePickerRollerView) Content() []View { return nil }

func (rv *datePickerRollerView) GenerateLayout(w *Window) Layout {
	cfg := &rv.cfg
	sel := cfg.SelectedDate

	var drums []View
	switch cfg.DisplayMode {
	case RollerDayMonthYear:
		drums = append(drums,
			rollerDrum(cfg, "day", sel.Day(), 1,
				datePickerDaysInMonth(int(sel.Month()), sel.Year()),
				rollerDayFormat, 50),
			rollerDrum(cfg, "month", int(sel.Month()), 1, 12,
				rollerMonthFormat(cfg.LongMonths), 80),
			rollerDrum(cfg, "year", sel.Year(),
				cfg.MinYear, cfg.MaxYear, rollerYearFormat, 60),
		)
	case RollerMonthDayYear:
		drums = append(drums,
			rollerDrum(cfg, "month", int(sel.Month()), 1, 12,
				rollerMonthFormat(cfg.LongMonths), 80),
			rollerDrum(cfg, "day", sel.Day(), 1,
				datePickerDaysInMonth(int(sel.Month()), sel.Year()),
				rollerDayFormat, 50),
			rollerDrum(cfg, "year", sel.Year(),
				cfg.MinYear, cfg.MaxYear, rollerYearFormat, 60),
		)
	case RollerMonthYear:
		drums = append(drums,
			rollerDrum(cfg, "month", int(sel.Month()), 1, 12,
				rollerMonthFormat(cfg.LongMonths), 80),
			rollerDrum(cfg, "year", sel.Year(),
				cfg.MinYear, cfg.MaxYear, rollerYearFormat, 60),
		)
	case RollerYearOnly:
		drums = append(drums,
			rollerDrum(cfg, "year", sel.Year(),
				cfg.MinYear, cfg.MaxYear, rollerYearFormat, 60),
		)
	}

	onChange := cfg.OnChange
	minYear := cfg.MinYear
	maxYear := cfg.MaxYear
	selectedDate := cfg.SelectedDate

	return GenerateViewLayout(container(ContainerCfg{
		ID:       cfg.ID,
		IDFocus:  cfg.IDFocus,
		Color:    cfg.Color,
		MinWidth: cfg.MinWidth,
		MaxWidth: cfg.MaxWidth,
		Padding:  PaddingSmall,
		Spacing:  Some(SpacingSmall),
		HAlign:   HAlignCenter,
		VAlign:   VAlignMiddle,
		axis:     AxisLeftToRight,
		OnKeyDown: func(_ *Layout, e *Event, w *Window) {
			rollerOnKeyDown(onChange, selectedDate,
				minYear, maxYear, e, w)
		},
		Content: drums,
	}), w)
}

// rollerDrum builds a single drum column showing visibleItems.
func rollerDrum(
	cfg *DatePickerRollerCfg, name string,
	value, minVal, maxVal int,
	format func(int) string, drumWidth float32,
) View {
	vis := cfg.VisibleItems
	half := vis / 2
	ts := cfg.TextStyle

	var items []View
	for i := range vis {
		offset := i - half
		v := value + offset
		// Wrap within range.
		span := maxVal - minVal + 1
		for v < minVal {
			v += span
		}
		for v > maxVal {
			v -= span
		}

		label := format(v)
		itemTS := ts
		if offset == 0 {
			itemTS.Size = ts.Size + 4
		} else {
			dist := offset
			if dist < 0 {
				dist = -dist
			}
			alpha := uint8(150)
			if dist > 1 {
				alpha = 80
			}
			itemTS.Color = RGBA(
				ts.Color.R, ts.Color.G, ts.Color.B, alpha)
		}

		centeredTS := itemTS
		centeredTS.Align = TextAlignCenter
		items = append(items, Row(ContainerCfg{
			Width:  drumWidth,
			Height: cfg.ItemHeight,
			HAlign: HAlignCenter,
			VAlign: VAlignMiddle,
			Content: []View{
				Text(TextCfg{
					Text:      label,
					TextStyle: centeredTS,
					Sizing:    FillFit,
				}),
			},
		}))
	}

	onChange := cfg.OnChange
	selectedDate := cfg.SelectedDate
	minYear := cfg.MinYear
	maxYear := cfg.MaxYear

	return Column(ContainerCfg{
		Width:   drumWidth,
		Padding: PaddingNone,
		Content: items,
		OnScroll: func(_ *Layout, w *Window) {
			// Scroll dispatches via the view's scroll events,
			// so keyboard is used instead for simplicity.
		},
		OnKeyDown: func(_ *Layout, e *Event, w *Window) {
			switch name {
			case "day":
				if e.KeyCode == KeyUp {
					rollerAdjustDay(-1, selectedDate, onChange, w)
					e.IsHandled = true
				} else if e.KeyCode == KeyDown {
					rollerAdjustDay(1, selectedDate, onChange, w)
					e.IsHandled = true
				}
			case "month":
				if e.KeyCode == KeyUp {
					rollerAdjustMonth(-1, selectedDate, onChange, w)
					e.IsHandled = true
				} else if e.KeyCode == KeyDown {
					rollerAdjustMonth(1, selectedDate, onChange, w)
					e.IsHandled = true
				}
			case "year":
				if e.KeyCode == KeyUp {
					rollerAdjustYear(-1, selectedDate, minYear, maxYear,
						onChange, w)
					e.IsHandled = true
				} else if e.KeyCode == KeyDown {
					rollerAdjustYear(1, selectedDate, minYear, maxYear,
						onChange, w)
					e.IsHandled = true
				}
			}
		},
	})
}

// rollerOnKeyDown handles keyboard navigation for the roller.
func rollerOnKeyDown(
	onChange func(time.Time, *Window),
	sel time.Time, minYear, maxYear int,
	e *Event, w *Window,
) {
	if onChange == nil {
		return
	}
	switch {
	case e.Modifiers == ModShift && e.KeyCode == KeyUp:
		rollerAdjustDay(-1, sel, onChange, w)
		e.IsHandled = true
	case e.Modifiers == ModShift && e.KeyCode == KeyDown:
		rollerAdjustDay(1, sel, onChange, w)
		e.IsHandled = true
	case e.Modifiers == ModAlt && e.KeyCode == KeyUp:
		rollerAdjustMonth(-1, sel, onChange, w)
		e.IsHandled = true
	case e.Modifiers == ModAlt && e.KeyCode == KeyDown:
		rollerAdjustMonth(1, sel, onChange, w)
		e.IsHandled = true
	case e.Modifiers == ModNone && e.KeyCode == KeyUp:
		rollerAdjustYear(-1, sel, minYear, maxYear, onChange, w)
		e.IsHandled = true
	case e.Modifiers == ModNone && e.KeyCode == KeyDown:
		rollerAdjustYear(1, sel, minYear, maxYear, onChange, w)
		e.IsHandled = true
	}
}

func rollerAdjustDay(
	delta int, sel time.Time,
	onChange func(time.Time, *Window), w *Window,
) {
	if onChange == nil {
		return
	}
	newDate := sel.AddDate(0, 0, delta)
	onChange(newDate, w)
}

func rollerAdjustMonth(
	delta int, sel time.Time,
	onChange func(time.Time, *Window), w *Window,
) {
	if onChange == nil {
		return
	}
	newDate := sel.AddDate(0, delta, 0)
	// Clamp day to new month's max.
	dim := datePickerDaysInMonth(int(newDate.Month()), newDate.Year())
	if newDate.Day() > dim {
		newDate = time.Date(newDate.Year(), newDate.Month(), dim,
			0, 0, 0, 0, time.Local)
	}
	onChange(newDate, w)
}

func rollerAdjustYear(
	delta int, sel time.Time,
	minYear, maxYear int,
	onChange func(time.Time, *Window), w *Window,
) {
	if onChange == nil {
		return
	}
	newYear := sel.Year() + delta
	if newYear < minYear || newYear > maxYear {
		return
	}
	dim := datePickerDaysInMonth(int(sel.Month()), newYear)
	day := sel.Day()
	if day > dim {
		day = dim
	}
	newDate := time.Date(newYear, sel.Month(), day,
		0, 0, 0, 0, time.Local)
	onChange(newDate, w)
}

func rollerDayFormat(v int) string   { return fmt.Sprintf("%02d", v) }
func rollerYearFormat(v int) string  { return fmt.Sprintf("%d", v) }

func rollerMonthFormat(long bool) func(int) string {
	return func(v int) string {
		idx := v - 1
		if idx < 0 || idx >= 12 {
			return ""
		}
		if long {
			return guiLocale.MonthsFull[idx]
		}
		return guiLocale.MonthsShort[idx]
	}
}

func applyRollerDefaults(cfg *DatePickerRollerCfg) {
	if cfg.MinYear == 0 {
		cfg.MinYear = 1900
	}
	if cfg.MaxYear == 0 {
		cfg.MaxYear = 2100
	}
	if cfg.ItemHeight == 0 {
		cfg.ItemHeight = 32
	}
	if cfg.VisibleItems == 0 {
		cfg.VisibleItems = 5
	}
	if cfg.VisibleItems%2 == 0 {
		cfg.VisibleItems++
	}
	if cfg.Color == (Color{}) {
		cfg.Color = guiTheme.ColorBackground
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
	if cfg.SelectedDate.IsZero() {
		cfg.SelectedDate = time.Now()
	}
}
