package gui

import "testing"

func TestThemeMaker(t *testing.T) {
	cfg := ThemeCfg{
		Name:            "test",
		ColorBackground: RGB(48, 48, 48),
		ColorPanel:      RGB(64, 64, 64),
		ColorInterior:   RGB(74, 74, 74),
		ColorHover:      RGB(84, 84, 84),
		ColorFocus:      RGB(94, 94, 94),
		ColorActive:     RGB(104, 104, 104),
		ColorBorder:     RGB(100, 100, 100),
		ColorSelect:     RGB(65, 105, 225),
		TextStyleDef:    TextStyle{Color: RGB(225, 225, 225), Size: 16},
		PaddingSmall:    PaddingSmall,
		PaddingMedium:   PaddingMedium,
		PaddingLarge:    PaddingLarge,
		Padding:         PaddingMedium,
		Radius:          RadiusMedium,
		RadiusSmall:     RadiusSmall,
		RadiusMedium:    RadiusMedium,
		RadiusLarge:     RadiusLarge,
		SpacingSmall:    SpacingSmall,
		SpacingMedium:   SpacingMedium,
		SpacingLarge:    SpacingLarge,
		SizeTextTiny:    SizeTextTiny,
		SizeTextXSmall:  SizeTextXSmall,
		SizeTextSmall:   SizeTextSmall,
		SizeTextMedium:  SizeTextMedium,
		SizeTextLarge:   SizeTextLarge,
		SizeTextXLarge:  SizeTextXLarge,
		SizeScrollbar:   7,
		SizeScrollbarMin: 20,
		SizeRadio:       16,
		SizeSwitchWidth:  36,
		SizeSwitchHeight: 22,
		ScrollMultiplier: 20,
		ScrollDeltaLine:  1,
		ScrollDeltaPage:  10,
	}
	theme := ThemeMaker(cfg)
	if theme.Name != "test" {
		t.Errorf("name = %q", theme.Name)
	}
	if theme.ButtonStyle.Color != cfg.ColorInterior {
		t.Error("button color mismatch")
	}
	if theme.N1.Size != SizeTextXLarge {
		t.Errorf("N1.Size = %f", theme.N1.Size)
	}
}

func TestSetTheme(t *testing.T) {
	// Save all globals modified by SetTheme.
	savedBtn := DefaultButtonStyle
	savedText := DefaultTextStyle
	savedCtr := DefaultContainerStyle
	savedRect := DefaultRectangleStyle
	savedInp := DefaultInputStyle
	savedSb := DefaultScrollbarStyle
	savedRad := DefaultRadioStyle
	savedSw := DefaultSwitchStyle
	savedTog := DefaultToggleStyle
	savedSel := DefaultSelectStyle
	savedLb := DefaultListBoxStyle
	defer func() {
		DefaultButtonStyle = savedBtn
		DefaultTextStyle = savedText
		DefaultContainerStyle = savedCtr
		DefaultRectangleStyle = savedRect
		DefaultInputStyle = savedInp
		DefaultScrollbarStyle = savedSb
		DefaultRadioStyle = savedRad
		DefaultSwitchStyle = savedSw
		DefaultToggleStyle = savedTog
		DefaultSelectStyle = savedSel
		DefaultListBoxStyle = savedLb
	}()

	theme := Theme{
		ButtonStyle: ButtonStyle{Color: Red},
	}
	SetTheme(theme)
	if DefaultButtonStyle.Color != Red {
		t.Error("SetTheme should update DefaultButtonStyle")
	}
}

func TestWithColors(t *testing.T) {
	theme := Theme{
		ColorHover: RGB(1, 1, 1),
		ButtonStyle: ButtonStyle{
			ColorHover:       RGB(1, 1, 1),
			ColorBorderFocus: RGB(2, 2, 2),
		},
	}
	newHover := RGB(200, 200, 200)
	updated := theme.WithColors(ColorOverrides{
		ColorHover: &newHover,
	})
	if updated.ColorHover != newHover {
		t.Error("theme hover not updated")
	}
	if updated.ButtonStyle.ColorHover != newHover {
		t.Error("button hover not propagated")
	}
}

func TestAdjustFontSize(t *testing.T) {
	cfg := ThemeCfg{
		TextStyleDef:   TextStyle{Color: RGB(225, 225, 225), Size: 16},
		SizeTextTiny:   10,
		SizeTextXSmall: 12,
		SizeTextSmall:  14,
		SizeTextMedium: 16,
		SizeTextLarge:  20,
		SizeTextXLarge: 24,
		Radius:         RadiusMedium,
		RadiusSmall:    RadiusSmall,
		RadiusMedium:   RadiusMedium,
		RadiusLarge:    RadiusLarge,
		SpacingMedium:  SpacingMedium,
		PaddingMedium:  PaddingMedium,
		SizeScrollbar:  7,
		SizeScrollbarMin: 20,
		SizeRadio:      16,
		SizeSwitchWidth: 36,
		SizeSwitchHeight: 22,
	}
	theme := ThemeMaker(cfg)
	bigger, err := theme.AdjustFontSize(2, 8, 32)
	if err != nil {
		t.Fatal(err)
	}
	if bigger.SizeTextMedium != 18 {
		t.Errorf("medium = %f, want 18", bigger.SizeTextMedium)
	}
	_, err = theme.AdjustFontSize(100, 1, 32)
	if err == nil {
		t.Error("should error on out of range")
	}
	_, err = theme.AdjustFontSize(1, 0, 32)
	if err == nil {
		t.Error("should error on minSize < 1")
	}
}

func TestWithButtonStyle(t *testing.T) {
	theme := Theme{}
	s := ButtonStyle{Color: Blue}
	updated := theme.WithButtonStyle(s)
	if updated.ButtonStyle.Color != Blue {
		t.Error("WithButtonStyle not applied")
	}
}
