package main

import "github.com/mike-ward/go-gui/gui"

func demoRow(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Evenly spaced row:", TextStyle: t.B3}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					demoBox("A", t.ColorActive),
					demoBox("B", t.ColorSelect),
					demoBox("C", t.ColorHover),
				},
			}),
			gui.Text(gui.TextCfg{Text: "Right-aligned:", TextStyle: t.B3}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				HAlign:  gui.HAlignRight,
				Content: []gui.View{
					demoBox("X", t.ColorActive),
					demoBox("Y", t.ColorSelect),
				},
			}),
			gui.Text(gui.TextCfg{Text: "Center-aligned:", TextStyle: t.B3}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				HAlign:  gui.HAlignCenter,
				Content: []gui.View{
					demoBox("1", t.ColorActive),
					demoBox("2", t.ColorSelect),
				},
			}),
		},
	})
}

func demoColumn(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(16)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FitFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Top (default)", TextStyle: t.B3}),
					demoBox("A", t.ColorActive),
					demoBox("B", t.ColorSelect),
					demoBox("C", t.ColorHover),
				},
			}),
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FitFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				HAlign:  gui.HAlignCenter,
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Centered", TextStyle: t.B3}),
					demoBox("X", t.ColorActive),
					demoBox("Y", t.ColorSelect),
				},
			}),
		},
	})
}

func demoWrapPanel(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	colors := []gui.Color{t.ColorActive, t.ColorSelect, t.ColorHover, t.ColorFocus, t.ColorBorder}
	views := make([]gui.View, 12)
	for i := range views {
		views[i] = demoBoxSized("W", colors[i%len(colors)], 70, 40)
	}
	return gui.Wrap(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		Content: views,
	})
}

func demoOverflowPanel(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	views := make([]gui.View, 20)
	for i := range views {
		views[i] = demoBoxSized("Item", t.ColorActive, 100, 30)
	}
	return gui.Column(gui.ContainerCfg{
		Sizing:   gui.FillFixed,
		Height:   150,
		Overflow: true,
		IDScroll: 100,
		Spacing:  gui.Some(float32(4)),
		Padding:  gui.Some(gui.PaddingNone),
		Content:  views,
	})
}

func demoExpandPanel(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.ExpandPanel(gui.ExpandPanelCfg{
				ID:   "expand-1",
				Open: app.ExpandOpen,
				Head: gui.Text(gui.TextCfg{Text: "Click to expand", TextStyle: t.B3}),
				Content: gui.Column(gui.ContainerCfg{
					Sizing:  gui.FillFit,
					Padding: gui.Some(gui.NewPadding(8, 0, 8, 0)),
					Content: []gui.View{
						gui.Text(gui.TextCfg{
							Text:      "This content is revealed when the panel is expanded.",
							TextStyle: t.N3,
						}),
					},
				}),
				OnToggle: func(w *gui.Window) {
					a := gui.State[ArcadeApp](w)
					a.ExpandOpen = !a.ExpandOpen
				},
			}),
		},
	})
}

func demoSidebar(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Button(gui.ButtonCfg{
				ID:      "btn-sidebar-toggle",
				Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Toggle Sidebar", TextStyle: t.N3}),
				},
				OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
					a := gui.State[ArcadeApp](w)
					a.SidebarOpen = !a.SidebarOpen
					e.IsHandled = true
				},
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFixed,
				Height:  200,
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					w.Sidebar(gui.SidebarCfg{
						ID:    "demo-sidebar",
						Open:  app.SidebarOpen,
						Width: 200,
						Content: []gui.View{
							gui.Column(gui.ContainerCfg{
								Sizing:  gui.FillFill,
								Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
								Spacing: gui.Some(float32(8)),
								Content: []gui.View{
									gui.Text(gui.TextCfg{Text: "Sidebar", TextStyle: t.B4}),
									gui.Text(gui.TextCfg{
										Text:      "Slide-out panel content.",
										TextStyle: t.N3,
									}),
								},
							}),
						},
					}),
					gui.Column(gui.ContainerCfg{
						Sizing:  gui.FillFill,
						Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
						Color:   t.ColorPanel,
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Main content area", TextStyle: t.N3}),
						},
					}),
				},
			}),
		},
	})
}

func demoSplitter(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)
	return gui.Splitter(gui.SplitterCfg{
		ID:          "demo-splitter",
		Orientation: gui.SplitterHorizontal,
		Sizing:      gui.FillFixed,
		Ratio:       app.SplitterState.Ratio,
		Collapsed:   app.SplitterState.Collapsed,
		OnChange: func(r float32, c gui.SplitterCollapsed, _ *gui.Event, w *gui.Window) {
			a := gui.State[ArcadeApp](w)
			a.SplitterState.Ratio = r
			a.SplitterState.Collapsed = c
		},
		First: gui.SplitterPaneCfg{
			MinSize: 100,
			Content: []gui.View{
				gui.Column(gui.ContainerCfg{
					Sizing:  gui.FillFill,
					Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
					Color:   t.ColorPanel,
					Content: []gui.View{
						gui.Text(gui.TextCfg{Text: "First Pane", TextStyle: t.B3}),
					},
				}),
			},
		},
		Second: gui.SplitterPaneCfg{
			MinSize: 100,
			Content: []gui.View{
				gui.Column(gui.ContainerCfg{
					Sizing:  gui.FillFill,
					Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
					Color:   t.ColorInterior,
					Content: []gui.View{
						gui.Text(gui.TextCfg{Text: "Second Pane", TextStyle: t.B3}),
					},
				}),
			},
		},
		ShowCollapseButtons: true,
	})
}

func demoScrollbar(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	views := make([]gui.View, 30)
	for i := range views {
		views[i] = gui.Text(gui.TextCfg{Text: "Scrollable item", TextStyle: t.N3})
	}
	return gui.Column(gui.ContainerCfg{
		Sizing:        gui.FillFixed,
		Height:        200,
		IDScroll:      101,
		Padding:       gui.Some(gui.PaddingNone),
		Spacing:       gui.Some(float32(4)),
		ScrollbarCfgY: &gui.ScrollbarCfg{GapEdge: 4},
		Content:       views,
	})
}
