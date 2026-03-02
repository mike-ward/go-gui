package gui

import "github.com/mike-ward/go-glyph"

// Theme default colors (dark theme).
var (
	colorBackgroundDark = RGB(48, 48, 48)
	colorPanelDark      = RGB(64, 64, 64)
	colorInteriorDark   = RGB(74, 74, 74)
	colorHoverDark      = RGB(84, 84, 84)
	colorFocusDark      = RGB(94, 94, 94)
	colorActiveDark     = RGB(104, 104, 104)
	colorBorderDark     = RGB(100, 100, 100)
	colorSelectDark     = RGB(65, 105, 225)
	colorTextDark       = RGB(225, 225, 225)
)

// Radius constants.
const (
	RadiusNone   float32 = 0
	RadiusSmall  float32 = 3.5
	RadiusMedium float32 = 5.5
	RadiusLarge  float32 = 7.5
)

// Text size constants.
const (
	SizeTextMedium float32 = 16
	SizeTextTiny   float32 = SizeTextMedium - 6
	SizeTextXSmall float32 = SizeTextMedium - 4
	SizeTextSmall  float32 = SizeTextMedium - 2
	SizeTextLarge  float32 = SizeTextMedium + 4
	SizeTextXLarge float32 = SizeTextMedium + 8
	SizeBorderDef  float32 = 1.5
)

// Spacing constants.
const (
	SpacingSmall  float32 = 5
	SpacingMedium float32 = 10
	SpacingLarge  float32 = 15
)

// TextStyle defines text rendering properties.
type TextStyle struct {
	Family        string
	Color         Color
	BgColor       Color
	Size          float32
	LineSpacing   float32
	LetterSpacing float32
	Align         TextAlignment
	Underline     bool
	Strikethrough bool
	Typeface      glyph.Typeface
	StrokeWidth   float32
	StrokeColor   Color
	Features      *glyph.FontFeatures
}

// ToGlyphStyle converts a gui TextStyle to a glyph.TextStyle.
func (ts TextStyle) ToGlyphStyle() glyph.TextStyle {
	return glyph.TextStyle{
		FontName:      ts.Family,
		Color:         glyph.Color(ts.Color),
		BgColor:       glyph.Color(ts.BgColor),
		Size:          ts.Size,
		Features:      ts.Features,
		Underline:     ts.Underline,
		Strikethrough: ts.Strikethrough,
		Typeface:      ts.Typeface,
		StrokeWidth:   ts.StrokeWidth,
		StrokeColor:   glyph.Color(ts.StrokeColor),
	}
}

// ButtonStyle defines button visual properties.
type ButtonStyle struct {
	Color            Color
	ColorHover       Color
	ColorFocus       Color
	ColorClick       Color
	ColorBorder      Color
	ColorBorderFocus Color
	Padding          Padding
	SizeBorder       float32
	Radius           float32
	BlurRadius       float32
	Shadow           *BoxShadow
	Gradient         *GradientDef
}

// ContainerStyle defines container visual properties.
type ContainerStyle struct {
	Color          Color
	ColorBorder    Color
	Padding        Padding
	Radius         float32
	BlurRadius     float32
	Spacing        float32
	SizeBorder     float32
	Shadow         *BoxShadow
	Gradient       *GradientDef
	BorderGradient *GradientDef
}

// RectangleStyle defines rectangle visual properties.
type RectangleStyle struct {
	Color          Color
	ColorBorder    Color
	Radius         float32
	BlurRadius     float32
	SizeBorder     float32
	Shadow         *BoxShadow
	Gradient       *GradientDef
	BorderGradient *GradientDef
}

// Default styles (dark theme).
var (
	DefaultTextStyle = TextStyle{
		Color: colorTextDark,
		Size:  SizeTextMedium,
	}

	DefaultButtonStyle = ButtonStyle{
		Color:            colorInteriorDark,
		ColorHover:       colorHoverDark,
		ColorFocus:       colorActiveDark,
		ColorClick:       colorActiveDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		Padding:          PaddingButton,
		SizeBorder:       SizeBorderDef,
		Radius:           RadiusMedium,
	}

	DefaultContainerStyle = ContainerStyle{
		Color:       ColorTransparent,
		ColorBorder: ColorTransparent,
		Padding:     PaddingMedium,
		Radius:      RadiusMedium,
		Spacing:     SpacingMedium,
		SizeBorder:  SizeBorderDef,
	}

	DefaultRectangleStyle = RectangleStyle{
		Color:       ColorTransparent,
		ColorBorder: colorBorderDark,
		Radius:      RadiusMedium,
		SizeBorder:  SizeBorderDef,
	}

	DefaultDataGridStyle DataGridStyle
)

// DataGridStyle defines data grid visual properties.
type DataGridStyle struct {
	ColorBackground   Color
	ColorHeader       Color
	ColorHeaderHover  Color
	ColorFilter       Color
	ColorQuickFilter  Color
	ColorRowHover     Color
	ColorRowAlt       Color
	ColorRowSelected  Color
	ColorBorder       Color
	ColorResizeHandle Color
	ColorResizeActive Color
	PaddingCell       Padding
	PaddingHeader     Padding
	PaddingFilter     Padding
	SizeBorder        float32
	Radius            float32
	TextStyle         TextStyle
	TextStyleHeader   TextStyle
	TextStyleFilter   TextStyle
}
