//go:generate go run ./internal/gen/spinnerkinds/

package gui

// SvgSpinnerCfg configures a built-in SVG spinner.
//
// Phase 0 caveat: only assets that animate via rotation render
// animated. Other assets render as the static initial frame
// until the SMIL parser gains support for attribute animation
// and spline easing.
type SvgSpinnerCfg struct {
	ID              string
	Kind            SvgSpinnerKind
	Color           Color
	Width           float32
	Height          float32
	Sizing          Sizing
	Padding         Opt[Padding]
	MinWidth        float32
	MaxWidth        float32
	MinHeight       float32
	MaxHeight       float32
	OnClick         func(*Layout, *Event, *Window)
	A11YLabel       string
	A11YDescription string
}

// SvgSpinner renders a built-in animated SVG spinner identified
// by cfg.Kind. cfg.Color recolors monochrome assets via the
// SVG fill="currentColor" convention. When unset, defaults to
// the current theme's text color so spinners are visible on
// the standard background.
func SvgSpinner(cfg SvgSpinnerCfg) View {
	if uint16(cfg.Kind) >= svgSpinnerCount {
		cfg.Kind = 0
	}
	if !cfg.Color.IsSet() {
		cfg.Color = guiTheme.TextStyleDef.Color
	}
	width := cfg.Width
	height := cfg.Height
	if width <= 0 && height <= 0 {
		width = 48
		height = 48
	}
	return Svg(SvgCfg{
		ID:              cfg.ID,
		SvgData:         svgSpinnerData[cfg.Kind],
		Width:           width,
		Height:          height,
		Color:           cfg.Color,
		Sizing:          cfg.Sizing,
		Padding:         cfg.Padding,
		OnClick:         cfg.OnClick,
		A11YLabel:       cfg.A11YLabel,
		A11YDescription: cfg.A11YDescription,
	})
}

// SvgSpinnerName returns the asset basename for a given kind
// (e.g. "90-ring"). Useful for debugging and gallery labels.
func SvgSpinnerName(k SvgSpinnerKind) string {
	if uint16(k) >= svgSpinnerCount {
		return ""
	}
	return svgSpinnerName[k]
}

// SvgSpinnerCount returns the number of built-in spinner kinds.
func SvgSpinnerCount() int { return svgSpinnerCount }
