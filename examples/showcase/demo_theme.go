package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

func demoThemeGen(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ShowcaseApp](w)
	strategies := []string{"mono", "complement", "analogous", "triadic", "warm", "cool"}
	pickerColor := app.ThemeGenSeed
	if app.ThemeGenPickText {
		pickerColor = app.ThemeGenText
	}

	strategyViews := make([]gui.View, len(strategies))
	for i, strategy := range strategies {
		selected := app.ThemeGenStrategy == strategy
		color := t.ColorInterior
		textStyle := t.N2
		if selected {
			color = t.ColorActive
			textStyle.Color = gui.White
		}
		sv := strategy
		strategyViews[i] = gui.Button(gui.ButtonCfg{
			ID:      "strat-" + sv,
			Color:   color,
			Padding: gui.Some(gui.NewPadding(4, 10, 4, 10)),
			Radius:  gui.Some(float32(12)),
			Content: []gui.View{gui.Text(gui.TextCfg{Text: strategyLabel(sv), TextStyle: textStyle})},
			OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
				gui.State[ShowcaseApp](w).ThemeGenStrategy = sv
				applyGenTheme(w)
				e.IsHandled = true
			},
		})
	}

	title := "Pick a seed color to generate a full theme."
	if app.ThemeGenName != "" {
		title = app.ThemeGenName
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: title, TextStyle: t.N3}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(16)),
				Padding: gui.Some(gui.PaddingNone),
				VAlign:  gui.VAlignTop,
				Content: []gui.View{
					gui.Column(gui.ContainerCfg{
						Sizing:  gui.FitFit,
						Spacing: gui.Some(float32(10)),
						Padding: gui.Some(gui.PaddingNone),
						Content: []gui.View{
							gui.ColorPicker(gui.ColorPickerCfg{
								ID:    "theme-gen-cp",
								Color: pickerColor,
								OnColorChange: func(c gui.Color, _ *gui.Event, w *gui.Window) {
									app := gui.State[ShowcaseApp](w)
									if app.ThemeGenPickText {
										app.ThemeGenText = c
									} else {
										app.ThemeGenSeed = c
									}
									applyGenTheme(w)
								},
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
								Width:  140,
								Sizing: gui.FixedFit,
								OnChange: func(v float32, _ *gui.Event, w *gui.Window) {
									gui.State[ShowcaseApp](w).ThemeGenTint = v
									applyGenTheme(w)
								},
							}),
							gui.Row(gui.ContainerCfg{
								Sizing:  gui.FillFit,
								Spacing: gui.Some(float32(12)),
								Padding: gui.Some(gui.PaddingNone),
								Content: []gui.View{
									gui.Column(gui.ContainerCfg{
										Sizing:  gui.FitFit,
										Spacing: gui.Some(float32(6)),
										Padding: gui.Some(gui.PaddingNone),
										Content: []gui.View{
											gui.Text(gui.TextCfg{Text: "Radius", TextStyle: t.N3}),
											gui.NumericInput(gui.NumericInputCfg{
												ID:       "theme-gen-radius",
												IDFocus:  9180,
												Text:     app.ThemeGenRadiusText,
												Value:    float64p(float64(app.ThemeGenRadius)),
												Decimals: 1,
												Min:      float64p(0),
												Max:      float64p(30),
												Width:    80,
												Sizing:   gui.FixedFit,
												OnTextChanged: func(_ *gui.Layout, text string, w *gui.Window) {
													gui.State[ShowcaseApp](w).ThemeGenRadiusText = text
												},
												OnValueCommit: func(_ *gui.Layout, value *float64, text string, w *gui.Window) {
													app := gui.State[ShowcaseApp](w)
													app.ThemeGenRadiusText = text
													if value != nil {
														app.ThemeGenRadius = float32(*value)
														applyGenTheme(w)
													}
												},
											}),
										},
									}),
									gui.Column(gui.ContainerCfg{
										Sizing:  gui.FitFit,
										Spacing: gui.Some(float32(6)),
										Padding: gui.Some(gui.PaddingNone),
										Content: []gui.View{
											gui.Text(gui.TextCfg{Text: "Border", TextStyle: t.N3}),
											gui.NumericInput(gui.NumericInputCfg{
												ID:       "theme-gen-border",
												IDFocus:  9181,
												Text:     app.ThemeGenBorderText,
												Value:    float64p(float64(app.ThemeGenBorder)),
												Decimals: 1,
												Min:      float64p(0),
												Max:      float64p(10),
												Width:    80,
												Sizing:   gui.FixedFit,
												OnTextChanged: func(_ *gui.Layout, text string, w *gui.Window) {
													gui.State[ShowcaseApp](w).ThemeGenBorderText = text
												},
												OnValueCommit: func(_ *gui.Layout, value *float64, text string, w *gui.Window) {
													app := gui.State[ShowcaseApp](w)
													app.ThemeGenBorderText = text
													if value != nil {
														app.ThemeGenBorder = float32(*value)
														applyGenTheme(w)
													}
												},
											}),
										},
									}),
								},
							}),
						},
					}),
					gui.Column(gui.ContainerCfg{
						Sizing:  gui.FillFit,
						Spacing: gui.Some(float32(10)),
						Padding: gui.Some(gui.PaddingNone),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Palette", TextStyle: t.B3}),
							gui.Wrap(gui.ContainerCfg{
								Sizing:  gui.FillFit,
								Spacing: gui.Some(float32(4)),
								Padding: gui.Some(gui.PaddingNone),
								Content: strategyViews,
							}),
							gui.Checkbox(gui.ToggleCfg{
								ID:       "theme-gen-pick-text",
								Label:    "Edit text color",
								Selected: app.ThemeGenPickText,
								OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
									gui.State[ShowcaseApp](w).ThemeGenPickText = !gui.State[ShowcaseApp](w).ThemeGenPickText
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
											w.SetTheme(gui.ThemeDarkBordered)
											syncThemeGenFromCfg(gui.State[ShowcaseApp](w), gui.ThemeDarkBorderedCfg)
											e.IsHandled = true
										},
									}),
									gui.Button(gui.ButtonCfg{
										ID:      "btn-reset-light",
										Padding: gui.Some(gui.NewPadding(6, 12, 6, 12)),
										Content: []gui.View{gui.Text(gui.TextCfg{Text: "Reset Light", TextStyle: t.N3})},
										OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
											w.SetTheme(gui.ThemeLightBordered)
											syncThemeGenFromCfg(gui.State[ShowcaseApp](w), gui.ThemeLightBorderedCfg)
											e.IsHandled = true
										},
									}),
								},
							}),
							gui.Row(gui.ContainerCfg{
								Sizing:  gui.FillFit,
								Spacing: gui.Some(float32(8)),
								Padding: gui.Some(gui.PaddingNone),
								Content: []gui.View{
									gui.Button(gui.ButtonCfg{
										ID:      "btn-theme-save",
										Padding: gui.Some(gui.NewPadding(6, 12, 6, 12)),
										Content: []gui.View{gui.Text(gui.TextCfg{Text: "Save Theme", TextStyle: t.N3})},
										OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
											w.NativeSaveDialog(gui.NativeSaveDialogCfg{
												Title:            "Save Theme",
												DefaultName:      "theme.json",
												DefaultExtension: "json",
												Filters: []gui.NativeFileFilter{
													{Name: "JSON", Extensions: []string{"json"}},
												},
												ConfirmOverwrite: true,
												OnDone: func(result gui.NativeDialogResult, w *gui.Window) {
													if result.Status != gui.DialogOK || len(result.Paths) == 0 {
														return
													}
													app := gui.State[ShowcaseApp](w)
													cfg := generateThemeCfg(
														app.ThemeGenSeed,
														app.ThemeGenStrategy,
														gui.CurrentTheme().TitlebarDark,
														app.ThemeGenTint,
														app.ThemeGenText,
														app.ThemeGenRadius,
														app.ThemeGenBorder,
													)
													path := result.Paths[0].Path
													if err := themeCfgSave(path, cfg); err != nil {
														app.ThemeGenName = err.Error()
														return
													}
													app.ThemeGenName = filepath.Base(path)
												},
											})
											e.IsHandled = true
										},
									}),
									gui.Button(gui.ButtonCfg{
										ID:      "btn-theme-load",
										Padding: gui.Some(gui.NewPadding(6, 12, 6, 12)),
										Content: []gui.View{gui.Text(gui.TextCfg{Text: "Load Theme", TextStyle: t.N3})},
										OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
											w.NativeOpenDialog(gui.NativeOpenDialogCfg{
												Title: "Load Theme",
												Filters: []gui.NativeFileFilter{
													{Name: "JSON", Extensions: []string{"json"}},
												},
												OnDone: func(result gui.NativeDialogResult, w *gui.Window) {
													if result.Status != gui.DialogOK || len(result.Paths) == 0 {
														return
													}
													path := result.Paths[0].Path
													cfg, err := themeCfgLoad(path)
													if err != nil {
														gui.State[ShowcaseApp](w).ThemeGenName = err.Error()
														return
													}
													w.SetTheme(gui.ThemeMaker(cfg))
													app := gui.State[ShowcaseApp](w)
													syncThemeGenFromCfg(app, cfg)
													app.ThemeGenName = filepath.Base(path)
												},
											})
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

func strategyLabel(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func syncThemeGenFromCfg(app *ShowcaseApp, cfg gui.ThemeCfg) {
	app.ThemeGenSeed = cfg.ColorSelect
	app.ThemeGenTint = 0
	app.ThemeGenStrategy = "mono"
	app.ThemeGenRadius = cfg.Radius
	app.ThemeGenRadiusText = floatString(cfg.Radius)
	app.ThemeGenBorder = cfg.SizeBorder
	app.ThemeGenBorderText = floatString(cfg.SizeBorder)
	app.ThemeGenText = cfg.TextStyleDef.Color
	app.ThemeGenPickText = false
}

func applyGenTheme(w *gui.Window) {
	app := gui.State[ShowcaseApp](w)
	cfg := generateThemeCfg(
		app.ThemeGenSeed,
		app.ThemeGenStrategy,
		gui.CurrentTheme().TitlebarDark,
		app.ThemeGenTint,
		app.ThemeGenText,
		app.ThemeGenRadius,
		app.ThemeGenBorder,
	)
	w.SetTheme(gui.ThemeMaker(cfg))
}

func generateThemeCfg(seed gui.Color, strategy string, isDark bool, tint float32, textColor gui.Color, radius, border float32) gui.ThemeCfg {
	h, s, _ := seed.ToHSV()
	tintFactor := tint / 100.0

	ph := h
	ah := h
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
			TextStyleDef: gui.TextStyle{
				Family: gui.ThemeDarkCfg.TextStyleDef.Family,
				Color:  textColor,
				Size:   gui.ThemeDarkCfg.TextStyleDef.Size,
			},
			TitlebarDark: true,
			SizeBorder:   border,
			Radius:       radius,
			RadiusSmall:  radius * 0.64,
			RadiusMedium: radius,
			RadiusLarge:  radius * 1.36,
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
		TextStyleDef: gui.TextStyle{
			Family: gui.ThemeLightCfg.TextStyleDef.Family,
			Color:  textColor,
			Size:   gui.ThemeLightCfg.TextStyleDef.Size,
		},
		TitlebarDark: false,
		SizeBorder:   border,
		Radius:       radius,
		RadiusSmall:  radius * 0.64,
		RadiusMedium: radius,
		RadiusLarge:  radius * 1.36,
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
