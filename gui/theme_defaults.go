package gui

// Light theme color vars.
var (
	colorBackgroundLight  = RGB(225, 225, 225)
	colorPanelLight       = RGB(205, 205, 215)
	colorInteriorLight    = RGB(195, 195, 215)
	colorHoverLight       = RGB(185, 185, 215)
	colorFocusLight       = RGB(175, 175, 215)
	colorActiveLight      = RGB(165, 165, 215)
	colorBorderLight      = RGB(135, 135, 165)
	colorSelectLight      = RGB(65, 105, 225)
	colorBorderFocusLight = RGB(0, 0, 165)
	colorTextLight        = RGB(32, 32, 32)
)

// Scroll constants.
const (
	scrollMultiplier float32 = 20
	scrollDeltaLine  float32 = 1
	scrollDeltaPage  float32 = 10
)

// baseCfg returns the shared sizing/spacing/widget-size fields
// common to all preset themes.
func baseCfg() ThemeCfg {
	return ThemeCfg{
		MonoFontFamily:   defaultMonoFontFamily,
		Padding:          PaddingMedium,
		PaddingSmall:     PaddingSmall,
		PaddingMedium:    PaddingMedium,
		PaddingLarge:     PaddingLarge,
		Radius:           RadiusMedium,
		RadiusSmall:      RadiusSmall,
		RadiusMedium:     RadiusMedium,
		RadiusLarge:      RadiusLarge,
		SpacingSmall:     SpacingSmall,
		SpacingMedium:    SpacingMedium,
		SpacingLarge:     SpacingLarge,
		SizeTextTiny:     SizeTextTiny,
		SizeTextXSmall:   SizeTextXSmall,
		SizeTextSmall:    SizeTextSmall,
		SizeTextMedium:   SizeTextMedium,
		SizeTextLarge:    SizeTextLarge,
		SizeTextXLarge:   SizeTextXLarge,
		ScrollMultiplier: scrollMultiplier,
		ScrollDeltaLine:  scrollDeltaLine,
		ScrollDeltaPage:  scrollDeltaPage,
		SizeSwitchWidth:  36,
		SizeSwitchHeight: 22,
		SizeRadio:        16,
		SizeScrollbar:    7,
		SizeScrollbarMin: 20,
		SizeProgressBar:  20,
		SizeSlider:       6,
		SizeSliderThumb:  16,
	}
}

// baseDarkCfg returns the common dark ThemeCfg.
func baseDarkCfg() ThemeCfg {
	cfg := baseCfg()
	cfg.Name = "dark"
	cfg.ColorBackground = colorBackgroundDark
	cfg.ColorPanel = colorPanelDark
	cfg.ColorInterior = colorInteriorDark
	cfg.ColorHover = colorHoverDark
	cfg.ColorFocus = colorFocusDark
	cfg.ColorActive = colorActiveDark
	cfg.ColorBorder = colorBorderDark
	cfg.ColorBorderFocus = colorSelectDark
	cfg.ColorSelect = colorSelectDark
	cfg.TitlebarDark = true
	cfg.TextStyleDef = DefaultTextStyle
	cfg.ColorError = RGBA(218, 54, 51, 255)
	return cfg
}

// Preset theme configs and themes.
var (
	ThemeDarkCfg ThemeCfg
	ThemeDark    Theme

	ThemeDarkNoPaddingCfg ThemeCfg
	ThemeDarkNoPadding    Theme

	ThemeDarkBorderedCfg ThemeCfg
	ThemeDarkBordered    Theme

	ThemeLightCfg ThemeCfg
	ThemeLight    Theme

	ThemeLightNoPaddingCfg ThemeCfg
	ThemeLightNoPadding    Theme

	ThemeLightBorderedCfg ThemeCfg
	ThemeLightBordered    Theme

	ThemeBlueBorderedCfg ThemeCfg
	ThemeBlueBordered    Theme
)

func init() {
	// Dark.
	ThemeDarkCfg = baseDarkCfg()
	ThemeDark = ThemeMaker(ThemeDarkCfg)

	// Dark no padding.
	ThemeDarkNoPaddingCfg = baseDarkCfg()
	ThemeDarkNoPaddingCfg.Name = "dark-no-padding"
	ThemeDarkNoPaddingCfg.Padding = PaddingNone
	ThemeDarkNoPaddingCfg.SizeBorder = 0
	ThemeDarkNoPaddingCfg.Radius = RadiusNone
	ThemeDarkNoPadding = ThemeMaker(ThemeDarkNoPaddingCfg)

	// Dark bordered.
	ThemeDarkBorderedCfg = baseDarkCfg()
	ThemeDarkBorderedCfg.Name = "dark-bordered"
	ThemeDarkBorderedCfg.SizeBorder = SizeBorderDef
	ThemeDarkBordered = ThemeMaker(ThemeDarkBorderedCfg)

	// Light.
	ThemeLightCfg = baseCfg()
	ThemeLightCfg.Name = "light"
	ThemeLightCfg.ColorBackground = colorBackgroundLight
	ThemeLightCfg.ColorPanel = colorPanelLight
	ThemeLightCfg.ColorInterior = colorInteriorLight
	ThemeLightCfg.ColorHover = colorHoverLight
	ThemeLightCfg.ColorFocus = colorFocusLight
	ThemeLightCfg.ColorActive = colorActiveLight
	ThemeLightCfg.ColorBorder = colorBorderLight
	ThemeLightCfg.ColorBorderFocus = colorBorderFocusLight
	ThemeLightCfg.ColorSelect = colorSelectLight
	ThemeLightCfg.ColorError = RGBA(200, 40, 40, 255)
	ThemeLightCfg.TextStyleDef = TextStyle{
		Family: defaultFontFamily,
		Color:  colorTextLight,
		Size:   SizeTextMedium,
	}
	ThemeLight = ThemeMaker(ThemeLightCfg)

	// Light no padding.
	ThemeLightNoPaddingCfg = ThemeLightCfg
	ThemeLightNoPaddingCfg.Name = "light-no-padding"
	ThemeLightNoPaddingCfg.Padding = PaddingNone
	ThemeLightNoPaddingCfg.SizeBorder = 0
	ThemeLightNoPaddingCfg.Radius = RadiusNone
	ThemeLightNoPadding = ThemeMaker(ThemeLightNoPaddingCfg)

	// Light bordered.
	ThemeLightBorderedCfg = ThemeLightCfg
	ThemeLightBorderedCfg.Name = "light-bordered"
	ThemeLightBorderedCfg.SizeBorder = SizeBorderDef
	ThemeLightBordered = ThemeMaker(ThemeLightBorderedCfg)

	// Blue bordered.
	ThemeBlueBorderedCfg = baseCfg()
	ThemeBlueBorderedCfg.Name = "blue-dark-bordered"
	ThemeBlueBorderedCfg.ColorBackground = ColorFromString("#151C30")
	ThemeBlueBorderedCfg.ColorPanel = ColorFromString("#1C243F")
	ThemeBlueBorderedCfg.ColorInterior = ColorFromString("#202A49")
	ThemeBlueBorderedCfg.ColorHover = ColorFromString("#243054")
	ThemeBlueBorderedCfg.ColorFocus = ColorFromString("#29365E")
	ThemeBlueBorderedCfg.ColorActive = ColorFromString("#2D3C68")
	ThemeBlueBorderedCfg.ColorBorder = ColorFromString("#364263")
	ThemeBlueBorderedCfg.ColorBorderFocus = ColorFromString("#617AC3")
	ThemeBlueBorderedCfg.ColorSelect = ColorFromString("#3E65D8")
	ThemeBlueBorderedCfg.ColorError = RGBA(218, 54, 51, 255)
	ThemeBlueBorderedCfg.TitlebarDark = true
	ThemeBlueBorderedCfg.TextStyleDef = TextStyle{
		Family: defaultFontFamily,
		Color:  ColorFromString("#E1E1E1"),
		Size:   SizeTextMedium,
	}
	ThemeBlueBorderedCfg.SizeBorder = SizeBorderDef
	ThemeBlueBordered = ThemeMaker(ThemeBlueBorderedCfg)

	// Register all preset themes.
	ThemeRegister(ThemeDark)
	ThemeRegister(ThemeDarkNoPadding)
	ThemeRegister(ThemeDarkBordered)
	ThemeRegister(ThemeLight)
	ThemeRegister(ThemeLightNoPadding)
	ThemeRegister(ThemeLightBordered)
	ThemeRegister(ThemeBlueBordered)

	// Set default active theme to dark.
	guiTheme = ThemeDark
}
