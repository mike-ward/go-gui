# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```
go test ./...          # run all tests (headless, ~0.25s)
go test ./gui/... -run TestFoo  # run single test
go vet ./...           # static analysis
golangci-lint run      # full lint (govet, staticcheck, errcheck, gosimple, unused, gofmt, goimports, revive)
go build ./...         # build all packages
go run ./examples/get_started/  # run the example app (requires SDL2)
```

## Architecture

Immediate-mode pipeline — no virtual DOM, no diffing:

```
View fn → GenerateViewLayout() → Layout tree
  → layoutArrange() (Fit/Fixed/Grow sizing)
  → renderLayout() → []RenderCmd
  → Backend (SDL2 + Metal/OpenGL)
```

### Packages

- `gui/` — core package: all widget factories, layout engine, theme system,
  animation subsystem, event dispatch, state management (~160 .go files)
- `gui/backend/sdl2/` — SDL2 backend; implements `TextMeasurer`, `SvgParser`,
  `NativePlatform`; wires into window via `sdl2.New(w)`
- `gui/backend/metal/` — Metal rendering backend (macOS)
- `gui/backend/gl/` — OpenGL rendering backend
- `gui/backend/filedialog/` — native file dialog support
- `gui/backend/printdialog/` — native print dialog support
- `gui/backend/internal/` — shared backend internals
- `gui/backend/test/` — headless no-op backend used by all unit tests
- `examples/` — 25 example apps (get_started, showcase, calculator, todo,
  snake, markdown, custom_shader, draw_canvas, etc.)

### Core Types

- `Layout` — tree node with `*Shape`, `*Layout` parent, `[]Layout` children
- `Shape` — central renderable: position, size, color, type, events, text, effects
- `RenderCmd` — single draw operation (rect, text, circle, image, SVG, …)
- `Window` — top-level container; holds typed state slot, layout tree, animations
- `View` — interface satisfied by `*Layout`; widget factories return `*Layout`

### State Management

One typed slot per window — no globals, no closures:

```go
w := gui.NewWindow(gui.WindowCfg{State: &App{}})
app := gui.State[App](w)  // type-asserts; panics if wrong type
```

### Sizing

`Sizing` is a combined axis enum: `SizingFit`, `SizingFixed`, `SizingGrow`,
`SizingFitFixed`, `SizingFixedFixed`, `SizingGrowGrow`, `SizingFixedGrow`,
`SizingGrowFixed`. Convenience aliases: `FitFit`, `FixedFixed`, `GrowGrow`, etc.

### Widgets

All widgets accept a `*Cfg` struct (zero-initializable). Event callbacks share
the signature `func(*Layout, *Event, *Window)`. `IDFocus uint32 > 0` opts a
widget into tab-order focus.

### External Dependencies

- `glyph` — text shaping/rendering library; local replace directive
  points to `../go-glyph` (`~/Documents/github/go-glyph`).
  For any text-related work, consult glyph first to check if it already
  provides the needed functionality. Only create new text handling
  routines when glyph does not supply them.

### Injected Interfaces

Backend injects at startup; nil in tests:
- `TextMeasurer` — glyph metrics for layout
- `SvgParser` — SVG parse + tessellate
- `NativePlatform` — native dialogs, notifications, print, a11y, IME, titlebar

### Key Implementation Notes

- `(*Layout).spacing()` counts only visible children (`ShapeType != ShapeNone`,
  `!Float`, `!OverDraw`) — fence-post gap calculation
- Shape text fields live in `Shape.TC` (`*ShapeTextConfig`), not on `Shape`
- `ContainerCfg.Title`/`TitleBG` render a group-box label in the top border
  (floating eraser + text, like HTML fieldset). `TitleBG` should match the
  parent's background color to erase the border behind the title.
- `Children []Layout` holds values; parents are pointers — avoids cycles
- `StateMap` (keyed by namespace constants like `nsOverflow`, `nsSvgCache`) is
  the per-window typed key-value store used by widgets for internal state
- `AmendLayout` hook on shapes runs after sizing to reposition overlay elements
  (color picker circles, splitter handles, etc.) or manage hover state.
  Layout uses absolute coordinates — moving a parent in `AmendLayout` does NOT
  move children. Use the float system (`FloatAnchor`/`FloatTieOff`/`FloatOffset`)
  for positioning elements that have children.
- Event callbacks must set `e.IsHandled = true` when the event is consumed to
  prevent further propagation
