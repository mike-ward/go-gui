package gui

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
)
