package gui

import (
	"fmt"
	"strconv"
	"time"
)

// DatePickerRollerDisplayMode controls which drums are shown.
type DatePickerRollerDisplayMode uint8

// DatePickerRollerDisplayMode constants.
const (
	RollerDayMonthYear DatePickerRollerDisplayMode = iota // DD MMM YYYY
	RollerMonthDayYear                                    // MMM DD YYYY
	RollerMonthYear                                       // MMM YYYY
	RollerYearOnly                                        // YYYY
)

// DatePickerRollerCfg configures a roller-style date picker.
type DatePickerRollerCfg struct {
	ID              string
	A11YLabel       string
	A11YDescription string
	IDFocus         uint32
	SelectedDate    time.Time
	DisplayMode     DatePickerRollerDisplayMode
	MinYear         int
	MaxYear         int
	ItemHeight      float32
	VisibleItems    int // must be odd
	MinWidth        float32
	MaxWidth        float32
	WidthDay        float32
	WidthMonth      float32
	WidthYear       float32
	LongMonths      bool // true = "January", false = "Jan"
	WrapYear        bool
	Color           Color
	ColorBorder     Color
	ColorBorderFocus Color
	SizeBorder      Opt[float32]
	Radius          Opt[float32]
	Padding         Opt[Padding]
	TextStyle       TextStyle
	OnChange        func(time.Time, *Window)
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
	var drumNames []string
	switch cfg.DisplayMode {
	case RollerDayMonthYear:
		drumNames = []string{"day", "month", "year"}
		drums = append(drums,
			rollerDrum(cfg, "day", sel.Day(), 1,
				datePickerDaysInMonth(int(sel.Month()), sel.Year()),
				rollerDayFormat, cfg.WidthDay, true),
			rollerDrum(cfg, "month", int(sel.Month()), 1, 12,
				rollerMonthFormat(cfg.LongMonths), cfg.WidthMonth, true),
			rollerDrum(cfg, "year", sel.Year(),
				cfg.MinYear, cfg.MaxYear, rollerYearFormat, cfg.WidthYear, cfg.WrapYear),
		)
	case RollerMonthDayYear:
		drumNames = []string{"month", "day", "year"}
		drums = append(drums,
			rollerDrum(cfg, "month", int(sel.Month()), 1, 12,
				rollerMonthFormat(cfg.LongMonths), cfg.WidthMonth, true),
			rollerDrum(cfg, "day", sel.Day(), 1,
				datePickerDaysInMonth(int(sel.Month()), sel.Year()),
				rollerDayFormat, cfg.WidthDay, true),
			rollerDrum(cfg, "year", sel.Year(),
				cfg.MinYear, cfg.MaxYear, rollerYearFormat, cfg.WidthYear, cfg.WrapYear),
		)
	case RollerMonthYear:
		drumNames = []string{"month", "year"}
		drums = append(drums,
			rollerDrum(cfg, "month", int(sel.Month()), 1, 12,
				rollerMonthFormat(cfg.LongMonths), cfg.WidthMonth, true),
			rollerDrum(cfg, "year", sel.Year(),
				cfg.MinYear, cfg.MaxYear, rollerYearFormat, cfg.WidthYear, cfg.WrapYear),
		)
	case RollerYearOnly:
		drumNames = []string{"year"}
		drums = append(drums,
			rollerDrum(cfg, "year", sel.Year(),
				cfg.MinYear, cfg.MaxYear, rollerYearFormat, cfg.WidthYear, cfg.WrapYear),
		)
	}

	onChange := cfg.OnChange
	minYear := cfg.MinYear
	maxYear := cfg.MaxYear
	selectedDate := cfg.SelectedDate
	mode := cfg.DisplayMode
	wrapYear := cfg.WrapYear

	return GenerateViewLayout(container(ContainerCfg{
		ID:          cfg.ID,
		IDFocus:     cfg.IDFocus,
		A11YRole:    AccessRoleDateField,
		A11YLabel:   a11yLabel(cfg.A11YLabel, "Date Roller"),
		Color:       cfg.Color,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  cfg.SizeBorder,
		Radius:      cfg.Radius,
		MinWidth:    cfg.MinWidth,
		MaxWidth:    cfg.MaxWidth,
		Padding:     cfg.Padding,
		Spacing:     Some(SpacingSmall),
		HAlign:    HAlignCenter,
		VAlign:    VAlignMiddle,
		axis:      AxisLeftToRight,
		OnKeyDown: func(_ *Layout, e *Event, w *Window) {
			rollerOnKeyDown(onChange, selectedDate,
				minYear, maxYear, e, w, mode, wrapYear)
		},
		OnClick: func(_ *Layout, e *Event, w *Window) {
			if cfg.IDFocus > 0 {
				w.SetIDFocus(cfg.IDFocus)
				e.IsHandled = true
			}
		},
		AmendLayout: func(lo *Layout, w *Window) {
			if cfg.IDFocus > 0 && w.IsFocus(cfg.IDFocus) {
				lo.Shape.ColorBorder = cfg.ColorBorderFocus
			}
			lo.Shape.Events.OnMouseScroll = func(
				_ *Layout, e *Event, w *Window,
			) {
				e.IsHandled = true
				if onChange == nil {
					return
				}
				delta := 1
				if e.ScrollY > 0 {
					delta = -1
				}
				for i, child := range lo.Children {
					if i < len(drumNames) &&
						child.Shape.PointInShape(
							e.MouseX, e.MouseY) {
						rollerDrumAdjust(drumNames[i],
							delta, selectedDate,
							minYear, maxYear, onChange, w, wrapYear)
						return
					}
				}
			}
		},
		Content: drums,
	}), w)
}

// rollerDrum builds a single drum column showing visibleItems.
func rollerDrum(
	cfg *DatePickerRollerCfg, name string,
	value, minVal, maxVal int,
	format func(int) string, drumWidth float32,
	wrap bool,
) View {
	vis := cfg.VisibleItems
	half := vis / 2
	ts := cfg.TextStyle
	onChange := cfg.OnChange
	selectedDate := cfg.SelectedDate
	minYear := cfg.MinYear
	maxYear := cfg.MaxYear

	var items []View
	for i := range vis {
		offset := i - half
		v := value + offset

		if wrap {
			// Wrap within range.
			span := maxVal - minVal + 1
			for v < minVal {
				v += span
			}
			for v > maxVal {
				v -= span
			}
		}

		label := format(v)
		if v < minVal || v > maxVal {
			label = " "
		}

		itemTS := ts
		if offset == 0 {
			itemTS.Size = ts.Size + 2
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
		itemCfg := ContainerCfg{
			Width:   drumWidth,
			Height:  cfg.ItemHeight,
			Padding: NoPadding,
			HAlign:  HAlignCenter,
			VAlign:  VAlignMiddle,
			Content: []View{
				Text(TextCfg{
					Text:      label,
					TextStyle: centeredTS,
					Sizing:    FillFit,
				}),
			},
		}
		itemClickable := offset != 0 && (v >= minVal && v <= maxVal || wrap)
		if itemClickable {
			clickDelta := offset
			itemCfg.OnClick = func(
				_ *Layout, _ *Event, w *Window,
			) {
				rollerDrumAdjust(name, clickDelta,
					selectedDate, minYear, maxYear,
					onChange, w, wrap)
			}
		}
		items = append(items, Row(itemCfg))
	}

	return Column(ContainerCfg{
		Width:   drumWidth,
		Padding: NoPadding,
		Content: items,
	})
}

func rollerDrumAdjust(
	name string, delta int, sel time.Time,
	minYear, maxYear int,
	onChange func(time.Time, *Window), w *Window,
	wrapYear bool,
) {
	switch name {
	case "day":
		rollerAdjustDay(delta, sel, minYear, maxYear, onChange, w)
	case "month":
		rollerAdjustMonth(delta, sel, minYear, maxYear, onChange, w)
	case "year":
		rollerAdjustYear(delta, sel, minYear, maxYear, onChange, w, wrapYear)
	}
}

// rollerOnKeyDown handles keyboard navigation for the roller.
func rollerOnKeyDown(
	onChange func(time.Time, *Window),
	sel time.Time, minYear, maxYear int,
	e *Event, w *Window,
	mode DatePickerRollerDisplayMode,
	wrapYear bool,
) {
	if onChange == nil {
		return
	}
	switch {
	case e.Modifiers == ModNone && e.KeyCode == KeyUp:
		switch mode {
		case RollerYearOnly:
			rollerAdjustYear(-1, sel, minYear, maxYear, onChange, w, wrapYear)
		case RollerMonthYear:
			rollerAdjustMonth(-1, sel, minYear, maxYear, onChange, w)
		default:
			rollerAdjustDay(-1, sel, minYear, maxYear, onChange, w)
		}
		e.IsHandled = true
	case e.Modifiers == ModNone && e.KeyCode == KeyDown:
		switch mode {
		case RollerYearOnly:
			rollerAdjustYear(1, sel, minYear, maxYear, onChange, w, wrapYear)
		case RollerMonthYear:
			rollerAdjustMonth(1, sel, minYear, maxYear, onChange, w)
		default:
			rollerAdjustDay(1, sel, minYear, maxYear, onChange, w)
		}
		e.IsHandled = true
	case e.Modifiers == ModShift && e.KeyCode == KeyUp:
		rollerAdjustYear(-1, sel, minYear, maxYear, onChange, w, wrapYear)
		e.IsHandled = true
	case e.Modifiers == ModShift && e.KeyCode == KeyDown:
		rollerAdjustYear(1, sel, minYear, maxYear, onChange, w, wrapYear)
		e.IsHandled = true
	case e.Modifiers == ModAlt && e.KeyCode == KeyUp:
		rollerAdjustMonth(-1, sel, minYear, maxYear, onChange, w)
		e.IsHandled = true
	case e.Modifiers == ModAlt && e.KeyCode == KeyDown:
		rollerAdjustMonth(1, sel, minYear, maxYear, onChange, w)
		e.IsHandled = true
	}
}

func rollerAdjustDay(
	delta int, sel time.Time,
	minYear, maxYear int,
	onChange func(time.Time, *Window), w *Window,
) {
	if onChange == nil {
		return
	}
	newDate := sel.AddDate(0, 0, delta)
	if newDate.Year() < minYear || newDate.Year() > maxYear {
		return
	}
	onChange(newDate, w)
}

func rollerAdjustMonth(
	delta int, sel time.Time,
	minYear, maxYear int,
	onChange func(time.Time, *Window), w *Window,
) {
	if onChange == nil {
		return
	}
	// Explicitly calculate to handle month-end clamping (e.g. Jan 31 -> Feb 28).
	y, m, d := sel.Date()
	nm := int(m) + delta
	ny := y
	for nm < 1 {
		nm += 12
		ny--
	}
	for nm > 12 {
		nm -= 12
		ny++
	}

	if ny < minYear || ny > maxYear {
		return
	}

	dim := datePickerDaysInMonth(nm, ny)
	nd := d
	if nd > dim {
		nd = dim
	}
	newDate := time.Date(ny, time.Month(nm), nd,
		sel.Hour(), sel.Minute(), sel.Second(), sel.Nanosecond(),
		sel.Location())
	onChange(newDate, w)
}

func rollerAdjustYear(
	delta int, sel time.Time,
	minYear, maxYear int,
	onChange func(time.Time, *Window), w *Window,
	wrap bool,
) {
	if onChange == nil {
		return
	}
	newYear := sel.Year() + delta
	if wrap {
		span := maxYear - minYear + 1
		for newYear < minYear {
			newYear += span
		}
		for newYear > maxYear {
			newYear -= span
		}
	} else {
		if newYear < minYear || newYear > maxYear {
			return
		}
	}
	dim := datePickerDaysInMonth(int(sel.Month()), newYear)
	day := sel.Day()
	if day > dim {
		day = dim
	}
	newDate := time.Date(newYear, sel.Month(), day,
		sel.Hour(), sel.Minute(), sel.Second(), sel.Nanosecond(),
		sel.Location())
	onChange(newDate, w)
}

func rollerDayFormat(v int) string  { return fmt.Sprintf("%02d", v) }
func rollerYearFormat(v int) string { return strconv.Itoa(v) }

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
		cfg.ItemHeight = 24
	}
	if cfg.VisibleItems == 0 {
		cfg.VisibleItems = 3
	}
	if cfg.VisibleItems%2 == 0 {
		cfg.VisibleItems++
	}
	if cfg.WidthDay == 0 {
		cfg.WidthDay = 40
	}
	if cfg.WidthMonth == 0 {
		if cfg.LongMonths {
			cfg.WidthMonth = 100
		} else {
			cfg.WidthMonth = 64
		}
	}
	if cfg.WidthYear == 0 {
		cfg.WidthYear = 52
	}
	if !cfg.Color.IsSet() {
		cfg.Color = guiTheme.ColorBackground
	}
	d := &DefaultDatePickerStyle
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = d.ColorBorder
	}
	if !cfg.ColorBorderFocus.IsSet() {
		cfg.ColorBorderFocus = d.ColorBorderFocus
	}
	if !cfg.SizeBorder.IsSet() {
		cfg.SizeBorder = Some(d.SizeBorder)
	}
	if !cfg.Radius.IsSet() {
		cfg.Radius = Some(d.RadiusBorder)
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(PaddingSmall)
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
	if cfg.SelectedDate.IsZero() {
		cfg.SelectedDate = time.Now()
	}
}
