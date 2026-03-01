package gui

// RectangleCfg configures a rectangle. Rectangles can be
// filled, outlined, colored, and have radius corners.
type RectangleCfg struct {
	ID             string
	Sizing         Sizing
	Color          Color
	ColorBorder    Color
	Gradient       *GradientDef
	BorderGradient *GradientDef
	Shadow         *BoxShadow
	Width          float32
	Height         float32
	MinWidth       float32
	MinHeight      float32
	MaxHeight      float32
	Radius         float32
	BlurRadius     float32
	SizeBorder     float32
	Disabled       bool
	Invisible      bool
}

// Rectangle draws a rectangle. Technically a container with no
// children, axis, or padding.
func Rectangle(cfg RectangleCfg) View {
	return container(ContainerCfg{
		ID:             cfg.ID,
		Width:          cfg.Width,
		Height:         cfg.Height,
		MinWidth:       cfg.Width,
		MinHeight:      cfg.Height,
		Sizing:         cfg.Sizing,
		Disabled:       cfg.Disabled,
		Invisible:      cfg.Invisible,
		Color:          cfg.Color,
		ColorBorder:    cfg.ColorBorder,
		Gradient:       cfg.Gradient,
		BorderGradient: cfg.BorderGradient,
		Shadow:         cfg.Shadow,
		BlurRadius:     cfg.BlurRadius,
		Padding:        PaddingNone,
		Radius:         cfg.Radius,
		SizeBorder:     cfg.SizeBorder,
		Spacing:        0,
	})
}
