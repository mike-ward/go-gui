package main

import (
	"strings"
	"time"

	"github.com/mike-ward/go-gui/gui"
)

const (
	groupWelcome   = "welcome"
	groupAll       = "all"
	groupText      = "text"
	groupInput     = "input"
	groupSelection = "selection"
	groupData      = "data"
	groupGraphics  = "graphics"
	groupNav       = "navigation"
	groupLayout    = "layout"
	groupFeedback  = "feedback"
	groupOverlays  = "overlays"
)

var showcaseBreadcrumbPath = []gui.BreadcrumbItemCfg{
	gui.NewBreadcrumbItem("home", "Home", nil),
	gui.NewBreadcrumbItem("docs", "Docs", nil),
	gui.NewBreadcrumbItem("guide", "Guide", nil),
	gui.NewBreadcrumbItem("page", "Getting Started", nil),
}

// ShowcaseApp holds all state for the showcase example.
type ShowcaseApp struct {
	LocaleIndex       int
	NavQuery          string
	SelectedGroup     string
	SelectedComponent string
	ShowDocs          bool

	ButtonClicks    int
	ButtonCopyUntil time.Time

	InputText      string
	InputPassword  string
	InputPhone     string
	InputExpiry    string
	InputMultiline string

	ToggleA    bool
	CheckboxA  bool
	SwitchA    bool
	RadioValue string

	SelectValue     []string
	ListBoxSelected []string
	ComboboxValue   string
	RangeValue      float32

	NumericENText        string
	NumericDEText        string
	NumericCurrencyText  string
	NumericPercentText   string
	NumericPlainText     string
	NumericENValue       *float64
	NumericDEValue       *float64
	NumericCurrencyValue *float64
	NumericPercentValue  *float64
	NumericPlainValue    *float64

	DatePickerDates []time.Time
	InputDate       time.Time
	RollerDate      time.Time

	ColorPickerColor gui.Color
	ColorPickerHSV   bool

	Form FormModel

	BCSelected  string
	BCPath      []gui.BreadcrumbItemCfg
	TabSelected string

	SplitterMainState   gui.SplitterState
	SplitterDetailState gui.SplitterState
	SidebarOpen         bool
	ExpandOpen          bool

	DialogResult string
	NotifyResult string

	PrintingLastPath string
	PrintingStatus   string

	PaletteAction string

	AnimTweenX    float32
	AnimSpringX   float32
	AnimKeyframeX float32

	DataGridQuery       gui.GridQueryState
	DataGridSelection   gui.GridSelection
	TableSortBy      int
	TableBorderStyle string
	TableMultiSelect bool
	TableSelected    map[int]bool
	DataSourceQuery     gui.GridQueryState
	DataSourceSelection gui.GridSelection
	DataSource          gui.DataGridDataSource

	TreeSelected  string
	TreeLazyNodes map[string][]gui.TreeNodeCfg

	DragListItems []gui.ListBoxOption
	DragTabItems  []gui.TabItemCfg
	DragTabSel    string
	DragTreeNodes []gui.TreeNodeCfg

	ThemeGenSeed       gui.Color
	ThemeGenStrategy   string
	ThemeGenTint       float32
	ThemeGenRadius     float32
	ThemeGenRadiusText string
	ThemeGenBorder     float32
	ThemeGenBorderText string
	ThemeGenPickText   bool
	ThemeGenText       gui.Color
	ThemeGenName       string
}

func newShowcaseApp() *ShowcaseApp {
	value1234 := 1234.5
	valuePct := 0.125

	return &ShowcaseApp{
		SelectedGroup:        groupAll,
		SelectedComponent:    "welcome",
		InputMultiline:       "Now is the time for all good men to come to the aid of their country",
		RadioValue:           "go",
		RangeValue:           50,
		NumericENText:        "1,234.50",
		NumericENValue:       &value1234,
		NumericDEText:        "1.234,50",
		NumericDEValue:       cloneFloatPtr(&value1234),
		NumericCurrencyText:  "$1,234.50",
		NumericCurrencyValue: cloneFloatPtr(&value1234),
		NumericPercentText:   "12.50%",
		NumericPercentValue:  &valuePct,
		InputDate:            time.Now(),
		RollerDate:           time.Now(),
		ColorPickerColor:     gui.RGBA(255, 85, 0, 255),
		BCSelected:           "page",
		BCPath:               append([]gui.BreadcrumbItemCfg(nil), showcaseBreadcrumbPath...),
		TabSelected:          "overview",
		SplitterMainState:    gui.SplitterState{Ratio: 0.30},
		SplitterDetailState:  gui.SplitterState{Ratio: 0.55},
		SidebarOpen:          true,
		ThemeGenSeed:         gui.ThemeDarkBorderedCfg.ColorSelect,
		ThemeGenStrategy:     "mono",
		ThemeGenRadius:       gui.ThemeDarkBorderedCfg.Radius,
		ThemeGenRadiusText:   floatString(gui.ThemeDarkBorderedCfg.Radius),
		ThemeGenBorder:       gui.ThemeDarkBorderedCfg.SizeBorder,
		ThemeGenBorderText:   floatString(gui.ThemeDarkBorderedCfg.SizeBorder),
		ThemeGenText:         gui.ThemeDarkBorderedCfg.TextStyleDef.Color,
		DataGridSelection: gui.GridSelection{
			SelectedRowIDs: map[string]bool{},
		},
		TableBorderStyle: "all",
		DataSourceSelection: gui.GridSelection{
			SelectedRowIDs: map[string]bool{},
		},
		TreeLazyNodes: make(map[string][]gui.TreeNodeCfg),
		DragListItems: []gui.ListBoxOption{
			gui.NewListBoxOption("apple", "Apple", ""),
			gui.NewListBoxOption("banana", "Banana", ""),
			gui.NewListBoxOption("cherry", "Cherry", ""),
			gui.NewListBoxOption("date", "Date", ""),
			gui.NewListBoxOption("elderberry", "Elderberry", ""),
			gui.NewListBoxOption("fig", "Fig", ""),
		},
		DragTabItems: []gui.TabItemCfg{
			gui.NewTabItem("alpha", "Alpha", nil),
			gui.NewTabItem("beta", "Beta", nil),
			gui.NewTabItem("gamma", "Gamma", nil),
			gui.NewTabItem("delta", "Delta", nil),
		},
		DragTabSel: "alpha",
		DragTreeNodes: []gui.TreeNodeCfg{
			{ID: "src", Text: "src", Nodes: []gui.TreeNodeCfg{
				{ID: "main.go", Text: "main.go"},
				{ID: "util.go", Text: "util.go"},
				{ID: "app.go", Text: "app.go"},
			}},
			{ID: "docs", Text: "docs", Nodes: []gui.TreeNodeCfg{
				{ID: "README.md", Text: "README.md"},
				{ID: "GUIDE.md", Text: "GUIDE.md"},
			}},
			{ID: "tests", Text: "tests"},
			{ID: "build", Text: "build"},
		},
	}
}

func cloneFloatPtr(v *float64) *float64 {
	if v == nil {
		return nil
	}
	copyValue := *v
	return &copyValue
}

func demoGroups() []DemoGroup {
	return []DemoGroup{
		{Key: groupWelcome, Label: "Welcome"},
		{Key: groupInput, Label: "Input"},
		{Key: groupSelection, Label: "Selection"},
		{Key: groupData, Label: "Data Display"},
		{Key: groupLayout, Label: "Layout"},
		{Key: groupNav, Label: "Navigation"},
		{Key: groupFeedback, Label: "Feedback"},
		{Key: groupOverlays, Label: "Overlays"},
		{Key: groupText, Label: "Text"},
		{Key: groupGraphics, Label: "Graphics"},
	}
}

// DemoEntry describes a single showcase destination.
type DemoEntry struct {
	ID      string
	Label   string
	Group   string
	Summary string
	Tags    []string
}

// DemoGroup describes a catalog group in the left pane.
type DemoGroup struct {
	Key   string
	Label string
}

func demoEntries() []DemoEntry {
	return []DemoEntry{
		{ID: "welcome", Label: "Welcome", Group: groupWelcome, Summary: "Start here for a quick introduction to go-gui and this showcase.", Tags: []string{"start", "intro", "overview"}},
		{ID: "doc_get_started", Label: "Get Started", Group: groupWelcome, Summary: "Step-by-step guide to building your first go-gui application.", Tags: []string{"guide", "tutorial", "setup"}},
		{ID: "doc_animations", Label: "Animations", Group: groupWelcome, Summary: "Guide to tween, spring, keyframe, and transition APIs.", Tags: []string{"doc", "animation", "tween", "spring"}},
		{ID: "doc_architecture", Label: "Architecture", Group: groupWelcome, Summary: "Internal architecture and design decisions of the framework.", Tags: []string{"doc", "design", "internals", "structure"}},
		{ID: "doc_containers", Label: "Containers", Group: groupWelcome, Summary: "Row, column, wrap, canvas, and circle container reference.", Tags: []string{"doc", "container", "row", "column", "wrap", "layout"}},
		{ID: "doc_custom_widgets", Label: "Custom Widgets", Group: groupWelcome, Summary: "Build third-party widgets via composition or View implementation.", Tags: []string{"doc", "widget", "custom", "extend"}},
		{ID: "doc_data_grid", Label: "Data Grid", Group: groupWelcome, Summary: "Data grid component documentation and usage patterns.", Tags: []string{"doc", "grid", "table", "data"}},
		{ID: "doc_forms", Label: "Forms", Group: groupWelcome, Summary: "Form validation model and field adapter documentation.", Tags: []string{"doc", "forms", "validation", "async"}},
		{ID: "doc_gradients", Label: "Gradients", Group: groupWelcome, Summary: "Guide to linear and radial gradient APIs.", Tags: []string{"doc", "gradient", "linear", "radial"}},
		{ID: "doc_layout_algorithm", Label: "Layout Algorithm", Group: groupWelcome, Summary: "How the layout engine measures and arranges views.", Tags: []string{"doc", "layout", "sizing", "algorithm"}},
		{ID: "doc_locales", Label: "Locales", Group: groupWelcome, Summary: "Locale bundles, translation, and runtime language switching.", Tags: []string{"doc", "locale", "i18n", "translation", "rtl"}},
		{ID: "doc_markdown", Label: "Markdown", Group: groupWelcome, Summary: "Guide to the markdown renderer and its options.", Tags: []string{"doc", "markdown", "renderer"}},
		{ID: "doc_native_dialogs", Label: "Native Dialogs", Group: groupWelcome, Summary: "Guide to native file, save, and alert dialog APIs.", Tags: []string{"doc", "dialog", "native", "file"}},
		{ID: "doc_performance", Label: "Performance", Group: groupWelcome, Summary: "Performance optimization tips and best practices.", Tags: []string{"doc", "performance", "optimization", "speed"}},
		{ID: "doc_printing", Label: "Printing", Group: groupWelcome, Summary: "Guide to PDF export and native print dialog APIs.", Tags: []string{"doc", "print", "pdf", "export"}},
		{ID: "doc_shaders", Label: "Shaders", Group: groupWelcome, Summary: "Guide to custom fragment shader integration.", Tags: []string{"doc", "shader", "glsl", "metal"}},
		{ID: "doc_splitter", Label: "Splitter", Group: groupWelcome, Summary: "Guide to resizable split panel APIs.", Tags: []string{"doc", "splitter", "panel", "resize"}},
		{ID: "doc_svg", Label: "SVG", Group: groupWelcome, Summary: "Guide to SVG rendering and inline SVG APIs.", Tags: []string{"doc", "svg", "vector", "path"}},
		{ID: "doc_tables", Label: "Tables", Group: groupWelcome, Summary: "Table component documentation and column configuration.", Tags: []string{"doc", "table", "columns", "data"}},
		{ID: "doc_tree", Label: "Tree View", Group: groupWelcome, Summary: "Guide to tree view configuration, lazy loading, and keyboard navigation.", Tags: []string{"doc", "tree", "hierarchy", "lazy", "virtualized"}},
		{ID: "doc_themes", Label: "Themes", Group: groupWelcome, Summary: "Theme system: presets, custom themes, JSON, runtime switching.", Tags: []string{"doc", "theme", "color", "style"}},

		{ID: "color_picker", Label: "Color Picker", Group: groupInput, Summary: "Pick RGBA and optional HSV values.", Tags: []string{"color", "hsv", "rgba"}},
		{ID: "date_picker", Label: "Date Picker", Group: groupInput, Summary: "Select one or many dates from a calendar.", Tags: []string{"calendar", "dates", "input"}},
		{ID: "date_picker_roller", Label: "Date Picker Roller", Group: groupInput, Summary: "Roll wheel-style month/day/year controls.", Tags: []string{"date", "roller", "time"}},
		{ID: "input", Label: "Input", Group: groupInput, Summary: "Single-line, password, and multiline text input.", Tags: []string{"text", "textarea", "password"}},
		{ID: "input_date", Label: "Input Date", Group: groupInput, Summary: "Text input with date picker dropdown.", Tags: []string{"date", "input", "calendar"}},
		{ID: "numeric_input", Label: "Numeric Input", Group: groupInput, Summary: "Locale-aware number input with step controls.", Tags: []string{"number", "decimal", "locale", "spinner"}},
		{ID: "forms", Label: "Forms", Group: groupInput, Summary: "Form runtime with sync and async validation states recreated in example code.", Tags: []string{"form", "validation", "async", "touched", "dirty"}},

		{ID: "listbox", Label: "List Box", Group: groupSelection, Summary: "Single and multi-select list options.", Tags: []string{"list", "multi", "select"}},
		{ID: "radio", Label: "Radio", Group: groupSelection, Summary: "Single radio control.", Tags: []string{"option", "boolean", "choice"}},
		{ID: "radio_group", Label: "Radio Button Group", Group: groupSelection, Summary: "Mutually exclusive options in row or column.", Tags: []string{"group", "options", "select"}},
		{ID: "range_slider", Label: "Range Slider", Group: groupSelection, Summary: "Drag horizontal or vertical value controls.", Tags: []string{"slider", "value", "range"}},
		{ID: "drag_reorder", Label: "Drag Reorder", Group: groupSelection, Summary: "Drag-to-reorder items in lists, tabs, and trees.", Tags: []string{"drag", "reorder", "list", "tabs", "tree", "keyboard"}},
		{ID: "combobox", Label: "Combobox", Group: groupSelection, Summary: "Single-select with typeahead filtering.", Tags: []string{"dropdown", "filter", "typeahead", "autocomplete"}},
		{ID: "select", Label: "Select", Group: groupSelection, Summary: "Dropdown with optional multi-select.", Tags: []string{"dropdown", "pick", "options"}},
		{ID: "switch", Label: "Switch", Group: groupSelection, Summary: "On/off switch control.", Tags: []string{"toggle", "boolean", "control"}},
		{ID: "toggle", Label: "Toggle", Group: groupSelection, Summary: "Checkbox-style and icon toggles.", Tags: []string{"checkbox", "boolean", "control"}},

		{ID: "image", Label: "Image", Group: groupGraphics, Summary: "Render local image assets.", Tags: []string{"photo", "asset", "media"}},
		{ID: "rectangle", Label: "Rectangle", Group: groupGraphics, Summary: "Draw colored shapes with border and radius.", Tags: []string{"shape", "primitive", "box"}},
		{ID: "svg", Label: "SVG", Group: groupGraphics, Summary: "Render vector graphics from SVG strings.", Tags: []string{"vector", "icon", "path"}},
		{ID: "printing", Label: "Printing", Group: groupGraphics, Summary: "Export current view to PDF and open native print dialog.", Tags: []string{"print", "pdf", "export"}},
		{ID: "animations", Label: "Animations", Group: groupGraphics, Summary: "Tween, spring, and layout transition samples.", Tags: []string{"motion", "tween", "spring"}},
		{ID: "gradient", Label: "Gradients", Group: groupGraphics, Summary: "Linear and radial gradient fills.", Tags: []string{"gradient", "linear", "radial", "fill"}},
		{ID: "box_shadows", Label: "Box Shadows", Group: groupGraphics, Summary: "Shadow presets with spread behavior notes.", Tags: []string{"shadow", "depth", "blur"}},
		{ID: "shader", Label: "Custom Shaders", Group: groupGraphics, Summary: "Custom fragment shaders for dynamic fills.", Tags: []string{"shader", "glsl", "metal"}},
		{ID: "theme_gen", Label: "Theme", Group: groupGraphics, Summary: "Generate a theme from a seed color, tint level, and palette strategy.", Tags: []string{"theme", "color", "palette", "generator"}},

		{ID: "markdown", Label: "Markdown", Group: groupText, Summary: "Render markdown into styled rich content.", Tags: []string{"docs", "text", "rich"}},
		{ID: "rtf", Label: "Rich Text Format", Group: groupText, Summary: "Mixed styles, links, and inline rich runs.", Tags: []string{"rich text", "link", "style"}},
		{ID: "text", Label: "Text", Group: groupText, Summary: "Typography, gradients, outlines, and curved text.", Tags: []string{"font", "type", "styles", "gradient", "outline", "stroke", "curve"}},
		{ID: "icons", Label: "Icons", Group: groupText, Summary: "Icon font catalog and glyph references.", Tags: []string{"icon", "font", "glyph"}},

		{ID: "table", Label: "Table", Group: groupData, Summary: "Declarative and sortable table data.", Tags: []string{"rows", "columns", "csv"}},
		{ID: "data_grid", Label: "Data Grid", Group: groupData, Summary: "Controlled virtualized grid for interactive tabular data.", Tags: []string{"grid", "virtualized", "data"}},
		{ID: "data_source", Label: "Data Source", Group: groupData, Summary: "Async data-source backed grid with CRUD.", Tags: []string{"async", "pagination", "crud", "source"}},
		{ID: "tree", Label: "Tree View", Group: groupData, Summary: "Hierarchical expandable node display.", Tags: []string{"nodes", "hierarchy", "outline"}},

		{ID: "breadcrumb", Label: "Breadcrumb", Group: groupNav, Summary: "Trail navigation with optional content panels.", Tags: []string{"breadcrumb", "navigation", "trail", "path"}},
		{ID: "menus", Label: "Menus + Menubar", Group: groupNav, Summary: "Nested menus, separators, and custom menu items.", Tags: []string{"menu", "menubar", "submenu"}},
		{ID: "scrollbar", Label: "Scrollable Containers", Group: groupNav, Summary: "Bind scrollable layouts to shared scroll ids.", Tags: []string{"scrollbar", "scroll", "container"}},
		{ID: "splitter", Label: "Splitter", Group: groupNav, Summary: "Resizable panes with drag, keyboard, and collapse.", Tags: []string{"split", "pane", "resize"}},
		{ID: "tab_control", Label: "Tab Control", Group: groupNav, Summary: "Switch content panels with keyboard-friendly tabs.", Tags: []string{"tabs", "navigation", "panes"}},

		{ID: "button", Label: "Button", Group: groupFeedback, Summary: "Trigger actions with click and keyboard focus.", Tags: []string{"action", "press", "click"}},
		{ID: "progress_bar", Label: "Progress Bar", Group: groupFeedback, Summary: "Determinate and indeterminate progress indicators.", Tags: []string{"progress", "loader", "status"}},
		{ID: "pulsar", Label: "Pulsar", Group: groupFeedback, Summary: "Animated pulse indicator with optional icons.", Tags: []string{"pulse", "loading", "indicator"}},
		{ID: "toast", Label: "Toast", Group: groupFeedback, Summary: "Non-blocking notifications with auto-dismiss and actions.", Tags: []string{"notification", "alert", "severity", "stack"}},
		{ID: "badge", Label: "Badge", Group: groupFeedback, Summary: "Numeric and colored pill labels for counts and status.", Tags: []string{"badge", "count", "status", "pill", "label"}},
		{ID: "native_notification", Label: "Native Notification", Group: groupFeedback, Summary: "OS-level notifications on macOS, Windows, and Linux.", Tags: []string{"notification", "native", "os", "alert", "push"}},

		{ID: "dialog", Label: "Dialog", Group: groupOverlays, Summary: "Message, confirm, prompt, and custom dialogs.", Tags: []string{"modal", "confirm", "prompt"}},
		{ID: "expand_panel", Label: "Expand Panel", Group: groupOverlays, Summary: "Collapsible region with custom header and content.", Tags: []string{"accordion", "collapse", "panel"}},
		{ID: "command_palette", Label: "Command Palette", Group: groupOverlays, Summary: "Keyboard-first searchable action list.", Tags: []string{"command", "search", "palette", "keyboard"}},
		{ID: "tooltip", Label: "Tooltip", Group: groupOverlays, Summary: "Hover hints with custom placement and content.", Tags: []string{"hover", "hint", "floating"}},
		{ID: "inspector", Label: "Inspector", Group: groupOverlays, Summary: "Dev-mode layout tree and property inspector.", Tags: []string{"inspector", "debug", "devtools", "layout", "tree"}},

		{ID: "row", Label: "Row", Group: groupLayout, Summary: "Horizontal container arranging children left-to-right.", Tags: []string{"row", "horizontal", "container", "layout"}},
		{ID: "column_demo", Label: "Column", Group: groupLayout, Summary: "Vertical container arranging children top-to-bottom.", Tags: []string{"column", "vertical", "container", "layout"}},
		{ID: "wrap_panel", Label: "Wrap Panel", Group: groupLayout, Summary: "Flow layout that wraps children to the next line.", Tags: []string{"wrap", "flow", "reflow", "layout"}},
		{ID: "overflow_panel", Label: "Overflow Panel", Group: groupLayout, Summary: "Row that hides non-fitting children in a dropdown.", Tags: []string{"overflow", "toolbar", "responsive", "layout"}},
		{ID: "sidebar", Label: "Sidebar", Group: groupLayout, Summary: "Animated panel that slides in and out.", Tags: []string{"sidebar", "panel", "slide", "layout"}},
	}
}

func entryMatchesQuery(entry DemoEntry, query string) bool {
	if query == "" {
		return true
	}
	q := strings.ToLower(query)
	if strings.Contains(strings.ToLower(entry.ID), q) ||
		strings.Contains(strings.ToLower(entry.Label), q) ||
		strings.Contains(strings.ToLower(entry.Summary), q) ||
		strings.Contains(strings.ToLower(entry.Group), q) {
		return true
	}
	for _, tag := range entry.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

func filteredEntries(app *ShowcaseApp) []DemoEntry {
	out := make([]DemoEntry, 0, len(demoEntries()))
	for _, entry := range demoEntries() {
		if app.SelectedGroup != groupAll && entry.Group != app.SelectedGroup {
			continue
		}
		if !entryMatchesQuery(entry, app.NavQuery) {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func hasEntry(entries []DemoEntry, selected string) bool {
	for _, entry := range entries {
		if entry.ID == selected {
			return true
		}
	}
	return false
}

func selectedEntry(entries []DemoEntry, selected string) DemoEntry {
	for _, entry := range entries {
		if entry.ID == selected {
			return entry
		}
	}
	if len(entries) == 0 {
		return DemoEntry{}
	}
	return entries[0]
}

func preferredComponentForGroup(_ string, entries []DemoEntry) string {
	if len(entries) == 0 {
		return ""
	}
	best := entries[0]
	for _, entry := range entries[1:] {
		if entrySortBefore(entry, best) {
			best = entry
		}
	}
	return best.ID
}

func entrySortBefore(a, b DemoEntry) bool {
	aPin := entryPin(a.ID)
	bPin := entryPin(b.ID)
	if aPin != bPin {
		return aPin < bPin
	}
	aLabel := strings.ToLower(a.Label)
	bLabel := strings.ToLower(b.Label)
	if aLabel != bLabel {
		return aLabel < bLabel
	}
	return strings.ToLower(a.ID) < strings.ToLower(b.ID)
}

func entryPin(id string) int {
	switch id {
	case "welcome":
		return 0
	case "doc_get_started":
		return 1
	default:
		return 2
	}
}

func groupLabel(key string) string {
	for _, group := range demoGroups() {
		if group.Key == key {
			return group.Label
		}
	}
	if key == groupAll {
		return "All"
	}
	return key
}
