package gui

// ButtonCfg configures a clickable button. Without an OnClick
// handler it functions as bubble text (no mouse interaction).
type ButtonCfg struct {
	ID               string
	Color            Color
	ColorHover       Color
	ColorFocus       Color
	ColorClick       Color
	ColorBorder      Color
	ColorBorderFocus Color
	Padding          Opt[Padding]
	SizeBorder       Opt[float32]
	BlurRadius       float32
	Shadow           *BoxShadow
	Gradient         *GradientDef
	Content          []View
	OnClick          func(*Layout, *Event, *Window)
	OnHover          func(*Layout, *Event, *Window)
	Float            bool
	FloatAnchor      FloatAttach
	FloatTieOff      FloatAttach
	FloatOffsetX     float32
	FloatOffsetY     float32
	Radius           Opt[float32]
	IDFocus          uint32
	HAlign           HorizontalAlign
	VAlign           VerticalAlign
	Disabled         bool
	Invisible        bool

	// Sizing
	Sizing    Sizing
	Width     float32
	Height    float32
	MinWidth  float32
	MaxWidth  float32
	MinHeight float32
	MaxHeight float32

	// Accessibility
	A11YRole        AccessRole
	A11YState       AccessState
	A11YLabel       string
	A11YDescription string
}

// Button creates a clickable button. Delegates to Row with
// amend_layout for focus coloring and on_hover for cursor/color
// state changes.
func Button(cfg ButtonCfg) View {
	// Apply defaults from button style.
	applyButtonDefaults(&cfg)

	d := &DefaultButtonStyle
	sizeBorder := cfg.SizeBorder.Get(d.SizeBorder)
	radius := cfg.Radius.Get(d.Radius)

	// Capture values for closures.
	colorHover := cfg.ColorHover
	colorClick := cfg.ColorClick
	colorFocus := cfg.ColorFocus
	colorBorderFocus := cfg.ColorBorderFocus
	userOnHover := cfg.OnHover
	onClick := cfg.OnClick

	a11yRole := cfg.A11YRole
	if a11yRole == AccessRoleNone {
		a11yRole = AccessRoleButton
	}

	return Row(ContainerCfg{
		ID:              cfg.ID,
		IDFocus:         cfg.IDFocus,
		A11YRole:        a11yRole,
		A11YState:       cfg.A11YState,
		A11YLabel:       cfg.A11YLabel,
		A11YDescription: cfg.A11YDescription,
		Color:           cfg.Color,
		ColorBorder:     cfg.ColorBorder,
		SizeBorder:      Some(sizeBorder),
		BlurRadius:      cfg.BlurRadius,
		Shadow:          cfg.Shadow,
		Gradient:        cfg.Gradient,
		Padding:         cfg.Padding,
		Radius:          Some(radius),
		Width:           cfg.Width,
		Height:          cfg.Height,
		MinWidth:        cfg.MinWidth,
		MaxWidth:        cfg.MaxWidth,
		MinHeight:       cfg.MinHeight,
		MaxHeight:       cfg.MaxHeight,
		Sizing:          cfg.Sizing,
		Disabled:        cfg.Disabled,
		Invisible:       cfg.Invisible,
		HAlign:          cfg.HAlign,
		VAlign:          cfg.VAlign,
		Float:           cfg.Float,
		FloatAnchor:     cfg.FloatAnchor,
		FloatTieOff:     cfg.FloatTieOff,
		FloatOffsetX:    cfg.FloatOffsetX,
		FloatOffsetY:    cfg.FloatOffsetY,
		OnClick:         onClick,
		OnChar:          spacebarToClick(onClick),
		OnKeyDown:       enterToClick(onClick),
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
		OnHover: func(layout *Layout, e *Event, w *Window) {
			if !layout.Shape.HasEvents() ||
				layout.Shape.Events.OnClick == nil {
				return
			}
			w.SetMouseCursor(CursorPointingHand)
			if !w.IsFocus(layout.Shape.IDFocus) {
				layout.Shape.Color = colorHover
			}
			if e.MouseButton == MouseLeft {
				layout.Shape.Color = colorClick
			}
			if userOnHover != nil {
				userOnHover(layout, e, w)
			}
		},
		Content: cfg.Content,
	})
}

func applyButtonDefaults(cfg *ButtonCfg) {
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
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(d.Padding)
	}
	if cfg.HAlign == HAlignStart {
		cfg.HAlign = HAlignCenter
	}
	if cfg.VAlign == VAlignTop {
		cfg.VAlign = VAlignMiddle
	}
}
