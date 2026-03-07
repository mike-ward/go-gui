package gui

import "time"

// tooltipState tracks active tooltip bounds and ID.
type tooltipState struct {
	bounds       DrawClip  // absolute run bounds (mouse check)
	floatOffsetX float32   // run-relative X for float popup
	floatOffsetY float32   // run-relative Y for float popup
	blockKey     uint64    // FNV hash of owning RichText
	id           string
	hoverID      string    // trigger currently hovered
	hoverStart   time.Time // when hover began
	text         string    // RTF tooltip content
}

// clearText resets RTF-sourced tooltip state. No-op when
// text is empty so WithTooltip-managed state is preserved.
func (ts *tooltipState) clearText() {
	if ts.text == "" {
		return
	}
	ts.hoverID = ""
	ts.hoverStart = time.Time{}
	ts.id = ""
	ts.text = ""
	ts.bounds = DrawClip{}
	ts.floatOffsetX = 0
	ts.floatOffsetY = 0
	ts.blockKey = 0
}

// TooltipCfg configures a tooltip popup.
type TooltipCfg struct {
	ID           string
	Color        Color
	ColorHover   Color
	ColorBorder  Color
	Padding      Opt[Padding]
	TextStyle    TextStyle
	Content      []View
	Delay        time.Duration
	Radius       float32
	RadiusBorder float32
	SizeBorder   float32
	OffsetX      float32
	OffsetY      float32
	FloatZIndex  int
	Anchor       FloatAttach
	TieOff       FloatAttach
}

// Tooltip creates a floating tooltip view.
func Tooltip(cfg TooltipCfg) View {
	applyTooltipDefaults(&cfg)
	return Column(ContainerCfg{
		ID:            cfg.ID,
		Float:         true,
		FloatAutoFlip: true,
		FloatAnchor:   cfg.Anchor,
		FloatTieOff:  cfg.TieOff,
		FloatOffsetX: cfg.OffsetX,
		FloatOffsetY: cfg.OffsetY,
		FloatZIndex:  cfg.FloatZIndex,
		Color:        cfg.Color,
		ColorBorder:  cfg.ColorBorder,
		SizeBorder:   Some(cfg.SizeBorder),
		Radius:       Some(cfg.Radius),
		Padding:      cfg.Padding,
		MaxWidth:     300,
		Content:      cfg.Content,
	})
}

// AnimationTooltip returns an Animate that checks mouse position
// after a delay and activates the tooltip if the mouse is still
// inside the trigger bounds.
func AnimationTooltip(cfg TooltipCfg) *Animate {
	delay := cfg.Delay
	if delay == 0 {
		delay = DefaultTooltipStyle.Delay
	}
	id := cfg.ID
	return &Animate{
		AnimateID: "___tooltip___",
		Delay:     delay,
		Callback: func(_ *Animate, w *Window) {
			b := w.viewState.tooltip.bounds
			mx := w.viewState.mousePosX
			my := w.viewState.mousePosY
			if mx >= b.X && my >= b.Y &&
				mx < b.X+b.Width && my < b.Y+b.Height {
				w.viewState.tooltip.id = id
			}
		},
	}
}

func applyTooltipDefaults(cfg *TooltipCfg) {
	d := &DefaultTooltipStyle
	if !cfg.Color.IsSet() {
		cfg.Color = d.Color
	}
	if !cfg.ColorHover.IsSet() {
		cfg.ColorHover = d.ColorHover
	}
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = d.ColorBorder
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(d.Padding)
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = d.TextStyle
	}
	if cfg.Delay == 0 {
		cfg.Delay = d.Delay
	}
	if cfg.Radius == 0 {
		cfg.Radius = d.Radius
	}
	if cfg.RadiusBorder == 0 {
		cfg.RadiusBorder = d.RadiusBorder
	}
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = d.SizeBorder
	}
	if cfg.OffsetX == 0 {
		cfg.OffsetX = -3
	}
	if cfg.OffsetY == 0 {
		cfg.OffsetY = -3
	}
	if cfg.Anchor == FloatTopLeft {
		cfg.Anchor = FloatBottomCenter
	}
	if cfg.TieOff == FloatTopLeft {
		cfg.TieOff = FloatTopCenter
	}
}

// WithTooltipCfg configures a tooltip wrapper.
type WithTooltipCfg struct {
	ID      string
	Text    string
	Delay   time.Duration
	Anchor  FloatAttach
	TieOff  FloatAttach
	Content []View
}

// WithTooltip wraps content and shows a tooltip on hover after
// a delay. Tooltip state is managed via AmendLayout.
func WithTooltip(w *Window, cfg WithTooltipCfg) View {
	tipID := cfg.ID
	if tipID == "" {
		tipID = cfg.Text
	}

	delay := cfg.Delay
	if delay == 0 {
		delay = DefaultTooltipStyle.Delay
	}

	anchor := cfg.Anchor
	if anchor == FloatTopLeft {
		anchor = FloatBottomCenter
	}
	tieOff := cfg.TieOff
	if tieOff == FloatTopLeft {
		tieOff = FloatTopCenter
	}

	content := make([]View, 0, len(cfg.Content)+1)
	content = append(content, cfg.Content...)

	if w.viewState.tooltip.id == tipID {
		content = append(content, Tooltip(TooltipCfg{
			ID:     tipID + "_popup",
			Anchor: anchor,
			TieOff: tieOff,
			Content: []View{
				Text(TextCfg{
					Text:      cfg.Text,
					TextStyle: DefaultTooltipStyle.TextStyle,
					Mode:      TextModeWrap,
				}),
			},
		}))
	}

	return Column(ContainerCfg{
		A11YRole:        AccessRoleGroup,
		A11YDescription: cfg.Text,
		Content:         content,
		AmendLayout:     withTooltipAmend(tipID, delay),
	})
}

// withTooltipAmend returns the AmendLayout callback for a
// WithTooltip wrapper.
func withTooltipAmend(
	tipID string, delay time.Duration,
) func(*Layout, *Window) {
	return func(l *Layout, w *Window) {
		ts := &w.viewState.tooltip
		mx := w.viewState.mousePosX
		my := w.viewState.mousePosY
		inside := mx >= l.Shape.X && my >= l.Shape.Y &&
			mx < l.Shape.X+l.Shape.Width &&
			my < l.Shape.Y+l.Shape.Height

		switch {
		case inside && ts.hoverID == "":
			ts.hoverID = tipID
			ts.hoverStart = time.Now()
			w.animationAdd(&Animate{
				AnimateID: "___tooltip___",
				Delay:     delay,
				Callback: func(_ *Animate, w *Window) {
					if w.viewState.tooltip.hoverID == tipID {
						w.viewState.tooltip.id = tipID
					}
				},
			})

		case inside && ts.hoverID == tipID &&
			time.Since(ts.hoverStart) >= delay:
			ts.id = tipID

		case !inside && ts.hoverID == tipID:
			ts.hoverID = ""
			ts.hoverStart = time.Time{}
			ts.id = ""
		}
	}
}
