package main

import "github.com/mike-ward/go-gui/gui"

var svgSpinnerCategories = []string{
	"Rings & Circles",
	"Dots",
	"Bars",
	"Loaders",
	"Blocks",
	"Pulse",
	"Miscellaneous",
}

var svgSpinnerGroups = map[string][]gui.SvgSpinnerKind{
	"Rings & Circles": {
		gui.SvgSpinner90Ring,
		gui.SvgSpinner90RingWithBg,
		gui.SvgSpinner180Ring,
		gui.SvgSpinner180RingWithBg,
		gui.SvgSpinner270Ring,
		gui.SvgSpinner270RingWithBg,
		gui.SvgSpinnerCircles,
		gui.SvgSpinnerEclipse,
		gui.SvgSpinnerEclipseHalf,
		gui.SvgSpinnerLoader2,
		gui.SvgSpinnerOval,
		gui.SvgSpinnerRingResize,
		gui.SvgSpinnerRings,
		gui.SvgSpinnerSpinner,
		gui.SvgSpinnerSpinnerDouble,
		gui.SvgSpinnerSpinningCircles,
		gui.SvgSpinnerTailSpin,
	},
	"Dots": {
		gui.SvgSpinner12DotsScaleRotate,
		gui.SvgSpinner3DotsBounce,
		gui.SvgSpinner3DotsFade,
		gui.SvgSpinner3DotsMove,
		gui.SvgSpinner3DotsRotate,
		gui.SvgSpinner3DotsScale,
		gui.SvgSpinner3DotsScaleMiddle,
		gui.SvgSpinner6DotsRotate,
		gui.SvgSpinner6DotsScale,
		gui.SvgSpinner6DotsScaleMiddle,
		gui.SvgSpinner8DotsRotate,
		gui.SvgSpinnerBallTriangle,
		gui.SvgSpinnerDotRevolve,
	},
	"Bars": {
		gui.SvgSpinnerBars,
		gui.SvgSpinnerBarsFade,
		gui.SvgSpinnerBarsRotateFade,
		gui.SvgSpinnerBarsScale,
		gui.SvgSpinnerBarsScaleFade,
		gui.SvgSpinnerBarsScaleMiddle,
		gui.SvgSpinnerHorizontalBar,
	},
	"Loaders": {
		gui.SvgSpinnerLoader1,
		gui.SvgSpinnerLoader3,
		gui.SvgSpinnerLoader4,
		gui.SvgSpinnerLoader5,
		gui.SvgSpinnerLoader6,
		gui.SvgSpinnerLoader7,
		gui.SvgSpinnerLoader8,
		gui.SvgSpinnerLoader9,
		gui.SvgSpinnerLoader10,
	},
	"Blocks": {
		gui.SvgSpinnerBlocksScale,
		gui.SvgSpinnerBlocksShuffle2,
		gui.SvgSpinnerBlocksShuffle3,
		gui.SvgSpinnerBlocksWave,
	},
	"Pulse": {
		gui.SvgSpinnerPuff,
		gui.SvgSpinnerPulse,
		gui.SvgSpinnerPulse3,
		gui.SvgSpinnerPulseMultiple,
		gui.SvgSpinnerPulseRing,
		gui.SvgSpinnerPulseRings2,
		gui.SvgSpinnerPulseRings3,
		gui.SvgSpinnerPulseRingsMultiple,
	},
	"Miscellaneous": {
		gui.SvgSpinnerAudio,
		gui.SvgSpinnerBouncingBall,
		gui.SvgSpinnerClock,
		gui.SvgSpinnerGrid,
		gui.SvgSpinnerHearts,
		gui.SvgSpinnerTadpole,
		gui.SvgSpinnerWifi,
		gui.SvgSpinnerWifiFade,
		gui.SvgSpinnerWindToy,
	},
}

func demoSvgSpinner(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	t := gui.CurrentTheme()

	category := app.SvgSpinnerCategory
	if _, ok := svgSpinnerGroups[category]; !ok {
		category = svgSpinnerCategories[0]
	}

	kinds := svgSpinnerGroups[category]
	cells := make([]gui.View, len(kinds))
	for i, k := range kinds {
		cells[i] = gui.Column(gui.ContainerCfg{
			Sizing:     gui.FitFit,
			Padding:    gui.Some(t.PaddingMedium),
			SizeBorder: gui.NoBorder,
			HAlign:     gui.HAlignCenter,
			Spacing:    gui.SomeF(6),
			Content: []gui.View{
				gui.SvgSpinner(gui.SvgSpinnerCfg{
					ID:     "svg-spin-" + gui.SvgSpinnerName(k),
					Kind:   k,
					Width:  72,
					Height: 72,
				}),
				gui.Text(gui.TextCfg{
					Text:      gui.SvgSpinnerName(k),
					TextStyle: t.N4,
				}),
			},
		})
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:     gui.FillFit,
		Padding:    gui.NoPadding,
		SizeBorder: gui.NoBorder,
		Spacing:    gui.SomeF(12),
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:     gui.FillFit,
				Padding:    gui.NoPadding,
				SizeBorder: gui.NoBorder,
				Spacing:    gui.SomeF(8),
				VAlign:     gui.VAlignMiddle,
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text:      "Category",
						TextStyle: t.N3,
					}),
					gui.Select(gui.SelectCfg{
						ID:       "svg-spinner-category",
						Selected: []string{category},
						Options:  svgSpinnerCategories,
						OnSelect: func(sel []string, _ *gui.Event, w *gui.Window) {
							if len(sel) > 0 {
								gui.State[ShowcaseApp](w).SvgSpinnerCategory = sel[0]
							}
						},
					}),
				},
			}),
			gui.Wrap(gui.ContainerCfg{
				Sizing:     gui.FillFit,
				Spacing:    gui.SomeF(8),
				Padding:    gui.NoPadding,
				SizeBorder: gui.NoBorder,
				Content:    cells,
			}),
		},
	})
}
