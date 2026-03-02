package gui

import "time"

// tooltipState tracks active tooltip bounds and ID.
type tooltipState struct {
	bounds DrawClip
	id     string
}

// TooltipCfg configures a tooltip popup.
type TooltipCfg struct {
	ID           string
	Color        Color
	ColorHover   Color
	ColorBorder  Color
	Padding      Padding
	TextStyle    TextStyle
	Content      []View
	Delay        time.Duration
	Radius       float32
	RadiusBorder float32
	SizeBorder   float32
	OffsetX      float32
	OffsetY      float32
	Anchor       FloatAttach
	TieOff       FloatAttach
}

// Tooltip creates a floating tooltip view.
func Tooltip(cfg TooltipCfg) View {
	applyTooltipDefaults(&cfg)
	return Row(ContainerCfg{
		ID:           cfg.ID,
		Float:        true,
		FloatAnchor:  cfg.Anchor,
		FloatTieOff:  cfg.TieOff,
		FloatOffsetX: cfg.OffsetX,
		FloatOffsetY: cfg.OffsetY,
		Color:        cfg.Color,
		ColorBorder:  cfg.ColorBorder,
		SizeBorder:   Some(cfg.SizeBorder),
		Radius:       Some(cfg.Radius),
		Padding:      Some(cfg.Padding),
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
	if cfg.Color == (Color{}) {
		cfg.Color = d.Color
	}
	if cfg.ColorHover == (Color{}) {
		cfg.ColorHover = d.ColorHover
	}
	if cfg.ColorBorder == (Color{}) {
		cfg.ColorBorder = d.ColorBorder
	}
	if cfg.Padding == (Padding{}) {
		cfg.Padding = d.Padding
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
}
