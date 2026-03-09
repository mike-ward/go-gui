package main

import "github.com/mike-ward/go-gui/gui"

func demoDoc(w *gui.Window, source string) gui.View {
	return showcaseMarkdownPanel(w, "showcase-inline-doc", source)
}

func componentDoc(id string) string {
	switch id {
	case "native_notification":
		id = "notification"
	case "column_demo":
		id = "column"
	}
	if doc, ok := widgetDocs[id]; ok {
		return doc
	}
	return ""
}

var widgetDocs = map[string]string{
	// Feedback
	"button": `# Button

Trigger actions with click and keyboard focus.

## Usage

` + "```go" + `
gui.Button(gui.ButtonCfg{
    ID:      "submit",
    Content: []gui.View{gui.Text(gui.TextCfg{Text: "Submit"})},
    OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
        e.IsHandled = true
    },
})
` + "```" + `

## Key Properties

| Property | Type | Description |
|----------|------|-------------|
| Content | []View | Child views inside the button |
| Padding | Opt[Padding] | Inner padding |
| Color | Color | Background color |
| ColorHover | Color | Hover background |
| ColorClick | Color | Click background |
| Radius | Opt[float32] | Corner radius |

## Events

| Callback | Signature | Fired when |
|----------|-----------|------------|
| OnClick | func(*Layout, *Event, *Window) | Button clicked |
`,

	"progress_bar": `# Progress Bar

Determinate and indeterminate progress indicators.

## Usage

` + "```go" + `
gui.ProgressBar(gui.ProgressBarCfg{
    ID:    "loading",
    Value: 0.75,
})
` + "```" + `

## Key Properties

| Property | Type | Description |
|----------|------|-------------|
| Value | float32 | Progress 0.0–1.0 |
| Indeterminate | bool | Animated indeterminate mode |
| Color | Color | Bar fill color |
`,

	"pulsar": `# Pulsar

Animated pulse indicator for loading states.

## Usage

` + "```go" + `
w.Pulsar()
` + "```" + `
`,

	"toast": `# Toast

Non-blocking notifications with severity and auto-dismiss.

## Usage

` + "```go" + `
w.Toast(gui.ToastCfg{
    Title:    "Saved",
    Body:     "Document saved.",
})
` + "```" + `

## API

| Method | Description |
|--------|-------------|
| w.Toast(cfg) | Show toast |
| w.ToastDismiss(id) | Dismiss specific toast |
`,

	"badge": `# Badge

Numeric and colored pill labels for counts and status.

## Usage

` + "```go" + `
gui.Badge(gui.BadgeCfg{Label: "5", Variant: gui.BadgeInfo})
` + "```" + `

## Variants

| Variant | Use case |
|---------|----------|
| BadgeDefault | Custom color |
| BadgeInfo | Informational |
| BadgeSuccess | Positive status |
| BadgeWarning | Needs attention |
| BadgeError | Critical |
`,

	// Input
	"input": `# Input

Single-line, password, and multiline text input with optional mask.

## Usage

` + "```go" + `
gui.Input(gui.InputCfg{
    ID:          "name",
    Text:        app.Name,
    Placeholder: "Enter name...",
    OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
        gui.State[App](w).Name = s
    },
})
` + "```" + `

## Key Properties

| Property | Type | Description |
|----------|------|-------------|
| Text | string | Current text value |
| Placeholder | string | Hint text |
| IsPassword | bool | Mask input |
| Mode | InputMode | InputSingleLine or InputMultiline |
| MaskPreset | MaskPreset | MaskPhoneUS, etc. |
| Height | float32 | Multiline height |
`,

	"numeric_input": `# Numeric Input

Locale-aware number input with step buttons and arrow keys.

## Usage

` + "```go" + `
gui.NumericInput(gui.NumericInputCfg{
    ID:          "qty",
    Placeholder: "Enter number",
})
` + "```" + `
`,

	"color_picker": `# Color Picker

Interactive HSV color selection wheel with value slider.

## Usage

` + "```go" + `
gui.ColorPicker(gui.ColorPickerCfg{
    ID:    "cp",
    Color: app.Color,
    OnColorChange: func(c gui.Color, _ *gui.Event, w *gui.Window) {
        gui.State[App](w).Color = c
    },
})
` + "```" + `
`,

	"date_picker": `# Date Picker

Calendar-style date selection with month/year navigation.

## Usage

` + "```go" + `
gui.DatePicker(gui.DatePickerCfg{ID: "dp"})
` + "```" + `
`,

	"date_picker_roller": `# Date Picker Roller

Rolling drum-style date selection.

## Usage

` + "```go" + `
gui.DatePickerRoller(gui.DatePickerRollerCfg{ID: "dpr"})
` + "```" + `
`,

	"input_date": `# Input Date

Text input with calendar popup for date entry.

## Usage

` + "```go" + `
gui.InputDate(gui.InputDateCfg{ID: "id", Sizing: gui.FillFit})
` + "```" + `
`,

	"forms": `# Forms

Combine inputs, labels, and buttons into form layouts.

Use ` + "`labeledRow`" + ` helpers to align labels with inputs in a
consistent grid.
`,

	// Selection
	"toggle": `# Toggle

Checkbox-style toggles with labels.

## Usage

` + "```go" + `
gui.Toggle(gui.ToggleCfg{
    ID:       "accept",
    Label:    "Accept terms",
    Selected: app.Accepted,
    OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
        s := gui.State[App](w)
        s.Accepted = !s.Accepted
    },
})
` + "```" + `

## Events

| Callback | Signature | Fired when |
|----------|-----------|------------|
| OnClick | func(*Layout, *Event, *Window) | Toggle clicked |
`,

	"switch": `# Switch

On/off switch control with animated thumb.

## Usage

` + "```go" + `
gui.Switch(gui.SwitchCfg{
    ID:       "feature",
    Label:    "Enable feature",
    Selected: app.Enabled,
    OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
        s := gui.State[App](w)
        s.Enabled = !s.Enabled
    },
})
` + "```" + `
`,

	"radio": `# Radio

Single radio button for selecting one option from a group.

## Usage

` + "```go" + `
gui.Radio(gui.RadioCfg{
    ID:       "opt-go",
    Label:    "Go",
    Selected: app.Lang == "go",
    OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
        gui.State[App](w).Lang = "go"
    },
})
` + "```" + `
`,

	"radio_group": `# Radio Button Group

Grouped radio buttons in row or column layout.

## Usage

` + "```go" + `
gui.RadioButtonGroupColumn(gui.RadioButtonGroupCfg{
    Value: app.Lang,
    Options: []gui.RadioOption{
        gui.NewRadioOption("Go", "go"),
        gui.NewRadioOption("Rust", "rust"),
    },
    OnSelect: func(v string, w *gui.Window) {
        gui.State[App](w).Lang = v
    },
})
` + "```" + `
`,

	"select": `# Select

Dropdown with optional multi-select.

## Usage

` + "```go" + `
gui.Select(gui.SelectCfg{
    ID:       "lang",
    Selected: app.Selected,
    Options:  []string{"Go", "Rust", "Zig"},
    OnSelect: func(sel []string, _ *gui.Event, w *gui.Window) {
        gui.State[App](w).Selected = sel
    },
})
` + "```" + `
`,

	"listbox": `# List Box

Single and multi-select scrollable list.

## Usage

` + "```go" + `
gui.ListBox(gui.ListBoxCfg{
    ID:          "lb",
    Multiple:    true,
    SelectedIDs: app.Selected,
    Data: []gui.ListBoxOption{
        gui.NewListBoxOption("go", "Go", "go"),
    },
    OnSelect: func(ids []string, _ *gui.Event, w *gui.Window) {
        gui.State[App](w).Selected = ids
    },
})
` + "```" + `
`,

	"combobox": `# Combobox

Editable dropdown with type-ahead filtering.

## Usage

` + "```go" + `
gui.Combobox(gui.ComboboxCfg{
    ID:      "cb",
    Value:   app.Value,
    Options: []string{"Go", "Rust", "Zig"},
    OnSelect: func(v string, _ *gui.Event, w *gui.Window) {
        gui.State[App](w).Value = v
    },
})
` + "```" + `
`,

	"range_slider": `# Range Slider

Drag horizontal value control.

## Usage

` + "```go" + `
gui.RangeSlider(gui.RangeSliderCfg{
    ID:    "vol",
    Value: app.Volume,
    Min:   0, Max: 100,
    OnChange: func(v float32, _ *gui.Event, w *gui.Window) {
        gui.State[App](w).Volume = v
    },
})
` + "```" + `
`,

	// Data
	"table": `# Table

Sortable data table from string arrays.

## Usage

` + "```go" + `
cfg := gui.TableCfgFromData([][]string{
    {"Name", "Age"},   // header row
    {"Alice", "30"},
    {"Bob", "25"},
})
cfg.ID = "my-table"
w.Table(cfg)
` + "```" + `
`,

	"data_grid": `# Data Grid

Full-featured grid with sorting, filtering, paging, and column chooser.

## Usage

` + "```go" + `
w.DataGrid(gui.DataGridCfg{
    ID:       "grid",
    PageSize: 10,
    Columns: []gui.GridColumnCfg{
        {ID: "name", Title: "Name", Sortable: true},
    },
    Rows: []gui.GridRow{
        {ID: "1", Cells: map[string]string{"name": "Alice"}},
    },
})
` + "```" + `
`,

	"data_source": `# Data Source

Async data-source backed grid with CRUD operations.

## Usage

` + "```go" + `
source := gui.NewInMemoryDataSource([]gui.GridRow{
    {
        ID: "1",
        Cells: map[string]string{
            "name": "Alice", "team": "Core", "status": "Open",
        },
    },
})

w.DataGrid(gui.DataGridCfg{
    ID:              "ds-grid",
    Columns:         showcaseDataGridColumns(),
    DataSource:      source,
    PaginationKind:  gui.GridPaginationCursor,
    PageLimit:       50,
    ShowQuickFilter: true,
    ShowCRUDToolbar: true,
})
` + "```" + `

## DataGridDataSource

Backends can provide:

- Cursor or offset pagination
- Simulated or real loading latency
- Create, update, and delete mutations
- Runtime stats from ` + "`w.DataGridSourceStats(id)`" + `

This showcase uses ` + "`gui.NewInMemoryDataSource`" + ` to mirror the original source-backed grid demo.
`,

	// Text
	"text": `# Text

Theme typography sizes, weights, and styles.

## Usage

` + "```go" + `
gui.Text(gui.TextCfg{Text: "Hello", TextStyle: t.N3})
gui.Text(gui.TextCfg{Text: "Bold", TextStyle: t.B4})
` + "```" + `

## Style Shortcuts

| Prefix | Meaning | Sizes |
|--------|---------|-------|
| N | Normal | N1–N6 |
| B | Bold | B1–B6 |
| I | Italic | I1–I6 |
| M | Monospace | M1–M6 |
| Icon | Icon font | Icon1–Icon6 |
`,

	"rtf": `# Rich Text

Mixed styles within a single text block using TextStyle fields.

Supports bold, italic, underline, strikethrough, and custom colors.
`,

	"markdown": `# Markdown

Render markdown strings with syntax highlighting, tables, and blockquotes.

## Usage

` + "```go" + `
w.Markdown(gui.MarkdownCfg{
    Source: "# Hello\n**Bold** and *italic*",
    Style:  gui.DefaultMarkdownStyle(),
})
` + "```" + `
`,

	// Graphics
	"svg": `# SVG

Scalable vector graphics from file or inline string.

## Usage

` + "```go" + `
gui.Svg(gui.SvgCfg{
    SvgData: "<svg>...</svg>",
    Width: 100, Height: 100,
})
` + "```" + `
`,

	"image": `# Image

Display image files (PNG, JPEG, etc.).

## Usage

` + "```go" + `
gui.Image(gui.ImageCfg{
    Src:   "photo.png",
    Width: 200, Height: 150,
})
` + "```" + `
`,

	"gradient": `# Gradients

Linear gradients with configurable direction and color stops.

## Usage

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Gradient: &gui.GradientDef{
        Direction: gui.GradientToRight,
        Stops: []gui.GradientStop{
            {Pos: 0, Color: gui.ColorFromString("#3b82f6")},
            {Pos: 1, Color: gui.ColorFromString("#8b5cf6")},
        },
    },
})
` + "```" + `
`,

	"box_shadows": `# Box Shadows

Drop shadows on containers with offset, blur, and color.

## Usage

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Shadow: &gui.BoxShadow{
        OffsetX: 4, OffsetY: 4,
        BlurRadius: 16,
        Color: gui.RGBA(0, 0, 0, 80),
    },
})
` + "```" + `
`,

	"rectangle": `# Rectangle

Styled containers used as visual shapes: sharp, rounded, bordered, pill.

Set ` + "`Radius`" + ` to control corner rounding. Use ` + "`ColorBorder`" + ` and
` + "`SizeBorder`" + ` for outlined rectangles.
`,

	"icons": `# Icons

Icon font catalog rendered from ` + "`gui.IconLookup`" + ` using the theme Icon styles.

## Usage

` + "```go" + `
gui.Text(gui.TextCfg{Text: gui.IconCheck, TextStyle: t.Icon4})

for name, glyph := range gui.IconLookup {
    _ = name
    _ = glyph
}
` + "```" + `

Use ` + "`gui.IconLookup`" + ` for programmatic access to the full icon catalog.
`,

	// Layout
	"row": `# Row

Horizontal container — children flow left to right.

## Usage

` + "```go" + `
gui.Row(gui.ContainerCfg{
    Spacing: gui.SomeF(8),
    Content: []gui.View{child1, child2},
})
` + "```" + `
`,

	"column": `# Column

Vertical container — children flow top to bottom.

## Usage

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Spacing: gui.SomeF(8),
    Content: []gui.View{child1, child2},
})
` + "```" + `
`,

	"wrap_panel": `# Wrap Panel

Horizontal flow that wraps to the next line when full.

## Usage

` + "```go" + `
gui.Wrap(gui.ContainerCfg{
    Spacing: gui.SomeF(4),
    Content: items,
})
` + "```" + `
`,

	"overflow_panel": `# Overflow Panel

Scrollable container for content that exceeds available space.

Enable scrolling with ` + "`IDScroll`" + ` and ` + "`ScrollbarCfgY`" + `.
`,

	"expand_panel": `# Expand Panel

Collapsible sections with animated expand/collapse.

## Usage

` + "```go" + `
gui.ExpandPanel(gui.ExpandPanelCfg{
    ID:    "ep",
    Title: "Details",
    Open:  app.Open,
    OnToggle: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
        s := gui.State[App](w)
        s.Open = !s.Open
        e.IsHandled = true
    },
    Content: []gui.View{...},
})
` + "```" + `
`,

	"sidebar": `# Sidebar

Slide-out panel overlaying main content.

## Usage

` + "```go" + `
w.Sidebar(gui.SidebarCfg{
    ID:   "sb",
    Open: app.SidebarOpen,
    Content: []gui.View{...},
    MainContent: []gui.View{...},
})
` + "```" + `
`,

	"splitter": `# Splitter

Resizable split panes with draggable divider.

## Usage

` + "```go" + `
gui.Splitter(gui.SplitterCfg{
    ID:    "split",
    State: app.SplitterState,
    OnStateChange: func(s gui.SplitterState, w *gui.Window) {
        gui.State[App](w).SplitterState = s
    },
    Content1: []gui.View{left},
    Content2: []gui.View{right},
})
` + "```" + `
`,

	"scrollbar": `# Scrollable Containers

Custom scrollbar styling via ScrollbarCfg on any container.

## Usage

` + "```go" + `
gui.Column(gui.ContainerCfg{
    IDScroll:      myScrollID,
    ScrollbarCfgY: &gui.ScrollbarCfg{GapEdge: 4},
    Content:       views,
})
` + "```" + `
`,

	// Navigation
	"breadcrumb": `# Breadcrumb

Trail navigation with clickable path segments.

## Usage

` + "```go" + `
gui.Breadcrumb(gui.BreadcrumbCfg{
    Items: []gui.BreadcrumbItemCfg{
        {Label: "Home", ID: "home"},
        {Label: "Settings", ID: "settings"},
    },
    OnSelect: func(id string, _ *gui.Event, w *gui.Window) {
        gui.State[App](w).Page = id
    },
})
` + "```" + `
`,

	"tab_control": `# Tab Control

Switch content panels with keyboard-friendly tabs.

## Usage

` + "```go" + `
gui.TabControl(gui.TabControlCfg{
    Selected: app.Tab,
    Items: []gui.TabItemCfg{
        {ID: "t1", Label: "Tab 1", Content: []gui.View{...}},
        {ID: "t2", Label: "Tab 2", Content: []gui.View{...}},
    },
    OnSelect: func(id string, w *gui.Window) {
        gui.State[App](w).Tab = id
    },
})
` + "```" + `
`,

	"menus": `# Menus + Menubar

Nested menus with separators, submenus, and keyboard shortcuts.

## Usage

` + "```go" + `
gui.Menubar(gui.MenubarCfg{
    Items: []gui.MenuItemCfg{
        {Label: "File", Items: []gui.MenuItemCfg{
            {Label: "New", ID: "new"},
            {Label: "Open", ID: "open"},
            {Separator: true},
            {Label: "Quit", ID: "quit"},
        }},
    },
    OnAction: func(id string, _ *gui.Event, w *gui.Window) {
        // handle menu action
    },
})
` + "```" + `
`,

	"command_palette": `# Command Palette

Quick command search with fuzzy filtering.

## Usage

` + "```go" + `
gui.CommandPalette(gui.CommandPaletteCfg{
    ID:       "cmd",
    IDFocus:  focusPalette,
    IDScroll: scrollPalette,
    Items:    items,
    OnAction: func(id string, _ *gui.Event, w *gui.Window) {
        // handle action
    },
})

// Toggle with Ctrl+K or programmatically:
gui.CommandPaletteToggle("cmd", focusPalette, w)
` + "```" + `
`,

	// Overlays
	"dialog": `# Dialog

Message, confirm, and custom modal dialogs.

## Usage

` + "```go" + `
w.Dialog(gui.DialogCfg{
    Title:      "Confirm",
    Body:       "Delete this item?",
    DialogType: gui.DialogConfirm,
    OnOkYes: func(w *gui.Window) {
        // confirmed
    },
    OnCancelNo: func(w *gui.Window) {
        // cancelled
    },
})
` + "```" + `

## Dialog Types

| Type | Buttons |
|------|---------|
| DialogMessage | OK |
| DialogConfirm | Yes / No |
`,

	"tooltip": `# Tooltip

Hover hints attached to any widget.

## Usage

` + "```go" + `
gui.WithTooltip(w, gui.WithTooltipCfg{
    Text: "Helpful hint",
    Content: []gui.View{
        gui.Button(gui.ButtonCfg{...}),
    },
})
` + "```" + `
`,

	// Animations
	"animations": `# Animations

Tween, spring, and keyframe animation APIs.

## Tween

` + "```go" + `
a := gui.NewTweenAnimation("id", from, to,
    func(v float32, w *gui.Window) { ... })
w.AnimationAdd(a)
` + "```" + `

## Spring

` + "```go" + `
a := gui.NewSpringAnimation("id",
    func(v float32, w *gui.Window) { ... })
a.SpringTo(from, to)
w.AnimationAdd(a)
` + "```" + `

## Keyframes

` + "```go" + `
a := gui.NewKeyframeAnimation("id",
    []gui.Keyframe{
        {At: 0, Value: 0, Easing: gui.EaseLinear},
        {At: 1, Value: 300, Easing: gui.EaseOutBounce},
    },
    func(v float32, w *gui.Window) { ... })
w.AnimationAdd(a)
` + "```" + `
`,

	// Theme
	"theme_gen": `# Theme Generator

Generate a complete theme from a seed color using HSV color theory.

Strategies: mono, complement, analogous, triadic, warm, cool.

Use the tint slider to control surface saturation.
`,

	// Notification
	"notification": `# Notification

Send native OS notifications from your application.

## Usage

` + "```go" + `
w.NativeNotification(gui.NativeNotificationCfg{
    Title: "App",
    Body:  "Task completed!",
    OnDone: func(r gui.NativeNotificationResult, w *gui.Window) {
        if r.Status == gui.NotificationOK { ... }
    },
})
` + "```" + `

## Result Status

| Status | Meaning |
|--------|---------|
| NotificationOK | Delivered |
| NotificationDenied | Permission denied |
| NotificationError | Platform error |
`,

	// Shader
	"shader": `# Custom Shader

Apply custom fragment shaders (Metal + GLSL) to any container.

## Usage

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Width: 300, Height: 200,
    Sizing: gui.FixedFixed,
    Shader: &gui.Shader{
        Metal: "float4(uv.x, uv.y, 0.5, 1.0)",
        GLSL:  "fragColor = vec4(uv.x, uv.y, 0.5, 1.0);",
        Params: []float32{0},
    },
})
` + "```" + `

Params (up to 16 floats) are passed to the shader. Animate them
with the animation system for dynamic effects.
`,

	// Printing
	"printing": `# Printing / PDF Export

Export the current window to PDF or send to the OS print dialog.

## Export PDF

` + "```go" + `
job := gui.NewPrintJob()
job.OutputPath = "/tmp/output.pdf"
job.Title = "My Document"
r := w.ExportPrintJob(job)
` + "```" + `

## Print

` + "```go" + `
job := gui.NewPrintJob()
job.Title = "My Document"
r := w.RunPrintJob(job)
` + "```" + `
`,

	"tree": `# Tree View

Hierarchical expandable node display with virtualization and lazy-loading support.

## Showcase Sections

- Basic tree
- Virtualized tree (scroll)
- Lazy-loading tree

## Sample Nodes

- Mammals
  - Lion
  - Cat
  - Human
- Birds
  - Condor
  - Eagle
    - Bald
    - Golden
    - Sea
  - Parrot
  - Robin

## Implemented Features

- Expand/collapse nested nodes
- Flat-row virtualization for large trees
- Lazy loading when folders expand
- Keyboard navigation with ` + "`Up`" + ` / ` + "`Down`" + ` / ` + "`Left`" + ` / ` + "`Right`" + ` / ` + "`Home`" + ` / ` + "`End`" + ` / ` + "`Enter`" + ` / ` + "`Space`" + `

## API

` + "```go" + `
gui.Tree(gui.TreeCfg{
    ID:       "project-tree",
    IDFocus:  2001,
    IDScroll: 2002,
    MaxHeight: 240,
    OnSelect: func(nodeID string, _ *gui.Event, w *gui.Window) {
        gui.State[AppState](w).SelectedNode = nodeID
    },
    OnLazyLoad: func(treeID, nodeID string, w *gui.Window) {
        // Load children and update state, then refresh.
    },
    Nodes: []gui.TreeNodeCfg{
        {
            ID:   "src",
            Text: "src",
            Icon: gui.IconFolder,
            Nodes: []gui.TreeNodeCfg{
                {Text: "main.go"},
                {Text: "view_tree.go"},
            },
        },
        {ID: "remote", Text: "remote", Lazy: true},
    },
})
` + "```" + `

` + "`TreeNodeCfg.ID`" + ` defaults to ` + "`Text`" + ` when omitted. Node IDs should be unique within a single tree.

## Deferred

Drag-reorder remains deferred and is documented separately on the Drag Reorder page.
`,

	// Drag Reorder
	"drag_reorder": `# Drag Reorder

Drag items to reorder within lists, tabs, and tree views. Keyboard shortcuts provide an accessible alternative to mouse dragging.

## Behaviors

- Drag items to reorder (5px threshold before activation)
- Use ` + "`Alt+Arrow`" + ` as a keyboard fallback
- Press ` + "`Escape`" + ` to cancel an active drag
- Tree reordering is scoped to siblings under the same parent
- FLIP animation on index change and drop
- Auto-scroll near container edges during drag

## Usage

` + "```go" + `
gui.ListBox(gui.ListBoxCfg{
    Reorderable: true,
    OnReorder: func(movedID, beforeID string, w *gui.Window) {
        from, to := gui.ReorderIndices(ids, movedID, beforeID)
        if from >= 0 { sliceMove(items, from, to) }
    },
})
` + "```" + `

The same ` + "`Reorderable`" + ` + ` + "`OnReorder`" + ` pattern applies to TabControl and Tree.

` + "`ReorderIndices`" + ` computes delete/insert indices from ` + "`(movedID, beforeID)`" + `.
` + "`beforeID`" + ` is ` + "`\"\"`" + ` when dropping at the end of the list.
`,

	// Locale
	"locale": `# Locale

Switch between registered locales to change date/number formatting
and UI strings (OK, Cancel, etc.).

Built-in: en-US, de-DE, fr-FR, es-ES, pt-BR, ja-JP, zh-CN, ko-KR, ar-SA, he-IL.
`,
}

// ── General doc pages (shown as standalone entries in the Docs group) ──

const docGetStarted = `# Getting Started

## Minimal App

` + "```go" + `
package main

import (
    "github.com/mike-ward/go-gui/gui"
    "github.com/mike-ward/go-gui/gui/backend"
)

type App struct{ Count int }

func main() {
    w := gui.NewWindow(gui.WindowCfg{
        State: &App{}, Title: "Counter",
        Width: 400, Height: 300,
        OnInit: func(w *gui.Window) { w.UpdateView(view) },
    })
    backend.Run(w)
}

func view(w *gui.Window) gui.View {
    app := gui.State[App](w)
    return gui.Column(gui.ContainerCfg{
        Sizing: gui.FillFill, HAlign: gui.HAlignCenter,
        VAlign: gui.VAlignMiddle, Spacing: gui.SomeF(8),
        Content: []gui.View{
            gui.Text(gui.TextCfg{Text: fmt.Sprintf("Count: %d", app.Count)}),
            gui.Button(gui.ButtonCfg{
                ID: "inc",
                Content: []gui.View{gui.Text(gui.TextCfg{Text: "+"})},
                OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
                    gui.State[App](w).Count++
                    e.IsHandled = true
                },
            }),
        },
    })
}
` + "```" + `

## Key Concepts

- **One state per window**: ` + "`gui.State[T](w)`" + ` returns ` + "`*T`" + `
- **Immediate mode**: the view function runs every frame
- **No virtual DOM**: layout tree is built fresh each frame
- **Event callbacks**: set ` + "`e.IsHandled = true`" + ` to consume
`

const docArchitecture = `# Architecture

## Rendering Pipeline

` + "```" + `
View fn → Layout tree → layoutArrange() → renderLayout() → []RenderCmd → Backend
` + "```" + `

The framework is **immediate-mode** — no virtual DOM, no diffing.

## State Management

One typed state slot per window.

` + "```go" + `
w := gui.NewWindow(gui.WindowCfg{State: &MyApp{}})
app := gui.State[MyApp](w)
` + "```" + `

## Sizing Model

| Constant | Width | Height |
|----------|-------|--------|
| FitFit | Fit | Fit |
| FillFill | Fill | Fill |
| FixedFixed | Fixed | Fixed |
| FillFit | Fill | Fit |
| FixedFill | Fixed | Fill |

## Core Types

- **Layout** — tree node with Shape, parent, children
- **Shape** — renderable: position, size, color, type, events
- **RenderCmd** — single draw operation sent to backend
- **View** — interface satisfied by *Layout
`

const docContainers = `# Containers

## Row, Column, Wrap

| Container | Axis | Behavior |
|-----------|------|----------|
| Row | Horizontal | Left to right |
| Column | Vertical | Top to bottom |
| Wrap | Horizontal | Wraps when full |

## Alignment

- **HAlign**: Left (default), Center, Right
- **VAlign**: Top (default), Middle, Bottom

## Spacing & Padding

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Spacing: gui.SomeF(8),
    Padding: gui.Some(gui.NewPadding(16, 16, 16, 16)),
})
` + "```" + `

## Scrolling

` + "```go" + `
gui.Column(gui.ContainerCfg{
    IDScroll:      myScrollID,
    ScrollbarCfgY: &gui.ScrollbarCfg{GapEdge: 4},
})
` + "```" + `
`

const docThemes = `# Themes

## Built-in Themes

ThemeDark, ThemeDarkBordered, ThemeDarkNoPadding,
ThemeLight, ThemeLightBordered, ThemeLightNoPadding

## Switching

` + "```go" + `
gui.SetTheme(gui.ThemeLight)
w.SetTheme(gui.ThemeLight)
` + "```" + `

## Custom Themes

` + "```go" + `
cfg := gui.ThemeCfg{
    Name:            "custom",
    ColorBackground: gui.ColorFromHSV(220, 0.15, 0.19),
    TextStyleDef:    gui.ThemeDarkCfg.TextStyleDef,
    TitlebarDark:    true,
}
gui.SetTheme(gui.ThemeMaker(cfg))
` + "```" + `

## Text Styles

N1–N6 (normal), B1–B6 (bold), I1–I6 (italic), M1–M6 (mono), Icon1–Icon6
`

const docAnimations = `# Animations

## Tween

` + "```go" + `
a := gui.NewTweenAnimation("id", from, to,
    func(v float32, w *gui.Window) { ... })
w.AnimationAdd(a)
` + "```" + `

## Spring

` + "```go" + `
a := gui.NewSpringAnimation("id",
    func(v float32, w *gui.Window) { ... })
a.SpringTo(from, to)
w.AnimationAdd(a)
` + "```" + `

## Keyframes

` + "```go" + `
a := gui.NewKeyframeAnimation("id",
    []gui.Keyframe{
        {At: 0, Value: 0, Easing: gui.EaseLinear},
        {At: 1, Value: 300, Easing: gui.EaseOutBounce},
    },
    func(v float32, w *gui.Window) { ... })
w.AnimationAdd(a)
` + "```" + `

## Easings

EaseLinear, EaseInQuad, EaseOutQuad, EaseInOutQuad,
EaseInCubic, EaseOutCubic, EaseInOutCubic,
EaseInBack, EaseOutBack, EaseOutElastic, EaseOutBounce
`

const docLocales = `# Locales

## Built-in

en-US, de-DE, fr-FR, es-ES, pt-BR, ja-JP, zh-CN, ko-KR, ar-SA, he-IL

## Switching

` + "```go" + `
w.SetLocaleID("de-DE")
gui.LocaleAutoDetect()
` + "```" + `

## Formatting

` + "```go" + `
locale := gui.CurrentLocale()
gui.LocaleFormatDate(time.Now(), locale.Date.ShortDate)
` + "```" + `

## Custom Locales

` + "```go" + `
locale, _ := gui.LocaleParse(jsonString)
gui.LocaleRegister(locale)
` + "```" + `
`

const docCustomWidgets = `# Custom Widgets

Build new widgets by composing existing ones. There is no special
widget registration — just return a ` + "`View`" + ` from a function.

## Pattern

` + "```go" + `
type MyWidgetCfg struct {
    ID    string
    Label string
    Value float32
}

func MyWidget(cfg MyWidgetCfg) gui.View {
    t := gui.CurrentTheme()
    return gui.Column(gui.ContainerCfg{
        Sizing: gui.FillFit,
        Content: []gui.View{
            gui.Text(gui.TextCfg{Text: cfg.Label, TextStyle: t.B3}),
            gui.ProgressBar(gui.ProgressBarCfg{
                Percent: cfg.Value,
                Sizing:  gui.FillFit,
            }),
        },
    })
}
` + "```" + `

## Guidelines

- Accept a ` + "`*Cfg`" + ` struct (zero-initializable)
- Use ` + "`Opt[T]`" + ` for optional numeric fields
- Event callbacks: ` + "`func(*Layout, *Event, *Window)`" + `
- Set ` + "`e.IsHandled = true`" + ` to consume events
- Prefer ` + "`Column`" + ` over raw ` + "`container()`" + ` for correct sizing
`

const docDataGrid = `# Data Grid

Full-featured grid with sorting, filtering, paging, column reorder,
and column chooser.

## Basic Setup

` + "```go" + `
w.DataGrid(gui.DataGridCfg{
    ID:       "grid",
    PageSize: 10,
    Columns: []gui.GridColumnCfg{
        {ID: "name", Title: "Name", Width: 150,
         Sortable: true, Filterable: true, Reorderable: true},
    },
    Rows: []gui.GridRow{
        {ID: "1", Cells: map[string]string{"name": "Alice"}},
    },
    ShowQuickFilter:   true,
    ShowColumnChooser: true,
})
` + "```" + `

## Column Options

| Field | Type | Description |
|-------|------|-------------|
| Sortable | bool | Enable column sort |
| Filterable | bool | Enable column filter |
| Reorderable | bool | Allow drag reorder |
| Resizable | bool | Allow resize |
| Editable | bool | Inline editing |
| Pin | GridColumnPin | Freeze left/right |
`

const docForms = `# Forms

Combine inputs, labels, and containers into form layouts.

## Labeled Row Pattern

` + "```go" + `
func labeledRow(t gui.Theme, label string, input gui.View) gui.View {
    return gui.Row(gui.ContainerCfg{
        Sizing: gui.FillFit, VAlign: gui.VAlignMiddle,
        Spacing: gui.SomeF(8),
        Content: []gui.View{
            gui.Text(gui.TextCfg{Text: label, TextStyle: t.B3,
                Sizing: gui.FixedFit}),
            input,
        },
    })
}
` + "```" + `

## Fieldset Grouping

Use ` + "`ContainerCfg.Title`" + ` for HTML-fieldset-style group boxes:

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Title:       "Personal Info",
    TitleBG:     t.ColorBackground,
    ColorBorder: t.ColorBorder,
    SizeBorder:  gui.SomeF(1),
})
` + "```" + `

## Masked Input

` + "```go" + `
gui.Input(gui.InputCfg{
    MaskPreset: gui.MaskPhoneUS,
    Placeholder: "(555) 000-0000",
})
` + "```" + `
`

const docGradients = `# Gradients

Apply linear gradients to any container.

## Directions

GradientToRight, GradientToLeft, GradientToTop, GradientToBottom,
GradientToTopRight, GradientToTopLeft, GradientToBottomRight,
GradientToBottomLeft

## Usage

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Gradient: &gui.GradientDef{
        Direction: gui.GradientToRight,
        Stops: []gui.GradientStop{
            {Pos: 0, Color: gui.ColorFromString("#3b82f6")},
            {Pos: 1, Color: gui.ColorFromString("#8b5cf6")},
        },
    },
})
` + "```" + `

## Border Gradient

` + "```go" + `
gui.Column(gui.ContainerCfg{
    BorderGradient: &gui.GradientDef{...},
    SizeBorder: gui.SomeF(2),
})
` + "```" + `
`

const docLayoutAlgorithm = `# Layout Algorithm

The layout engine runs in two passes: measure then arrange.

## Sizing Modes

| Mode | Behavior |
|------|----------|
| Fit | Shrink to content |
| Fixed | Use explicit Width/Height |
| Fill (Grow) | Expand to fill parent |

Combined as two axes: FitFit, FillFill, FixedFixed, FillFit, etc.

## Measure Pass

1. Fit children measured first (leaf to root)
2. Fixed children use their declared size
3. Fill children split remaining space proportionally

## Arrange Pass

1. Children positioned per axis direction (Row=horizontal, Column=vertical)
2. Spacing inserted between visible children
3. Alignment (HAlign, VAlign) offsets children within the container
4. AmendLayout hooks run for overlay repositioning

## Key Rules

- ` + "`spacing()`" + ` counts only visible children (not ShapeNone, Float, OverDraw)
- Fill children in an AxisNone container get 0 size
- Prefer Column/Row over raw container() for correct Fill distribution
`

const docMarkdownGuide = `# Markdown

Render markdown strings with full CommonMark support.

## Usage

` + "```go" + `
w.Markdown(gui.MarkdownCfg{
    Source: "# Hello\n**Bold** and *italic*",
    Style:  gui.DefaultMarkdownStyle(),
})
` + "```" + `

## Supported Features

- Headings (H1--H6)
- **Bold**, *italic*, ~~strikethrough~~, ` + "`inline code`" + `
- Links and images
- Ordered and unordered lists
- Code blocks with syntax highlighting
- Blockquotes
- Tables (GFM)
- Horizontal rules

## Custom Styles

` + "```go" + `
style := gui.DefaultMarkdownStyle()
style.H1 = gui.TextStyle{Size: 32, Color: myColor}
` + "```" + `
`

const docNativeDialogs = `# Native Dialogs

Access OS-native file open, save, and folder dialogs.

## Open File

` + "```go" + `
np := w.NativePlatformBackend()
r := np.ShowOpenDialog("Open", "", []string{".go", ".txt"}, false)
if len(r.Paths) > 0 {
    path := r.Paths[0].Path
}
` + "```" + `

## Save File

` + "```go" + `
r := np.ShowSaveDialog("Save", "", "file.txt", ".txt", nil, true)
` + "```" + `

## Select Folder

` + "```go" + `
r := np.ShowFolderDialog("Select Folder", "")
` + "```" + `

## Result

` + "```go" + `
type PlatformDialogResult struct {
    Status       NativeDialogStatus
    Paths        []PlatformPath
    ErrorCode    string
    ErrorMessage string
}
` + "```" + `
`

const docPerformance = `# Performance

Tips for keeping go-gui apps fast and allocation-free.

## Reduce Allocations

- Reuse slices across frames (store in state, reset with ` + "`[:0]`" + `)
- Avoid ` + "`fmt.Sprintf`" + ` in hot paths; use ` + "`strconv`" + ` instead
- Prefer value types over pointers for small structs

## StateMap

Per-window typed key-value store for widget internal state.
Keyed by namespace constants (nsOverflow, nsSvgCache, etc.).
Avoids globals and closures.

## Immediate-Mode Patterns

- The view function runs every frame -- keep it fast
- No virtual DOM or diffing; layout tree built fresh each frame
- Avoid expensive computations in view functions; cache in state
- Event callbacks should set ` + "`e.IsHandled = true`" + ` early

## Profiling

Use Go's built-in pprof. The hottest path is typically
` + "`layoutArrange()`" + ` -> ` + "`renderLayout()`" + `.
`

const docPrinting = `# Printing

Export the current window to PDF or send to the OS print dialog.

## Export PDF

` + "```go" + `
job := gui.NewPrintJob()
job.OutputPath = "/tmp/output.pdf"
job.Title = "My Document"
r := w.ExportPrintJob(job)
if r.ErrorMessage != "" {
    // handle error
}
` + "```" + `

## Print via OS Dialog

` + "```go" + `
job := gui.NewPrintJob()
job.Title = "My Document"
r := w.RunPrintJob(job)
` + "```" + `

## PrintJob Options

| Field | Type | Description |
|-------|------|-------------|
| Paper | PaperSize | A4, Letter, etc. |
| Orientation | PrintOrientation | Portrait / Landscape |
| Margins | PrintMargins | Top/Right/Bottom/Left |
| Copies | int | Number of copies |
| Duplex | PrintDuplexMode | Simplex / LongEdge / ShortEdge |
`

const docShaders = `# Custom Shaders

Apply custom fragment shaders to any container. Provide both
Metal (MSL) and GLSL bodies for cross-platform support.

## Usage

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Width: 300, Height: 200,
    Sizing: gui.FixedFixed,
    Shader: &gui.Shader{
        Metal: "return float4(uv.x, uv.y, 0.5, 1.0);",
        GLSL:  "fragColor = vec4(uv.x, uv.y, 0.5, 1.0);",
        Params: []float32{0},
    },
})
` + "```" + `

## Parameters

Up to 16 float32 values passed as ` + "`params[]`" + ` (Metal) or
` + "`u_params[]`" + ` (GLSL). Animate them with the animation system.

## Available Uniforms

| Metal | GLSL | Description |
|-------|------|-------------|
| uniforms.size | u_size | Container size (vec2) |
| in.position | gl_FragCoord | Fragment position |
| params[i] | u_params[i] | Custom parameters |
`

const docSplitterGuide = `# Splitter

Resizable split panes with draggable divider.

## Usage

` + "```go" + `
gui.Splitter(gui.SplitterCfg{
    ID:          "split",
    Orientation: gui.SplitterHorizontal,
    Sizing:      gui.FillFixed,
    Ratio:       app.SplitterState.Ratio,
    Collapsed:   app.SplitterState.Collapsed,
    OnChange: func(r float32, c gui.SplitterCollapsed,
        _ *gui.Event, w *gui.Window) {
        a := gui.State[App](w)
        a.SplitterState.Ratio = r
        a.SplitterState.Collapsed = c
    },
    First:  gui.SplitterPaneCfg{MinSize: 100, Content: [...]},
    Second: gui.SplitterPaneCfg{MinSize: 100, Content: [...]},
    ShowCollapseButtons: true,
    DoubleClickCollapse: true,
})
` + "```" + `

## Key Properties

| Property | Description |
|----------|-------------|
| IDFocus | Enables keyboard control |
| DoubleClickCollapse | Collapse on divider double-click |
| ShowCollapseButtons | Show collapse arrows |
| DragStep | Arrow key step size |
| DragStepLarge | Shift+Arrow step size |
`

const docSvg = `# SVG

Render inline SVG strings or load from files.

## Inline SVG

` + "```go" + `
gui.Svg(gui.SvgCfg{
    SvgData: "<svg viewBox='0 0 100 100'>...</svg>",
    Width: 100, Height: 100,
})
` + "```" + `

## TextPath

Render text along a curve using the SVG textPath element:

` + "```xml" + `
<svg viewBox="0 0 200 100">
  <path id="curve" d="M10,80 Q100,10 190,80" fill="none"/>
  <text><textPath href="#curve">Text on a path</textPath></text>
</svg>
` + "```" + `

The SVG parser tessellates paths and renders via the backend.
`

const docTables = `# Tables

Two table widgets: Table (simple) and DataGrid (full-featured).

## Table

Build from string arrays using ` + "`TableCfgFromData`" + `:

` + "```go" + `
cfg := gui.TableCfgFromData([][]string{
    {"Name", "Age"},
    {"Alice", "30"},
    {"Bob", "25"},
})
cfg.ID = "my-table"
w.Table(cfg)
` + "```" + `

## DataGrid

For sorting, filtering, paging, column reorder:

` + "```go" + `
w.DataGrid(gui.DataGridCfg{
    ID:       "grid",
    PageSize: 10,
    Columns:  []gui.GridColumnCfg{...},
    Rows:     []gui.GridRow{...},
    ShowQuickFilter:   true,
    ShowColumnChooser: true,
})
` + "```" + `

See the Data Grid Guide doc page for full column options.
`
