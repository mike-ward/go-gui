package gui

// ExpandPanelCfg configures an expand panel. It consists of a
// header (always visible) and content (visible when expanded).
type ExpandPanelCfg struct {
	ID        string
	Head      View
	Content   View
	Open      bool
	OnToggle  func(*Window)
	Sizing    Sizing
	Color     Color
	ColorHover Color
	ColorClick Color
	ColorBorder Color
	Padding    Opt[Padding]
	SizeBorder Opt[float32]
	Radius     Opt[float32]
	MinWidth   float32
	MaxWidth   float32
	MinHeight  float32
	MaxHeight  float32

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// ExpandPanel creates an expandable panel view.
func ExpandPanel(cfg ExpandPanelCfg) View {
	if !cfg.Color.IsSet() {
		cfg.Color = guiTheme.ExpandPanelStyle.Color
	}
	if !cfg.ColorHover.IsSet() {
		cfg.ColorHover = guiTheme.ExpandPanelStyle.ColorHover
	}
	if !cfg.ColorClick.IsSet() {
		cfg.ColorClick = guiTheme.ExpandPanelStyle.ColorClick
	}
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = guiTheme.ExpandPanelStyle.ColorBorder
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(guiTheme.ExpandPanelStyle.Padding)
	}
	sizeBorder := cfg.SizeBorder.Get(guiTheme.ExpandPanelStyle.SizeBorder)
	radius := cfg.Radius.Get(guiTheme.ExpandPanelStyle.Radius)

	onToggle := cfg.OnToggle
	colorHover := cfg.ColorHover
	colorClick := cfg.ColorClick

	a11yState := AccessState(0)
	if cfg.Open {
		a11yState = AccessStateExpanded
	}

	arrowText := "▼"
	if cfg.Open {
		arrowText = "▲"
	}

	return Column(ContainerCfg{
		ID:              cfg.ID,
		A11YRole:        AccessRoleDisclosure,
		A11YState:       a11yState,
		A11YLabel:       cfg.A11YLabel,
		A11YDescription: cfg.A11YDescription,
		Color:           cfg.Color,
		ColorBorder:     cfg.ColorBorder,
		SizeBorder:      Some(sizeBorder),
		Padding:         cfg.Padding,
		Radius:          Some(radius),
		Sizing:          cfg.Sizing,
		MinWidth:        cfg.MinWidth,
		MaxWidth:        cfg.MaxWidth,
		MinHeight:       cfg.MinHeight,
		MaxHeight:       cfg.MaxHeight,
		Spacing:         Some[float32](0),
		Content: []View{
			Row(ContainerCfg{
				Padding: Some(PaddingNone),
				Sizing:  FillFit,
				VAlign:  VAlignMiddle,
				Content: []View{
					cfg.Head,
					Row(ContainerCfg{
						Padding: Some(NewPadding(0, PadMedium, 0, 0)),
						Content: []View{
							Text(TextCfg{
								Text:      arrowText,
								TextStyle: guiTheme.N3,
							}),
						},
					}),
				},
				OnClick: func(_ *Layout, e *Event, w *Window) {
					if onToggle != nil {
						onToggle(w)
						e.IsHandled = true
					}
				},
				OnChar: func(_ *Layout, e *Event, w *Window) {
					if e.CharCode == CharSpace && onToggle != nil {
						onToggle(w)
						e.IsHandled = true
					}
				},
				OnHover: func(layout *Layout, e *Event, w *Window) {
					w.SetMouseCursorPointingHand()
					layout.Shape.Color = colorHover
					if e.MouseButton == MouseLeft {
						layout.Shape.Color = colorClick
					}
					e.IsHandled = true
				},
			}),
			Column(ContainerCfg{
				Invisible: !cfg.Open,
				Padding:   Some(PaddingNone),
				Sizing:    FillFit,
				Spacing:   Some[float32](0),
				Content: []View{
					cfg.Content,
				},
			}),
		},
	})
}
