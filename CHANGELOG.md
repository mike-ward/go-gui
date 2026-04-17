# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.12.3] - 2026-04-17

### Fixed

- `renderDrawCanvas` now emits images before triangle batches and
  text, so `DrawCanvas` consumers that compose tile backgrounds with
  `DrawContext` overlays get the correct z-order. Previously images
  painted on top of every batch/text in the same canvas — invisible
  in unit tests that only inspect `Texts()`/`Batches()` but user-
  visible once a tile-map demo ran in a window

### Changed

- SDL2 / GL / Metal backends now forward high-resolution
  `MouseWheelEvent.PreciseX` / `PreciseY` for smooth-scroll devices
  (trackpad pixel-scroll, Magic Mouse, high-res wheels), falling
  back to integer `X`/`Y` when the precise field is zero or the SDL
  runtime predates 2.0.18. Enables sub-integer scroll deltas in
  consumers that accumulate fractional `ScrollY`

## [v0.12.2] - 2026-04-16

### Added

- Image download pipeline now handles remote URLs for
  `DrawContext.Image`. Shared `ResolveImageSrc(w, src)` resolves
  http/https URLs to local cache paths, schedules background
  downloads when uncached, and returns "" while in flight.
  `gui.Image` and `emitDrawCanvasImages` both route through it so
  DrawCanvas tiles render after the first fetch
- `WindowCfg.ImageFetcher` hook: apps can supply a custom HTTP
  client to set User-Agent, auth headers, or route through a
  shared pool. Default fetcher sends `User-Agent: go-gui/vX.Y.Z`
  so providers (e.g. OSM) can identify traffic
- `WindowCfg.MaxImageDownloads`: process-wide cap on concurrent
  image downloads. Defaults to 6; first-window-wins for sizing
- Exported `Version` const tracks the module tag

### Fixed

- HTTP status codes are now checked before the body is written to
  disk. Non-200 responses (4xx/5xx) no longer poison the cache
  with error-page payloads

### Changed

- `downloadImage` dropped the HEAD pre-flight and validates
  size/content-type on the GET response. Single round trip per
  fetch

### Performance

- `ResolveImageSrc` caches the URL→path mapping per window so
  already-resolved tiles skip the `MkdirAll` + `Stat` syscalls
  each frame. Critical for DrawCanvas-based tile maps that render
  dozens of images per frame at 60fps

## [v0.12.1] - 2026-04-16

### Added

- DrawCanvas: `DrawContext.Image(x, y, w, h, src, bgOpacity, bgColor)`
  draws images inside the canvas via the same deferred-emit pipeline
  as text. `src` accepts the same forms as `ImageCfg.Src` (local path,
  http/https URL, data URL)
- DrawCanvas: `DrawCanvasCfg.IDFocus` and `OnKeyDown` enable keyboard
  focus and key event handling. A11Y role flips to button when the
  canvas is focusable

## [v0.12.0] - 2026-04-15

### Added

- Time-travel debugging: opt-in via WindowCfg.DebugTimeTravel. User state
  implements Snapshotter (Snapshot/Restore; optional Size). Framework
  captures a snapshot after every dispatched event; scrubber window
  auto-spawns alongside the app window with a slider, step buttons
  (first/prev/next/last), cause label, counter, freeze toggle, and
  keyboard shortcuts (arrows, home/end, space, esc)
- Window.Now() virtual clock: returns pinned snapshot timestamp during
  scrub, live time otherwise; use in view fns that render clock-driven
  data so scrubbed frames match their snapshot
- Window.EnableHistory(maxBytes), HistoryLen(), OpenDebugWindow(),
  Freeze/Resume/IsFrozen, PostRestore(idx) public API
- RegisterNamespaceSnapshot(ns): widget authors opt additional StateMap
  namespaces into scrub restore; scroll (nsScrollX/nsScrollY) and
  widget-local focus (nsInputFocus, nsListBoxFocus, nsTreeFocus) are
  pre-registered
- BoundedMap.cloneAny/restoreAny: type-preserving snapshot through an
  interface so whitelisted namespaces rewind without reflection
- examples/time_travel: counter demo wiring Snapshotter + DebugTimeTravel

### Hardening

- Snapshotter.Size() capped at 1 GiB to prevent totalBytes overflow
- Slider NaN/Inf rejected before int conversion in the scrubber
- BoundedMap restore recovers from type-assertion panics so a single
  out-of-sync namespace doesn't break the rest of the scrub
- Parent-window title truncated before composing the scrubber title

### Notes

- Read-only scrub only: rewinding state does not un-do past side effects
  (HTTP requests, file writes, sounds)
- Requires multi-window mode (App + App.OpenWindow). Single-window apps
  log a notice and no-op
- Zero-cost when disabled: nil-history check short-circuits the hot
  path with no allocation

## [v0.11.0] - 2026-04-14

### Added

- WindowCfg.OnCloseRequest hook: intercept OS window-close and app-quit
  events for save/discard/cancel prompts. Callback owns calling
  Window.Close() to proceed or doing nothing to cancel. Dispatch
  extracted into DispatchCloseRequest / DispatchQuitRequest helpers
  shared by sdl2/gl/metal backends.

## [v0.10.0] - 2026-04-14

### Added

- DockNode/SplitterState JSON serialization: struct tags, text-marshaled
  enums (DockNodeKind, DockSplitDir, SplitterOrientation, SplitterCollapsed),
  DockNodeSanitize for post-unmarshal hardening
- Showcase docs: new dock_layout component entry, splitter serialization section

### Changed

- SplitterStateNormalize handles NaN/Inf ratios and invalid Collapsed values
- Modernize: sync.OnceFunc/OnceValue, slices.SortStableFunc, cmp.Compare

## [v0.9.9] - 2026-04-13

### Fixed

- Metal backend: native Cocoa file-drop bridge bypasses go-sdl2 crash
  (SDL_free on Cocoa-allocated string); per-window callback map for
  multi-window support

## [v0.9.8] - 2026-04-13

### Added

- File-drop event support: OnFileDrop callback on Container, DrawCanvas,
  and EventHandlers; SDL2 backend maps DropEvent to EventFileDropped

### Changed

- Rename EventFilesDropped → EventFileDropped (singular)

## [v0.9.7] - 2026-04-13

### Changed

- Bump go-glyph dependency from v1.6.4 to v1.6.5

## [v0.9.6] - 2026-04-12

### Changed

- Deduplicate helpers across gui/ (asciiLower, f64Clamp, FNV-1a hash,
  skipLayoutChild, shapeBounds, emitClipCmd, cpInputColumn,
  progressBarCenterLabel, finishDiagramFetch, baseCfg)
- Replace `fmt.Sprintf` with `strconv` in hot paths (data grid, inspector,
  a11y, data source)
- Eliminate per-frame heap allocations: gesture Event scratch pool, defer
  removal in render/image opacity, rotateCoordsInverse float path,
  stack-array cellContent, inspector cache map reuse
- Convert copy-paste spinner tests to table-driven
- Remove redundant state and unnecessary comments
- 42 files changed, −278 net lines

## [v0.9.5] - 2026-04-11

### Added

- `Window.FrameCount() uint64` accessor for the monotonic frame
  counter; lets widgets detect repeat callbacks within a render cycle

## [v0.9.4] - 2026-04-11

### Added

- `Window.SetTitle(string)` + `Window.SetTitleFn(func(string))` — dynamic
  OS window title updates. Wired in sdl2, metal, and gl backends via
  `sdl.Window.SetTitle`
- Input hardening on `SetTitle`: 4 KiB cap, UTF-8-safe truncation,
  embedded-NUL stripping; no-alloc fast path for clean short inputs

## [v0.9.3] - 2026-04-10

### Added

- `NativeSaveDiscardDialog` — Save / Don't Save / Cancel alert for
  unsaved-changes flows
- Native menubar: route macOS app-menu "About" through `OnAction`

### Changed

- License: PolyForm NC 1.0 → MIT

### Fixed

- Solitaire example: replace double-click auto-move with right-click
- CI: brew upgrade harfbuzz/pango text stack on macOS; checkout
  go-glyph for test job and use local replace directive

## [v0.9.2] - 2026-04-09

### Added

- `Window.TextMeasurer()` accessor for downstream widgets that need
  direct access to the backend measurer

### Fixed

- Drop `t.Parallel` on tests mutating `guiTheme.ScrollMultiplier`
  (race-avoidance)

## [v0.9.1] - 2026-04-08

### Changed

- Bump `github.com/mike-ward/go-glyph` to v1.6.4
- Bump `golang.org/x/sys` to v0.43.0

## [v0.9.0] - 2026-04-07

### Added

- `gui/highlight` subpackage: chroma-backed syntax highlighter with curated lexer set (go, python, js/ts, rust, c/cpp, java, ruby, shell, html, css, json, yaml, toml, sql, markdown, diff, dockerfile, make) and DoS caps (256KB source, 100k tokens)
- `MarkdownStyle.CodeHighlighter` field: optional highlighter for fenced code blocks; nil preserves parser's built-in tokenizer
- `MarkdownStyle.CodeTypeColor`, `CodeFunctionColor`, `CodeBuiltinColor` palette fields
- Showcase: component docs, welcome, data grid features, markdown demo, and inspector overlay all use `highlight.Default()`

## [v0.8.0] - 2026-04-06

### Added

- `Spinner` widget: animated mathematical curve loading indicator with 21 named `CurveType` constants (rose, lissajous, hypotrochoid, butterfly, cardioid, lemniscate, epitrochoid, heart wave, spiral, fourier and variants)
- Spinner particle-trail rendering via `DrawCanvas` with faint ghost path outline
- Spinner optional slow rotation (`Rotate` field, 30s per revolution)
- Spinner `Opt[float32]` params (ParamA/B/D) for custom curve tuning
- DrawContext: `QuadBezier`, `CubicBezier` drawing primitives
- DrawCanvas: `OnMouseUp` event
- `ClearNamespace` and `ClearDrawCanvasCache` for targeted cache flush
- Mouse button state in motion events; `OnMouseMove` on `DrawCanvas`
- `OnMouseLeave` event and `RequestRedraw()` for tooltip support
- Showcase: Spinner demo with all 21 curves, varied colors, and rotation examples

### Fixed

- Table column auto-sizing; DrawRecorder `Text()` fall-through
- Live resize redraw on Windows (SDL event watcher)
- Mutex safety: defer Unlock, add missing lock in `ClearViewState`
- gofmt alignment in theme_defaults const blocks

### Changed

- Bump go-gl/gl to 2025-03-31 snapshot
- Bump go-glyph v1.6.1 → v1.6.2
- Set default font to Segoe UI on Windows

## [v0.7.0] - 2026-04-02

### Breaking

- `GridPaginationCursor`/`GridPaginationOffset` iota values shifted; new `GridPaginationNone` (0) added
- `Color.Over` returns `ColorTransparent` (set=true) instead of zero `Color` when both inputs are fully transparent
- `executeFocusCallback`/`executeMouseCallback` removed unused debug string parameter

### Fixed

- Race: synchronize `guiTheme` and `Default*Style` globals with `sync.RWMutex`
- Race: `App.Broadcast` no longer holds lock during user callback (deadlock)
- Race: metal a11y buffers protected with mutex
- Race: SDL2 resize event watcher allocates per-callback instead of sharing pointer
- Bug: layout overflow hides visible children when Float/OverDraw interleaved
- Bug: Fill distribution subtracts OverDraw widths never added to parent
- Bug: stencil depth decrement without matching increment at depth 255
- Bug: masked input edits skip undo/redo stack
- Bug: `InputDate.OnSelect` passes nil `*Event` to callback
- Bug: `queueOnValue` missing nil function guard
- Bug: `ColorFromHSVA` produces wrong colors for negative hue
- Bug: data grid OnHover closure captures stale window pointer
- Correctness: `renderImage`/`renderShape` use defer for shape color restore
- Correctness: SVG render checks `rectIntersection` ok before drawing
- Correctness: `render_validate` checks NaN/Inf/nil for gradient, shadow, blur, shader, rotate
- Correctness: `WithColors` borderFocus falls back to theme-level `ColorSelect`
- Correctness: `WithColors` updates SkeletonStyle and InspectorStyle
- Correctness: `AdjustFontSize` clamps each sub-size individually
- Correctness: `SetTheme` syncs `DefaultInspectorStyle`
- Correctness: `ColorFilterCompose` nil-checks inputs
- Correctness: scroll handlers set `IsHandled` and use shape-relative coords
- Correctness: gesture emits rotate `Began` before first `Changed`
- Correctness: `InMemoryDataSource.Capabilities` acquires read lock
- Correctness: `effectivePaginationKind` returns `GridPaginationNone` when unsupported
- Correctness: dock tree nil Root guard
- Correctness: `bounded_map` eviction handles tombstone-only runs
- Fix: variable shadowing in gesture, data_source, data_source_orm, locale_bundle, view_listbox
- Fix: date-dependent nil panic in TestDatePickerSubElementClickFocus
- Fix: wrap bench missing pool reset; raise CI alert threshold to 200%

### Added

- `GridPaginationNone` constant for unsupported pagination
- `WithInspectorStyle` theme builder
- `StrSourceChanged` locale field
- Data grid CRUD source-change detection and toolbar indicator

### Changed

- Replace `intMin`/`intMax` with Go builtin `min`/`max` (33 call sites)
- Replace `fmt.Sprintf` with `strconv` in per-frame data grid/source paths
- `f32IsFinite` uses bit-pattern check instead of float64 round-trip
- `ColorFilter` factories return pointers to package-level singletons
- `Shortcut.String()` pre-allocates buffer
- `contentWidth`/`contentHeight` skip Float and ShapeNone children, matching `spacing()`
- Move test-only helpers from production files to `_test.go` files
- `native_print` uses defer for lock/unlock
- Document animation spring divergence threshold and zero-delay repeat behavior

## [v0.6.0] - 2026-04-01

### Added

- DrawContext: `Text`, `TextWidth`, `FontHeight` for canvas text rendering
- DrawContext: `FilledRoundedRect`, `RoundedRect` for rounded-corner rectangles
- DrawContext: `DashedLine`, `DashedPolyline` for dashed stroke patterns
- DrawContext: `PolylineJoined` for polylines with miter joins at vertices
- DrawContext: `Texts()`, `Batches()` accessors for testing canvas output
- Render pipeline emits `RenderText` commands from `DrawCanvas`
- Showcase: updated draw canvas demo with line chart (joined polyline, dashed grid, text labels) and bar chart (rounded bars, dashed reference line)
