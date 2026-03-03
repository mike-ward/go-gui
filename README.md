# go-gui

![Go version](https://img.shields.io/badge/go-1.26%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![CI](https://github.com/mike-ward/go-gui/actions/workflows/ci.yml/badge.svg)

Cross-platform immediate mode GUI framework for Go.

## Overview

go-gui is an immediate-mode GUI framework ported from the
[V-language gui library](https://github.com/mike-ward/gui). Every frame, a
View function returns a tree of Layout nodes. The layout engine sizes and
positions them. The backend converts the result into draw commands and
renders them to the screen. No virtual DOM, no diffing — the tree is
rebuilt each frame.

State is held in a typed slot on the Window (`State[T](w)`), not in closures
or global variables. Views are pure functions: given the same window state
they produce the same layout. This makes reasoning about UI straightforward
and testing easy — the `gui/backend/test` no-op backend runs all layout
and widget logic headlessly.

go-gui descends from the V-language gui framework and preserves its design
philosophy: minimal ceremony, composable widgets, and a clear separation
between layout and rendering.

## Features

- 50+ built-in widgets (buttons, inputs, sliders, tables, trees, tabs, …)
- Stateless view model — views are plain functions
- Full animation subsystem with keyframe and spring animations
- GPU-accelerated rendering via SDL2 + Metal/OpenGL shaders
- DataGrid with virtualization, sorting, grouping, inline editing,
  CSV/TSV/XLSX/PDF export, and async data sources
- Markdown and RTF views with syntax highlighting
- SVG loading, caching, and tessellation
- Dock layout with drag-and-drop and tab groups
- IME support and accessibility tree
- Native dialogs, notifications, and print/PDF
- Locale and i18n support
- Theme system with built-in dark/light variants and custom theme support

## Requirements

- Go 1.26+
- SDL2 development libraries (see below)
- [go-glyph](https://github.com/mike-ward/go-glyph) — auto-fetched via
  `go get`

### Installing SDL2

| Platform         | Command                            |
| ---------------- | ---------------------------------- |
| macOS (Homebrew) | `brew install sdl2`                |
| Ubuntu / Debian  | `sudo apt-get install libsdl2-dev` |
| Fedora / RHEL    | `sudo dnf install SDL2-devel`      |
| Arch Linux       | `sudo pacman -S sdl2`              |
| Windows (MSYS2)  | `pacman -S mingw-w64-x86_64-SDL2`  |
| Windows (vcpkg)  | `vcpkg install sdl2`               |
## Installation

    go get github.com/mike-ward/go-gui

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/mike-ward/go-gui/gui"
    sdl2 "github.com/mike-ward/go-gui/gui/backend/sdl2"
)

// App holds all mutable state for the application.
type App struct {
    Clicks int
}

func main() {
    // Choose a built-in theme before opening the window.
    gui.SetTheme(gui.ThemeDarkBordered)

    w := gui.NewWindow(gui.WindowCfg{
        State:  &App{},  // typed state slot; retrieve with gui.State[App](w)
        Title:  "get_started",
        Width:  300,
        Height: 300,
        OnInit: func(w *gui.Window) {
            w.UpdateView(mainView) // register the view function
        },
    })

    sdl2.Run(w) // blocks until the window is closed
}

// mainView is called every frame. It returns a layout tree.
func mainView(w *gui.Window) gui.View {
    ww, wh := w.WindowSize()
    app := gui.State[App](w)

    return gui.Column(gui.ContainerCfg{
        Width:   float32(ww),
        Height:  float32(wh),
        Sizing:  gui.FixedFixed,
        HAlign:  gui.HAlignCenter,
        VAlign:  gui.VAlignMiddle,
        Spacing: 10,
        Content: []gui.View{
            gui.Text(gui.TextCfg{
                Text:      "Hello go-gui!",
                TextStyle: gui.CurrentTheme().B1,
            }),
            gui.Button(gui.ButtonCfg{
                IDFocus: 1,
                Content: []gui.View{
                    gui.Text(gui.TextCfg{
                        Text: fmt.Sprintf("%d Clicks", app.Clicks),
                    }),
                },
                OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
                    gui.State[App](w).Clicks++
                },
            }),
        },
    })
}
```

See [`examples/get_started/`](examples/get_started/) for the full runnable
version.

## Core Concepts

### Views and Layouts

A *View* is any value that satisfies the `gui.View` interface — in practice,
a `*gui.Layout` returned by a widget factory such as `gui.Button(...)` or
`gui.Column(...)`.

The View *function* registered with `w.UpdateView(fn)` is called every frame.
It returns the root of the layout tree. The engine then:

1. Runs the sizing pass (`layoutArrange`) to compute widths, heights, and
   positions.
2. Walks the tree with `renderLayout` to produce a `[]RenderCmd` slice.
3. Hands the commands to the backend for drawing.

### State Management

Per-window state is stored in a typed slot:

```go
type Counter struct{ N int }

w := gui.NewWindow(gui.WindowCfg{State: &Counter{}})
// Inside any callback or view:
s := gui.State[Counter](w)
s.N++
```

`State[T]` performs a single type-assertion; it panics if the stored value is
not a `*T`.

### Themes

```go
gui.SetTheme(gui.ThemeDark)          // set globally before NewWindow
t := gui.CurrentTheme()              // read anywhere
t.B1                                 // bold large text style
t.N1                                 // normal text style
```

Custom themes are built with `gui.ThemeMaker` and registered via
`gui.RegisterTheme`.

### Event Handling

Events are wired through `Cfg` structs on each widget:

```go
gui.Button(gui.ButtonCfg{
    IDFocus: 1,                       // tab-order index
    OnClick: func(l *gui.Layout, e *gui.Event, w *gui.Window) { … },
    OnHover: func(l *gui.Layout, e *gui.Event, w *gui.Window) { … },
})
gui.Input(gui.InputCfg{
    OnKeyDown:    func(l *gui.Layout, e *gui.Event, w *gui.Window) { … },
    OnCharInput:  func(l *gui.Layout, e *gui.Event, w *gui.Window) { … },
    OnTextCommit: func(l *gui.Layout, e *gui.Event, w *gui.Window) { … },
})
```

### Widget Catalogue

#### Layout

| Widget        | Factory                           | Description                |
| ------------- | --------------------------------- | -------------------------- |
| Row           | `Row(ContainerCfg)`               | Horizontal flex container  |
| Column        | `Column(ContainerCfg)`            | Vertical flex container    |
| Spacer        | `Spacer(SpacerCfg)`               | Flexible blank space       |
| ExpandPanel   | `ExpandPanel(ExpandPanelCfg)`     | Collapsible section        |
| Splitter      | `Split(SplitterCfg)`              | Resizable two-pane split   |
| DockLayout    | `DockLayout(DockCfg)`             | Drag-and-drop dock areas   |
| OverflowPanel | `OverflowPanel(OverflowPanelCfg)` | Wraps overflowing children |

#### Input

| Widget           | Factory                                 | Description                  |
| ---------------- | --------------------------------------- | ---------------------------- |
| Button           | `Button(ButtonCfg)`                     | Clickable button             |
| Input            | `Input(InputCfg)`                       | Single-line text field       |
| Checkbox         | `Checkbox(CheckboxCfg)`                 | Boolean toggle               |
| RadioButtonGroup | `RadioButtonGroup(RadioButtonGroupCfg)` | Mutually-exclusive options   |
| Select           | `Select(SelectCfg)`                     | Dropdown selector            |
| Combobox         | `Combobox(ComboboxCfg)`                 | Editable dropdown            |
| RangeSlider      | `RangeSlider(RangeSliderCfg)`           | Min/max range picker         |
| InputDate        | `InputDate(InputDateCfg)`               | Date field with picker       |
| ColorPicker      | `ColorPicker(ColorPickerCfg)`           | RGBA color picker            |
| CommandPalette   | `CommandPalette(CommandPaletteCfg)`     | Fuzzy-search command palette |

#### Display

| Widget       | Factory                         | Description                   |
| ------------ | ------------------------------- | ----------------------------- |
| Text         | `Text(TextCfg)`                 | Styled text label             |
| Badge        | `Badge(BadgeCfg)`               | Notification badge            |
| ProgressBar  | `ProgressBar(ProgressBarCfg)`   | Determinate/indeterminate bar |
| Pulsar       | `Pulsar(PulsarCfg)`             | Blinking cursor indicator     |
| DrawCanvas   | `DrawCanvas(DrawCanvasCfg)`     | Custom-draw surface           |
| Image        | `Image(ImageCfg)`               | Raster image view             |
| SvgView      | `SvgView(SvgViewCfg)`           | SVG vector image view         |
| MarkdownView | `MarkdownView(MarkdownViewCfg)` | Rendered Markdown             |
| RtfView      | `RtfView(RtfViewCfg)`           | Rendered RTF                  |

#### Data

| Widget           | Factory                                 | Description                     |
| ---------------- | --------------------------------------- | ------------------------------- |
| ListBox          | `ListBox(ListBoxCfg)`                   | Scrollable item list            |
| Table            | `Table(TableCfg)`                       | Static data table               |
| DataGrid         | `DataGrid(DataGridCfg)`                 | Virtualized grid with full CRUD |
| DatePicker       | `DatePicker(DatePickerCfg)`             | Calendar date picker            |
| DatePickerRoller | `DatePickerRoller(DatePickerRollerCfg)` | Drum-style date roller          |

#### Navigation

| Widget     | Factory                     | Description                   |
| ---------- | --------------------------- | ----------------------------- |
| TabControl | `Tabs(TabControlCfg)`       | Tabbed panel                  |
| Breadcrumb | `Breadcrumb(BreadcrumbCfg)` | Navigational breadcrumb trail |
| Menu       | `Menu(MenuCfg)`             | Pull-down menu bar            |
| Sidebar    | `Sidebar(SidebarCfg)`       | Collapsible side navigation   |

#### Overlay

| Widget  | Factory               | Description            |
| ------- | --------------------- | ---------------------- |
| Dialog  | `Dialog(DialogCfg)`   | Modal dialog           |
| Toast   | `Toast(ToastCfg)`     | Transient notification |
| Tooltip | `Tooltip(TooltipCfg)` | Hover tooltip          |

## Backend

go-gui separates rendering concerns into injectable interfaces:

| Interface        | Purpose                                    |
| ---------------- | ------------------------------------------ |
| `TextMeasurer`   | Measure glyph extents for layout           |
| `SvgParser`      | Parse and tessellate SVG files             |
| `NativePlatform` | Native dialogs, notifications, print, a11y |

The SDL2 backend (`gui/backend/sdl2`) implements all three and wires itself
into the window on `sdl2.New(w)`. Custom backends implement the interfaces
and call the corresponding `w.Set*` methods.

The headless test backend (`gui/backend/test`) provides a no-op
implementation used by all unit tests.

## Architecture

```
View fn
  │
  ▼
GenerateViewLayout()
  │
  ▼
Layout tree
  │
  ├─ layoutArrange() ──── sizing pass (Fit/Fixed/Grow axes)
  │
  ▼
renderLayout()
  │
  ▼
[]RenderCmd
  │
  ▼
Backend dispatch loop ── Metal / OpenGL / SDL2 renderer
```

## Contributing

1. Install Go 1.26+ and SDL2 development libraries (see above).
2. Clone the repo.
3. Run tests:

        go test ./...

4. Run static analysis:

        go vet ./...

## License

MIT — see [LICENSE](LICENSE).
