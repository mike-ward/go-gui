# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```
go test ./...          # run all tests (headless, ~0.25s)
go test ./gui/... -run TestFoo  # run single test
go vet ./...           # static analysis
golangci-lint run ./gui/...  # full lint (govet, staticcheck, errcheck, gosimple, unused, gofmt, goimports, revive)
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

## Coding Conventions

- **No variable shadowing.** Never use `:=` to redeclare a variable that
  already exists in an outer scope. Use `=` to assign to the existing
  variable, or choose a distinct name.
- **Clean lint and format.** All committed code must pass `golangci-lint run ./...`
  and `gofmt` with zero issues. Run both before proposing a commit.

# context-mode — MANDATORY routing rules

You have context-mode MCP tools available. These rules are NOT optional — they protect your context window from flooding. A single unrouted command can dump 56 KB into context and waste the entire session.

## BLOCKED commands — do NOT attempt these

### curl / wget — BLOCKED
Any Bash command containing `curl` or `wget` is intercepted and replaced with an error message. Do NOT retry.
Instead use:
- `ctx_fetch_and_index(url, source)` to fetch and index web pages
- `ctx_execute(language: "javascript", code: "const r = await fetch(...)")` to run HTTP calls in sandbox

### Inline HTTP — BLOCKED
Any Bash command containing `fetch('http`, `requests.get(`, `requests.post(`, `http.get(`, or `http.request(` is intercepted and replaced with an error message. Do NOT retry with Bash.
Instead use:
- `ctx_execute(language, code)` to run HTTP calls in sandbox — only stdout enters context

### WebFetch — BLOCKED
WebFetch calls are denied entirely. The URL is extracted and you are told to use `ctx_fetch_and_index` instead.
Instead use:
- `ctx_fetch_and_index(url, source)` then `ctx_search(queries)` to query the indexed content

## REDIRECTED tools — use sandbox equivalents

### Bash (>20 lines output)
Bash is ONLY for: `git`, `mkdir`, `rm`, `mv`, `cd`, `ls`, `npm install`, `pip install`, and other short-output commands.
For everything else, use:
- `ctx_batch_execute(commands, queries)` — run multiple commands + search in ONE call
- `ctx_execute(language: "shell", code: "...")` — run in sandbox, only stdout enters context

### Read (for analysis)
If you are reading a file to **Edit** it → Read is correct (Edit needs content in context).
If you are reading to **analyze, explore, or summarize** → use `ctx_execute_file(path, language, code)` instead. Only your printed summary enters context. The raw file content stays in the sandbox.

### Grep (large results)
Grep results can flood context. Use `ctx_execute(language: "shell", code: "grep ...")` to run searches in sandbox. Only your printed summary enters context.

## Tool selection hierarchy

1. **GATHER**: `ctx_batch_execute(commands, queries)` — Primary tool. Runs all commands, auto-indexes output, returns search results. ONE call replaces 30+ individual calls.
2. **FOLLOW-UP**: `ctx_search(queries: ["q1", "q2", ...])` — Query indexed content. Pass ALL questions as array in ONE call.
3. **PROCESSING**: `ctx_execute(language, code)` | `ctx_execute_file(path, language, code)` — Sandbox execution. Only stdout enters context.
4. **WEB**: `ctx_fetch_and_index(url, source)` then `ctx_search(queries)` — Fetch, chunk, index, query. Raw HTML never enters context.
5. **INDEX**: `ctx_index(content, source)` — Store content in FTS5 knowledge base for later search.

## Subagent routing

When spawning subagents (Agent/Task tool), the routing block is automatically injected into their prompt. Bash-type subagents are upgraded to general-purpose so they have access to MCP tools. You do NOT need to manually instruct subagents about context-mode.

## Output constraints

- Keep responses under 500 words.
- Write artifacts (code, configs, PRDs) to FILES — never return them as inline text. Return only: file path + 1-line description.
- When indexing content, use descriptive source labels so others can `ctx_search(source: "label")` later.

## ctx commands

| Command | Action |
|---------|--------|
| `ctx stats` | Call the `ctx_stats` MCP tool and display the full output verbatim |
| `ctx doctor` | Call the `ctx_doctor` MCP tool, run the returned shell command, display as checklist |
| `ctx upgrade` | Call the `ctx_upgrade` MCP tool, run the returned shell command, display as checklist |
