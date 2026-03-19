package gui

import "errors"

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
	t.ListBoxStyle.ColorBorder = border
	t.ListBoxStyle.ColorSelect = sel

	t.TreeStyle.ColorHover = hover
	t.TreeStyle.ColorFocus = focus
	if o.ColorBorder != nil {
		t.TreeStyle.ColorBorder = border
	}

	t.ScrollbarStyle.ColorThumb = active
	t.RectangleStyle.ColorBorder = border

	t.DialogStyle.Color = panel
	t.DialogStyle.ColorBorder = border
	t.DialogStyle.ColorBorderFocus = borderFocus

	t.ToastStyle.Color = panel
	t.ToastStyle.ColorBorder = border
	t.ToastStyle.ColorInfo = sel

	t.TooltipStyle.Color = interior
	t.TooltipStyle.ColorBorder = border

	t.BadgeStyle.Color = active
	t.BadgeStyle.ColorInfo = sel

	t.ExpandPanelStyle.Color = panel
	t.ExpandPanelStyle.ColorHover = hover
	t.ExpandPanelStyle.ColorClick = active
	t.ExpandPanelStyle.ColorBorder = border

	t.ProgressBarStyle.Color = interior
	t.ProgressBarStyle.ColorBar = sel
	t.ProgressBarStyle.ColorBorder = border
	t.ProgressBarStyle.TextBackground = ColorTransparent

	t.SliderStyle.Color = interior
	t.SliderStyle.ColorClick = active
	t.SliderStyle.ColorThumb = panel
	t.SliderStyle.ColorLeft = sel
	t.SliderStyle.ColorFocus = sel
	t.SliderStyle.ColorHover = hover
	t.SliderStyle.ColorBorder = border
	t.SliderStyle.ColorBorderFocus = borderFocus

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

	t.ComboboxStyle.Color = interior
	t.ComboboxStyle.ColorHover = hover
	t.ComboboxStyle.ColorFocus = interior
	t.ComboboxStyle.ColorBorder = border
	t.ComboboxStyle.ColorBorderFocus = borderFocus
	t.ComboboxStyle.ColorHighlight = sel

	t.CommandPaletteStyle.Color = panel
	t.CommandPaletteStyle.ColorBorder = border
	t.CommandPaletteStyle.ColorHighlight = sel

	t.MenubarStyle.Color = interior
	t.MenubarStyle.ColorHover = hover
	t.MenubarStyle.ColorFocus = focus
	t.MenubarStyle.ColorBorder = border
	t.MenubarStyle.ColorBorderFocus = borderFocus
	t.MenubarStyle.ColorSelect = sel

	t.DatePickerStyle.Color = interior
	t.DatePickerStyle.ColorHover = hover
	t.DatePickerStyle.ColorFocus = focus
	t.DatePickerStyle.ColorClick = active
	t.DatePickerStyle.ColorBorder = border
	t.DatePickerStyle.ColorBorderFocus = borderFocus
	t.DatePickerStyle.ColorSelect = sel

	t.ColorPickerStyle.Color = interior
	t.ColorPickerStyle.ColorBorder = border
	t.ColorPickerStyle.ColorBorderFocus = borderFocus

	t.DataGridStyle.ColorBackground = interior
	t.DataGridStyle.ColorHeader = panel
	t.DataGridStyle.ColorHeaderHover = hover
	t.DataGridStyle.ColorFilter = interior
	t.DataGridStyle.ColorQuickFilter = panel
	t.DataGridStyle.ColorRowHover = hover
	t.DataGridStyle.ColorRowSelected = sel
	t.DataGridStyle.ColorBorder = border
	t.DataGridStyle.ColorResizeHandle = border
	t.DataGridStyle.ColorResizeActive = sel

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
