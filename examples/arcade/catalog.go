package main

import "github.com/mike-ward/go-gui/gui"

func catalogPanel(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)
	entries := filteredEntries(app)

	return gui.Column(gui.ContainerCfg{
		Width:   catalogWidth,
		Sizing:  gui.FixedFill,
		Color:   t.ColorPanel,
		Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
		Spacing: gui.Some(float32(8)),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Arcade", TextStyle: t.B5}),
			searchInput(app),
			groupPicker(app),
			line(),
			gui.Column(gui.ContainerCfg{
				IDScroll:      scrollCatalog,
				Sizing:        gui.FillFill,
				Padding:       gui.Some(gui.PaddingNone),
				Spacing:       gui.Some(float32(2)),
				ScrollbarCfgY: &gui.ScrollbarCfg{GapEdge: 4},
				Content:       catalogRows(entries, app),
			}),
			line(),
			bottomBar(),
		},
	})
}

func searchInput(app *ArcadeApp) gui.View {
	return gui.Input(gui.InputCfg{
		ID:          "nav-search",
		IDFocus:     focusSearch,
		Sizing:      gui.FillFit,
		Text:        app.NavQuery,
		Placeholder: "Search...",
		OnTextChanged: func(_ *gui.Layout, text string, w *gui.Window) {
			gui.State[ArcadeApp](w).NavQuery = text
		},
	})
}

func groupPicker(app *ArcadeApp) gui.View {
	t := gui.CurrentTheme()
	groups := demoGroups()
	views := make([]gui.View, len(groups))
	for i, g := range groups {
		selected := app.SelectedGroup == g.Key
		color := t.ColorInterior
		textStyle := t.N2
		if selected {
			color = t.ColorActive
			textStyle.Color = gui.RGB(255, 255, 255)
		}
		gKey := g.Key
		views[i] = gui.Button(gui.ButtonCfg{
			ID:               "grp-" + gKey,
			Color:            color,
			ColorBorder:      color,
			ColorBorderFocus: color,
			Radius:           gui.Some(float32(10)),
			Padding:          gui.Some(gui.NewPadding(4, 8, 4, 8)),
			Content: []gui.View{
				gui.Text(gui.TextCfg{Text: g.Label, TextStyle: textStyle}),
			},
			OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
				gui.State[ArcadeApp](w).SelectedGroup = gKey
				e.IsHandled = true
			},
		})
	}
	return gui.Wrap(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.Some(gui.PaddingNone),
		Spacing: gui.Some(float32(4)),
		Content: views,
	})
}

func catalogRows(entries []DemoEntry, app *ArcadeApp) []gui.View {
	t := gui.CurrentTheme()
	var views []gui.View
	currentGroup := ""

	for _, entry := range entries {
		if entry.Group != currentGroup && entry.Group != "welcome" {
			currentGroup = entry.Group
			views = append(views, gui.Column(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.Some(gui.NewPadding(8, 0, 2, 0)),
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text:      groupLabel(currentGroup),
						TextStyle: t.B2,
					}),
				},
			}))
		}

		selected := app.SelectedComponent == entry.ID
		color := gui.ColorTransparent
		if selected {
			color = t.ColorSelect
		}
		id := entry.ID
		label := entry.Label
		views = append(views, gui.Button(gui.ButtonCfg{
			ID:               "cat-" + id,
			Sizing:           gui.FillFit,
			Color:            color,
			ColorHover:       t.ColorHover,
			ColorClick:       t.ColorActive,
			ColorFocus:       color,
			ColorBorder:      gui.ColorTransparent,
			ColorBorderFocus: gui.ColorTransparent,
			Radius:           gui.Some(float32(4)),
			Padding:          gui.Some(gui.NewPadding(4, 10, 4, 10)),
			HAlign:           gui.HAlignLeft,
			Content: []gui.View{
				gui.Text(gui.TextCfg{Text: label, TextStyle: t.N3}),
			},
			OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
				gui.State[ArcadeApp](w).SelectedComponent = id
				e.IsHandled = true
			},
		}))
	}
	return views
}

func bottomBar() gui.View {
	t := gui.CurrentTheme()
	icon := gui.IconMoon
	if t.TitlebarDark {
		icon = gui.IconSunnyO
	}
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.Some(gui.PaddingNone),
		HAlign:  gui.HAlignRight,
		Content: []gui.View{
			gui.Button(gui.ButtonCfg{
				ID:               "theme-toggle",
				Color:            gui.ColorTransparent,
				ColorBorder:      gui.ColorTransparent,
				ColorBorderFocus: t.ColorActive,
				Padding:          gui.Some(gui.NewPadding(6, 8, 6, 8)),
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text:      icon,
						TextStyle: t.Icon3,
					}),
				},
				OnClick: func(_ *gui.Layout, e *gui.Event, _ *gui.Window) {
					if gui.CurrentTheme().TitlebarDark {
						gui.SetTheme(gui.ThemeLight)
					} else {
						gui.SetTheme(gui.ThemeDark)
					}
					e.IsHandled = true
				},
			}),
		},
	})
}
