package gui

import "time"

// InputDateCfg configures a date input with dropdown calendar.
type InputDateCfg struct {
	ID                   string
	Date                 time.Time
	Placeholder          string
	AllowedWeekdays      []DatePickerWeekdays
	AllowedMonths        []DatePickerMonths
	AllowedYears         []int
	AllowedDates         []time.Time
	OnSelect             func([]time.Time, *Event, *Window)
	OnEnter              func(*Layout, *Event, *Window)
	WeekdaysLen          DatePickerWeekdayLen
	TextStyle            TextStyle
	PlaceholderStyle     TextStyle
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
	Sizing               Sizing
	Width                float32
	Height               float32
	MinWidth             float32
	MaxWidth             float32
	Disabled             bool
	Invisible            bool
	SelectMultiple       bool
	HideTodayIndicator   bool
	MondayFirstDayOfWeek bool
	ShowAdjacentMonths   bool

	A11YLabel       string
	A11YDescription string
}

type inputDateView struct {
	cfg InputDateCfg
}

// InputDate creates a date input field with a dropdown calendar.
func InputDate(cfg InputDateCfg) View {
	applyInputDateDefaults(&cfg)
	return &inputDateView{cfg: cfg}
}

func (idv *inputDateView) Content() []View { return nil }

func (idv *inputDateView) GenerateLayout(w *Window) Layout {
	cfg := &idv.cfg

	isOpen := StateReadOr[string, bool](w, nsInputDate, cfg.ID, false)
	cfgID := cfg.ID

	// Format date for display.
	dateText := ""
	if !cfg.Date.IsZero() {
		dateText = LocaleFormatDate(cfg.Date,
			guiLocale.Date.ShortDate)
	}

	displayText := dateText
	ts := cfg.TextStyle
	if displayText == "" && cfg.Placeholder != "" {
		displayText = cfg.Placeholder
		ts = cfg.PlaceholderStyle
	}

	var content []View

	// Date text + calendar icon button.
	content = append(content,
		Row(ContainerCfg{
			Sizing:  FillFit,
			Padding: Some(PaddingNone),
			Spacing: Some(SpacingSmall),
			VAlign:  VAlignMiddle,
			Content: []View{
				Text(TextCfg{
					Text:      displayText,
					TextStyle: ts,
					Sizing:    FillFit,
				}),
				Button(ButtonCfg{
					Disabled: cfg.Disabled,
					Content: []View{Text(TextCfg{
						Text: "\U0001F4C5",
					})},
					OnClick: func(_ *Layout, e *Event, w *Window) {
						inputDateToggle(cfgID, w)
						e.IsHandled = true
					},
				}),
			},
		}),
	)

	// Floating date picker.
	if isOpen {
		content = append(content, Column(ContainerCfg{
			Float:        true,
			FloatAnchor:  FloatBottomLeft,
			FloatTieOff:  FloatTopLeft,
			Padding:      Some(PaddingNone),
			FloatOffsetY: -cfg.SizeBorder.Get(0),
			Content: []View{
				DatePicker(DatePickerCfg{
					ID:                   cfgID + ".picker",
					Dates:                []time.Time{cfg.Date},
					AllowedWeekdays:      cfg.AllowedWeekdays,
					AllowedMonths:        cfg.AllowedMonths,
					AllowedYears:         cfg.AllowedYears,
					AllowedDates:         cfg.AllowedDates,
					WeekdaysLen:          cfg.WeekdaysLen,
					TextStyle:            cfg.TextStyle,
					Color:                cfg.Color,
					ColorHover:           cfg.ColorHover,
					ColorFocus:           cfg.ColorFocus,
					ColorClick:           cfg.ColorClick,
					ColorBorder:          cfg.ColorBorder,
					ColorBorderFocus:     cfg.ColorBorderFocus,
					ColorSelect:          cfg.ColorSelect,
					SizeBorder:           cfg.SizeBorder,
					CellSpacing:          cfg.CellSpacing,
					Radius:               cfg.Radius,
					RadiusBorder:         cfg.RadiusBorder,
					SelectMultiple:       cfg.SelectMultiple,
					HideTodayIndicator:   cfg.HideTodayIndicator,
					MondayFirstDayOfWeek: cfg.MondayFirstDayOfWeek,
					ShowAdjacentMonths:   cfg.ShowAdjacentMonths,
					OnSelect: func(dates []time.Time, e *Event, w *Window) {
						inputDateClose(cfgID, w)
						if cfg.OnSelect != nil {
							cfg.OnSelect(dates, e, w)
						}
					},
				}),
			},
		}))
	}

	col := &containerView{
		cfg: ContainerCfg{
			ID:          cfg.ID,
			IDFocus:     cfg.IDFocus,
			A11YRole:    AccessRoleDateField,
			A11YLabel:   a11yLabel(cfg.A11YLabel, "Date Input"),
			Color:       cfg.Color,
			ColorBorder: cfg.ColorBorder,
			SizeBorder:  cfg.SizeBorder,
			Radius:      cfg.RadiusBorder,
			Padding:     cfg.Padding,
			Sizing:      cfg.Sizing,
			Width:       cfg.Width,
			Height:      cfg.Height,
			MinWidth:    cfg.MinWidth,
			MaxWidth:    cfg.MaxWidth,
			Disabled:    cfg.Disabled,
			Invisible:   cfg.Invisible,
			Content:     content,
			axis:        AxisTopToBottom,
			AmendLayout: func(layout *Layout, w *Window) {
				if w.IsFocus(cfg.IDFocus) {
					layout.Shape.ColorBorder = cfg.ColorBorderFocus
				}
			},
			OnKeyDown: func(_ *Layout, e *Event, w *Window) {
				if isOpen && e.KeyCode == KeyEscape {
					inputDateClose(cfgID, w)
					e.IsHandled = true
				} else if !isOpen && (e.KeyCode == KeySpace ||
					e.KeyCode == KeyEnter) {
					inputDateOpen(cfgID, w)
					e.IsHandled = true
				}
			},
		},
		content:   content,
		shapeType: ShapeRectangle,
	}
	return GenerateViewLayout(col, w)
}

func inputDateToggle(id string, w *Window) {
	sm := StateMap[string, bool](w, nsInputDate, capModerate)
	cur, _ := sm.Get(id)
	sm.Set(id, !cur)
	w.UpdateWindow()
}

func inputDateOpen(id string, w *Window) {
	sm := StateMap[string, bool](w, nsInputDate, capModerate)
	sm.Set(id, true)
	w.UpdateWindow()
}

func inputDateClose(id string, w *Window) {
	sm := StateMap[string, bool](w, nsInputDate, capModerate)
	sm.Set(id, false)
	w.UpdateWindow()
}

func applyInputDateDefaults(cfg *InputDateCfg) {
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
		cfg.Padding = Some(PaddingSmall)
	}
	sizeBorder := cfg.SizeBorder.Get(d.SizeBorder)
	cellSpacing := cfg.CellSpacing.Get(d.CellSpacing)
	radius := cfg.Radius.Get(d.Radius)
	radiusBorder := cfg.RadiusBorder.Get(d.RadiusBorder)
	cfg.SizeBorder = Some(sizeBorder)
	cfg.CellSpacing = Some(cellSpacing)
	cfg.Radius = Some(radius)
	cfg.RadiusBorder = Some(radiusBorder)
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = d.TextStyle
	}
	if cfg.PlaceholderStyle == (TextStyle{}) {
		cfg.PlaceholderStyle = TextStyle{
			Color: RGBA(
				d.TextStyle.Color.R,
				d.TextStyle.Color.G,
				d.TextStyle.Color.B, 100),
			Size: d.TextStyle.Size,
		}
	}
}
