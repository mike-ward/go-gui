package main

import "github.com/mike-ward/go-gui/gui"

func demoDialog(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ShowcaseApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(8),
				Padding: gui.NoPadding,
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
									gui.State[ShowcaseApp](w).DialogResult = "OK"
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
									gui.State[ShowcaseApp](w).DialogResult = "Yes"
								},
								OnCancelNo: func(w *gui.Window) {
									gui.State[ShowcaseApp](w).DialogResult = "No"
								},
							})
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:      "btn-dialog-prompt",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Prompt", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							w.Dialog(gui.DialogCfg{
								Title:      "Input Required",
								Body:       "Enter your name:",
								DialogType: gui.DialogPrompt,
								OnReply: func(reply string, w *gui.Window) {
									gui.State[ShowcaseApp](w).DialogResult = "Reply: " + reply
								},
								OnCancelNo: func(w *gui.Window) {
									gui.State[ShowcaseApp](w).DialogResult = "Cancelled"
								},
							})
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:      "btn-dialog-custom",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Custom", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							w.Dialog(gui.DialogCfg{
								Title:      "Custom Dialog",
								DialogType: gui.DialogCustom,
								CustomContent: []gui.View{
									gui.Column(gui.ContainerCfg{
										Sizing:  gui.FillFit,
										Spacing: gui.SomeF(8),
										Padding: gui.Some(gui.NewPadding(8, 8, 8, 8)),
										Content: []gui.View{
											gui.Text(gui.TextCfg{
												Text:      "This dialog has custom content.",
												TextStyle: t.N3,
											}),
											gui.ProgressBar(gui.ProgressBarCfg{
												Percent:  0.65,
												TextShow: true,
												Sizing:   gui.FillFit,
											}),
										},
									}),
								},
								OnOkYes: func(w *gui.Window) {
									gui.State[ShowcaseApp](w).DialogResult = "Custom OK"
								},
							})
							e.IsHandled = true
						},
					}),
				},
			}),
			line(),
			sectionLabel(t, "Native File Dialogs"),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(8),
				Padding: gui.NoPadding,
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						ID:      "btn-open-file",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: gui.IconFolder + " Open", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							np := w.NativePlatformBackend()
							if np == nil {
								gui.State[ShowcaseApp](w).DialogResult = "No native platform"
								e.IsHandled = true
								return
							}
							r := np.ShowOpenDialog("Open File", "", nil, false)
							a := gui.State[ShowcaseApp](w)
							if len(r.Paths) > 0 {
								a.DialogResult = "Opened: " + r.Paths[0].Path
							} else {
								a.DialogResult = "Open cancelled"
							}
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:      "btn-save-file",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: gui.IconDownload + " Save", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							np := w.NativePlatformBackend()
							if np == nil {
								gui.State[ShowcaseApp](w).DialogResult = "No native platform"
								e.IsHandled = true
								return
							}
							r := np.ShowSaveDialog("Save File", "", "untitled.txt", ".txt", nil, true)
							a := gui.State[ShowcaseApp](w)
							if len(r.Paths) > 0 {
								a.DialogResult = "Save: " + r.Paths[0].Path
							} else {
								a.DialogResult = "Save cancelled"
							}
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:      "btn-folder",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: gui.IconFolder + " Folder", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							np := w.NativePlatformBackend()
							if np == nil {
								gui.State[ShowcaseApp](w).DialogResult = "No native platform"
								e.IsHandled = true
								return
							}
							r := np.ShowFolderDialog("Select Folder", "")
							a := gui.State[ShowcaseApp](w)
							if len(r.Paths) > 0 {
								a.DialogResult = "Folder: " + r.Paths[0].Path
							} else {
								a.DialogResult = "Folder cancelled"
							}
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

func demoNotification(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ShowcaseApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Button(gui.ButtonCfg{
				ID:      "btn-notify",
				Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: gui.IconBell + " Send Notification", TextStyle: t.N3}),
				},
				OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
					w.NativeNotification(gui.NativeNotificationCfg{
						Title: "showcase",
						Body:  "Hello from go-gui showcase!",
						OnDone: func(r gui.NativeNotificationResult, w *gui.Window) {
							a := gui.State[ShowcaseApp](w)
							switch r.Status {
							case gui.NotificationOK:
								a.NotifyResult = "Notification sent"
							case gui.NotificationDenied:
								a.NotifyResult = "Permission denied"
							default:
								a.NotifyResult = "Error: " + r.ErrorMessage
							}
						},
					})
					e.IsHandled = true
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      app.NotifyResult,
				TextStyle: t.N3,
			}),
		},
	})
}

func demoTooltip(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(16),
		Padding: gui.NoPadding,
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
