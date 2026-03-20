package gui

import "testing"

func TestWithStyleMethodsSetsField(t *testing.T) {
	var base Theme

	// Verify a representative method properly sets its field.
	got := base.WithButtonStyle(ButtonStyle{Color: Color{R: 42}})
	if got.ButtonStyle.Color.R != 42 {
		t.Error("WithButtonStyle did not set ButtonStyle.Color")
	}
	// Verify non-target fields are unchanged.
	if got.Name != base.Name {
		t.Error("WithButtonStyle modified Name")
	}
}

func TestWithStyleMethodsCoverage(t *testing.T) {
	// Exercise every With* method for coverage.
	// Each method is a 2-statement setter; calling with any
	// value covers the full method body.
	var th Theme
	th = th.WithButtonStyle(ButtonStyle{})
	th = th.WithContainerStyle(ContainerStyle{})
	th = th.WithRectangleStyle(RectangleStyle{})
	th = th.WithTextStyle(TextStyle{})
	th = th.WithInputStyle(InputStyle{})
	th = th.WithScrollbarStyle(ScrollbarStyle{})
	th = th.WithRadioStyle(RadioStyle{})
	th = th.WithSwitchStyle(SwitchStyle{})
	th = th.WithToggleStyle(ToggleStyle{})
	th = th.WithSelectStyle(SelectStyle{})
	th = th.WithListBoxStyle(ListBoxStyle{})
	th = th.WithTreeStyle(TreeStyle{})
	th = th.WithDialogStyle(DialogStyle{})
	th = th.WithToastStyle(ToastStyle{})
	th = th.WithTooltipStyle(TooltipStyle{})
	th = th.WithBadgeStyle(BadgeStyle{})
	th = th.WithExpandPanelStyle(ExpandPanelStyle{})
	th = th.WithProgressBarStyle(ProgressBarStyle{})
	th = th.WithSliderStyle(SliderStyle{})
	th = th.WithTabControlStyle(TabControlStyle{})
	th = th.WithBreadcrumbStyle(BreadcrumbStyle{})
	th = th.WithSplitterStyle(SplitterStyle{})
	th = th.WithTableStyle(TableStyle{})
	th = th.WithComboboxStyle(ComboboxStyle{})
	th = th.WithCommandPaletteStyle(CommandPaletteStyle{})
	th = th.WithMenubarStyle(MenubarStyle{})
	th = th.WithDatePickerStyle(DatePickerStyle{})
	th = th.WithColorPickerStyle(ColorPickerStyle{})
	th = th.WithDataGridStyle(DataGridStyle{})
	_ = th
}

func TestWithTextStyleUsesTextStyleDefField(t *testing.T) {
	var base Theme
	got := base.WithTextStyle(TextStyle{Color: Color{G: 99}})
	if got.TextStyleDef.Color.G != 99 {
		t.Error("WithTextStyle should set TextStyleDef field")
	}
}
