# SVG Support

Reference for authoring custom SVG assets that render correctly in
go-gui. Covers the parser pipeline (`gui/svg`) used by `gui.Svg`,
`gui.SvgSpinner`, and any backend that loads `*.svg`.

Goal: faithful rendering of designer-authored static and animated
SVG. Not a complete SVG 1.1 / 2 implementation. Anything outside the
supported subset either renders as a static first frame or is
silently ignored.

If a feature is needed that this doc lists as unsupported, file an
issue with the source SVG attached.

## Pipeline

```
SVG text
  → xml_tree (DOM)
  → CSS extract (<style>, inline style="")
  → cascade walk → ComputedStyle per VectorPath
  → @keyframes compile + SMIL <animate*> compile → SvgAnimation list
  → tessellate (fill + stroke geometry)
  → render (per-frame: animation eval → transform → triangle list)
```

Parse is allocation-conscious; render is per-frame and lock-free.

## Document Structure

### Root `<svg>`

| Attribute              | Status    | Notes                                                                                                     |
| ---------------------- | --------- | --------------------------------------------------------------------------------------------------------- |
| `viewBox`              | Supported | Lowercase `viewbox` accepted (HTML quirk)                                                                 |
| `width` / `height`     | Supported | Numeric, `%`, length units (px stripped)                                                                  |
| `xmlns`                | Ignored   | Namespaces collapsed; xlink + svg both work                                                               |
| `preserveAspectRatio`  | Supported | All 9 align values + `meet`/`slice`. `none` falls back to `xMidYMid meet` (non-uniform stretch deferred). |
| `version`              | Ignored   | Treated as SVG 1.1                                                                                        |
| `aria-label`           | Supported | Surfaced via `SvgParsed.A11y.AriaLabel`                                                                   |
| `aria-roledescription` | Supported | Surfaced via `SvgParsed.A11y.AriaRoleDesc`                                                                |
| `aria-hidden`          | Supported | `"true"` → `SvgParsed.A11y.AriaHidden`                                                                    |

The viewBox dimension cap is `10000` per axis. Coordinates are
clamped at ±`1000000`. Documents larger than `100000` elements or
`100000` path segments are rejected.

### Shape Elements

| Element              | Status                                                |
| -------------------- | ----------------------------------------------------- |
| `<path>`             | Supported                                             |
| `<rect>`             | Supported (`rx`/`ry` honored)                         |
| `<circle>`           | Supported                                             |
| `<ellipse>`          | Supported                                             |
| `<line>`             | Supported                                             |
| `<polygon>`          | Supported                                             |
| `<polyline>`         | Supported                                             |
| `<g>` / `<a>`        | Group; transforms + style cascade                     |
| `<defs>`             | Holds gradients, clipPaths, filters, paths-by-id      |
| `<text>`             | Supported (see Text below)                            |
| `<tspan>`            | Supported (positioned text runs)                      |
| `<textPath>`         | Supported                                             |
| `<use>`              | Supported (href / xlink:href, x/y, attr override)     |
| `<symbol>`           | Supported (children inlined when targeted by `<use>`) |
| `<image>`            | **Not supported**                                     |
| `<switch>`           | **Not supported**                                     |
| `<foreignObject>`    | **Not supported**                                     |
| `<title>` / `<desc>` | Supported (parsed → `SvgParsed.A11y.Title/Desc`)      |

`<use href="#id">` (or `xlink:href`) is resolved by inlining a
clone of the referenced subtree at parse time. `x`/`y` attributes
on the use site become a `translate(x,y)` transform composed onto
the clone; presentation attrs (`fill`, `class`, `style`, ...) on
the use site cascade into the clone. Cycles are guarded by a
visited-set + depth-8 cap. `<symbol>` targets contribute their
children directly (the symbol wrapper is dropped); `<symbol>`
elements not targeted by any `<use>` render as defs entries (no
output). Symbol-level `viewBox` honoring is a future polish — for
now, layout the symbol's children directly in the symbol's
coordinate space.

### Path `d` Commands

All SVG path commands are supported, both absolute and relative:

| Command   | Notes                                      |
| --------- | ------------------------------------------ |
| `M` / `m` | Move to. Implicit `L` after the first pair |
| `L` / `l` | Line to                                    |
| `H` / `h` | Horizontal line                            |
| `V` / `v` | Vertical line                              |
| `C` / `c` | Cubic Bézier                               |
| `S` / `s` | Smooth cubic                               |
| `Q` / `q` | Quadratic Bézier                           |
| `T` / `t` | Smooth quadratic                           |
| `A` / `a` | Elliptical arc — flattened to cubics       |
| `Z` / `z` | Close subpath. `Z`-then-`M` is handled     |

`fill-rule` of `nonzero` (default) and `evenodd` are both honored by
the scanline tessellator.

## Paint and Stroke

### Colors

`fill`, `stroke`, `stop-color`, `color` accept:

- Named colors (full SVG 1.1 name set, ~150 names)
- Hex `#rgb`, `#rgba`, `#rrggbb`, `#rrggbbaa`
- `rgb(r, g, b)` and `rgba(r, g, b, a)`
- `currentColor` and `inherit`
- `none` (suppresses paint)

**Not supported:** `hsl()`, `hsla()`, `lab()`, `lch()`, `color()`,
`color-mix()`, system colors.

### Gradients

`<linearGradient>` is supported with:

- `gradientUnits="objectBoundingBox"` (default) and `userSpaceOnUse`
- `x1`/`y1`/`x2`/`y2` as numbers or `%`
- `<stop offset stop-color stop-opacity>` (offset honored as `%` or 0–1)
- `currentColor` stops (substituted at render-time tint)

`<radialGradient>` is supported with:

- `cx`/`cy`/`r`/`fx`/`fy` as numbers or `%` (defaults: 50%; `fx`/`fy` default to `cx`/`cy`)
- Same units + stop semantics as linear
- Simplified focal interpolation: parameter `t` is `distance / R`
  clamped to `[0,1]`. The full SVG cone-focused projection (subtle
  edge-falloff difference when `fx`/`fy` differ from `cx`/`cy`) is a
  future polish.

`spreadMethod` is honored on both linear and radial gradients:

- `pad` (default) — values outside [0,1] clamp to first/last stop.
- `reflect` — triangle wave; the gradient mirrors back and forth.
- `repeat` — sawtooth; the gradient wraps as a tile.

Internal stop-boundary subdivision still treats the gradient as
clamped — gradients with many stops will render correctly under
reflect/repeat at the wrap points but with slightly less anti-
aliasing of the discontinuity than pad mode achieves.

**Not supported:** `gradientTransform`, `<pattern>`,
`<meshgradient>`.

### Stroke

| Property            | Supported values                         |
| ------------------- | ---------------------------------------- |
| `stroke-width`      | length (px stripped, percentages parsed) |
| `stroke-linecap`    | `butt` (default), `round`, `square`      |
| `stroke-linejoin`   | `miter` (default), `round`, `bevel`      |
| `stroke-miterlimit` | Honored; default 4                       |
| `stroke-dasharray`  | Comma/space list, percentages converted  |
| `stroke-dashoffset` | Number or percentage                     |
| `stroke-opacity`    | 0.0 – 1.0                                |

`stroke-linejoin: arcs` and `miter-clip` fall back to `miter`.

### Opacity / Visibility

- `opacity`, `fill-opacity`, `stroke-opacity` — all supported,
  composed multiplicatively per spec
- `display: none` — element + descendants suppressed
- `visibility: hidden` — element suppressed (children may still
  render via the cascade if they re-enable it)

## Transforms

`transform` attribute and CSS `transform` property accept:

- `matrix(a b c d e f)`
- `translate(tx [ty])`
- `scale(sx [sy])`
- `rotate(angle [cx cy])`
- `skewX(angle)` / `skewY(angle)`

Multiple functions compose left-to-right per spec. Up to 100
functions per attribute.

`transform-origin` accepts numeric values, percentages, and the
keywords `left`/`center`/`right`/`top`/`bottom`. Resolved against
the element's bounding box at parse time.

**Not supported:** 3D transforms (`rotate3d`, `translate3d`,
`matrix3d`, `perspective`).

## Clipping and Filters

| Feature                                                                                | Status                                                       |
| -------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| `<clipPath>`                                                                           | Supported; one path per element via `clip-path="url(#id)"`   |
| `clipPathUnits`                                                                        | Ignored (treated as `userSpaceOnUse`)                        |
| `<mask>`                                                                               | **Not supported**                                            |
| `<filter>`                                                                             | Only `feGaussianBlur` (single child); `stdDeviation` honored |
| `feOffset`, `feFlood`, `feMerge`, `feColorMatrix`, `feComposite`, `feDropShadow`, etc. | **Not supported**                                            |
| `clip-rule`                                                                            | `evenodd` honored on clip subpaths                           |

Compound filter chains beyond `feGaussianBlur` are dropped (asset
renders without the filter).

## Text

`<text>` and `<tspan>` are tessellated through the glyph engine.

| Attribute     | Status                                    |
| ------------- | ----------------------------------------- |
| `x`, `y`      | Supported (single value; no list form)    |
| `dx`, `dy`    | Supported on tspan                        |
| `text-anchor` | `start` / `middle` / `end`                |
| `font-family` | Supported (resolved against system fonts) |
| `font-size`   | Supported (lengths)                       |
| `font-weight` | Numeric or `bold` (≥600 → bold)           |
| `font-style`  | `italic` honored                          |
| `fill`        | Solid + linearGradient + radialGradient   |
| `<textPath>`  | Supported                                 |

**Not supported:** `rotate` per-glyph, `lengthAdjust`, `textLength`,
`writing-mode`, bidi reordering controls beyond what the glyph
shaper supplies, `font-variant`.

## CSS

CSS lives in `<style>` blocks (any element scope) or inline
`style=""` attributes. Tokenizer is `tdewolff/parse/v2/css`.

### Selectors

| Selector                                                        | Status                                                                                                                                                                          |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Tag (`circle`)                                                  | Yes                                                                                                                                                                             |
| `#id`                                                           | Yes                                                                                                                                                                             |
| `.class`                                                        | Yes                                                                                                                                                                             |
| Compound (`circle.dot#a`)                                       | Yes                                                                                                                                                                             |
| Descendant (` `)                                                | Yes                                                                                                                                                                             |
| Child (`>`)                                                     | Yes                                                                                                                                                                             |
| Group (`,`)                                                     | Yes                                                                                                                                                                             |
| Universal (`*`)                                                 | Yes                                                                                                                                                                             |
| `:nth-child(an+b)`                                              | Yes                                                                                                                                                                             |
| `:root`                                                         | Yes (= `<svg>`)                                                                                                                                                                 |
| Sibling (`+`, `~`)                                              | Yes                                                                                                                                                                             |
| Attribute (`[name]`, `[name=v]`, `~=`, `\|=`, `^=`, `$=`, `*=`) | Yes                                                                                                                                                                             |
| `:hover` / `:focus`                                             | Selector parsed + matched; state driven via `SvgCfg.HoveredElementID` / `FocusedElementID`. Automatic mouse-event detection on the `Svg` widget itself remains deferred (v0.16) |
| `:not(inner)`                                                   | Yes (single-compound; comma-list deferred)                                                                                                                                      |
| `:active` and other pseudo-classes                              | **No**                                                                                                                                                                          |
| Pseudo-elements (`::before`)                                    | **No**                                                                                                                                                                          |

### Cascade

Full SVG-CSS cascade: animations > inline style > author rules
(by `!important`, specificity, source order) > presentation
attributes > inherited > initial.

### Custom Properties

`var(--name)` resolved against the element + ancestor chain.
`var(--x, fallback)` honored when `--x` is undefined; the fallback
itself may contain another `var()` call (recursive resolution
bounded at depth 32).

### `calc()`

Basic arithmetic: `+`, `-`, `*`, `/`, parenthesized subexpressions.
Units `px` and unitless. Mixed-unit operands are rejected per spec
(e.g. `calc(10px + 50%)` invalidates the declaration). Nested
`calc()` and `calc()` inside `var()` fallback are resolved
recursively. Resolution happens at parse time — the literal result
is substituted into the value string.

### `@media`

Only `@media (prefers-reduced-motion: reduce)` matches.
Other media queries never match — rules inside are dropped.

### Transitions

`transition`, `transition-property`, `transition-duration`, etc.
are **not supported**. Use `@keyframes` for animation.

## Animation

Two coexisting pipelines: SMIL and CSS. They animate the same paths
through one shared timeline, so a path can carry both kinds at
once.

### SMIL

| Element              | Status                                                                                                         |
| -------------------- | -------------------------------------------------------------------------------------------------------------- |
| `<animate>`          | Supported on attributes listed below                                                                           |
| `<animateTransform>` | `rotate`, `translate`, `scale` (`from`/`to` or `values`); TRS sandwich composition                             |
| `<animateMotion>`    | Supported with inline `path=` or `<mpath xlink:href>`. `rotate="auto"` honored. `keyPoints`/`keyTimes` honored |
| `<set>`              | Zero-duration discrete change. Default freeze (SMIL spec says remove — override with `fill="remove"`)          |
| `<discard>`          | **Ignored**                                                                                                    |

Animatable attributes via `<animate>`: `cx`, `cy`, `r`, `rx`, `ry`,
`x`, `y`, `width`, `height`, `opacity`, `fill-opacity`,
`stroke-opacity`, `stroke-dasharray`, `stroke-dashoffset`, `fill`,
`stroke`. Other `attributeName` values are ignored.

| Timing feature                             | Status                               |
| ------------------------------------------ | ------------------------------------ |
| `dur`, `begin`, `repeatCount`/`repeatDur`  | Yes                                  |
| `keyTimes`                                 | Yes                                  |
| `keySplines` (cubic-bezier)                | Yes                                  |
| `calcMode` `linear`, `discrete`            | Yes                                  |
| `calcMode` `paced`                         | Falls back to `linear`               |
| Syncbase `begin="other.end"` chains        | Yes                                  |
| `from` / `to` / `by` shorthand             | Yes                                  |
| `additive="sum"`, `accumulate="sum"`       | Yes (rotate is the well-tested case) |
| `min` / `max` dur clamps                   | Yes                                  |
| `restart="never"` / `"whenNotActive"`      | Yes                                  |
| `restart="always"`                         | Default                              |
| Wallclock begin (`begin="wallclock(...)"`) | **No**                               |
| Event-based begin (`begin="click"`, etc.)  | **No**                               |

Maximum 100 animations per document. Maximum 256 keyframes per
element.

### CSS Animation

| Property                    | Supported values                                                                                                            |
| --------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `animation-name`            | Identifier matching `@keyframes`                                                                                            |
| `animation-duration`        | seconds, `ms`                                                                                                               |
| `animation-delay`           | seconds, `ms`, negative allowed                                                                                             |
| `animation-iteration-count` | number, `infinite`                                                                                                          |
| `animation-direction`       | `normal`, `reverse`, `alternate`, `alternate-reverse`                                                                       |
| `animation-fill-mode`       | `none`, `forwards`, `backwards`, `both`                                                                                     |
| `animation-timing-function` | `linear`, `ease`, `ease-in`, `ease-out`, `ease-in-out`, `step-start`, `step-end`, `steps(n[, jump-…])`, `cubic-bezier(...)` |
| `animation` shorthand       | All of the above, comma-separated lists                                                                                     |
| `animation-play-state`      | **Not supported** (always plays)                                                                                            |

Multiple animations on one element are layered: each
property animates independently. `animation: a 1s, b 2s` produces
two records.

### Animatable CSS Properties

`opacity`, `fill-opacity`, `stroke-opacity`, `fill`, `stroke`,
`stroke-dashoffset`, `transform`. Color tweens lerp in sRGB.
`transform` keyframes are decomposed into rotate/translate/scale
sandwiches.

Other property names in `@keyframes` are silently dropped.

## Limits and Failure Modes

| Limit                | Value                   | Behavior on overflow            |
| -------------------- | ----------------------- | ------------------------------- |
| Elements per file    | 100,000                 | Parse aborts                    |
| Path segments        | 100,000                 | Parse aborts                    |
| Group nesting        | 32                      | Deeper groups flattened/dropped |
| Animation count      | 100                     | Extra animations ignored        |
| Keyframes per anim   | 256                     | Extra keyframes dropped         |
| Attribute size       | 1 MB                    | Attribute ignored               |
| Coordinate magnitude | ±1,000,000              | Clamped                         |
| ViewBox dimension    | 10,000                  | Clamped                         |
| Tessellator vertices | 2,048 per ear-clip pass | Pass aborts (CI safety)         |

Anything outside the supported subset of SMIL or CSS animation
renders as the static first frame — the geometry still draws.

## Hit-Testing

`(*TessellatedPath).ContainsPoint(px, py float32) bool` reports
whether a point in viewBox coordinates falls inside the path's
filled region. Even-odd vs nonzero is honored. Use with
`SvgParsed.Paths` after `Tessellate` to drive per-element click /
hover handlers from the host widget. The widget's own `OnClick`
fires for any hit on the SVG bounding box; per-element routing is
the caller's responsibility.

## Authoring Tips

- Use `fill="currentColor"` on monochrome assets so the widget's
  `Color` config tints them at render time.
- Leave `width`/`height` off the root and rely on `viewBox` so the
  layout decides the size.
- For curve-heavy assets, raise `SvgCfg.FlatnessTolerance` (in
  viewBox units) to trade visual fidelity for vertex count. Default
  0 keeps the renderer's built-in 0.15 floor.
- For interactive states, drive `SvgCfg.HoveredElementID` /
  `FocusedElementID` from the host widget (hit-test via
  `ContainsPoint`) and write `:hover` / `:focus` rules in the
  embedded `<style>`. Automatic mouse-event wiring is deferred.
- Prefer CSS custom properties (`--name` + `var(--name, fallback)`)
  with `calc()` to keep theming logic inside the asset rather than
  generating multiple SVG strings host-side.
- Prefer `@keyframes` over `<animate>` when both are equivalent —
  CSS animation has broader timing-function coverage.
- For motion paths, `<animateMotion>` with `rotate="auto"` is the
  only supported way to keep an element tangent to the path.
- Avoid `<filter>` chains beyond a single `feGaussianBlur`.
- Test under `prefers-reduced-motion: reduce` if reduced-motion
  variants matter — that media query _is_ matched.

## Diagnostics

- `go test ./gui/svg/... -run Phase` runs the phased golden tests.
  Each phase covers a feature slice (parser, transform, animation,
  CSS cascade, etc.).
- `gui/svg/testdata/css-spinners` and `gui/assets/svg-spinners` are
  the reference asset corpus — copy patterns from there.
- `examples/showcase` "Graphics → SVG" demo exercises gradient
  `spreadMethod` (pad / reflect / repeat), centered + focal radial
  gradients, and CSS selectors (class, attribute, sibling combinator,
  `:not()`, `var()` + `calc()`). The "Feedback → SVG Spinner" demo
  renders all 106 built-in spinners; both are useful as live
  regression views.
