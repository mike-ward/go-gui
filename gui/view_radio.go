package gui

// RadioCfg configures a radio button.
type RadioCfg struct {
	ID               string
	Label            string
	Color            Color
	ColorHover       Color
	ColorFocus       Color
	ColorClick       Color
	ColorBorder      Color
	ColorBorderFocus Color
	ColorSelect      Color
	ColorUnselect    Color
	Padding          Padding
	TextStyle        TextStyle
	OnClick          func(*Layout, *Event, *Window)
	Size             float32
	IDFocus          uint32
	SizeBorder       float32
	Disabled         bool
	Selected         bool
	Invisible        bool

	A11YLabel       string
	A11YDescription string
}

// Radio creates a radio button view.
func Radio(cfg RadioCfg) View {
	applyRadioDefaults(&cfg)

	colorBorderFocus := cfg.ColorBorderFocus
	circleColor := cfg.ColorUnselect
	if cfg.Selected {
		circleColor = cfg.ColorSelect
	}

	content := make([]View, 0, 2)
	content = append(content, Circle(ContainerCfg{
		Width:       cfg.Size,
		Height:      cfg.Size,
		Color:       circleColor,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  cfg.SizeBorder,
		Disabled:    cfg.Disabled,
		Invisible:   cfg.Invisible,
		Sizing:      FixedFixed,
		HAlign:      HAlignCenter,
		VAlign:      VAlignMiddle,
	}))

	if len(cfg.Label) > 0 {
		content = append(content, Row(ContainerCfg{
			Padding: NewPadding(0, PadXSmall, 0, 0),
			Content: []View{
				Text(TextCfg{Text: cfg.Label, TextStyle: cfg.TextStyle}),
			},
		}))
	}

	a11yState := AccessStateNone
	if cfg.Selected {
		a11yState = AccessStateSelected
	}

	return Row(ContainerCfg{
		ID:              cfg.ID,
		IDFocus:         cfg.IDFocus,
		Padding:         cfg.Padding,
		VAlign:          VAlignMiddle,
		A11YRole:        AccessRoleRadioButton,
		A11YState:       a11yState,
		A11YLabel:       a11yLabel(cfg.A11YLabel, cfg.Label),
		A11YDescription: cfg.A11YDescription,
		OnClick:         leftClickOnly(cfg.OnClick),
		OnChar:          spacebarToClick(cfg.OnClick),
		AmendLayout: func(layout *Layout, w *Window) {
			if layout.Shape.Disabled ||
				!layout.Shape.HasEvents() ||
				layout.Shape.Events.OnClick == nil {
				return
			}
			if len(layout.Children) == 0 {
				return
			}
			if w.IsFocus(layout.Shape.IDFocus) {
				layout.Children[0].Shape.ColorBorder = colorBorderFocus
			}
		},
		OnHover: func(_ *Layout, _ *Event, w *Window) {
			w.SetMouseCursor(CursorPointingHand)
		},
		Content: content,
	})
}

func applyRadioDefaults(cfg *RadioCfg) {
	d := &DefaultButtonStyle
	if cfg.Color == (Color{}) {
		cfg.Color = d.Color
	}
	if cfg.ColorHover == (Color{}) {
		cfg.ColorHover = d.ColorHover
	}
	if cfg.ColorFocus == (Color{}) {
		cfg.ColorFocus = d.ColorFocus
	}
	if cfg.ColorClick == (Color{}) {
		cfg.ColorClick = d.ColorClick
	}
	if cfg.ColorBorder == (Color{}) {
		cfg.ColorBorder = d.ColorBorder
	}
	if cfg.ColorBorderFocus == (Color{}) {
		cfg.ColorBorderFocus = d.ColorBorderFocus
	}
	if cfg.ColorSelect == (Color{}) {
		cfg.ColorSelect = colorSelectDark
	}
	if cfg.ColorUnselect == (Color{}) {
		cfg.ColorUnselect = colorInteriorDark
	}
	if cfg.Size == 0 {
		cfg.Size = SizeTextMedium
	}
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = SizeBorderDef
	}
	if cfg.Padding == (Padding{}) {
		cfg.Padding = PaddingSmall
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
}
