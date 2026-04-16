# Go-Gui Roadmap

Stuff I'm considering. No promises.

| Feature                               | Notes                                                               |
| ------------------------------------- | ------------------------------------------------------------------- |
| ~~Multi-window~~                      | ✅ Implemented — App manages N windows with cross-window messaging  |
| ~~System tray / menubar~~             | ✅ Implemented — native macOS menubar (NSMenu) + system tray        |
| ~~Touch / gesture input~~             | ✅ Implemented — tap, double-tap, long-press, pan, swipe, pinch, rotate |
| Embedded video/audio                  | No media playback widget                                            |
| ~~Skeleton / shimmer loading~~        | ✅ Implemented — Skeleton widget with gradient shimmer animation    |
| ~~Spell check integration~~           | ✅ Implemented — OS-level spell check via NSSpellChecker            |
| ~~Global hotkeys / shortcut manager~~ | ✅ Implemented — Command registry with global/non-global dispatch   |
| Autocomplete / suggestion list        | Text input with dropdown suggestions (Combobox may partially cover) |
| ~~Keyboard shortcut hints~~           | ✅ Implemented — MenuItemCfg.CommandID renders shortcut hints       |
| Native dark/light mode sync           | Auto-switch theme to match OS appearance                            |
| ~~Mobile target spike~~               | ✅ gesture model, safe area, virtual keyboard insets                |
| ~~Web target spike~~                  | ✅ Wasm renderer + browser clipboard/input backends                 |

## Charting / Graphing / Plotting (External Package)

Separate package built on top of `gui`

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
- [x] Canvas image primitive (`DrawContext.Image` for tiles/sprites)
- [x] Canvas keyboard focus (`DrawCanvasCfg.IDFocus` + `OnKeyDown`)
