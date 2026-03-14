package gui

// SliderCfg configures a single-value slider. Wraps RangeSlider
// with a simplified API.
type SliderCfg struct {
	ID         string
	Value      float32
	Min        float32
	Max        float32
	Step       float32
	OnChange   func(float32, *Event, *Window)
	Color      Color
	ColorThumb Color
	ColorFocus Color
	ColorHover Color
	Sizing     Sizing
	Width      float32
	Height     float32
	ThumbSize  float32
	Radius     Opt[float32]
	Padding    Opt[Padding]
	SizeBorder Opt[float32]
	IDFocus    uint32
	RoundValue bool
	Vertical   bool
	Disabled   bool
	Invisible  bool

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// Slider creates a single-value slider. Delegates to RangeSlider.
func Slider(cfg SliderCfg) View {
	return RangeSlider(RangeSliderCfg{
		ID:              cfg.ID,
		Value:           cfg.Value,
		Min:             cfg.Min,
		Max:             cfg.Max,
		Step:            cfg.Step,
		OnChange:        cfg.OnChange,
		Color:           cfg.Color,
		ColorThumb:      cfg.ColorThumb,
		ColorFocus:      cfg.ColorFocus,
		ColorHover:      cfg.ColorHover,
		Sizing:          cfg.Sizing,
		Width:           cfg.Width,
		Height:          cfg.Height,
		ThumbSize:       cfg.ThumbSize,
		Radius:          cfg.Radius,
		Padding:         cfg.Padding,
		SizeBorder:      cfg.SizeBorder,
		IDFocus:         cfg.IDFocus,
		RoundValue:      cfg.RoundValue,
		Vertical:        cfg.Vertical,
		Disabled:        cfg.Disabled,
		Invisible:       cfg.Invisible,
		A11YLabel:       cfg.A11YLabel,
		A11YDescription: cfg.A11YDescription,
	})
}
