package main

import (
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

func detailPanel(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	entries := filteredEntries(app)

	if len(entries) == 0 {
		return gui.Column(gui.ContainerCfg{
			IDScroll:   scrollDetail,
			Sizing:     gui.FillFill,
			SizeBorder: gui.NoBorder,
			Padding:    gui.Some(detailPanelPadding()),
			ScrollbarCfgY: &gui.ScrollbarCfg{
				GapEdge: 4,
			},
			Content: []gui.View{
				gui.Text(gui.TextCfg{
					Text:      "No component matches filter",
					TextStyle: gui.CurrentTheme().B2,
				}),
			},
		})
	}

	if !hasEntry(entries, app.SelectedComponent) {
		app.SelectedComponent = preferredComponentForGroup(app.SelectedGroup, entries)
	}

	entry := selectedEntry(entries, app.SelectedComponent)

	var content gui.View
	switch {
	case entry.ID == "":
		content = demoPlaceholder(gui.CurrentTheme(), "No demo configured")
	case app.ShowDocs && entry.Group != groupWelcome:
		style := gui.DefaultMarkdownStyle()
		style.H2 = gui.CurrentTheme().B3
		content = w.Markdown(gui.MarkdownCfg{
			ID:      "doc-" + entry.ID,
			Source:  componentDoc(entry.ID),
			Padding: gui.NoPadding,
			Style:   style,
		})
	default:
		content = componentDemo(w, entry.ID)
	}

	return gui.Column(gui.ContainerCfg{
		IDScroll:   scrollDetail,
		Sizing:     gui.FillFill,
		Color:      gui.CurrentTheme().ColorBackground,
		SizeBorder: gui.NoBorder,
		Padding:    gui.Some(detailPanelPadding()),
		Spacing:    gui.Some(gui.CurrentTheme().SpacingLarge),
		ScrollbarCfgY: &gui.ScrollbarCfg{
			GapEdge: 4,
		},
		Content: []gui.View{
			viewTitleBar(entry, app.ShowDocs),
			showcaseWrappedText(entry.Summary, gui.CurrentTheme().N3),
			content,
			line(),
			relatedExamplesFooter(entry.ID),
		},
	})
}

func detailPanelPadding() gui.Padding {
	base := gui.CurrentTheme().PaddingLarge
	base.Right += gui.CurrentTheme().ScrollbarStyle.Size + 4
	return base
}

func viewTitleBar(entry DemoEntry, showDocs bool) gui.View {
	titleContent := []gui.View{
		gui.Text(gui.TextCfg{Text: entry.Label, TextStyle: gui.CurrentTheme().B1}),
	}
	if entry.ID != "welcome" && !strings.HasPrefix(entry.ID, "doc_") {
		titleContent = append(titleContent,
			gui.Row(gui.ContainerCfg{
				Sizing:     gui.FillFit,
				Padding:    gui.NoPadding,
				SizeBorder: gui.NoBorder,
			}),
			docButton(showDocs),
		)
	}
	return gui.Column(gui.ContainerCfg{
		Sizing:     gui.FillFit,
		Spacing:    gui.NoSpacing,
		Padding:    gui.NoPadding,
		SizeBorder: gui.NoBorder,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:     gui.FillFit,
				Padding:    gui.NoPadding,
				SizeBorder: gui.NoBorder,
				VAlign:     gui.VAlignMiddle,
				Content:    titleContent,
			}),
			line(),
		},
	})
}

func docButton(showDocs bool) gui.View {
	color := gui.ColorTransparent
	if showDocs {
		color = gui.CurrentTheme().ColorActive
	}
	return gui.Button(gui.ButtonCfg{
		ID:         "btn-doc-toggle",
		A11YLabel:  "Toggle docs",
		Color:      color,
		SizeBorder: gui.NoBorder,
		Padding:    gui.Some(gui.NewPadding(4, 8, 4, 8)),
		Radius:     gui.SomeF(3),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: gui.IconBook, TextStyle: gui.CurrentTheme().Icon4}),
		},
		OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
			gui.State[ShowcaseApp](w).ShowDocs = !gui.State[ShowcaseApp](w).ShowDocs
			e.IsHandled = true
		},
	})
}

func componentDemo(w *gui.Window, id string) gui.View {
	switch id {
	case "welcome":
		return demoWelcome(w)

	case "button":
		return demoButton(w)
	case "input":
		return demoInput(w)
	case "toggle":
		return demoToggle(w)
	case "switch":
		return demoSwitch(w)
	case "radio":
		return demoRadio(w)
	case "radio_group":
		return demoRadioGroup(w)
	case "combobox":
		return demoCombobox(w)
	case "select":
		return demoSelect(w)
	case "listbox":
		return demoListBox(w)
	case "range_slider":
		return demoRangeSlider(w)
	case "progress_bar":
		return demoProgressBar(w)
	case "pulsar":
		return demoPulsar(w)
	case "toast":
		return demoToast(w)
	case "badge":
		return demoBadge(w)
	case "native_notification":
		return demoNotification(w)
	case "breadcrumb":
		return demoBreadcrumb(w)
	case "menus":
		return demoMenus(w)
	case "dialog":
		return demoDialog(w)
	case "tree":
		return demoTree(w)
	case "drag_reorder":
		return demoDragReorder(w)
	case "printing":
		return demoPrinting(w)
	case "text":
		return demoText(w)
	case "rtf":
		return demoRtf(w)
	case "table":
		return demoTable(w)
	case "data_grid":
		return demoDataGrid(w)
	case "data_source":
		return demoDataSource(w)
	case "date_picker":
		return demoDatePicker(w)
	case "input_date":
		return demoInputDate(w)
	case "numeric_input":
		return demoNumericInput(w)
	case "forms":
		return demoForms(w)
	case "date_picker_roller":
		return demoDatePickerRoller(w)
	case "svg":
		return demoSvg(w)
	case "image":
		return demoImage(w)
	case "expand_panel":
		return demoExpandPanel(w)
	case "icons":
		return demoIcons(w)
	case "gradient":
		return demoGradient(w)
	case "box_shadows":
		return demoBoxShadows(w)
	case "shader":
		return demoShader(w)
	case "animations":
		return demoAnimations(w)
	case "color_picker":
		return demoColorPicker(w)
	case "theme_gen":
		return demoThemeGen(w)
	case "markdown":
		return demoMarkdown(w)
	case "tab_control":
		return demoTabControl(w)
	case "command_palette":
		return demoCommandPalette(w)
	case "tooltip":
		return demoTooltip(w)
	case "inspector":
		return demoInspector(w)
	case "rectangle":
		return demoRectangle(w)
	case "scrollbar":
		return demoScrollbar(w)
	case "splitter":
		return demoSplitter(w)
	case "row":
		return demoRow(w)
	case "column_demo":
		return demoColumn(w)
	case "wrap_panel":
		return demoWrapPanel(w)
	case "overflow_panel":
		return demoOverflowPanel(w)
	case "sidebar":
		return demoSidebar(w)

	case "doc_get_started", "doc_animations", "doc_architecture", "doc_containers",
		"doc_custom_widgets", "doc_data_grid", "doc_forms", "doc_gradients",
		"doc_layout_algorithm", "doc_locales", "doc_markdown", "doc_native_dialogs",
		"doc_performance", "doc_printing", "doc_shaders", "doc_splitter",
		"doc_svg", "doc_tables", "doc_tree", "doc_themes":
		return demoDocPage(w, id)
	default:
		return demoPlaceholder(gui.CurrentTheme(), "Demo: "+id)
	}
}

func demoPlaceholder(t gui.Theme, text string) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Color:   t.ColorPanel,
		Padding: gui.Some(gui.NewPadding(24, 24, 24, 24)),
		Radius:  gui.SomeF(8),
		Content: []gui.View{
			showcaseWrappedText(text, t.N3),
		},
	})
}

func relatedExamplesFooter(id string) gui.View {
	return gui.Text(gui.TextCfg{
		Text:      "Related examples: " + relatedExamples(id),
		TextStyle: gui.CurrentTheme().N5,
		Mode:      gui.TextModeWrap,
	})
}
