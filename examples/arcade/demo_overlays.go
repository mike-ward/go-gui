package main

import "github.com/mike-ward/go-gui/gui"

func demoDialog(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						ID:      "btn-dialog-msg",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Message", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							w.Dialog(gui.DialogCfg{
								Title:      "Information",
								Body:       "This is a message dialog.",
								DialogType: gui.DialogMessage,
								OnOkYes: func(w *gui.Window) {
									gui.State[ArcadeApp](w).DialogResult = "OK"
								},
							})
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:      "btn-dialog-confirm",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Confirm", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							w.Dialog(gui.DialogCfg{
								Title:      "Confirm Action",
								Body:       "Are you sure you want to proceed?",
								DialogType: gui.DialogConfirm,
								OnOkYes: func(w *gui.Window) {
									gui.State[ArcadeApp](w).DialogResult = "Yes"
								},
								OnCancelNo: func(w *gui.Window) {
									gui.State[ArcadeApp](w).DialogResult = "No"
								},
							})
							e.IsHandled = true
						},
					}),
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      "Result: " + app.DialogResult,
				TextStyle: t.N3,
			}),
		},
	})
}

func demoTooltip(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(16)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.WithTooltip(w, gui.WithTooltipCfg{
				Text: "This is a tooltip!",
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						ID:      "btn-tooltip",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Hover me", TextStyle: t.N3}),
						},
					}),
				},
			}),
			gui.WithTooltip(w, gui.WithTooltipCfg{
				Text: "Another tooltip with more text",
				Content: []gui.View{
					gui.Badge(gui.BadgeCfg{
						Label:   "Tip",
						Variant: gui.BadgeInfo,
					}),
				},
			}),
		},
	})
}
