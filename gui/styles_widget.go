package gui

import "time"

// InputStyle defines input field visual properties.
type InputStyle struct {
	Color            Color
	ColorHover       Color
	ColorFocus       Color
	ColorClick       Color
	ColorBorder      Color
	ColorBorderFocus Color
	Padding          Padding
	SizeBorder       float32
	Radius           float32
	Shadow           *BoxShadow
	TextStyleNormal  TextStyle
	PlaceholderStyle TextStyle
}

// ScrollbarStyle defines scrollbar visual properties.
type ScrollbarStyle struct {
	Size            float32
	MinThumbSize    float32
	ColorThumb      Color
	ColorBackground Color
	Radius          float32
	RadiusThumb     float32
	GapEdge         float32
	GapEnd          float32
}

// RadioStyle defines radio button visual properties.
type RadioStyle struct {
	Size             float32
	Color            Color
	ColorHover       Color
	ColorFocus       Color
	ColorClick       Color
	ColorBorder      Color
	ColorBorderFocus Color
	ColorSelect      Color
	ColorUnselect    Color
	Padding          Padding
	SizeBorder       float32
	TextStyleNormal  TextStyle
}

// SwitchStyle defines switch toggle visual properties.
type SwitchStyle struct {
	SizeWidth        float32
	SizeHeight       float32
	Color            Color
	ColorClick       Color
	ColorFocus       Color
	ColorHover       Color
	ColorBorder      Color
	ColorBorderFocus Color
	ColorSelect      Color
	ColorUnselect    Color
	Padding          Padding
	SizeBorder       float32
	Radius           float32
	Shadow           *BoxShadow
	TextStyleNormal  TextStyle
}

// ToggleStyle defines toggle button visual properties.
type ToggleStyle struct {
	Color            Color
	ColorBorder      Color
	ColorBorderFocus Color
	ColorClick       Color
	ColorFocus       Color
	ColorHover       Color
	ColorSelect      Color
	Padding          Padding
	SizeBorder       float32
	Radius           float32
	Shadow           *BoxShadow
	TextStyleNormal  TextStyle
	TextStyleLabel   TextStyle
}

// SelectStyle defines select dropdown visual properties.
type SelectStyle struct {
	MinWidth         float32
	MaxWidth         float32
	Color            Color
	ColorHover       Color
	ColorFocus       Color
	ColorClick       Color
	ColorBorder      Color
	ColorBorderFocus Color
	ColorSelect      Color
	Padding          Padding
	SizeBorder       float32
	Radius           float32
	Shadow           *BoxShadow
	TextStyleNormal  TextStyle
	SubheadingStyle  TextStyle
	PlaceholderStyle TextStyle
}

// ListBoxStyle defines list box visual properties.
type ListBoxStyle struct {
	Color            Color
	ColorHover       Color
	ColorFocus       Color
	ColorBorder      Color
	ColorBorderFocus Color
	ColorSelect      Color
	Padding          Padding
	SizeBorder       float32
	Radius           float32
	Shadow           *BoxShadow
	TextStyleNormal  TextStyle
	SubheadingStyle  TextStyle
}

// DialogStyle defines dialog visual properties.
type DialogStyle struct {
	Color            Color
	ColorBorder      Color
	ColorBorderFocus Color
	Padding          Padding
	SizeBorder       float32
	Radius           float32
	RadiusBorder     float32
	BlurRadius       float32
	Shadow           *BoxShadow
	AlignButtons     HorizontalAlign
	TitleTextStyle   TextStyle
	TextStyle        TextStyle
}

// ToastAnchor specifies toast notification position.
type ToastAnchor uint8

const (
	ToastTopLeft ToastAnchor = iota
	ToastTopRight
	ToastBottomLeft
	ToastBottomRight
)

// ToastStyle defines toast notification visual properties.
type ToastStyle struct {
	MaxVisible  int
	Anchor      ToastAnchor
	Width       float32
	Margin      float32
	Spacing     float32
	AccentWidth float32
	Padding     Padding
	Radius      float32
	SizeBorder  float32
	Color       Color
	ColorBorder Color
	ColorInfo    Color
	ColorSuccess Color
	ColorWarning Color
	ColorError   Color
	TextStyle   TextStyle
	TitleStyle  TextStyle
	Shadow      *BoxShadow
}

// TooltipStyle defines tooltip visual properties.
type TooltipStyle struct {
	Delay            time.Duration
	Color            Color
	ColorHover       Color
	ColorFocus       Color
	ColorClick       Color
	ColorBorder      Color
	ColorBorderFocus Color
	Padding          Padding
	SizeBorder       float32
	Radius           float32
	RadiusBorder     float32
	Shadow           *BoxShadow
	TextStyle        TextStyle
}

// BadgeStyle defines badge visual properties.
type BadgeStyle struct {
	Color        Color
	ColorInfo    Color
	ColorSuccess Color
	ColorWarning Color
	ColorError   Color
	Padding      Padding
	Radius       float32
	TextStyle    TextStyle
	DotSize      float32
}

// ExpandPanelStyle defines expand panel visual properties.
type ExpandPanelStyle struct {
	Color        Color
	ColorHover   Color
	ColorClick   Color
	ColorBorder  Color
	Padding      Padding
	SizeBorder   float32
	Radius       float32
	RadiusBorder float32
}

// ProgressBarStyle defines progress bar visual properties.
type ProgressBarStyle struct {
	Size           float32
	Color          Color
	ColorBar       Color
	ColorBorder    Color
	TextBackground Color
	Padding        Padding
	TextPadding    Padding
	SizeBorder     float32
	Radius         float32
	TextShow       bool
	TextStyle      TextStyle
}

// RangeSliderStyle defines range slider visual properties.
type RangeSliderStyle struct {
	Size             float32
	ThumbSize        float32
	Color            Color
	ColorClick       Color
	ColorThumb       Color
	ColorLeft        Color
	ColorFocus       Color
	ColorHover       Color
	ColorBorder      Color
	ColorBorderFocus Color
	Padding          Padding
	SizeBorder       float32
	Radius           float32
}

// Default widget styles (dark theme).
var (
	DefaultInputStyle = InputStyle{
		Color:            colorInteriorDark,
		ColorHover:       colorHoverDark,
		ColorFocus:       colorActiveDark,
		ColorClick:       colorActiveDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		Padding:          PaddingSmall,
		SizeBorder:       SizeBorderDef,
		Radius:           RadiusMedium,
		TextStyleNormal:  DefaultTextStyle,
		PlaceholderStyle: TextStyle{
			Color: RGBA(colorTextDark.R, colorTextDark.G, colorTextDark.B, 100),
			Size:  SizeTextMedium,
		},
	}

	DefaultScrollbarStyle = ScrollbarStyle{
		Size:            7,
		MinThumbSize:    20,
		ColorThumb:      colorActiveDark,
		ColorBackground: ColorTransparent,
		Radius:          RadiusSmall,
		RadiusThumb:     RadiusSmall,
		GapEdge:         3,
		GapEnd:          2,
	}

	DefaultRadioStyle = RadioStyle{
		Size:             SizeTextMedium,
		Color:            colorPanelDark,
		ColorHover:       colorHoverDark,
		ColorFocus:       colorSelectDark,
		ColorClick:       colorActiveDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		ColorSelect:      colorSelectDark,
		ColorUnselect:    colorActiveDark,
		Padding:          PadAll(4),
		SizeBorder:       2.0,
		TextStyleNormal:  DefaultTextStyle,
	}

	DefaultSwitchStyle = SwitchStyle{
		SizeWidth:        36,
		SizeHeight:       22,
		Color:            colorPanelDark,
		ColorClick:       colorInteriorDark,
		ColorFocus:       colorFocusDark,
		ColorHover:       colorHoverDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		ColorSelect:      colorSelectDark,
		ColorUnselect:    colorActiveDark,
		Padding:          PaddingThree,
		SizeBorder:       SizeBorderDef,
		Radius:           RadiusLarge * 2,
		TextStyleNormal:  DefaultTextStyle,
	}

	DefaultToggleStyle = ToggleStyle{
		Color:            colorPanelDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		ColorClick:       colorInteriorDark,
		ColorFocus:       colorActiveDark,
		ColorHover:       colorHoverDark,
		ColorSelect:      colorInteriorDark,
		Padding:          NewPadding(1, 1, 1, 2),
		SizeBorder:       SizeBorderDef,
		Radius:           RadiusSmall,
		TextStyleNormal:  DefaultTextStyle,
		TextStyleLabel:   DefaultTextStyle,
	}

	DefaultSelectStyle = SelectStyle{
		MinWidth:         75,
		MaxWidth:         200,
		Color:            colorInteriorDark,
		ColorHover:       colorHoverDark,
		ColorFocus:       colorFocusDark,
		ColorClick:       colorActiveDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		ColorSelect:      colorSelectDark,
		Padding:          PaddingSmall,
		SizeBorder:       SizeBorderDef,
		Radius:           RadiusMedium,
		TextStyleNormal:  DefaultTextStyle,
		SubheadingStyle:  DefaultTextStyle,
		PlaceholderStyle: TextStyle{
			Color: RGBA(colorTextDark.R, colorTextDark.G, colorTextDark.B, 100),
			Size:  SizeTextMedium,
		},
	}

	DefaultListBoxStyle = ListBoxStyle{
		Color:            colorInteriorDark,
		ColorHover:       colorHoverDark,
		ColorFocus:       colorFocusDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		ColorSelect:      colorSelectDark,
		Padding:          PaddingButton,
		SizeBorder:       SizeBorderDef,
		Radius:           RadiusMedium,
		TextStyleNormal:  DefaultTextStyle,
		SubheadingStyle:  DefaultTextStyle,
	}

	DefaultDialogStyle = DialogStyle{
		Color:            colorPanelDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		Padding:          PaddingLarge,
		SizeBorder:       SizeBorderDef,
		Radius:           RadiusMedium,
		RadiusBorder:     RadiusMedium,
		AlignButtons:     HAlignCenter,
		TitleTextStyle: TextStyle{
			Color: colorTextDark,
			Size:  SizeTextLarge,
		},
		TextStyle: DefaultTextStyle,
	}

	DefaultToastStyle = ToastStyle{
		MaxVisible:   5,
		Anchor:       ToastBottomRight,
		Width:        260,
		Margin:       16,
		Spacing:      8,
		AccentWidth:  4,
		Padding:      PaddingMedium,
		Radius:       RadiusMedium,
		SizeBorder:   SizeBorderDef,
		Color:        colorPanelDark,
		ColorBorder:  colorBorderDark,
		ColorInfo:    colorSelectDark,
		ColorSuccess: RGBA(46, 160, 67, 255),
		ColorWarning: RGBA(210, 153, 34, 255),
		ColorError:   RGBA(218, 54, 51, 255),
		TextStyle:    DefaultTextStyle,
		TitleStyle: TextStyle{
			Color: colorTextDark,
			Size:  SizeTextMedium,
		},
	}

	DefaultTooltipStyle = TooltipStyle{
		Delay:            500 * time.Millisecond,
		Color:            colorInteriorDark,
		ColorHover:       colorHoverDark,
		ColorFocus:       colorActiveDark,
		ColorClick:       colorActiveDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		Padding:          PaddingSmall,
		SizeBorder:       SizeBorderDef,
		Radius:           RadiusSmall,
		RadiusBorder:     RadiusSmall,
		TextStyle:        DefaultTextStyle,
	}

	DefaultBadgeStyle = BadgeStyle{
		Color:        colorActiveDark,
		ColorInfo:    colorSelectDark,
		ColorSuccess: RGBA(46, 160, 67, 255),
		ColorWarning: RGBA(210, 153, 34, 255),
		ColorError:   RGBA(218, 54, 51, 255),
		Padding:      NewPadding(2, 8, 2, 8),
		Radius:       RadiusSmall,
		TextStyle:    DefaultTextStyle,
		DotSize:      8,
	}

	DefaultExpandPanelStyle = ExpandPanelStyle{
		Color:        colorPanelDark,
		ColorHover:   colorHoverDark,
		ColorClick:   colorActiveDark,
		ColorBorder:  colorBorderDark,
		Padding:      PaddingMedium,
		SizeBorder:   SizeBorderDef,
		Radius:       RadiusMedium,
		RadiusBorder: RadiusMedium,
	}

	DefaultProgressBarStyle = ProgressBarStyle{
		Size:           20,
		Color:          colorInteriorDark,
		ColorBar:       colorSelectDark,
		ColorBorder:    colorBorderDark,
		TextBackground: colorPanelDark,
		Padding:        PaddingNone,
		TextPadding:    NewPadding(1, 4, 1, 4),
		SizeBorder:     0,
		Radius:         RadiusSmall,
		TextShow:       true,
		TextStyle:      DefaultTextStyle,
	}

	DefaultRangeSliderStyle = RangeSliderStyle{
		Size:             20,
		ThumbSize:        16,
		Color:            colorInteriorDark,
		ColorClick:       colorActiveDark,
		ColorThumb:       colorPanelDark,
		ColorLeft:        colorSelectDark,
		ColorFocus:       colorSelectDark,
		ColorHover:       colorHoverDark,
		ColorBorder:      colorBorderDark,
		ColorBorderFocus: colorSelectDark,
		Padding:          PaddingNone,
		SizeBorder:       SizeBorderDef,
		Radius:           RadiusSmall,
	}
)
