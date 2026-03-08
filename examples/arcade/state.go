package main

import (
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

// ArcadeApp holds all state for the Arcade showcase.
type ArcadeApp struct {
	// Navigation
	NavQuery          string
	SelectedGroup     string
	SelectedComponent string

	// Docs toggle
	ShowDocs bool

	// Button
	ButtonClicks int

	// Input
	InputText, InputPassword, InputPhone, InputExpiry string
	InputMultiline                                    string

	// Toggle / Switch
	ToggleA, CheckboxA, SwitchA bool

	// Radio
	RadioValue string

	// Select
	SelectValue []string

	// ListBox
	ListBoxSelected []string

	// Combobox
	ComboboxValue string

	// Range slider
	RangeValue float32

	// Color picker
	ColorPickerColor gui.Color

	// Breadcrumb
	BCSelected string

	// Tabs
	TabSelected string

	// Splitter
	SplitterState gui.SplitterState

	// Sidebar
	SidebarOpen bool

	// Expand panel
	ExpandOpen bool

	// Dialog
	DialogResult string

	// Command palette
	PaletteAction string

	// Animations
	AnimTweenX    float32
	AnimSpringX   float32
	AnimKeyframeX float32

	// Theme generator
	ThemeGenSeed     gui.Color
	ThemeGenStrategy string
	ThemeGenTint     float32
}

func newArcadeApp() *ArcadeApp {
	return &ArcadeApp{
		SelectedGroup:     "all",
		SelectedComponent: "welcome",
		RangeValue:        50,
		RadioValue:        "go",
		BCSelected:        "home",
		TabSelected:       "tab1",
		SplitterState:     gui.SplitterState{Ratio: 0.5},
		ColorPickerColor:  gui.ColorFromString("#3b82f6"),
		ThemeGenSeed:      gui.ThemeDarkBorderedCfg.ColorSelect,
		ThemeGenStrategy:  "mono",
	}
}

// DemoEntry describes a single demonstrable component.
type DemoEntry struct {
	ID, Label, Group, Summary string
	Tags                      []string
}

// DemoGroup describes a category in the group picker.
type DemoGroup struct {
	Key, Label string
}

func demoGroups() []DemoGroup {
	return []DemoGroup{
		{"all", "All"},
		{"input", "Input"},
		{"selection", "Selection"},
		{"data", "Data"},
		{"text", "Text"},
		{"graphics", "Graphics"},
		{"layout", "Layout"},
		{"navigation", "Navigation"},
		{"feedback", "Feedback"},
		{"overlays", "Overlays"},
		{"animations", "Animations"},
		{"theme", "Theme"},
		{"locale", "Locale"},
		{"docs", "Docs"},
	}
}

func demoEntries() []DemoEntry {
	return []DemoEntry{
		{"welcome", "Welcome", "welcome", "Introduction to the Arcade showcase", nil},

		{"input", "Input", "input", "Text input fields", []string{"text", "field"}},
		{"numeric_input", "Numeric Input", "input", "Locale-aware numeric input", []string{"number", "currency"}},
		{"color_picker", "Color Picker", "input", "Interactive color selection", []string{"color", "hsv", "rgb"}},
		{"date_picker", "Date Picker", "input", "Calendar date selection", []string{"calendar", "date"}},
		{"date_picker_roller", "Date Picker Roller", "input", "Rolling date selection", []string{"date", "roller"}},
		{"input_date", "Input Date", "input", "Typed date input", []string{"date", "input"}},
		{"forms", "Forms", "input", "Form layout patterns", []string{"form", "validation"}},

		{"toggle", "Toggle", "selection", "Toggle buttons and checkboxes", []string{"checkbox", "check"}},
		{"switch", "Switch", "selection", "On/off switch control", []string{"on", "off"}},
		{"radio", "Radio", "selection", "Radio button selection", []string{"radio", "option"}},
		{"radio_group", "Radio Group", "selection", "Grouped radio buttons", []string{"radio", "group"}},
		{"select", "Select", "selection", "Dropdown selection", []string{"dropdown", "combo"}},
		{"listbox", "ListBox", "selection", "Scrollable list selection", []string{"list", "multi"}},
		{"combobox", "Combobox", "selection", "Editable dropdown", []string{"combo", "edit"}},
		{"range_slider", "Range Slider", "selection", "Numeric range selection", []string{"slider", "range"}},

		{"table", "Table", "data", "Sortable data table", []string{"grid", "sort"}},
		{"data_grid", "Data Grid", "data", "Full-featured data grid", []string{"grid", "paging", "filter"}},

		{"text", "Text", "text", "Text display styles", []string{"font", "style"}},
		{"rtf", "Rich Text", "text", "Rich text formatting", []string{"rich", "format"}},
		{"markdown", "Markdown", "text", "Markdown rendering", []string{"md", "document"}},

		{"svg", "SVG", "graphics", "Scalable vector graphics", []string{"vector", "draw"}},
		{"image", "Image", "graphics", "Image display", []string{"picture", "photo"}},
		{"gradient", "Gradient", "graphics", "Color gradients", []string{"linear", "radial"}},
		{"box_shadows", "Box Shadows", "graphics", "Shadow effects", []string{"shadow", "drop"}},
		{"rectangle", "Rectangle", "graphics", "Styled rectangles", []string{"rect", "shape"}},
		{"icons", "Icons", "graphics", "Icon font library", []string{"feather", "glyph"}},

		{"row", "Row", "layout", "Horizontal layout", []string{"horizontal", "flex"}},
		{"column", "Column", "layout", "Vertical layout", []string{"vertical", "flex"}},
		{"wrap_panel", "Wrap Panel", "layout", "Wrapping flow layout", []string{"wrap", "flow"}},
		{"overflow_panel", "Overflow Panel", "layout", "Scrollable overflow", []string{"scroll", "overflow"}},
		{"expand_panel", "Expand Panel", "layout", "Collapsible sections", []string{"collapse", "accordion"}},
		{"sidebar", "Sidebar", "layout", "Slide-out sidebar", []string{"drawer", "panel"}},
		{"splitter", "Splitter", "layout", "Resizable split panes", []string{"split", "resize"}},
		{"scrollbar", "Scrollbar", "layout", "Custom scrollbar styling", []string{"scroll", "bar"}},

		{"breadcrumb", "Breadcrumb", "navigation", "Path navigation", []string{"path", "trail"}},
		{"tab_control", "Tab Control", "navigation", "Tabbed content panels", []string{"tab", "panel"}},
		{"menus", "Menus", "navigation", "Menu bar with items", []string{"menu", "bar"}},
		{"command_palette", "Command Palette", "navigation", "Quick command search", []string{"command", "search", "palette"}},

		{"button", "Button", "feedback", "Clickable buttons", []string{"click", "action"}},
		{"progress_bar", "Progress Bar", "feedback", "Progress indicators", []string{"progress", "loading"}},
		{"pulsar", "Pulsar", "feedback", "Pulsing activity indicator", []string{"pulse", "loading"}},
		{"toast", "Toast", "feedback", "Toast notifications", []string{"notification", "message"}},
		{"badge", "Badge", "feedback", "Status badges", []string{"count", "status"}},

		{"dialog", "Dialog", "overlays", "Modal dialogs", []string{"modal", "popup"}},
		{"tooltip", "Tooltip", "overlays", "Hover tooltips", []string{"hover", "hint"}},

		{"animations", "Animations", "animations", "Tween, spring, and keyframe animations", []string{"tween", "spring", "keyframe"}},

		{"theme_gen", "Theme Generator", "theme", "Generate themes from a seed color", []string{"theme", "color", "generate"}},
		{"locale", "Locale", "locale", "Locale switching and formatting", []string{"i18n", "language", "format"}},

		{"doc_get_started", "Getting Started", "docs", "Minimal app setup and key concepts", []string{"start", "setup", "tutorial"}},
		{"doc_architecture", "Architecture", "docs", "Rendering pipeline and core types", []string{"pipeline", "layout", "render"}},
		{"doc_containers", "Containers", "docs", "Row, Column, Wrap, scrolling", []string{"row", "column", "wrap"}},
		{"doc_themes", "Themes", "docs", "Built-in themes and custom theme creation", []string{"theme", "style", "color"}},
		{"doc_animations", "Animations", "docs", "Tween, spring, and keyframe APIs", []string{"tween", "spring", "keyframe"}},
		{"doc_locales", "Locales", "docs", "Locale registration and date/number formatting", []string{"i18n", "locale", "format"}},
	}
}

func filteredEntries(app *ArcadeApp) []DemoEntry {
	all := demoEntries()
	if app.SelectedGroup == "all" && app.NavQuery == "" {
		return all
	}
	result := make([]DemoEntry, 0, len(all))
	for _, e := range all {
		if app.SelectedGroup != "all" && e.Group != app.SelectedGroup {
			continue
		}
		if app.NavQuery != "" && !entryMatchesQuery(e, app.NavQuery) {
			continue
		}
		result = append(result, e)
	}
	return result
}

func entryMatchesQuery(e DemoEntry, q string) bool {
	q = strings.ToLower(q)
	if strings.Contains(strings.ToLower(e.ID), q) ||
		strings.Contains(strings.ToLower(e.Label), q) ||
		strings.Contains(strings.ToLower(e.Summary), q) {
		return true
	}
	for _, tag := range e.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

func groupLabel(key string) string {
	for _, g := range demoGroups() {
		if g.Key == key {
			return g.Label
		}
	}
	return key
}
