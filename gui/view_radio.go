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
	Padding          Opt[Padding]
	TextStyle        TextStyle
	OnClick          func(*Layout, *Event, *Window)
	Size             Opt[float32]
	IDFocus          uint32
	SizeBorder       Opt[float32]
	Disabled         bool
	Selected         bool
	Invisible        bool

	A11YLabel       string
	A11YDescription string
}

// Radio creates a radio button view.
func Radio(cfg RadioCfg) View {
	applyRadioDefaults(&cfg)

	dr := &DefaultRadioStyle
	size := cfg.Size.Get(dr.Size)
	sizeBorder := cfg.SizeBorder.Get(dr.SizeBorder)

	colorBorderFocus := cfg.ColorBorderFocus
	circleColor := cfg.ColorUnselect
	if cfg.Selected {
		circleColor = cfg.ColorSelect
	}

	content := make([]View, 0, 2)
	content = append(content, Circle(ContainerCfg{
		Width:       size,
		Height:      size,
		Color:       circleColor,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  Some(sizeBorder),
		Disabled:    cfg.Disabled,
		Invisible:   cfg.Invisible,
		Sizing:      FixedFixed,
		HAlign:      HAlignCenter,
		VAlign:      VAlignMiddle,
	}))

	if len(cfg.Label) > 0 {
		content = append(content, Row(ContainerCfg{
			Padding: Some(NewPadding(0, PadXSmall, 0, 0)),
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
		cfg.ColorSelect = DefaultRadioStyle.ColorSelect
	}
	if !cfg.ColorUnselect.IsSet() {
		cfg.ColorUnselect = DefaultRadioStyle.ColorUnselect
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(PaddingNone)
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
}
