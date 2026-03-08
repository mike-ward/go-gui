package main

import "github.com/mike-ward/go-gui/gui"

func detailPanel(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)

	var current DemoEntry
	for _, e := range demoEntries() {
		if e.ID == app.SelectedComponent {
			current = e
			break
		}
	}

	return gui.Column(gui.ContainerCfg{
		IDScroll:      scrollDetail,
		Sizing:        gui.FillFill,
		Color:         t.ColorBackground,
		Padding:       gui.Some(gui.NewPadding(16, 24, 16, 24)),
		Spacing:       gui.Some(float32(12)),
		ScrollbarCfgY: &gui.ScrollbarCfg{GapEdge: 4},
		Content: []gui.View{
			viewTitleBar(current),
			gui.Text(gui.TextCfg{Text: current.Summary, TextStyle: t.N3}),
			line(),
			componentDemo(w, app.SelectedComponent),
		},
	})
}

func viewTitleBar(entry DemoEntry) gui.View {
	t := gui.CurrentTheme()
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.Some(gui.PaddingNone),
		VAlign:  gui.VAlignMiddle,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: entry.Label, TextStyle: t.B5}),
		},
	})
}

func componentDemo(w *gui.Window, id string) gui.View {
	switch id {
	case "welcome":
		return demoWelcome(w)

	// Input
	case "input":
		return demoInput(w)
	case "numeric_input":
		return demoNumericInput(w)
	case "color_picker":
		return demoColorPicker(w)
	case "date_picker":
		return demoDatePicker(w)
	case "date_picker_roller":
		return demoDatePickerRoller(w)
	case "input_date":
		return demoInputDate(w)
	case "forms":
		return demoForms(w)

	// Selection
	case "toggle":
		return demoToggle(w)
	case "switch":
		return demoSwitch(w)
	case "radio":
		return demoRadio(w)
	case "radio_group":
		return demoRadioGroup(w)
	case "select":
		return demoSelect(w)
	case "listbox":
		return demoListBox(w)
	case "combobox":
		return demoCombobox(w)
	case "range_slider":
		return demoRangeSlider(w)

	// Data
	case "table":
		return demoTable(w)
	case "data_grid":
		return demoDataGrid(w)

	// Text
	case "text":
		return demoText(w)
	case "rtf":
		return demoRtf(w)
	case "markdown":
		return demoMarkdown(w)

	// Graphics
	case "svg":
		return demoSvg(w)
	case "image":
		return demoImage(w)
	case "gradient":
		return demoGradient(w)
	case "box_shadows":
		return demoBoxShadows(w)
	case "rectangle":
		return demoRectangle(w)
	case "icons":
		return demoIcons(w)

	// Layout
	case "row":
		return demoRow(w)
	case "column":
		return demoColumn(w)
	case "wrap_panel":
		return demoWrapPanel(w)
	case "overflow_panel":
		return demoOverflowPanel(w)
	case "expand_panel":
		return demoExpandPanel(w)
	case "sidebar":
		return demoSidebar(w)
	case "splitter":
		return demoSplitter(w)
	case "scrollbar":
		return demoScrollbar(w)

	// Navigation
	case "breadcrumb":
		return demoBreadcrumb(w)
	case "tab_control":
		return demoTabControl(w)
	case "menus":
		return demoMenus(w)
	case "command_palette":
		return demoCommandPalette(w)

	// Feedback
	case "button":
		return demoButton(w)
	case "progress_bar":
		return demoProgressBar(w)
	case "pulsar":
		return demoPulsar(w)
	case "toast":
		return demoToast(w)
	case "badge":
		return demoBadge(w)

	// Overlays
	case "dialog":
		return demoDialog(w)
	case "tooltip":
		return demoTooltip(w)

	// Animations
	case "animations":
		return demoAnimations(w)

	// Theme
	case "theme_gen":
		return demoThemeGen(w)

	// Locale
	case "locale":
		return demoLocale(w)

	// Docs
	case "doc_get_started":
		return demoDoc(w, docGetStarted)
	case "doc_architecture":
		return demoDoc(w, docArchitecture)
	case "doc_containers":
		return demoDoc(w, docContainers)
	case "doc_themes":
		return demoDoc(w, docThemes)
	case "doc_animations":
		return demoDoc(w, docAnimations)
	case "doc_locales":
		return demoDoc(w, docLocales)

	default:
		return demoPlaceholder(gui.CurrentTheme(), "Demo: "+id)
	}
}

func demoPlaceholder(t gui.Theme, text string) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Color:   t.ColorPanel,
		Padding: gui.Some(gui.NewPadding(24, 24, 24, 24)),
		Radius:  gui.Some(float32(8)),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: text, TextStyle: t.N4}),
		},
	})
}
