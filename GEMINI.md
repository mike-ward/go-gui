# GEMINI.md - go-gui Project Context

## Project Overview
**go-gui** is a cross-platform, immediate-mode GUI framework for Go. It follows a design philosophy of minimal ceremony and composable widgets, separating layout logic from rendering backends.

- **Primary Language:** Go 1.26+
- **Architecture:** Immediate-mode pipeline (No virtual DOM or diffing).
- **Key Technologies:** SDL2, Metal (macOS), OpenGL, [go-glyph](https://github.com/mike-ward/go-glyph) (text shaping).

## Architecture & Data Flow
The framework operates on a per-frame pipeline:
1. **View Function:** User-provided function returns a tree of `Layout` nodes.
2. **Layout Engine:** `layoutArrange()` computes sizes and positions using Fit/Fixed/Grow axes.
3. **Render Walk:** `renderLayout()` traverses the tree to produce a flat list of `RenderCmd` objects.
4. **Backend Dispatch:** The selected backend (Metal, GL, or SDL2) renders the commands to the screen.

## Core Concepts

### State Management
Window state is stored in a single typed slot on the `Window` object, retrieved via `gui.State[T](w)`. 
- **No Globals:** State is localized to the window.
- **Type Safety:** `State[T]` performs a type assertion; it panics if the type is incorrect.
- **Internal State:** Widgets use a `StateMap` (keyed by namespace constants) for transient internal state like scroll positions or hover effects.

### Layout & Sizing
- **Layout Tree:** Composed of `Layout` nodes, each containing a `Shape` (the central renderable unit).
- **Sizing Axis:** Uses a combined enum (e.g., `SizingFitFixed`, `SizingGrowGrow`) to define how children are sized relative to parents.
- **Spacing:** Calculation ignores invisible or floating shapes.

### Backends & Injected Interfaces
The core logic is backend-agnostic. Backends inject implementations for:
- `TextMeasurer`: Font metrics and glyph layout.
- `SvgParser`: SVG parsing and tessellation.
- `NativePlatform`: Dialogs, notifications, and accessibility.

## Building and Running

### Requirements
- **macOS:** `brew install sdl2 freetype harfbuzz pango fontconfig`
- **Linux:** `libsdl2-dev`, `libfreetype6-dev`, etc. (see README for details)

### Key Commands
- **Build All:** `go build ./...`
- **Run Example:** `go run ./examples/get_started/`
- **Test:** `go test ./...` (Runs headlessly using `gui/backend/test`)
- **Lint:** `golangci-lint run`

## Development Conventions

### Widget Implementation
- **Factories:** Widgets are created via factory functions (e.g., `gui.Button(cfg)`) that return `*gui.Layout`.
- **Configuration:** All widgets accept a `*Cfg` struct.
- **Event Callbacks:** Signature is `func(*Layout, *Event, *Window)`. 
- **Event Consumption:** Callbacks must set `e.IsHandled = true` to stop event propagation.

### Focus & Input
- **IDFocus:** A `uint32 > 0` assigned to a widget to opt it into the tab-order focus system.
- **Text Handling:** Rely on `go-glyph` for complex text operations.

### Styling & Themes
- Themes are global but can be overridden.
- Use `gui.CurrentTheme()` to access standard styles (e.g., `t.B1` for bold text).

## Project Structure
- `gui/`: Core logic, layout engine, and widget definitions.
- `gui/backend/`: Platform-specific rendering implementations (Metal, GL, SDL2).
- `examples/`: Comprehensive library of usage examples (Calculator, Snake, DataGrid, etc.).
- `docs/`: Additional documentation and architecture diagrams.
