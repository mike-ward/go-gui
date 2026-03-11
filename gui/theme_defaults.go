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

// baseDarkCfg returns the common dark ThemeCfg.
func baseDarkCfg() ThemeCfg {
	return ThemeCfg{
		Name:                 "dark",
		ColorBackground:      colorBackgroundDark,
		ColorPanel:           colorPanelDark,
		ColorInterior:        colorInteriorDark,
		ColorHover:           colorHoverDark,
		ColorFocus:           colorFocusDark,
		ColorActive:          colorActiveDark,
		ColorBorder:          colorBorderDark,
		ColorBorderFocus:     colorSelectDark,
		ColorSelect:          colorSelectDark,
		TitlebarDark:         true,
		TextStyleDef:         DefaultTextStyle,
		MonoFontFamily:       defaultMonoFontFamily,
		Padding:              PaddingMedium,
		PaddingSmall:         PaddingSmall,
		PaddingMedium:        PaddingMedium,
		PaddingLarge:         PaddingLarge,
		Radius:               RadiusMedium,
		RadiusSmall:          RadiusSmall,
		RadiusMedium:         RadiusMedium,
		RadiusLarge:          RadiusLarge,
		SpacingSmall:         SpacingSmall,
		SpacingMedium:        SpacingMedium,
		SpacingLarge:         SpacingLarge,
		SizeTextTiny:         SizeTextTiny,
		SizeTextXSmall:       SizeTextXSmall,
		SizeTextSmall:        SizeTextSmall,
		SizeTextMedium:       SizeTextMedium,
		SizeTextLarge:        SizeTextLarge,
		SizeTextXLarge:       SizeTextXLarge,
		ScrollMultiplier:     scrollMultiplier,
		ScrollDeltaLine:      scrollDeltaLine,
		ScrollDeltaPage:      scrollDeltaPage,
		SizeSwitchWidth:      36,
		SizeSwitchHeight:     22,
		SizeRadio:            16,
		SizeScrollbar:        7,
		SizeScrollbarMin:     20,
		SizeProgressBar:      20,
		SizeRangeSlider:      6,
		SizeRangeSliderThumb: 16,
	}
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
	ThemeLightCfg = ThemeCfg{
		Name:             "light",
		ColorBackground:  colorBackgroundLight,
		ColorPanel:       colorPanelLight,
		ColorInterior:    colorInteriorLight,
		ColorHover:       colorHoverLight,
		ColorFocus:       colorFocusLight,
		ColorActive:      colorActiveLight,
		ColorBorder:      colorBorderLight,
		ColorBorderFocus: colorBorderFocusLight,
		ColorSelect:      colorSelectLight,
		TextStyleDef: TextStyle{
			Color: colorTextLight,
			Size:  SizeTextMedium,
		},
		MonoFontFamily:       defaultMonoFontFamily,
		Padding:              PaddingMedium,
		PaddingSmall:         PaddingSmall,
		PaddingMedium:        PaddingMedium,
		PaddingLarge:         PaddingLarge,
		Radius:               RadiusMedium,
		RadiusSmall:          RadiusSmall,
		RadiusMedium:         RadiusMedium,
		RadiusLarge:          RadiusLarge,
		SpacingSmall:         SpacingSmall,
		SpacingMedium:        SpacingMedium,
		SpacingLarge:         SpacingLarge,
		SizeTextTiny:         SizeTextTiny,
		SizeTextXSmall:       SizeTextXSmall,
		SizeTextSmall:        SizeTextSmall,
		SizeTextMedium:       SizeTextMedium,
		SizeTextLarge:        SizeTextLarge,
		SizeTextXLarge:       SizeTextXLarge,
		ScrollMultiplier:     scrollMultiplier,
		ScrollDeltaLine:      scrollDeltaLine,
		ScrollDeltaPage:      scrollDeltaPage,
		SizeSwitchWidth:      36,
		SizeSwitchHeight:     22,
		SizeRadio:            16,
		SizeScrollbar:        7,
		SizeScrollbarMin:     20,
		SizeProgressBar:      20,
		SizeRangeSlider:      6,
		SizeRangeSliderThumb: 16,
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
	ThemeBlueBorderedCfg = ThemeCfg{
		Name:             "blue-dark-bordered",
		ColorBackground:  ColorFromString("#151C30"),
		ColorPanel:       ColorFromString("#1C243F"),
		ColorInterior:    ColorFromString("#202A49"),
		ColorHover:       ColorFromString("#243054"),
		ColorFocus:       ColorFromString("#29365E"),
		ColorActive:      ColorFromString("#2D3C68"),
		ColorBorder:      ColorFromString("#364263"),
		ColorBorderFocus: ColorFromString("#617AC3"),
		ColorSelect:      ColorFromString("#3E65D8"),
		TitlebarDark:     true,
		TextStyleDef: TextStyle{
			Color: ColorFromString("#E1E1E1"),
			Size:  16,
		},
		SizeBorder:           SizeBorderDef,
		Padding:              PaddingMedium,
		PaddingSmall:         PaddingSmall,
		PaddingMedium:        PaddingMedium,
		PaddingLarge:         PaddingLarge,
		Radius:               RadiusMedium,
		RadiusSmall:          RadiusSmall,
		RadiusMedium:         RadiusMedium,
		RadiusLarge:          RadiusLarge,
		SpacingSmall:         SpacingSmall,
		SpacingMedium:        SpacingMedium,
		SpacingLarge:         SpacingLarge,
		SizeTextTiny:         SizeTextTiny,
		SizeTextXSmall:       SizeTextXSmall,
		SizeTextSmall:        SizeTextSmall,
		SizeTextMedium:       SizeTextMedium,
		SizeTextLarge:        SizeTextLarge,
		SizeTextXLarge:       SizeTextXLarge,
		ScrollMultiplier:     scrollMultiplier,
		ScrollDeltaLine:      scrollDeltaLine,
		ScrollDeltaPage:      scrollDeltaPage,
		SizeSwitchWidth:      36,
		SizeSwitchHeight:     22,
		SizeRadio:            16,
		SizeScrollbar:        7,
		SizeScrollbarMin:     20,
		SizeProgressBar:      20,
		SizeRangeSlider:      6,
		SizeRangeSliderThumb: 16,
	}
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
