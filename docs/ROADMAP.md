# Go-Gui Roadmap

Stuff I'm considering. No promises.

| Feature                               | Notes                                                               |
| ------------------------------------- | ------------------------------------------------------------------- |
| **Multi-window**                      | Single window only. Qt/Flutter/SwiftUI all support N windows        |
| **Charts**                            | No bar/line/pie/scatter/area. DrawCanvas exists but no chart API    |
| **System tray / menubar**             | No tray icon, no native menu bar integration                        |
| **Touch / gesture input**             | No pinch, swipe, rotate, long-press. Required for tablet/mobile     |
| **Embedded video/audio**              | No media playback widget                                            |
| **Skeleton / shimmer loading**        | Placeholder UI during async loads                                   |
| ~~**Spell check integration**~~       | ✅ Implemented — OS-level spell check via NSSpellChecker            |
| **Global hotkeys / shortcut manager** | Centralized keybinding registry                                     |
| **Constraint-based layout**           | Auto-layout style constraints (complementary to flex)               |
| **Autocomplete / suggestion list**    | Text input with dropdown suggestions (Combobox may partially cover) |
| **Keyboard shortcut hints**           | Show accelerator keys in menus/tooltips                             |
| **Native dark/light mode sync**       | Auto-switch theme to match OS appearance                            |
| Mobile target spike                   | gesture model, safe area, virtual keyboard insets                   |
| Web target spike                      | Wasm renderer + browser clipboard/input backends                    |

## Charting / Graphing / Plotting (External Package)

Separate package built on top of `gui`. Requires a canvas View in
the framework — a View that exposes a draw callback with direct
access to the GPU drawing primitives within a clipped layout region.

### Framework Prerequisites

- [x] Canvas View: layout node with `on_draw` callback providing polyline, filled-polygon, and arc primitives
- [x] Retained geometry buffer in canvas (avoid re-tessellation when only transform/pan/zoom changes)
- [x] Text measurement API (`get_text_width`, `line_height`)
- [x] Text rotation (`TextStyle.rotation_radians`)
- [x] Rectangular clipping (`clip: true` on containers)
- [x] Mouse events (hover, click, scroll, mouse_lock for drag)
- [x] Cursor control (crosshair, pointer, resize)
- [x] Floating overlays / tooltips
- [x] Gradient fills on shapes
- [x] Animation stack (tween, spring, keyframe)
- [x] Custom fragment shaders

### Chart Types (P0)

- [ ] Line chart (polyline, markers, multiple series)
- [ ] Bar chart (vertical/horizontal, grouped, stacked)
- [ ] Area chart (filled polyline, stacked)
- [ ] Pie / donut chart (arc segments, labels, explode)
- [ ] Scatter plot (point clouds, bubble variant)

### Chart Types (P1)

- [ ] Candlestick / OHLC (financial)
- [ ] Gauge / radial progress
- [ ] Heatmap (grid cells, color scale)
- [ ] Radar / spider chart
- [ ] Histogram (bin computation, density overlay)

### Axes + Scales

- [ ] Linear, logarithmic, time, and category scales
- [ ] Auto tick generation with label collision avoidance
- [ ] Axis title, grid lines, minor grid lines
- [ ] Multi-axis support (dual Y)
- [ ] Locale-aware number/date formatting on tick labels

### Chart Interaction

- [ ] Hover tooltip with nearest-point snapping
- [ ] Crosshair / guideline on hover
- [ ] Click-to-select data point / series
- [ ] Zoom (scroll wheel + drag-to-zoom box)
- [ ] Pan (mouse_lock drag)
- [ ] Legend toggle (show/hide series)

### Animation + Transitions

- [ ] Animated data entry (bars grow, lines draw-on)
- [ ] Smooth data update transitions (morph old → new)
- [ ] Series add/remove animation

### Data Model

- [ ] Typed series: `[]f64`, `[]TimeValue`, `[]XY`
- [ ] Lazy / streaming data provider interface
- [ ] Auto domain/range from data with optional overrides

### Theming + Style

- [ ] Inherit `gui` theme colors (foreground, background, accent)
- [ ] Configurable color palettes per chart
- [ ] Consistent text styles with framework `TextStyle`
