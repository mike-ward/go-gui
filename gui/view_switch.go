package gui

// SwitchCfg configures a pill-shaped toggle switch.
type SwitchCfg struct {
	ID               string
	Label            string
	Color            Color
	ColorFocus       Color
	ColorHover       Color
	ColorClick       Color
	ColorBorder      Color
	ColorBorderFocus Color
	ColorSelect      Color
	ColorUnselect    Color
	Padding          Padding
	SizeBorder       float32
	TextStyle        TextStyle
	OnClick          func(*Layout, *Event, *Window)
	Width            float32
	Height           float32
	Radius           float32
	IDFocus          uint32
	Disabled         bool
	Invisible        bool
	Selected         bool

	A11YLabel       string
	A11YDescription string
}

// Switch creates a pill-shaped toggle switch.
func Switch(cfg SwitchCfg) View {
	applySwitchDefaults(&cfg)

	thumbColor := cfg.ColorUnselect
	if cfg.Selected {
		thumbColor = cfg.ColorSelect
	}
	circleSize := cfg.Height - cfg.Padding.Height() - (cfg.SizeBorder * 2)

	colorFocus := cfg.ColorFocus
	colorBorderFocus := cfg.ColorBorderFocus
	colorHover := cfg.ColorHover
	colorClick := cfg.ColorClick

	hAlign := HAlignStart
	if cfg.Selected {
		hAlign = HAlignEnd
	}

	content := make([]View, 0, 2)
	content = append(content, Row(ContainerCfg{
		ID:          cfg.ID,
		Width:       cfg.Width,
		Height:      cfg.Height,
		Sizing:      FixedFit,
		Color:       cfg.Color,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  cfg.SizeBorder,
		Radius:      cfg.Radius,
		Disabled:    cfg.Disabled,
		Invisible:   cfg.Invisible,
		Padding:     cfg.Padding,
		HAlign:      hAlign,
		VAlign:      VAlignMiddle,
		Content: []View{
			Circle(ContainerCfg{
				Color:  thumbColor,
				Width:  circleSize,
				Height: circleSize,
				Sizing: FixedFixed,
			}),
		},
	}))
	if len(cfg.Label) > 0 {
		content = append(content,
			Text(TextCfg{Text: cfg.Label, TextStyle: cfg.TextStyle}))
	}

	a11yState := AccessStateNone
	if cfg.Selected {
		a11yState = AccessStateChecked
	}

	return Row(ContainerCfg{
		IDFocus:         cfg.IDFocus,
		Padding:         PaddingNone,
		A11YRole:        AccessRoleSwitchToggle,
		A11YState:       a11yState,
		A11YLabel:       a11yLabel(cfg.A11YLabel, cfg.Label),
		A11YDescription: cfg.A11YDescription,
		OnChar:          spacebarToClick(cfg.OnClick),
		OnClick:         leftClickOnly(cfg.OnClick),
		OnHover: func(layout *Layout, e *Event, w *Window) {
			w.SetMouseCursor(CursorPointingHand)
			if len(layout.Children) > 0 {
				layout.Children[0].Shape.Color = colorHover
				if e.MouseButton == MouseLeft {
					layout.Children[0].Shape.Color = colorClick
				}
			}
		},
		AmendLayout: func(layout *Layout, w *Window) {
			if layout.Shape.Disabled ||
				!layout.Shape.HasEvents() ||
				layout.Shape.Events.OnClick == nil {
				return
			}
			if w.IsFocus(layout.Shape.IDFocus) {
				layout.Shape.Color = colorFocus
				layout.Shape.ColorBorder = colorBorderFocus
			}
		},
		Content: content,
	})
}

func applySwitchDefaults(cfg *SwitchCfg) {
	d := &DefaultButtonStyle
	if cfg.Color == (Color{}) {
		cfg.Color = d.Color
	}
	if cfg.ColorFocus == (Color{}) {
		cfg.ColorFocus = d.ColorFocus
	}
	if cfg.ColorHover == (Color{}) {
		cfg.ColorHover = d.ColorHover
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
	if cfg.Width == 0 {
		cfg.Width = 40
	}
	if cfg.Height == 0 {
		cfg.Height = 22
	}
	if cfg.Radius == 0 {
		cfg.Radius = 11
	}
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = SizeBorderDef
	}
	if cfg.Padding == (Padding{}) {
		cfg.Padding = PaddingTwo
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
}
