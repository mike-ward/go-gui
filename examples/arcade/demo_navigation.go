package main

import "github.com/mike-ward/go-gui/gui"

func demoBreadcrumb(w *gui.Window) gui.View {
	app := gui.State[ArcadeApp](w)
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Breadcrumb(gui.BreadcrumbCfg{
				ID: "breadcrumb-demo",
				Items: []gui.BreadcrumbItemCfg{
					gui.NewBreadcrumbItem("home", "Home", nil),
					gui.NewBreadcrumbItem("docs", "Docs", nil),
					gui.NewBreadcrumbItem("api", "API Reference", nil),
					gui.NewBreadcrumbItem("widgets", "Widgets", nil),
				},
				Selected: app.BCSelected,
				OnSelect: func(id string, _ *gui.Event, w *gui.Window) {
					gui.State[ArcadeApp](w).BCSelected = id
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      "Selected: " + app.BCSelected,
				TextStyle: t.N3,
			}),
		},
	})
}

func demoTabControl(w *gui.Window) gui.View {
	app := gui.State[ArcadeApp](w)
	t := gui.CurrentTheme()
	return gui.TabControl(gui.TabControlCfg{
		ID: "tab-demo",
		Items: []gui.TabItemCfg{
			gui.NewTabItem("tab1", "Overview", []gui.View{
				gui.Column(gui.ContainerCfg{
					Sizing:  gui.FillFit,
					Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
					Content: []gui.View{
						gui.Text(gui.TextCfg{Text: "Overview tab content.", TextStyle: t.N3}),
					},
				}),
			}),
			gui.NewTabItem("tab2", "Details", []gui.View{
				gui.Column(gui.ContainerCfg{
					Sizing:  gui.FillFit,
					Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
					Content: []gui.View{
						gui.Text(gui.TextCfg{Text: "Details tab content.", TextStyle: t.N3}),
					},
				}),
			}),
			gui.NewTabItem("tab3", "Settings", []gui.View{
				gui.Column(gui.ContainerCfg{
					Sizing:  gui.FillFit,
					Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
					Content: []gui.View{
						gui.Text(gui.TextCfg{Text: "Settings tab content.", TextStyle: t.N3}),
					},
				}),
			}),
		},
		Selected: app.TabSelected,
		OnSelect: func(id string, _ *gui.Event, w *gui.Window) {
			gui.State[ArcadeApp](w).TabSelected = id
		},
	})
}

func demoMenus(w *gui.Window) gui.View {
	return gui.Menubar(w, gui.MenubarCfg{
		ID: "menubar-demo",
		Items: []gui.MenuItemCfg{
			gui.MenuSubmenu("file", "File", []gui.MenuItemCfg{
				gui.MenuItemText("new", "New"),
				gui.MenuItemText("open", "Open"),
				gui.MenuSeparator(),
				gui.MenuItemText("save", "Save"),
				gui.MenuItemText("save-as", "Save As..."),
			}),
			gui.MenuSubmenu("edit", "Edit", []gui.MenuItemCfg{
				gui.MenuItemText("undo", "Undo"),
				gui.MenuItemText("redo", "Redo"),
				gui.MenuSeparator(),
				gui.MenuItemText("cut", "Cut"),
				gui.MenuItemText("copy", "Copy"),
				gui.MenuItemText("paste", "Paste"),
			}),
			gui.MenuSubmenu("help", "Help", []gui.MenuItemCfg{
				gui.MenuItemText("about", "About"),
			}),
		},
		Action: func(id string, _ *gui.Event, w *gui.Window) {
			w.Toast(gui.ToastCfg{Title: "Menu", Body: "Action: " + id})
		},
	})
}

func demoCommandPalette(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)

	const paletteID = "cmd-palette"
	const paletteFocus uint32 = 500

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Button(gui.ButtonCfg{
				ID:      "btn-palette",
				Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Open Command Palette", TextStyle: t.N3}),
				},
				OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
					gui.CommandPaletteToggle(paletteID, paletteFocus, w)
					e.IsHandled = true
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      "Last action: " + app.PaletteAction,
				TextStyle: t.N3,
			}),
			gui.CommandPalette(gui.CommandPaletteCfg{
				ID:          paletteID,
				IDFocus:     paletteFocus,
				Placeholder: "Type a command...",
				Items: []gui.CommandPaletteItem{
					{ID: "new-file", Label: "New File", Icon: gui.IconPlus, Group: "File"},
					{ID: "open-file", Label: "Open File", Icon: gui.IconBook, Group: "File"},
					{ID: "save", Label: "Save", Icon: gui.IconDownload, Group: "File"},
					{ID: "toggle-theme", Label: "Toggle Theme", Icon: gui.IconMoon, Group: "View"},
					{ID: "search", Label: "Search", Icon: gui.IconSearch, Group: "Edit"},
				},
				OnAction: func(id string, _ *gui.Event, w *gui.Window) {
					gui.State[ArcadeApp](w).PaletteAction = id
				},
			}),
		},
	})
}
