package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mike-ward/go-gui/gui"
)

func demoRotatedBox(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	labels := []string{"0°", "90°", "180°", "270°"}
	colors := []gui.Color{t.ColorActive, t.ColorSelect, t.ColorHover, t.ColorFocus}
	boxes := make([]gui.View, 4)
	for i := range boxes {
		boxes[i] = gui.Column(gui.ContainerCfg{
			Sizing:  gui.FitFit,
			Spacing: gui.SomeF(4),
			Padding: gui.NoPadding,
			HAlign:  gui.HAlignCenter,
			Content: []gui.View{
				gui.Text(gui.TextCfg{Text: labels[i], TextStyle: t.B4}),
				gui.RotatedBox(gui.RotatedBoxCfg{
					QuarterTurns: i,
					Content:      demoBoxSized("R", colors[i], 80, 50),
				}),
			},
		})
	}
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(16),
				Padding: gui.NoPadding,
				Content: boxes,
			}),
		},
	})
}

func demoRow(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Evenly spaced row:", TextStyle: t.B3}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(8),
				Padding: gui.NoPadding,
				Content: []gui.View{
					demoBox("A", t.ColorActive),
					demoBox("B", t.ColorSelect),
					demoBox("C", t.ColorHover),
				},
			}),
			gui.Text(gui.TextCfg{Text: "Right-aligned:", TextStyle: t.B3}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(8),
				Padding: gui.NoPadding,
				HAlign:  gui.HAlignRight,
				Content: []gui.View{
					demoBox("X", t.ColorActive),
					demoBox("Y", t.ColorSelect),
				},
			}),
			gui.Text(gui.TextCfg{Text: "Center-aligned:", TextStyle: t.B3}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(8),
				Padding: gui.NoPadding,
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
		Spacing: gui.SomeF(16),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FitFit,
				Spacing: gui.SomeF(8),
				Padding: gui.NoPadding,
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Top (default)", TextStyle: t.B3}),
					demoBox("A", t.ColorActive),
					demoBox("B", t.ColorSelect),
					demoBox("C", t.ColorHover),
				},
			}),
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FitFit,
				Spacing: gui.SomeF(8),
				Padding: gui.NoPadding,
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
		Spacing: gui.SomeF(8),
		Padding: gui.NoPadding,
		Content: views,
	})
}

func demoOverflowPanel(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	colors := []gui.Color{
		t.ColorActive, t.Cfg.ColorSuccess, t.Cfg.ColorWarning, t.Cfg.ColorError,
		t.ColorActive, t.Cfg.ColorSuccess, t.Cfg.ColorWarning, t.Cfg.ColorError,
	}
	items := make([]gui.OverflowItem, len(colors))
	for i := range items {
		label := fmt.Sprintf("Item %d", i+1)
		items[i] = gui.OverflowItem{
			ID: label,
			View: gui.Button(gui.ButtonCfg{
				Color:   colors[i],
				Content: []gui.View{gui.Text(gui.TextCfg{Text: label})},
			}),
			Text:   label,
			Action: func(*gui.MenuItemCfg, *gui.Event, *gui.Window) {},
		}
	}
	return gui.OverflowPanel(w, gui.OverflowPanelCfg{
		ID:      "overflow_panel_demo",
		IDFocus: 200,
		Items:   items,
	})
}

func demoExpandPanel(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ShowcaseApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.ExpandPanel(gui.ExpandPanelCfg{
				ID:   "expand-1",
				Open: app.ExpandOpen,
				Head: gui.Text(gui.TextCfg{Text: "Click to expand", TextStyle: t.B3}),
				Content: gui.Column(gui.ContainerCfg{
					Sizing:  gui.FillFit,
					Padding: gui.SomeP(8, 0, 8, 0),
					Content: []gui.View{
						gui.Text(gui.TextCfg{
							Text:      "This content is revealed when the panel is expanded.",
							TextStyle: t.N3,
						}),
					},
				}),
				OnToggle: func(w *gui.Window) {
					a := gui.State[ShowcaseApp](w)
					a.ExpandOpen = !a.ExpandOpen
				},
			}),
		},
	})
}

func demoSidebar(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ShowcaseApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Button(gui.ButtonCfg{
				ID:      "btn-sidebar-toggle",
				Padding: gui.SomeP(8, 16, 8, 16),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Toggle Sidebar", TextStyle: t.N3}),
				},
				OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
					a := gui.State[ShowcaseApp](w)
					a.SidebarOpen = !a.SidebarOpen
					e.IsHandled = true
				},
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFixed,
				Height:  200,
				Padding: gui.NoPadding,
				Content: []gui.View{
					w.Sidebar(gui.SidebarCfg{
						ID:    "demo-sidebar",
						Open:  app.SidebarOpen,
						Width: 200,
						Content: []gui.View{
							gui.Column(gui.ContainerCfg{
								Sizing:  gui.FillFill,
								Padding: gui.SomeP(12, 12, 12, 12),
								Spacing: gui.SomeF(8),
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
						Padding: gui.SomeP(12, 12, 12, 12),
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
	app := gui.State[ShowcaseApp](w)
	t := gui.CurrentTheme()
	mainRatio := int(app.SplitterMainState.Ratio * 100)
	detailRatio := int(app.SplitterDetailState.Ratio * 100)

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(t.SpacingSmall),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Drag handle. Focus splitter, then use arrow keys. Shift+arrow moves faster.",
				TextStyle: t.N5,
				Mode:      gui.TextModeWrap,
			}),
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("Main %d%% (%s), detail %d%% (%s).", mainRatio, splitterCollapsedLabel(app.SplitterMainState.Collapsed), detailRatio, splitterCollapsedLabel(app.SplitterDetailState.Collapsed)),
				TextStyle: t.N5,
				Mode:      gui.TextModeWrap,
			}),
			gui.Row(gui.ContainerCfg{
				Height:      373,
				Sizing:      gui.FillFixed,
				Color:       t.ColorPanel,
				ColorBorder: t.ColorBorder,
				SizeBorder:  gui.SomeF(1),
				Radius:      gui.Some(t.RadiusSmall),
				Padding:     gui.NoPadding,
				Content: []gui.View{
					showcaseSplitterMain(w),
				},
			}),
		},
	})
}

func showcaseSplitterMain(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	return gui.Splitter(gui.SplitterCfg{
		ID:          "catalog-splitter-main",
		IDFocus:     9160,
		Sizing:      gui.FillFill,
		Orientation: gui.SplitterHorizontal,
		Ratio:       gui.SomeF(app.SplitterMainState.Ratio),
		Collapsed:   app.SplitterMainState.Collapsed,
		OnChange:    onShowcaseSplitterMainChange,
		First: gui.SplitterPaneCfg{
			MinSize: 140,
			MaxSize: 340,
			Content: []gui.View{
				showcaseSplitterPane("Project", "- src\n- docs\n- tests", gui.CornflowerBlue),
			},
		},
		Second: gui.SplitterPaneCfg{
			MinSize: 220,
			Content: []gui.View{
				showcaseSplitterDetail(w),
			},
		},
	})
}

func showcaseSplitterDetail(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	return gui.Splitter(gui.SplitterCfg{
		ID:                  "catalog-splitter-detail",
		IDFocus:             9161,
		Orientation:         gui.SplitterVertical,
		Sizing:              gui.FillFill,
		HandleSize:          gui.SomeF(10),
		ShowCollapseButtons: true,
		Ratio:               gui.SomeF(app.SplitterDetailState.Ratio),
		Collapsed:           app.SplitterDetailState.Collapsed,
		OnChange:            onShowcaseSplitterDetailChange,
		First: gui.SplitterPaneCfg{
			MinSize: 110,
			Content: []gui.View{
				showcaseSplitterPane("Editor", "Top pane. Home/End collapses pane.", gui.Green),
			},
		},
		Second: gui.SplitterPaneCfg{
			MinSize: 90,
			Content: []gui.View{
				showcaseSplitterPane("Preview", "Bottom pane. Drag or use keyboard.", gui.Orange),
			},
		},
	})
}

func showcaseSplitterPane(title, note string, accent gui.Color) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Padding: gui.SomeP(10, 10, 10, 10),
		Spacing: gui.SomeF(6),
		Color:   t.ColorPanel,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.NoPadding,
				VAlign:  gui.VAlignMiddle,
				Spacing: gui.SomeF(8),
				Content: []gui.View{
					gui.Row(gui.ContainerCfg{
						Width:   8,
						Height:  8,
						Sizing:  gui.FixedFixed,
						Color:   accent,
						Padding: gui.NoPadding,
						Radius:  gui.SomeF(4),
					}),
					gui.Text(gui.TextCfg{Text: title, TextStyle: t.B5}),
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      note,
				TextStyle: t.N5,
				Mode:      gui.TextModeWrap,
			}),
			gui.Rectangle(gui.RectangleCfg{
				Sizing:      gui.FillFill,
				Color:       t.ColorBackground,
				ColorBorder: t.ColorBorder,
				SizeBorder:  1,
				Radius:      t.RadiusSmall,
			}),
		},
	})
}

func onShowcaseSplitterMainChange(ratio float32, collapsed gui.SplitterCollapsed, _ *gui.Event, w *gui.Window) {
	app := gui.State[ShowcaseApp](w)
	app.SplitterMainState = gui.SplitterStateNormalize(gui.SplitterState{
		Ratio:     ratio,
		Collapsed: collapsed,
	})
}

func onShowcaseSplitterDetailChange(ratio float32, collapsed gui.SplitterCollapsed, _ *gui.Event, w *gui.Window) {
	app := gui.State[ShowcaseApp](w)
	app.SplitterDetailState = gui.SplitterStateNormalize(gui.SplitterState{
		Ratio:     ratio,
		Collapsed: collapsed,
	})
}

func splitterCollapsedLabel(collapsed gui.SplitterCollapsed) string {
	switch collapsed {
	case gui.SplitterCollapseFirst:
		return "first"
	case gui.SplitterCollapseSecond:
		return "second"
	default:
		return "none"
	}
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
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			sectionLabel(t, "Vertical Scrollbar"),
			gui.Column(gui.ContainerCfg{
				Sizing:        gui.FillFixed,
				Height:        180,
				Overflow:      true,
				IDScroll:      101,
				Padding:       gui.NoPadding,
				Spacing:       gui.SomeF(4),
				ScrollbarCfgY: &gui.ScrollbarCfg{GapEdge: 4},
				Content:       vItems,
			}),
			sectionLabel(t, "Horizontal Scrollbar"),
			gui.Row(gui.ContainerCfg{
				Sizing:        gui.FillFixed,
				Height:        50,
				Overflow:      true,
				IDScroll:      102,
				ScrollMode:    gui.ScrollHorizontalOnly,
				Padding:       gui.NoPadding,
				Spacing:       gui.SomeF(4),
				ScrollbarCfgX: &gui.ScrollbarCfg{GapEdge: 4},
				Content:       hItems,
			}),
		},
	})
}

func demoPrinting(w *gui.Window) gui.View {
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
						ID:      "btn-export-pdf",
						Padding: gui.SomeP(8, 16, 8, 16),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: gui.IconExport, TextStyle: t.N3}),
							gui.Text(gui.TextCfg{Text: "Export PDF", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							a := gui.State[ShowcaseApp](w)
							outPath := filepath.Join(os.TempDir(), "showcase_export.pdf")
							job := gui.NewPrintJob()
							job.OutputPath = outPath
							job.Title = "showcase Export"
							r := w.ExportPrintJob(job)
							if r.ErrorMessage != "" {
								a.PrintingStatus = "Error: " + r.ErrorMessage
							} else {
								a.PrintingLastPath = r.Path
								a.PrintingStatus = "Exported: " + r.Path
							}
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:      "btn-print",
						Padding: gui.SomeP(8, 16, 8, 16),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: gui.IconPrint, TextStyle: t.N3}),
							gui.Text(gui.TextCfg{Text: "Print", TextStyle: t.N3}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							a := gui.State[ShowcaseApp](w)
							job := gui.NewPrintJob()
							job.Title = "showcase Print"
							r := w.RunPrintJob(job)
							if r.ErrorMessage != "" {
								a.PrintingStatus = "Error: " + r.ErrorMessage
							} else {
								a.PrintingStatus = "Print sent"
							}
							e.IsHandled = true
						},
					}),
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      app.PrintingStatus,
				TextStyle: t.N3,
			}),
		},
	})
}
