package gui

import "strconv"

// BadgeVariant selects the badge color preset.
type BadgeVariant uint8

const (
	BadgeDefault BadgeVariant = iota
	BadgeInfo
	BadgeSuccess
	BadgeWarning
	BadgeError
)

// BadgeCfg configures a Badge view.
type BadgeCfg struct {
	Label     string
	Variant   BadgeVariant
	Max       int // 0 = no cap; shows "max+" when exceeded
	Dot       bool
	Color     Color
	Padding   Padding
	Radius    float32
	TextStyle TextStyle
	DotSize   float32

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// Badge creates a badge view. Dot mode renders a fixed-size
// circle; labeled mode renders text inside a rounded row.
func Badge(cfg BadgeCfg) View {
	if cfg.Color == (Color{}) {
		cfg.Color = guiTheme.BadgeStyle.Color
	}
	if cfg.Padding == (Padding{}) {
		cfg.Padding = guiTheme.BadgeStyle.Padding
	}
	if cfg.Radius == 0 {
		cfg.Radius = guiTheme.BadgeStyle.Radius
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = guiTheme.BadgeStyle.TextStyle
	}
	if cfg.DotSize == 0 {
		cfg.DotSize = guiTheme.BadgeStyle.DotSize
	}

	style := guiTheme.BadgeStyle
	bg := cfg.Color
	switch cfg.Variant {
	case BadgeInfo:
		bg = style.ColorInfo
	case BadgeSuccess:
		bg = style.ColorSuccess
	case BadgeWarning:
		bg = style.ColorWarning
	case BadgeError:
		bg = style.ColorError
	}

	if cfg.Dot {
		sz := cfg.DotSize
		return Row(ContainerCfg{
			A11YLabel: a11yLabel(cfg.A11YLabel, "status"),
			Color:     bg,
			Radius:    sz / 2,
			Width:     sz,
			Height:    sz,
			Sizing:    FixedFixed,
			Padding:   PaddingNone,
		})
	}

	label := badgeLabel(cfg.Label, cfg.Max)
	return Row(ContainerCfg{
		A11YLabel: a11yLabel(cfg.A11YLabel, label),
		Color:     bg,
		Radius:    cfg.Radius,
		Sizing:    FitFit,
		Padding:   cfg.Padding,
		HAlign:    HAlignCenter,
		VAlign:    VAlignMiddle,
		Content: []View{
			Text(TextCfg{
				Text:      label,
				TextStyle: cfg.TextStyle,
			}),
		},
	})
}

func badgeLabel(label string, max int) string {
	if max <= 0 {
		return label
	}
	n := 0
	for _, c := range label {
		if c < '0' || c > '9' {
			return label
		}
		n = n*10 + int(c-'0')
	}
	if len(label) == 0 {
		return label
	}
	if n > max {
		return strconv.Itoa(max) + "+"
	}
	return label
}
