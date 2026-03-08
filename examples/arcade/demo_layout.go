package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mike-ward/go-gui/gui"
)

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

	ratioText := fmt.Sprintf("Ratio: %.0f%% / %.0f%%",
		app.SplitterState.Ratio*100, (1-app.SplitterState.Ratio)*100)
	collapseText := ""
	switch app.SplitterState.Collapsed {
	case gui.SplitterCollapseFirst:
		collapseText = " (first collapsed)"
	case gui.SplitterCollapseSecond:
		collapseText = " (second collapsed)"
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      ratioText + collapseText,
				TextStyle: t.N2,
			}),
			gui.Text(gui.TextCfg{
				Text:      "Drag divider, use collapse buttons, or double-click to collapse. Shift+Arrow for keyboard control.",
				TextStyle: t.N2,
			}),
			gui.Splitter(gui.SplitterCfg{
				ID:                  "demo-splitter",
				IDFocus:             51,
				Orientation:         gui.SplitterHorizontal,
				Sizing:              gui.FillFixed,
				Ratio:               app.SplitterState.Ratio,
				Collapsed:           app.SplitterState.Collapsed,
				DoubleClickCollapse: true,
				ShowCollapseButtons: true,
				OnChange: func(r float32, c gui.SplitterCollapsed, _ *gui.Event, w *gui.Window) {
					a := gui.State[ArcadeApp](w)
					a.SplitterState.Ratio = r
					a.SplitterState.Collapsed = c
				},
				First: gui.SplitterPaneCfg{
					MinSize: 80,
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
					MinSize: 80,
					Content: []gui.View{
						// Nested vertical splitter inside the second pane
						gui.Splitter(gui.SplitterCfg{
							ID:                  "demo-splitter-nested",
							Orientation:         gui.SplitterVertical,
							Sizing:              gui.FillFill,
							Ratio:               app.SplitterState2.Ratio,
							Collapsed:           app.SplitterState2.Collapsed,
							ShowCollapseButtons: true,
							OnChange: func(r float32, c gui.SplitterCollapsed, _ *gui.Event, w *gui.Window) {
								a := gui.State[ArcadeApp](w)
								a.SplitterState2.Ratio = r
								a.SplitterState2.Collapsed = c
							},
							First: gui.SplitterPaneCfg{
								MinSize: 40,
								Content: []gui.View{
									gui.Column(gui.ContainerCfg{
										Sizing:  gui.FillFill,
										Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
										Color:   t.ColorInterior,
										Content: []gui.View{
											gui.Text(gui.TextCfg{Text: "Top", TextStyle: t.B3}),
										},
									}),
								},
							},
							Second: gui.SplitterPaneCfg{
								MinSize: 40,
								Content: []gui.View{
									gui.Column(gui.ContainerCfg{
										Sizing:  gui.FillFill,
										Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
										Color:   t.ColorPanel,
										Content: []gui.View{
											gui.Text(gui.TextCfg{Text: "Bottom", TextStyle: t.B3}),
										},
									}),
								},
							},
						}),
					},
				},
			}),
		},
	})
}

func demoScrollbar(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()

	vItems := make([]gui.View, 30)
	for i := range vItems {
		vItems[i] = gui.Text(gui.TextCfg{
			Text:      fmt.Sprintf("Vertical item %d", i+1),
			TextStyle: t.N3,
		})
	}

	hItems := make([]gui.View, 15)
	for i := range hItems {
		hItems[i] = demoBoxSized(fmt.Sprintf("H%d", i+1), t.ColorActive, 80, 30)
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			sectionLabel(t, "Vertical Scrollbar"),
			gui.Column(gui.ContainerCfg{
				Sizing:        gui.FillFixed,
				Height:        180,
				IDScroll:      101,
				Padding:       gui.Some(gui.PaddingNone),
				Spacing:       gui.Some(float32(4)),
				ScrollbarCfgY: &gui.ScrollbarCfg{GapEdge: 4},
				Content:       vItems,
			}),
			sectionLabel(t, "Horizontal Scrollbar"),
			gui.Row(gui.ContainerCfg{
				Sizing:        gui.FillFixed,
				Height:        50,
				IDScroll:      102,
				Padding:       gui.Some(gui.PaddingNone),
				Spacing:       gui.Some(float32(4)),
				ScrollbarCfgX: &gui.ScrollbarCfg{GapEdge: 4},
				Content:       hItems,
			}),
		},
	})
}

func demoPrinting(w *gui.Window) gui.View {
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
						ID:      "btn-export-pdf",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: gui.IconDownload + " Export PDF", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							a := gui.State[ArcadeApp](w)
							outPath := filepath.Join(os.TempDir(), "arcade_export.pdf")
							job := gui.NewPrintJob()
							job.OutputPath = outPath
							job.Title = "Arcade Export"
							r := w.ExportPrintJob(job)
							if r.ErrorMessage != "" {
								a.PrintResult = "Error: " + r.ErrorMessage
							} else {
								a.PrintResult = "Exported: " + r.Path
							}
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:      "btn-print",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: gui.IconLayout + " Print", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							a := gui.State[ArcadeApp](w)
							job := gui.NewPrintJob()
							job.Title = "Arcade Print"
							r := w.RunPrintJob(job)
							if r.ErrorMessage != "" {
								a.PrintResult = "Error: " + r.ErrorMessage
							} else {
								a.PrintResult = "Print sent"
							}
							e.IsHandled = true
						},
					}),
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      app.PrintResult,
				TextStyle: t.N3,
			}),
		},
	})
}
