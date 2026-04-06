# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
