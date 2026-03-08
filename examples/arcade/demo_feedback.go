package main

import (
	"fmt"

	"github.com/mike-ward/go-gui/gui"
)

func demoButton(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("Clicks: %d", app.ButtonClicks),
				TextStyle: t.N4,
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				VAlign:  gui.VAlignMiddle,
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						ID:      "btn-primary",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Primary", TextStyle: t.B3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							gui.State[ArcadeApp](w).ButtonClicks++
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:          "btn-outline",
						Color:       gui.ColorTransparent,
						ColorBorder: t.ColorActive,
						SizeBorder:  gui.Some(float32(1)),
						Padding:     gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Outlined", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							gui.State[ArcadeApp](w).ButtonClicks++
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:       "btn-disabled",
						Padding:  gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Disabled: true,
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Disabled", TextStyle: t.N3}),
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:      "btn-icon",
						Padding: gui.Some(gui.NewPadding(8, 12, 8, 12)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{
								Text:      gui.IconHeart,
								TextStyle: t.Icon3,
							}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							gui.State[ArcadeApp](w).ButtonClicks++
							e.IsHandled = true
						},
					}),
				},
			}),
		},
	})
}

func demoProgressBar(_ *gui.Window) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.ProgressBar(gui.ProgressBarCfg{
				Percent:  0.25,
				TextShow: true,
				Sizing:   gui.FillFit,
			}),
			gui.ProgressBar(gui.ProgressBarCfg{
				Percent:  0.50,
				TextShow: true,
				Sizing:   gui.FillFit,
			}),
			gui.ProgressBar(gui.ProgressBarCfg{
				Percent:  0.75,
				TextShow: true,
				Sizing:   gui.FillFit,
			}),
			gui.ProgressBar(gui.ProgressBarCfg{
				Indefinite: true,
				Sizing:     gui.FillFit,
			}),
		},
	})
}

func demoPulsar(w *gui.Window) gui.View {
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.Some(gui.PaddingNone),
		Spacing: gui.Some(float32(8)),
		VAlign:  gui.VAlignMiddle,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Loading", TextStyle: gui.CurrentTheme().N3}),
			gui.Pulsar(gui.PulsarCfg{}, w),
		},
	})
}

func demoBadge(_ *gui.Window) gui.View {
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		VAlign:  gui.VAlignMiddle,
		Content: []gui.View{
			gui.Badge(gui.BadgeCfg{Label: "Default"}),
			gui.Badge(gui.BadgeCfg{Label: "Info", Variant: gui.BadgeInfo}),
			gui.Badge(gui.BadgeCfg{Label: "Success", Variant: gui.BadgeSuccess}),
			gui.Badge(gui.BadgeCfg{Label: "Warning", Variant: gui.BadgeWarning}),
			gui.Badge(gui.BadgeCfg{Label: "Error", Variant: gui.BadgeError}),
			gui.Badge(gui.BadgeCfg{Label: "42", Max: 99}),
			gui.Badge(gui.BadgeCfg{Dot: true, Variant: gui.BadgeError}),
		},
	})
}

func demoToast(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Button(gui.ButtonCfg{
				ID:      "btn-toast",
				Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Show Toast", TextStyle: t.N3}),
				},
				OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
					w.Toast(gui.ToastCfg{Title: "Arcade", Body: "Hello from Arcade!"})
					e.IsHandled = true
				},
			}),
			gui.Button(gui.ButtonCfg{
				ID:      "btn-toast-dismiss",
				Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Dismiss All", TextStyle: t.N3}),
				},
				OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
					w.ToastDismissAll()
					e.IsHandled = true
				},
			}),
		},
	})
}
