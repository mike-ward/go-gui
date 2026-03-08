package main

import (
	"fmt"
	"math"

	"github.com/mike-ward/go-gui/gui"
)

func demoThemeGen(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)

	strategies := []string{"mono", "complement", "analogous", "triadic", "warm", "cool"}

	strategyViews := make([]gui.View, len(strategies))
	for i, s := range strategies {
		selected := app.ThemeGenStrategy == s
		color := t.ColorInterior
		ts := t.N2
		if selected {
			color = t.ColorActive
			ts.Color = gui.RGB(255, 255, 255)
		}
		sv := s
		strategyViews[i] = gui.Button(gui.ButtonCfg{
			ID:      "strat-" + sv,
			Color:   color,
			Padding: gui.Some(gui.NewPadding(4, 10, 4, 10)),
			Radius:  gui.Some(float32(12)),
			Content: []gui.View{gui.Text(gui.TextCfg{Text: sv, TextStyle: ts})},
			OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
				gui.State[ArcadeApp](w).ThemeGenStrategy = sv
				applyGenTheme(w)
				e.IsHandled = true
			},
		})
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Pick a seed color to generate a full theme.",
				TextStyle: t.N3,
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(16)),
				Padding: gui.Some(gui.PaddingNone),
				VAlign:  gui.VAlignTop,
				Content: []gui.View{
					gui.ColorPicker(gui.ColorPickerCfg{
						ID:    "theme-gen-cp",
						Color: app.ThemeGenSeed,
						OnColorChange: func(c gui.Color, _ *gui.Event, w *gui.Window) {
							gui.State[ArcadeApp](w).ThemeGenSeed = c
							applyGenTheme(w)
						},
					}),
					gui.Column(gui.ContainerCfg{
						Sizing:  gui.FillFit,
						Spacing: gui.Some(float32(10)),
						Padding: gui.Some(gui.PaddingNone),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Strategy", TextStyle: t.B3}),
							gui.Wrap(gui.ContainerCfg{
								Sizing:  gui.FillFit,
								Spacing: gui.Some(float32(4)),
								Padding: gui.Some(gui.PaddingNone),
								Content: strategyViews,
							}),
							gui.Text(gui.TextCfg{
								Text:      fmt.Sprintf("Tint: %.0f%%", app.ThemeGenTint),
								TextStyle: t.N3,
							}),
							gui.RangeSlider(gui.RangeSliderCfg{
								ID:     "theme-gen-tint",
								Value:  app.ThemeGenTint,
								Min:    0,
								Max:    100,
								Sizing: gui.FillFit,
								OnChange: func(v float32, _ *gui.Event, w *gui.Window) {
									gui.State[ArcadeApp](w).ThemeGenTint = v
									applyGenTheme(w)
								},
							}),
							gui.Row(gui.ContainerCfg{
								Sizing:  gui.FillFit,
								Spacing: gui.Some(float32(8)),
								Padding: gui.Some(gui.PaddingNone),
								Content: []gui.View{
									gui.Button(gui.ButtonCfg{
										ID:      "btn-reset-dark",
										Padding: gui.Some(gui.NewPadding(6, 12, 6, 12)),
										Content: []gui.View{gui.Text(gui.TextCfg{Text: "Reset Dark", TextStyle: t.N3})},
										OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
											gui.SetTheme(gui.ThemeDarkBordered)
											syncThemeGenFromCfg(gui.State[ArcadeApp](w), gui.ThemeDarkBorderedCfg)
											e.IsHandled = true
										},
									}),
									gui.Button(gui.ButtonCfg{
										ID:      "btn-reset-light",
										Padding: gui.Some(gui.NewPadding(6, 12, 6, 12)),
										Content: []gui.View{gui.Text(gui.TextCfg{Text: "Reset Light", TextStyle: t.N3})},
										OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
											gui.SetTheme(gui.ThemeLightBordered)
											syncThemeGenFromCfg(gui.State[ArcadeApp](w), gui.ThemeLightBorderedCfg)
											e.IsHandled = true
										},
									}),
								},
							}),
						},
					}),
				},
			}),
		},
	})
}

func syncThemeGenFromCfg(app *ArcadeApp, cfg gui.ThemeCfg) {
	app.ThemeGenSeed = cfg.ColorSelect
	app.ThemeGenTint = 0
	app.ThemeGenStrategy = "mono"
}

func applyGenTheme(w *gui.Window) {
	app := gui.State[ArcadeApp](w)
	isDark := gui.CurrentTheme().TitlebarDark
	cfg := generateThemeCfg(app.ThemeGenSeed, app.ThemeGenStrategy, isDark, app.ThemeGenTint)
	w.SetTheme(gui.ThemeMaker(cfg))
}

func generateThemeCfg(seed gui.Color, strategy string, isDark bool, tint float32) gui.ThemeCfg {
	h, s, _ := seed.ToHSV()
	tintFactor := tint / 100.0

	ph := h // primary hue (surfaces)
	ah := h // accent hue (interactive states)
	accentS := max(min(s, 1.0), 0.5)
	accentV := float32(0.85)
	if !isDark {
		accentV = 0.65
	}

	switch strategy {
	case "complement":
		ah = wrapHue(h + 180)
	case "analogous":
		ah = wrapHue(h + 30)
	case "triadic":
		ah = wrapHue(h + 120)
	case "warm":
		ph = float32(math.Mod(float64(h), 60))
		ah = ph + 15
	case "cool":
		ph = 180 + float32(math.Mod(float64(h), 90))
		ah = ph + 20
	}

	if isDark {
		sTint := max(min(s, 1.0), 0.3) * tintFactor
		return gui.ThemeCfg{
			Name:             "generated",
			ColorBackground:  gui.ColorFromHSV(ph, sTint, 0.19),
			ColorPanel:       gui.ColorFromHSV(ph, sTint, 0.25),
			ColorInterior:    gui.ColorFromHSV(ph, sTint, 0.29),
			ColorHover:       gui.ColorFromHSV(ph, sTint, 0.33),
			ColorFocus:       gui.ColorFromHSV(ah, sTint, 0.37),
			ColorActive:      gui.ColorFromHSV(ah, sTint, 0.41),
			ColorBorder:      gui.ColorFromHSV(ah, sTint*0.8, 0.39),
			ColorSelect:      gui.ColorFromHSV(ah, accentS, accentV),
			ColorBorderFocus: gui.ColorFromHSV(ah, accentS*0.7, accentV*0.9),
			TextStyleDef:     gui.ThemeDarkCfg.TextStyleDef,
			TitlebarDark:     true,
			SizeBorder:       gui.ThemeDarkBorderedCfg.SizeBorder,
			Radius:           gui.ThemeDarkBorderedCfg.Radius,
			RadiusSmall:      gui.ThemeDarkBorderedCfg.Radius * 0.64,
			RadiusMedium:     gui.ThemeDarkBorderedCfg.Radius,
			RadiusLarge:      gui.ThemeDarkBorderedCfg.Radius * 1.36,
		}
	}

	sTint := max(min(s, 1.0), 0.3) * tintFactor * 0.5
	return gui.ThemeCfg{
		Name:             "generated",
		ColorBackground:  gui.ColorFromHSV(ph, sTint*0.6, 0.96),
		ColorPanel:       gui.ColorFromHSV(ph, sTint, 0.90),
		ColorInterior:    gui.ColorFromHSV(ph, sTint, 0.86),
		ColorHover:       gui.ColorFromHSV(ph, sTint, 0.82),
		ColorFocus:       gui.ColorFromHSV(ah, sTint, 0.78),
		ColorActive:      gui.ColorFromHSV(ah, sTint, 0.74),
		ColorBorder:      gui.ColorFromHSV(ah, sTint*1.5, 0.55),
		ColorSelect:      gui.ColorFromHSV(ah, accentS, accentV*0.75),
		ColorBorderFocus: gui.ColorFromHSV(ah, accentS*0.8, accentV*0.6),
		TextStyleDef:     gui.ThemeLightCfg.TextStyleDef,
		TitlebarDark:     false,
		SizeBorder:       gui.ThemeLightBorderedCfg.SizeBorder,
		Radius:           gui.ThemeLightBorderedCfg.Radius,
		RadiusSmall:      gui.ThemeLightBorderedCfg.Radius * 0.64,
		RadiusMedium:     gui.ThemeLightBorderedCfg.Radius,
		RadiusLarge:      gui.ThemeLightBorderedCfg.Radius * 1.36,
	}
}

func wrapHue(h float32) float32 {
	for h >= 360 {
		h -= 360
	}
	for h < 0 {
		h += 360
	}
	return h
}
