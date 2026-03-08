package main

import "github.com/mike-ward/go-gui/gui"

func demoDoc(w *gui.Window, source string) gui.View {
	return w.Markdown(gui.MarkdownCfg{Source: source, Style: gui.DefaultMarkdownStyle()})
}

func componentDoc(id string) string {
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
gui.Table(cfg)
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

Feather icon font glyphs rendered as text with Icon styles.

## Usage

` + "```go" + `
gui.Text(gui.TextCfg{Text: gui.IconHome, TextStyle: t.Icon4})
` + "```" + `

Available icons: IconHome, IconSearch, IconHeart, IconStar, IconBell,
IconCalendar, IconCamera, IconClock, IconCloud, IconCode, IconDownload,
IconEdit, IconEye, IconFilter, IconGlobe, IconInfo, IconLayout,
IconPlus, IconTag, IconTrash, and more.
`,

	// Layout
	"row": `# Row

Horizontal container — children flow left to right.

## Usage

` + "```go" + `
gui.Row(gui.ContainerCfg{
    Spacing: gui.Some(float32(8)),
    Content: []gui.View{child1, child2},
})
` + "```" + `
`,

	"column": `# Column

Vertical container — children flow top to bottom.

## Usage

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Spacing: gui.Some(float32(8)),
    Content: []gui.View{child1, child2},
})
` + "```" + `
`,

	"wrap_panel": `# Wrap Panel

Horizontal flow that wraps to the next line when full.

## Usage

` + "```go" + `
gui.Wrap(gui.ContainerCfg{
    Spacing: gui.Some(float32(4)),
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
        VAlign: gui.VAlignMiddle, Spacing: gui.Some(float32(8)),
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
    Spacing: gui.Some(float32(8)),
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
