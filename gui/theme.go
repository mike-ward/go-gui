package gui

import (
	"errors"
	"time"
)

// guiTheme is the package-level active theme.
var guiTheme Theme

// Theme describes a complete GUI theme. Only styles for existing
// Go views are populated (Button, Container, Rectangle, Text,
// Input, Scrollbar, Radio, Switch, Toggle, Select, ListBox).
type Theme struct {
	Cfg             ThemeCfg
	Name            string
	ColorBackground Color
	ColorPanel      Color
	ColorInterior   Color
	ColorHover      Color
	ColorFocus      Color
	ColorActive     Color
	ColorBorder     Color
	ColorSelect     Color
	TitlebarDark    bool

	// Per-widget styles.
	ButtonStyle    ButtonStyle
	ContainerStyle ContainerStyle
	RectangleStyle RectangleStyle
	TextStyleDef   TextStyle
	InputStyle     InputStyle
	ScrollbarStyle ScrollbarStyle
	RadioStyle     RadioStyle
	SwitchStyle    SwitchStyle
	ToggleStyle    ToggleStyle
	SelectStyle    SelectStyle
	ListBoxStyle   ListBoxStyle
	DialogStyle    DialogStyle
	ToastStyle        ToastStyle
	TooltipStyle      TooltipStyle
	BadgeStyle        BadgeStyle
	ExpandPanelStyle  ExpandPanelStyle
	ProgressBarStyle  ProgressBarStyle
	RangeSliderStyle  RangeSliderStyle
	TabControlStyle   TabControlStyle
	BreadcrumbStyle   BreadcrumbStyle
	SplitterStyle     SplitterStyle
	TableStyle        TableStyle

	// Text size shortcuts (N = normal, B = bold).
	N1 TextStyle
	N2 TextStyle
	N3 TextStyle
	N4 TextStyle
	N5 TextStyle
	N6 TextStyle
	B1 TextStyle
	B2 TextStyle
	B3 TextStyle
	B4 TextStyle
	B5 TextStyle
	B6 TextStyle

	// Layout constants.
	PaddingSmall  Padding
	PaddingMedium Padding
	PaddingLarge  Padding
	SizeBorder    float32

	RadiusSmall  float32
	RadiusMedium float32
	RadiusLarge  float32

	SpacingSmall  float32
	SpacingMedium float32
	SpacingLarge  float32

	SizeTextTiny   float32
	SizeTextXSmall float32
	SizeTextSmall  float32
	SizeTextMedium float32
	SizeTextLarge  float32
	SizeTextXLarge float32

	ScrollMultiplier float32
	ScrollDeltaLine  float32
	ScrollDeltaPage  float32
}

// ThemeCfg is the configuration struct for ThemeMaker.
type ThemeCfg struct {
	Name             string
	ColorBackground  Color
	ColorPanel       Color
	ColorInterior    Color
	ColorHover       Color
	ColorFocus       Color
	ColorActive      Color
	ColorBorder      Color
	ColorBorderFocus Color
	ColorSelect      Color
	ColorSuccess     Color
	ColorWarning     Color
	ColorError       Color
	TitlebarDark     bool
	Fill             bool
	FillBorder       bool
	Padding          Padding
	SizeBorder       float32
	Radius           float32
	TextStyleDef     TextStyle

	PaddingSmall  Padding
	PaddingMedium Padding
	PaddingLarge  Padding

	RadiusSmall  float32
	RadiusMedium float32
	RadiusLarge  float32

	SpacingSmall  float32
	SpacingMedium float32
	SpacingLarge  float32

	SizeTextTiny   float32
	SizeTextXSmall float32
	SizeTextSmall  float32
	SizeTextMedium float32
	SizeTextLarge  float32
	SizeTextXLarge float32

	ScrollMultiplier float32
	ScrollDeltaLine  float32
	ScrollDeltaPage  float32

	SizeSwitchWidth      float32
	SizeSwitchHeight     float32
	SizeRadio            float32
	SizeScrollbar        float32
	SizeScrollbarMin     float32
	SizeProgressBar      float32
	SizeRangeSlider      float32
	SizeRangeSliderThumb float32
}

// ThemeMaker builds a full Theme from a ThemeCfg.
func ThemeMaker(cfg ThemeCfg) Theme {
	ts := cfg.TextStyleDef
	makeStyle := func(base TextStyle, size float32) TextStyle {
		s := base
		s.Size = size
		return s
	}

	borderFocus := cfg.ColorBorderFocus
	if borderFocus.Eq(Color{}) {
		borderFocus = cfg.ColorSelect
	}

	// Scrollbar radius: none if cfg.Radius is none.
	sbRadius := cfg.RadiusSmall
	if cfg.Radius == RadiusNone {
		sbRadius = RadiusNone
	}

	placeholderColor := RGBA(ts.Color.R, ts.Color.G, ts.Color.B, 100)

	theme := Theme{
		Cfg:             cfg,
		Name:            cfg.Name,
		ColorBackground: cfg.ColorBackground,
		ColorPanel:      cfg.ColorPanel,
		ColorInterior:   cfg.ColorInterior,
		ColorHover:      cfg.ColorHover,
		ColorFocus:      cfg.ColorFocus,
		ColorActive:     cfg.ColorActive,
		ColorBorder:     cfg.ColorBorder,
		ColorSelect:     cfg.ColorSelect,
		TitlebarDark:    cfg.TitlebarDark,

		ButtonStyle: ButtonStyle{
			Color:            cfg.ColorInterior,
			ColorHover:       cfg.ColorHover,
			ColorFocus:       cfg.ColorActive,
			ColorClick:       cfg.ColorFocus,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			Padding:          PaddingButton,
			SizeBorder:       cfg.SizeBorder,
			Radius:           cfg.Radius,
		},
		ContainerStyle: ContainerStyle{
			Color:       ColorTransparent,
			ColorBorder: ColorTransparent,
			Padding:     cfg.Padding,
			Radius:      cfg.Radius,
			Spacing:     cfg.SpacingMedium,
			SizeBorder:  cfg.SizeBorder,
		},
		RectangleStyle: RectangleStyle{
			Color:       ColorTransparent,
			ColorBorder: cfg.ColorBorder,
			Radius:      cfg.Radius,
			SizeBorder:  cfg.SizeBorder,
		},
		TextStyleDef: ts,
		InputStyle: InputStyle{
			Color:            cfg.ColorInterior,
			ColorHover:       cfg.ColorHover,
			ColorFocus:       cfg.ColorInterior,
			ColorClick:       cfg.ColorActive,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			Padding:          cfg.Padding,
			SizeBorder:       cfg.SizeBorder,
			Radius:           cfg.Radius,
			TextStyleNormal:  ts,
			PlaceholderStyle: TextStyle{
				Color: placeholderColor,
				Size:  ts.Size,
			},
		},
		ScrollbarStyle: ScrollbarStyle{
			Size:            cfg.SizeScrollbar,
			MinThumbSize:    cfg.SizeScrollbarMin,
			ColorThumb:      cfg.ColorActive,
			ColorBackground: ColorTransparent,
			Radius:          sbRadius,
			RadiusThumb:     sbRadius,
			GapEdge:         3,
			GapEnd:          2,
		},
		RadioStyle: RadioStyle{
			Size:             cfg.SizeRadio,
			Color:            cfg.ColorPanel,
			ColorHover:       cfg.ColorHover,
			ColorFocus:       cfg.ColorSelect,
			ColorClick:       cfg.ColorActive,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			ColorSelect:      cfg.ColorSelect,
			ColorUnselect:    cfg.ColorActive,
			Padding:          PadAll(4),
			SizeBorder:       cfg.SizeBorder,
			TextStyleNormal:  ts,
		},
		SwitchStyle: SwitchStyle{
			SizeWidth:        cfg.SizeSwitchWidth,
			SizeHeight:       cfg.SizeSwitchHeight,
			Color:            cfg.ColorPanel,
			ColorClick:       cfg.ColorInterior,
			ColorFocus:       cfg.ColorInterior,
			ColorHover:       cfg.ColorHover,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			ColorSelect:      cfg.ColorSelect,
			ColorUnselect:    cfg.ColorActive,
			Padding:          PaddingThree,
			SizeBorder:       cfg.SizeBorder,
			Radius:           RadiusLarge * 2,
			TextStyleNormal:  ts,
		},
		ToggleStyle: ToggleStyle{
			Color:            cfg.ColorPanel,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			ColorClick:       cfg.ColorInterior,
			ColorFocus:       cfg.ColorInterior,
			ColorHover:       cfg.ColorHover,
			ColorSelect:      cfg.ColorInterior,
			Padding:          NewPadding(1, 1, 1, 2),
			SizeBorder:       cfg.SizeBorder,
			Radius:           cfg.Radius,
			TextStyleNormal:  ts,
			TextStyleLabel:   ts,
		},
		SelectStyle: SelectStyle{
			MinWidth:         75,
			MaxWidth:         200,
			Color:            cfg.ColorInterior,
			ColorHover:       cfg.ColorHover,
			ColorFocus:       cfg.ColorFocus,
			ColorClick:       cfg.ColorActive,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			ColorSelect:      cfg.ColorSelect,
			Padding:          PaddingSmall,
			SizeBorder:       cfg.SizeBorder,
			Radius:           cfg.RadiusMedium,
			TextStyleNormal:  ts,
			SubheadingStyle:  ts,
			PlaceholderStyle: TextStyle{
				Color: placeholderColor,
				Size:  ts.Size,
			},
		},
		ListBoxStyle: ListBoxStyle{
			Color:            cfg.ColorInterior,
			ColorHover:       cfg.ColorHover,
			ColorFocus:       cfg.ColorFocus,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			ColorSelect:      cfg.ColorSelect,
			Padding:          cfg.Padding,
			SizeBorder:       cfg.SizeBorder,
			Radius:           cfg.Radius,
			TextStyleNormal:  ts,
			SubheadingStyle:  ts,
		},
		DialogStyle: DialogStyle{
			Color:            cfg.ColorPanel,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			Padding:          cfg.PaddingLarge,
			SizeBorder:       cfg.SizeBorder,
			Radius:           cfg.RadiusMedium,
			RadiusBorder:     cfg.RadiusMedium,
			AlignButtons:     HAlignCenter,
			TitleTextStyle:   makeStyle(ts, cfg.SizeTextLarge),
			TextStyle:        ts,
		},
		ToastStyle: ToastStyle{
			MaxVisible:   5,
			Anchor:       ToastBottomRight,
			Width:        260,
			Margin:       16,
			Spacing:      8,
			AccentWidth:  4,
			Padding:      cfg.PaddingMedium,
			Radius:       cfg.RadiusMedium,
			SizeBorder:   cfg.SizeBorder,
			Color:        cfg.ColorPanel,
			ColorBorder:  cfg.ColorBorder,
			ColorInfo:    cfg.ColorSelect,
			ColorSuccess: RGBA(46, 160, 67, 255),
			ColorWarning: RGBA(210, 153, 34, 255),
			ColorError:   RGBA(218, 54, 51, 255),
			TextStyle:    ts,
			TitleStyle:   makeStyle(ts, cfg.SizeTextMedium),
		},
		TooltipStyle: TooltipStyle{
			Delay:            500 * time.Millisecond,
			Color:            cfg.ColorInterior,
			ColorHover:       cfg.ColorHover,
			ColorFocus:       cfg.ColorActive,
			ColorClick:       cfg.ColorActive,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			Padding:          cfg.PaddingSmall,
			SizeBorder:       cfg.SizeBorder,
			Radius:           cfg.RadiusSmall,
			RadiusBorder:     cfg.RadiusSmall,
			TextStyle:        ts,
		},
		BadgeStyle: BadgeStyle{
			Color:        cfg.ColorActive,
			ColorInfo:    cfg.ColorSelect,
			ColorSuccess: RGBA(46, 160, 67, 255),
			ColorWarning: RGBA(210, 153, 34, 255),
			ColorError:   RGBA(218, 54, 51, 255),
			Padding:      NewPadding(2, 8, 2, 8),
			Radius:       cfg.RadiusSmall,
			TextStyle:    ts,
			DotSize:      8,
		},
		ExpandPanelStyle: ExpandPanelStyle{
			Color:        cfg.ColorPanel,
			ColorHover:   cfg.ColorHover,
			ColorClick:   cfg.ColorActive,
			ColorBorder:  cfg.ColorBorder,
			Padding:      cfg.PaddingMedium,
			SizeBorder:   cfg.SizeBorder,
			Radius:       cfg.RadiusMedium,
			RadiusBorder: cfg.RadiusMedium,
		},
		ProgressBarStyle: ProgressBarStyle{
			Size:           cfg.SizeProgressBar,
			Color:          cfg.ColorInterior,
			ColorBar:       cfg.ColorSelect,
			ColorBorder:    cfg.ColorBorder,
			TextBackground: cfg.ColorPanel,
			Padding:        PaddingNone,
			TextPadding:    NewPadding(1, 4, 1, 4),
			Radius:         cfg.RadiusSmall,
			TextShow:       true,
			TextStyle:      ts,
		},
		RangeSliderStyle: RangeSliderStyle{
			Size:             cfg.SizeRangeSlider,
			ThumbSize:        cfg.SizeRangeSliderThumb,
			Color:            cfg.ColorInterior,
			ColorClick:       cfg.ColorActive,
			ColorThumb:       cfg.ColorPanel,
			ColorLeft:        cfg.ColorSelect,
			ColorFocus:       cfg.ColorSelect,
			ColorHover:       cfg.ColorHover,
			ColorBorder:      cfg.ColorBorder,
			ColorBorderFocus: borderFocus,
			Padding:          PaddingNone,
			SizeBorder:       cfg.SizeBorder,
			Radius:           cfg.RadiusSmall,
		},
		TabControlStyle: TabControlStyle{
			Color:              cfg.ColorPanel,
			ColorBorder:        cfg.ColorBorder,
			ColorHeader:        ColorTransparent,
			ColorHeaderBorder:  ColorTransparent,
			ColorContent:       cfg.ColorPanel,
			ColorContentBorder: cfg.ColorBorder,
			ColorTab:           cfg.ColorInterior,
			ColorTabHover:      cfg.ColorHover,
			ColorTabFocus:      cfg.ColorFocus,
			ColorTabClick:      cfg.ColorActive,
			ColorTabSelected:   cfg.ColorSelect,
			ColorTabDisabled:   cfg.ColorPanel,
			ColorTabBorder:     cfg.ColorBorder,
			ColorTabBorderFocus: borderFocus,
			Padding:            PaddingNone,
			PaddingHeader:      PaddingNone,
			PaddingContent:     cfg.PaddingMedium,
			PaddingTab:         cfg.PaddingSmall,
			SizeBorder:         cfg.SizeBorder,
			SizeContentBorder:  cfg.SizeBorder,
			SizeTabBorder:      cfg.SizeBorder,
			Radius:             cfg.RadiusMedium,
			RadiusHeader:       cfg.RadiusSmall,
			RadiusContent:      cfg.RadiusMedium,
			RadiusTab:          cfg.RadiusSmall,
			RadiusTabBorder:    cfg.RadiusSmall,
			Spacing:            cfg.SpacingSmall,
			SpacingHeader:      cfg.SpacingSmall,
			TextStyle:          ts,
			TextStyleSelected:  ts,
			TextStyleDisabled: TextStyle{
				Color: RGBA(ts.Color.R, ts.Color.G, ts.Color.B, 130),
				Size:  ts.Size,
			},
		},
		BreadcrumbStyle: BreadcrumbStyle{
			Separator:          "/",
			Color:              ColorTransparent,
			ColorBorder:        ColorTransparent,
			ColorTrail:         ColorTransparent,
			ColorCrumb:         ColorTransparent,
			ColorCrumbHover:    cfg.ColorHover,
			ColorCrumbClick:    cfg.ColorActive,
			ColorCrumbSelected: ColorTransparent,
			ColorCrumbDisabled: ColorTransparent,
			ColorContent:       cfg.ColorPanel,
			ColorContentBorder: cfg.ColorBorder,
			Padding:            PaddingNone,
			PaddingTrail:       cfg.PaddingSmall,
			PaddingCrumb:       NewPadding(2, 4, 2, 4),
			PaddingContent:     cfg.PaddingMedium,
			Radius:             cfg.RadiusMedium,
			RadiusCrumb:        cfg.RadiusSmall,
			RadiusContent:      cfg.RadiusMedium,
			Spacing:            cfg.SpacingSmall,
			SpacingTrail:       cfg.SpacingSmall,
			SizeContentBorder:  cfg.SizeBorder,
			TextStyle:          ts,
			TextStyleSelected:  ts,
			TextStyleDisabled: TextStyle{
				Color: RGBA(ts.Color.R, ts.Color.G, ts.Color.B, 130),
				Size:  ts.Size,
			},
			TextStyleSeparator: TextStyle{
				Color: RGBA(ts.Color.R, ts.Color.G, ts.Color.B, 160),
				Size:  ts.Size,
			},
		},
		SplitterStyle: SplitterStyle{
			HandleSize:        9,
			DragStep:          0.02,
			DragStepLarge:     0.10,
			ColorHandle:       cfg.ColorInterior,
			ColorHandleHover:  cfg.ColorHover,
			ColorHandleActive: cfg.ColorActive,
			ColorHandleBorder: cfg.ColorBorder,
			ColorGrip:         cfg.ColorSelect,
			ColorButton:       cfg.ColorInterior,
			ColorButtonHover:  cfg.ColorHover,
			ColorButtonActive: cfg.ColorActive,
			ColorButtonIcon:   ts.Color,
			SizeBorder:        cfg.SizeBorder,
			Radius:            cfg.RadiusSmall,
			RadiusBorder:      cfg.RadiusSmall,
		},
		TableStyle: TableStyle{
			ColorBorder:        cfg.ColorBorder,
			ColorSelect:        cfg.ColorSelect,
			ColorHover:         cfg.ColorHover,
			CellPadding:        PaddingTwoFive,
			TextStyle:          ts,
			TextStyleHead:      ts,
			AlignHead:          HAlignCenter,
			ColumnWidthDefault: 50,
			ColumnWidthMin:     20,
		},

		// Layout constants.
		PaddingSmall:  cfg.PaddingSmall,
		PaddingMedium: cfg.PaddingMedium,
		PaddingLarge:  cfg.PaddingLarge,
		SizeBorder:    cfg.SizeBorder,

		RadiusSmall:  cfg.RadiusSmall,
		RadiusMedium: cfg.RadiusMedium,
		RadiusLarge:  cfg.RadiusLarge,

		SpacingSmall:  cfg.SpacingSmall,
		SpacingMedium: cfg.SpacingMedium,
		SpacingLarge:  cfg.SpacingLarge,

		SizeTextTiny:   cfg.SizeTextTiny,
		SizeTextXSmall: cfg.SizeTextXSmall,
		SizeTextSmall:  cfg.SizeTextSmall,
		SizeTextMedium: cfg.SizeTextMedium,
		SizeTextLarge:  cfg.SizeTextLarge,
		SizeTextXLarge: cfg.SizeTextXLarge,

		ScrollMultiplier: cfg.ScrollMultiplier,
		ScrollDeltaLine:  cfg.ScrollDeltaLine,
		ScrollDeltaPage:  cfg.ScrollDeltaPage,
	}

	// Text size shortcuts.
	normal := ts
	bold := ts // font variant resolution deferred
	theme.N1 = makeStyle(normal, theme.SizeTextXLarge)
	theme.N2 = makeStyle(normal, theme.SizeTextLarge)
	theme.N3 = ts
	theme.N4 = makeStyle(normal, theme.SizeTextSmall)
	theme.N5 = makeStyle(normal, theme.SizeTextXSmall)
	theme.N6 = makeStyle(normal, theme.SizeTextTiny)
	theme.B1 = makeStyle(bold, theme.SizeTextXLarge)
	theme.B2 = makeStyle(bold, theme.SizeTextLarge)
	theme.B3 = makeStyle(bold, theme.SizeTextMedium)
	theme.B4 = makeStyle(bold, theme.SizeTextSmall)
	theme.B5 = makeStyle(bold, theme.SizeTextXSmall)
	theme.B6 = makeStyle(bold, theme.SizeTextTiny)

	return theme
}

// CurrentTheme returns the active theme.
func CurrentTheme() Theme {
	return guiTheme
}

// SetTheme sets the active theme and updates all Default*Style vars.
func SetTheme(t Theme) {
	guiTheme = t
	DefaultTextStyle = t.TextStyleDef
	DefaultButtonStyle = t.ButtonStyle
	DefaultContainerStyle = t.ContainerStyle
	DefaultRectangleStyle = t.RectangleStyle
	DefaultInputStyle = t.InputStyle
	DefaultScrollbarStyle = t.ScrollbarStyle
	DefaultRadioStyle = t.RadioStyle
	DefaultSwitchStyle = t.SwitchStyle
	DefaultToggleStyle = t.ToggleStyle
	DefaultSelectStyle = t.SelectStyle
	DefaultListBoxStyle = t.ListBoxStyle
	DefaultDialogStyle = t.DialogStyle
	DefaultToastStyle = t.ToastStyle
	DefaultTooltipStyle = t.TooltipStyle
	DefaultBadgeStyle = t.BadgeStyle
	DefaultExpandPanelStyle = t.ExpandPanelStyle
	DefaultProgressBarStyle = t.ProgressBarStyle
	DefaultRangeSliderStyle = t.RangeSliderStyle
	DefaultTabControlStyle = t.TabControlStyle
	DefaultBreadcrumbStyle = t.BreadcrumbStyle
	DefaultSplitterStyle = t.SplitterStyle
	DefaultTableStyle = t.TableStyle
}

// With*Style methods for selective overrides.

func (t Theme) WithButtonStyle(s ButtonStyle) Theme {
	t.ButtonStyle = s
	return t
}

func (t Theme) WithContainerStyle(s ContainerStyle) Theme {
	t.ContainerStyle = s
	return t
}

func (t Theme) WithRectangleStyle(s RectangleStyle) Theme {
	t.RectangleStyle = s
	return t
}

func (t Theme) WithTextStyle(s TextStyle) Theme {
	t.TextStyleDef = s
	return t
}

func (t Theme) WithInputStyle(s InputStyle) Theme {
	t.InputStyle = s
	return t
}

func (t Theme) WithScrollbarStyle(s ScrollbarStyle) Theme {
	t.ScrollbarStyle = s
	return t
}

func (t Theme) WithRadioStyle(s RadioStyle) Theme {
	t.RadioStyle = s
	return t
}

func (t Theme) WithSwitchStyle(s SwitchStyle) Theme {
	t.SwitchStyle = s
	return t
}

func (t Theme) WithToggleStyle(s ToggleStyle) Theme {
	t.ToggleStyle = s
	return t
}

func (t Theme) WithSelectStyle(s SelectStyle) Theme {
	t.SelectStyle = s
	return t
}

func (t Theme) WithListBoxStyle(s ListBoxStyle) Theme {
	t.ListBoxStyle = s
	return t
}

func (t Theme) WithDialogStyle(s DialogStyle) Theme {
	t.DialogStyle = s
	return t
}

func (t Theme) WithToastStyle(s ToastStyle) Theme {
	t.ToastStyle = s
	return t
}

func (t Theme) WithTooltipStyle(s TooltipStyle) Theme {
	t.TooltipStyle = s
	return t
}

func (t Theme) WithBadgeStyle(s BadgeStyle) Theme {
	t.BadgeStyle = s
	return t
}

func (t Theme) WithExpandPanelStyle(s ExpandPanelStyle) Theme {
	t.ExpandPanelStyle = s
	return t
}

func (t Theme) WithProgressBarStyle(s ProgressBarStyle) Theme {
	t.ProgressBarStyle = s
	return t
}

func (t Theme) WithRangeSliderStyle(s RangeSliderStyle) Theme {
	t.RangeSliderStyle = s
	return t
}

func (t Theme) WithTabControlStyle(s TabControlStyle) Theme {
	t.TabControlStyle = s
	return t
}

func (t Theme) WithBreadcrumbStyle(s BreadcrumbStyle) Theme {
	t.BreadcrumbStyle = s
	return t
}

func (t Theme) WithSplitterStyle(s SplitterStyle) Theme {
	t.SplitterStyle = s
	return t
}

func (t Theme) WithTableStyle(s TableStyle) Theme {
	t.TableStyle = s
	return t
}

// ColorOverrides specifies which semantic colors to update across
// all widget styles. Nil pointers mean "keep existing".
type ColorOverrides struct {
	ColorBackground  *Color
	ColorPanel       *Color
	ColorInterior    *Color
	ColorHover       *Color
	ColorFocus       *Color
	ColorActive      *Color
	ColorBorder      *Color
	ColorBorderFocus *Color
	ColorSelect      *Color
}

func colorOr(override *Color, fallback Color) Color {
	if override != nil {
		return *override
	}
	return fallback
}

// WithColors returns a new Theme with the specified colors updated
// across all widget styles.
func (t Theme) WithColors(o ColorOverrides) Theme {
	bg := colorOr(o.ColorBackground, t.ColorBackground)
	panel := colorOr(o.ColorPanel, t.ColorPanel)
	interior := colorOr(o.ColorInterior, t.ColorInterior)
	hover := colorOr(o.ColorHover, t.ColorHover)
	focus := colorOr(o.ColorFocus, t.ColorFocus)
	active := colorOr(o.ColorActive, t.ColorActive)
	border := colorOr(o.ColorBorder, t.ColorBorder)
	borderFocus := colorOr(o.ColorBorderFocus, t.ButtonStyle.ColorBorderFocus)
	sel := colorOr(o.ColorSelect, t.ColorSelect)

	t.ColorBackground = bg
	t.ColorPanel = panel
	t.ColorInterior = interior
	t.ColorHover = hover
	t.ColorFocus = focus
	t.ColorActive = active
	t.ColorBorder = border
	t.ColorSelect = sel

	t.ButtonStyle.Color = interior
	t.ButtonStyle.ColorHover = hover
	t.ButtonStyle.ColorFocus = active
	t.ButtonStyle.ColorClick = focus
	t.ButtonStyle.ColorBorder = border
	t.ButtonStyle.ColorBorderFocus = borderFocus

	t.InputStyle.Color = interior
	t.InputStyle.ColorHover = hover
	t.InputStyle.ColorFocus = interior
	t.InputStyle.ColorClick = active
	t.InputStyle.ColorBorder = border
	t.InputStyle.ColorBorderFocus = borderFocus

	t.RadioStyle.Color = panel
	t.RadioStyle.ColorHover = hover
	t.RadioStyle.ColorFocus = sel
	t.RadioStyle.ColorBorder = border
	t.RadioStyle.ColorBorderFocus = borderFocus
	t.RadioStyle.ColorSelect = sel
	t.RadioStyle.ColorUnselect = active

	t.SwitchStyle.Color = panel
	t.SwitchStyle.ColorHover = hover
	t.SwitchStyle.ColorBorder = border
	t.SwitchStyle.ColorBorderFocus = borderFocus
	t.SwitchStyle.ColorSelect = sel
	t.SwitchStyle.ColorUnselect = active

	t.ToggleStyle.Color = panel
	t.ToggleStyle.ColorHover = hover
	t.ToggleStyle.ColorBorder = border
	t.ToggleStyle.ColorBorderFocus = borderFocus

	t.SelectStyle.Color = interior
	t.SelectStyle.ColorHover = hover
	t.SelectStyle.ColorFocus = focus
	t.SelectStyle.ColorClick = active
	t.SelectStyle.ColorBorder = border
	t.SelectStyle.ColorBorderFocus = borderFocus
	t.SelectStyle.ColorSelect = sel

	t.ListBoxStyle.Color = interior
	t.ListBoxStyle.ColorHover = hover
	t.ListBoxStyle.ColorFocus = focus
	t.ListBoxStyle.ColorBorder = border
	t.ListBoxStyle.ColorBorderFocus = borderFocus
	t.ListBoxStyle.ColorSelect = sel

	t.ScrollbarStyle.ColorThumb = active

	t.RectangleStyle.ColorBorder = border

	t.DialogStyle.Color = panel
	t.DialogStyle.ColorBorder = border
	t.DialogStyle.ColorBorderFocus = borderFocus

	t.ToastStyle.Color = panel
	t.ToastStyle.ColorBorder = border
	t.ToastStyle.ColorInfo = sel

	t.TooltipStyle.Color = interior
	t.TooltipStyle.ColorHover = hover
	t.TooltipStyle.ColorBorder = border
	t.TooltipStyle.ColorBorderFocus = borderFocus

	t.BadgeStyle.Color = active
	t.BadgeStyle.ColorInfo = sel

	t.ExpandPanelStyle.Color = panel
	t.ExpandPanelStyle.ColorHover = hover
	t.ExpandPanelStyle.ColorClick = active
	t.ExpandPanelStyle.ColorBorder = border

	t.ProgressBarStyle.Color = interior
	t.ProgressBarStyle.ColorBar = sel
	t.ProgressBarStyle.ColorBorder = border
	t.ProgressBarStyle.TextBackground = panel

	t.RangeSliderStyle.Color = interior
	t.RangeSliderStyle.ColorClick = active
	t.RangeSliderStyle.ColorThumb = panel
	t.RangeSliderStyle.ColorLeft = sel
	t.RangeSliderStyle.ColorFocus = sel
	t.RangeSliderStyle.ColorHover = hover
	t.RangeSliderStyle.ColorBorder = border
	t.RangeSliderStyle.ColorBorderFocus = borderFocus

	t.TabControlStyle.Color = panel
	t.TabControlStyle.ColorBorder = border
	t.TabControlStyle.ColorContent = panel
	t.TabControlStyle.ColorContentBorder = border
	t.TabControlStyle.ColorTab = interior
	t.TabControlStyle.ColorTabHover = hover
	t.TabControlStyle.ColorTabFocus = focus
	t.TabControlStyle.ColorTabClick = active
	t.TabControlStyle.ColorTabSelected = sel
	t.TabControlStyle.ColorTabDisabled = panel
	t.TabControlStyle.ColorTabBorder = border
	t.TabControlStyle.ColorTabBorderFocus = borderFocus

	t.BreadcrumbStyle.ColorCrumbHover = hover
	t.BreadcrumbStyle.ColorCrumbClick = active
	t.BreadcrumbStyle.ColorContent = panel
	t.BreadcrumbStyle.ColorContentBorder = border

	t.SplitterStyle.ColorHandle = interior
	t.SplitterStyle.ColorHandleHover = hover
	t.SplitterStyle.ColorHandleActive = active
	t.SplitterStyle.ColorHandleBorder = border
	t.SplitterStyle.ColorGrip = sel
	t.SplitterStyle.ColorButton = interior
	t.SplitterStyle.ColorButtonHover = hover
	t.SplitterStyle.ColorButtonActive = active

	t.TableStyle.ColorBorder = border
	t.TableStyle.ColorSelect = sel
	t.TableStyle.ColorHover = hover

	return t
}

// AdjustFontSize returns a new Theme with all font sizes adjusted
// by delta, clamped to [minSize, maxSize].
func (t Theme) AdjustFontSize(delta, minSize, maxSize float32) (Theme, error) {
	if minSize < 1 {
		return t, errors.New("minSize must be > 0")
	}
	cfg := t.Cfg
	newSize := cfg.TextStyleDef.Size + delta
	if newSize < minSize || newSize > maxSize {
		return t, errors.New("new font size out of range")
	}
	cfg.TextStyleDef.Size = newSize
	cfg.SizeTextTiny += delta
	cfg.SizeTextXSmall += delta
	cfg.SizeTextSmall += delta
	cfg.SizeTextMedium += delta
	cfg.SizeTextLarge += delta
	cfg.SizeTextXLarge += delta
	return ThemeMaker(cfg), nil
}
