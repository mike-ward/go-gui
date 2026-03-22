# Go-Gui

![Go version](https://img.shields.io/badge/go-1.26%2B-blue)
![License](https://img.shields.io/badge/license-PolyForm%20NC%201.0-blue)
![CI](https://github.com/mike-ward/go-gui/actions/workflows/ci.yml/badge.svg)

Cross-platform immediate-mode GUI framework for Go.

![showcase](assets/showcase.png)

## Overview

Go-Gui is an immediate-mode GUI framework ported from the
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

Go-Gui descends from the V-language gui framework and preserves its design
philosophy: minimal ceremony, composable widgets, and a clear separation
between layout and rendering.

## Features

- 50+ built-in widgets (buttons, inputs, sliders, tables, trees, tabs, …)
- Stateless view model — views are plain functions
- Full animation subsystem with keyframe and spring animations
- GPU-accelerated rendering via SDL2 + Metal/OpenGL shaders
- Web/WASM backend — Canvas2D rendering with custom WebGL shaders,
  runs in any modern browser
- iOS backend — Metal rendering, UIKit windowing, touch events
- Command registry with global hotkeys, shortcut hints, and
  fuzzy-search command palette
- DataGrid with virtualization, sorting, grouping, inline editing,
  CSV/TSV/XLSX/PDF export, and async data sources
- Rich text input — multiline, click-to-cursor, drag-to-select,
  double-click word select, autoscroll, Home/End cycling
- Text selection and copy for read-only Text widgets
- Markdown and RTF views with syntax highlighting and code-block
  copy-to-clipboard
- SVG loading, caching, and tessellation
- Dock layout with drag-and-drop and tab groups
- ColorFilter post-processing — grayscale, sepia, brightness, contrast,
  hue rotate, invert, saturation, and composable filter chains
- ClipContents — stencil-based rounded-rect clip masking for containers
- RotatedBox — quarter-turn rotation for any widget subtree
- Embedded Feather icon font with themed icon text styles
- OS-level spell check for text inputs (macOS NSSpellChecker,
  Linux Hunspell)
- IME support and accessibility tree (macOS, Linux AT-SPI2)
- Multi-window support — open, close, and communicate across
  OS windows via App, Broadcast, and QueueCommand
- Native dialogs, notifications, and print/PDF
- Locale and i18n support
- Theme system with built-in dark/light variants and custom theme support

![gallery](assets/gallery.png)

## Requirements

- Go 1.26+
- SDL2 development libraries (see platform setup below)
- For GL backend: OpenGL 3.3 capable graphics drivers/runtime
- [go-glyph](https://github.com/mike-ward/go-glyph) — auto-fetched via
  `go get`

### Platform Setup

#### macOS (Homebrew)

```bash
brew install go pkg-config sdl2 freetype harfbuzz pango fontconfig
```

#### Ubuntu / Debian

```bash
sudo apt-get update
sudo apt-get install -y \
  golang \
  build-essential \
  pkg-config \
  libsdl2-dev \
  libfreetype6-dev \
  libharfbuzz-dev \
  libpango1.0-dev \
  libfontconfig1-dev
```

#### Fedora / RHEL

```bash
sudo dnf install -y golang gcc pkgconf-pkg-config \
  SDL2-devel freetype-devel harfbuzz-devel pango-devel fontconfig-devel
```

#### Arch Linux

```bash
sudo pacman -Syu --noconfirm go base-devel pkgconf \
  sdl2 freetype2 harfbuzz pango fontconfig
```

#### Windows (MSYS2 MinGW x64)

```bash
pacman -S --needed mingw-w64-x86_64-go mingw-w64-x86_64-gcc \
  mingw-w64-x86_64-pkgconf mingw-w64-x86_64-SDL2 \
  mingw-w64-x86_64-freetype mingw-w64-x86_64-harfbuzz \
  mingw-w64-x86_64-pango mingw-w64-x86_64-fontconfig
```

Then use the `MSYS2 MinGW x64` shell for `go build` / `go run`.

#### Windows (vcpkg toolchain)

```bash
vcpkg install sdl2:x64-windows freetype:x64-windows \
  harfbuzz:x64-windows pango:x64-windows fontconfig:x64-windows
```

Set `CGO_CFLAGS` and `CGO_LDFLAGS` to your vcpkg include/lib paths before building.

![Digital Rain Screenshot](assets/digital-rain.png)

## Installation

    go get github.com/mike-ward/go-gui

## Backend Selection

`backend.Run(w)` auto-selects Metal on macOS and OpenGL elsewhere:

```go
import "github.com/mike-ward/go-gui/gui/backend"

backend.Run(w) // Metal on macOS, GL on Linux/Windows
```

For multi-window applications, use `backend.RunApp`:

```go
app := gui.NewApp()
app.ExitMode = gui.ExitOnMainClose // or ExitOnLastClose (default)

w1 := gui.NewWindow(gui.WindowCfg{State: &Main{}, Title: "Main"})
w2 := gui.NewWindow(gui.WindowCfg{State: &Inspector{}, Title: "Inspector"})

backend.RunApp(app, w1, w2) // manages all windows
```

Open windows at runtime with `app.OpenWindow(cfg)`. Communicate across
windows with `app.Broadcast(fn)` or `other.QueueCommand(fn)`. See
[`examples/multi_window/`](examples/multi_window/) for a working example.

To force a specific backend, import it directly:

```go
import metal "github.com/mike-ward/go-gui/gui/backend/metal" // macOS only
import gl    "github.com/mike-ward/go-gui/gui/backend/gl"    // cross-platform
import sdl2  "github.com/mike-ward/go-gui/gui/backend/sdl2"  // software fallback
import web   "github.com/mike-ward/go-gui/gui/backend/web"   // WASM/browser
import ios   "github.com/mike-ward/go-gui/gui/backend/ios"   // iOS
```

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/mike-ward/go-gui/gui"
    "github.com/mike-ward/go-gui/gui/backend"
)

type App struct{ Clicks int }

func main() {
    gui.SetTheme(gui.ThemeDarkBordered)

    w := gui.NewWindow(gui.WindowCfg{
        State:  &App{},
        Title:  "get_started",
        Width:  300,
        Height: 300,
        OnInit: func(w *gui.Window) { w.UpdateView(mainView) },
    })

    backend.Run(w) // blocks until window closes
}

func mainView(w *gui.Window) gui.View {
    ww, wh := w.WindowSize()
    app := gui.State[App](w)

    return gui.Column(gui.ContainerCfg{
        Width:  float32(ww),
        Height: float32(wh),
        Sizing: gui.FixedFixed,
        HAlign: gui.HAlignCenter,
        VAlign: gui.VAlignMiddle,
        Content: []gui.View{
            gui.Text(gui.TextCfg{
                Text:      "Hello GUI!",
                TextStyle: gui.CurrentTheme().B1,
            }),
            gui.Button(gui.ButtonCfg{
                IDFocus: 1,
                Content: []gui.View{
                    gui.Text(gui.TextCfg{
                        Text: fmt.Sprintf("%d Clicks", app.Clicks),
                    }),
                },
                OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
                    gui.State[App](w).Clicks++
                    e.IsHandled = true
                },
            }),
        },
    })
}
```

See [`examples/get_started/`](examples/get_started/) for the full runnable
version. For the WASM/browser version, see [`examples/web_demo/`](examples/web_demo/).

### More Examples

| Directory                                                  | Description                                 |
| ---------------------------------------------------------- | ------------------------------------------- |
| [`animations`](examples/animations/)                       | Animation subsystem showcase                |
| [`benchmark`](examples/benchmark/)                         | Frame timing and allocation benchmarks      |
| [`blur_demo`](examples/blur_demo/)                         | Blur visual effect                          |
| [`calculator`](examples/calculator/)                       | Styled desktop calculator                   |
| [`color_picker`](examples/color_picker/)                   | Color picker widget                         |
| [`command_demo`](examples/command_demo/)                   | Command registry, hotkeys, command palette  |
| [`context_menu`](examples/context_menu/)                   | Right-click context menus                   |
| [`custom_shader`](examples/custom_shader/)                 | Custom GPU shader rendering                 |
| [`data_grid_data_source`](examples/data_grid_data_source/) | DataGrid with async data source             |
| [`date_picker_options`](examples/date_picker_options/)     | Date picker configurations                  |
| [`dialogs`](examples/dialogs/)                             | Native and custom dialogs                   |
| [`dock_layout`](examples/dock_layout/)                     | IDE-style docking panels with drag-and-drop |
| [`draw_canvas`](examples/draw_canvas/)                     | Custom-draw canvas surface                  |
| [`floating_layout`](examples/floating_layout/)             | Float-anchored overlay positioning          |
| [`gradient_demo`](examples/gradient_demo/)                 | OpenGL gradient rendering                   |
| [`ios_demo`](examples/ios_demo/)                           | iOS demo app (Metal + UIKit)                |
| [`listbox`](examples/listbox/)                             | ListBox widget demo                         |
| [`markdown`](examples/markdown/)                           | Markdown rendering with code-block copy     |
| [`menu_demo`](examples/menu_demo/)                         | Pull-down menu bar                          |
| [`multi_window`](examples/multi_window/)                   | Multi-window with cross-window messaging    |
| [`multiline_input`](examples/multiline_input/)             | Multiline text input                        |
| [`rotated_box`](examples/rotated_box/)                     | Quarter-turn rotation widget                |
| [`rtf`](examples/rtf/)                                     | RTF document viewer                         |
| [`scroll_demo`](examples/scroll_demo/)                     | Scrollable content layouts                  |
| [`shadow_demo`](examples/shadow_demo/)                     | Box shadow effects                          |
| [`showcase`](examples/showcase/)                           | Interactive widget showcase                 |
| [`snake`](examples/snake/)                                 | Snake game                                  |
| [`svg`](examples/svg/)                                     | SVG loading and display                     |
| [`todo`](examples/todo/)                                   | Classic todo app                            |
| [`web_demo`](examples/web_demo/)                           | Browser demo via WASM                       |

## Core Concepts

### Views and Layouts

A _View_ is any value that satisfies the `gui.View` interface — in practice,
a `*gui.Layout` returned by a widget factory such as `gui.Button(...)` or
`gui.Column(...)`.

The View _function_ registered with `w.UpdateView(fn)` is called every frame.
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
| Wrap          | `Wrap(ContainerCfg)`              | Flow-wrap container        |
| Canvas        | `Canvas(ContainerCfg)`            | Absolute-position canvas   |
| Circle        | `Circle(ContainerCfg)`            | Circular container         |
| Rectangle     | `Rectangle(RectangleCfg)`         | Styled rectangle           |
| ExpandPanel   | `ExpandPanel(ExpandPanelCfg)`     | Collapsible section        |
| Splitter      | `Split(SplitterCfg)`              | Resizable two-pane split   |
| DockLayout    | `DockLayout(DockCfg)`             | Drag-and-drop dock areas   |
| OverflowPanel | `OverflowPanel(OverflowPanelCfg)` | Wraps overflowing children |
| RotatedBox    | `RotatedBox(RotatedBoxCfg)`       | Quarter-turn rotation      |

#### Input

| Widget         | Factory                             | Description                  |
| -------------- | ----------------------------------- | ---------------------------- |
| Button         | `Button(ButtonCfg)`                 | Clickable button             |
| Input          | `Input(InputCfg)`                   | Single-line text field       |
| NumericInput   | `NumericInput(NumericInputCfg)`     | Numeric text field           |
| Checkbox       | `Checkbox(CheckboxCfg)`             | Boolean toggle               |
| Radio          | `Radio(RadioCfg)`                   | Mutually-exclusive options   |
| Select         | `Select(SelectCfg)`                 | Dropdown selector            |
| Combobox       | `Combobox(ComboboxCfg)`             | Editable dropdown            |
| Switch         | `Switch(SwitchCfg)`                 | On/off toggle switch         |
| Toggle         | `Toggle(ToggleCfg)`                 | Toggle button                |
| Slider         | `Slider(SliderCfg)`                 | Min/max range picker         |
| InputDate      | `InputDate(InputDateCfg)`           | Date field with picker       |
| ColorPicker    | `ColorPicker(ColorPickerCfg)`       | RGBA color picker            |
| Form           | `Form(FormCfg)`                     | Form with validation         |
| ThemePicker    | `ThemePicker(ThemePickerCfg)`       | Theme switcher               |
| CommandPalette | `CommandPalette(CommandPaletteCfg)` | Fuzzy-search command palette |

#### Display

| Widget      | Factory                       | Description                   |
| ----------- | ----------------------------- | ----------------------------- |
| Text        | `Text(TextCfg)`               | Styled text label             |
| Badge       | `Badge(BadgeCfg)`             | Notification badge            |
| ProgressBar | `ProgressBar(ProgressBarCfg)` | Determinate/indeterminate bar |
| Pulsar      | `Pulsar(PulsarCfg)`           | Blinking cursor indicator     |
| DrawCanvas  | `DrawCanvas(DrawCanvasCfg)`   | Custom-draw surface           |
| Image       | `Image(ImageCfg)`             | Raster image view             |
| Svg         | `Svg(SvgCfg)`                 | SVG vector image view         |
| Markdown    | `w.Markdown(MarkdownCfg)`     | Rendered Markdown             |
| RTF         | `RTF(RtfCfg)`                 | Rendered RTF                  |

#### Data

| Widget           | Factory                                 | Description                     |
| ---------------- | --------------------------------------- | ------------------------------- |
| ListBox          | `ListBox(ListBoxCfg)`                   | Scrollable item list            |
| Table            | `Table(TableCfg)`                       | Static data table               |
| DataGrid         | `w.DataGrid(DataGridCfg)`               | Virtualized grid with full CRUD |
| Tree             | `Tree(TreeCfg)`                         | Hierarchical tree view          |
| DatePicker       | `DatePicker(DatePickerCfg)`             | Calendar date picker            |
| DatePickerRoller | `DatePickerRoller(DatePickerRollerCfg)` | Drum-style date roller          |

#### Navigation

| Widget      | Factory                          | Description                   |
| ----------- | -------------------------------- | ----------------------------- |
| TabControl  | `Tabs(TabControlCfg)`            | Tabbed panel                  |
| Breadcrumb  | `Breadcrumb(BreadcrumbCfg)`      | Navigational breadcrumb trail |
| Menu        | `Menu(MenuCfg)`                  | Pull-down menu bar            |
| Menubar     | `Menubar(w, MenubarCfg)`         | Application menu bar          |
| ContextMenu | `ContextMenu(w, ContextMenuCfg)` | Right-click context menu      |
| Sidebar     | `w.Sidebar(SidebarCfg)`          | Collapsible side navigation   |

#### Overlay

| Widget      | Factory                          | Description            |
| ----------- | -------------------------------- | ---------------------- |
| Dialog      | `w.Dialog(DialogCfg)`            | Modal dialog           |
| Toast       | `w.Toast(ToastCfg)`              | Transient notification |
| WithTooltip | `WithTooltip(w, WithTooltipCfg)` | Hover tooltip wrapper  |

## Backend

Go-Gui separates rendering concerns into injectable interfaces:

| Interface        | Purpose                                    |
| ---------------- | ------------------------------------------ |
| `TextMeasurer`   | Measure glyph extents for layout           |
| `SvgParser`      | Parse and tessellate SVG files             |
| `NativePlatform` | Native dialogs, notifications, print, a11y |

The SDL2 backend (`gui/backend/sdl2`) implements all three and wires itself
into the window on `sdl2.New(w)`. The Web/WASM backend (`gui/backend/web`)
renders via Canvas2D with WebGL custom shaders. The iOS backend
(`gui/backend/ios`) uses Metal rendering and UIKit windowing with touch
event translation. Custom backends implement the interfaces and call the
corresponding `w.Set*` methods.

The headless test backend (`gui/backend/test`) provides a no-op
implementation used by all unit tests.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Application Layer                    │
│      examples/  ──  View fn(w *Window) *Layout          │
│                 gui.State[T](w) typed state slot        │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                  gui/ (core package)                    │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │   Widgets    │  │  State Mgmt  │  │   Animation   │  │
│  │  Button,Text │  │  StateMap    │  │   Subsystem   │  │
│  │  Container…  │  │  per-window  │  │               │  │
│  └──────┬───────┘  └──────┬───────┘  └───────┬───────┘  │
│         │                 │                  │          │
│  ┌──────▼─────────────────▼──────────────────▼───────┐  │
│  │              Layout Engine                        │  │
│  │  GenerateViewLayout() → Layout tree               │  │
│  │  layoutArrange() — Fit/Fixed/Grow sizing          │  │
│  │  renderLayout() → []RenderCmd                     │  │
│  └──────────────────────┬────────────────────────────┘  │
│                         │                               │
│  ┌──────────────────────▼────────────────────────────┐  │
│  │            Event Dispatch                         │  │
│  │  Mouse · Keyboard · Focus · Scroll                │  │
│  └───────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────┘
                         │
        ┌────────────────┼─────────────────┐
        │                │                 │
┌───────▼──────┐ ┌───────▼──────┐ ┌────────▼───────┐
│  TextMeasurer│ │  SvgParser   │ │ NativePlatform │
│  (interface) │ │  (interface) │ │  (interface)   │
└───────┬──────┘ └───────┬──────┘ └────────┬───────┘
        │                │                 │
┌───────▼────────────────▼─────────────────▼───────────┐
│               backend/sdl2/                          │
│  Injects interfaces at startup · Window management   │
├──────────────┬───────────────┬───────────────────────┤
│  backend/    │  backend/gl/  │  backend/filedialog/  │
│  metal/      │  OpenGL       │  backend/printdialog/ │
│  Metal(macOS)│               │                       │
├──────────────┼───────────────┼───────────────────────┤
│  backend/    │  backend/ios/ │  backend/spellcheck/  │
│  web/        │  Metal+UIKit  │  backend/atspi/       │
│  WASM+Canvas │  (iOS)        │  (Linux a11y)         │
└──────────────┴───────────────┴───────────────────────┘
        │
┌───────▼───────┐
│   go-glyph    │
│  Text shaping │
│  rendering    │
│  wrapping     │
└───────────────┘
```

**Pipeline (immediate-mode, no virtual DOM):**

```
View fn → Layout tree → layoutArrange() → renderLayout() → []RenderCmd → GPU
```

**Key types:** `Layout` (tree node), `Shape` (renderable), `RenderCmd` (draw op), `Window` (top-level + state slot)

## Contributing

1.  Install Go 1.26+ and SDL2 development libraries (see above).
2.  Clone the repo.
3.  Run tests:

        go test ./...

4.  Run static analysis:

        go vet ./...

## License

PolyForm Noncommercial 1.0.0 — see [LICENSE](LICENSE).
