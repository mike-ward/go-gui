# SVG Completeness Roadmap — v0.13.0 → v0.15.0

## Progress

| Milestone                | Status        | Date       | PR               |
| ------------------------ | ------------- | ---------- | ---------------- |
| v0.13.0 SVG completeness | ✅ code+tests | 2026-04-26 | (pending commit) |
| v0.14.0 CSS phase 2      | ☐ pending     | —          | —                |
| v0.15.0 Reuse + polish   | ☐ pending     | —          | —                |

**Per-task checklist (v0.13.0):**

- [x] A11y metadata
- [x] radialGradient
- [x] preserveAspectRatio (spec-correct default; `LegacyStretchAspect` flag dropped — no actual stretch was happening; existing default already `xMidYMid meet`)
- [x] ContainsPoint API
- [x] Visual examples: svg_a11y, svg_radial, svg_aspect, svg_hittest
- [x] Docs: README.md, docs/svg-support.md status flips, CHANGELOG.md

**Per-task checklist (v0.14.0):**

- [ ] Sibling + attr selectors
- [ ] :hover / :focus / :not()
- [ ] var() fallback
- [ ] calc()
- [ ] Visual examples: svg_css_selectors, svg_css_states, svg_css_vars
- [ ] Docs updates

**Per-task checklist (v0.15.0):**

- [ ] `<use>` element
- [ ] `<symbol>` element
- [ ] Gradient spreadMethod
- [ ] Tessellation tolerance tunable
- [ ] Visual examples: svg_use_symbol, svg_gradient_spread, svg_flatness
- [ ] Docs updates

After each milestone lands:

1. Flip status to ✅ done with date + PR link.
2. Tick per-task checkboxes.
3. Strike-through / remove tasks rolled into combined PRs.
4. Note any deferred carry-over to next milestone.

Commit roadmap update **in the same PR** that completes the milestone so
status + code stay in lockstep.

---

## Context

Survey of `gui/svg/` (parser, tessellator, renderer) + `gui/render_svg.go`
identified 33 spec/coverage gaps. This roadmap addresses the highest-leverage
subset across three milestones, in recommended order. Mask, complex filters,
foreignObject, switch, SMIL event/wallclock begin, sub-pixel AA, CSS
transition, animation-play-state are DEFERRED — out of scope for icon/spinner
use case, low ROI for effort.

Goal: bring SVG support from "icon-render-good" to "interactive + accessible

- spec-aware" for non-exotic SVGs. Unblocks hit-test (charts, clickable
  icons), screen-reader stories, and Material/Tailwind icon sets that lean on
  radial gradients and aspect-correct viewBoxes.

Sibling repos (go-glyph, go-charts, go-edit, go-kite) depend on gui SVG API
— every milestone scans positional struct literals + behavior breaks.

---

## v0.13.0 — "SVG completeness"

### 1. A11y metadata

- Extend `SvgParsed` (`gui/svg_parser.go:4`) with nested:
  ```go
  type SvgA11y struct {
      Title, Desc, AriaLabel, AriaRoleDesc string
      AriaHidden bool
  }
  ```
  Add `A11y SvgA11y` field. Nested keeps top-level struct surface tidy.
- Parse in `parseSvgWith()` (`gui/svg/xml.go`, after viewBox block ~line 70).
  Reuse `findAllByName(root, "title", &n)` + `"desc"` (existing helper at
  `gui/svg/xml_defs.go:31`). Take first non-defs match. Read `aria-label`,
  `role`, `aria-hidden` from `root.AttrMap`. `strings.TrimSpace` text content.
- Test: `gui/svg/xml_a11y_test.go` + fixtures `gui/svg/testdata/a11y_*.svg`.
  Fingerprint stability: zero-default for old SVGs, no render path consumes
  metadata yet → no golden churn.

### 2. radialGradient

- Scaffold ALREADY EXISTS — `SvgGradientDef` (`gui/svg_types.go:121-128`)
  has `CX, CY, R, FX, FY, IsRadial`. Just wire parser + renderer.
- Parser: extend `parseDefsGradients()` (`gui/svg/xml_defs.go:65-109`).
  After linear loop, `findAllByName(root, "radialGradient", &rnodes)`.
  Per node: parse `cx,cy,r` (default 50%), `fx,fy` (default cx,cy),
  `gradientUnits`, stops. Set `IsRadial=true`.
- Renderer: branch `resolveGradient()` (`gui/svg/tessellate.go:1065`) on
  `g.IsRadial`. New helper `projectOntoRadial(vx, vy, g) float32` returning
  `t = clamp(dist((vx,vy), focal) / R-effective, 0, 1)`. Reuse existing
  `interpolateGradient()` (`tessellate.go:1111`) for stops. Simplified
  linear focal→edge param; full cone-focused param noted as future polish.
- Test: `gui/svg/testdata/radial_*.svg` golden fingerprints.

### 3. preserveAspectRatio (spec-correct default)

- Parse: `gui/svg/xml.go:30-50`. New helper
  `parsePreserveAspectRatio(s string) (align uint8, slice bool)`. Default
  `xMidYMid meet`.
- Store on `VectorGraphic` + mirror to `SvgParsed`:
  ```go
  PreserveAlign uint8 // 0..8 = xMin/Mid/Max × yMin/Mid/Max; 9 = none
  PreserveSlice bool  // false=meet, true=slice
  ```
- Apply in `renderSvg()` (`gui/render_svg.go:35-42`) where viewBox `sx,sy`
  translate already lives. Compute `(scaleX, scaleY)` + alignment offset
  `(dx, dy)`; compose into existing translate. Reuse alignment math
  patterns from `gui/alignment.go`.
- No behavior break observed: prior renderer already centered with
  uniform scale (`xMidYMid meet`-equivalent). The `LegacyStretchAspect`
  escape hatch was dropped during implementation.
- Pre-merge: scan siblings for visual breakage:
  ```
  grep -rn "gui.SvgParsed{" ../go-glyph ../go-charts ../go-edit ../go-kite
  ```
  Open issues in each repo if visual deltas land.
- Test: `gui/render_svg_aspect_test.go` golden snapshots — meet/slice ×
  9 align.

### 4. ContainsPoint API

- Extend `TessellatedPath` (`gui/svg_types.go:9-24`):
  ```go
  MinX, MinY, MaxX, MaxY float32
  ```
  Populate during tessellation (one extra pass over `Triangles`).
- New file `gui/svg_hittest.go`:
  ```go
  func (tp *TessellatedPath) ContainsPoint(px, py float32) bool
  ```
  Inline barycentric (8 lines, mirrors internal `pointInTriangle` at
  `gui/svg/tessellate.go:1038-1060`). Fast-reject via bbox; iterate
  `Triangles` 6-floats stride. Tiny duplication keeps `gui` pkg from
  gaining `gui/svg` dep.
- Critical: invert `BaseTransX/Y/Scale/RotAngle` on `(px,py)` first when
  `HasBaseXform` set.
- Test: `gui/svg_hittest_test.go` — circle/rect/star primitives, edge
  cases on triangle borders.

### v0.13.0 verification

- `go test ./gui/...`
- `go test ./gui/svg/...`
- Sibling check: `cd ../go-glyph && go test ./...` (repeat go-charts,
  go-edit, go-kite).
- Breaking-literal scan:
  ```
  grep -rn "SvgParsed{" /Users/mikeward/Documents/github/go-gui/
  grep -rn "TessellatedPath{" /Users/mikeward/Documents/github/go-gui/
  ```

#### Visual demos (NEW examples in `examples/`)

- **`examples/svg_a11y/`** — loads SVG with `<title>`/`<desc>`/`aria-label`,
  prints metadata to overlay text widget. Confirms parser wired.
- **`examples/svg_radial/`** — grid of radial gradients: centered,
  off-center focal, varying R, multi-stop. Side-by-side with linear
  equivalents.
- **`examples/svg_aspect/`** — 3×3 grid of same SVG rendered at varying
  canvas aspect ratios; toggle button cycles `xMinYMin meet` →
  `xMidYMid meet` → `xMaxYMax meet` → `slice` variants → `none`.
- **`examples/svg_hittest/`** — load complex SVG, mouse-hover paints
  containing path, click logs `PathID`. Confirms `ContainsPoint` against
  curves/holes.
- Existing `examples/svg/` + `examples/svg_spinners/` re-run for
  regression.

---

## v0.14.0 — "CSS phase 2"

### 5. Sibling + attribute selectors

- File: `gui/css/parse.go`. Extend selector AST with `Combinator` enum
  (descendant/child/`+`/`~`) and `Attr{Name, Op, Value}`.
- Match engine: locate via `grep -n "func.*[Mm]atch" gui/css/`. Add
  attr-op handling (`=`, `~=`, `|=`, `^=`, `$=`, `*=`).
- Specificity: attr selectors → bump `b` count (a=0, b+=1).

### 6. CSS `:hover` / `:focus` / `:not()`

- Parse-time: `:not()` is a negation around inner selector — parser only.
- Runtime: `:hover` / `:focus` need state. Add `MatchState{Hover, Focus
bool}` per element. Hook `RecomputeStyleForElement(id, state)` in
  cascade.
- Re-tessellation cost concern: scope this milestone to **style-delta
  path** (color, opacity, stroke-width) — full retess only when
  geometry-affecting attrs change. Mirror existing SMIL-style update
  path.

### 7. `var(--x, fallback)`

- Tokenizer extension in css value parser. Resolve at cascade time —
  custom-prop lookup walks ancestor chain. Add `CustomProps
map[string]string` per cascade node.

### 8. `calc()` (basic)

- Tokenizer + RPN. Supported: `+ - * /`, `px`, unitless. Returns
  `CssLength`. Reject mixed units (spec-strict).

### v0.14.0 verification

- `go test ./gui/css/...`
- Golden cascade fixtures: `gui/css/testdata/*.css` + assertion JSON.
- Sibling re-check: `grep -rn "css.Selector" /Users/mikeward/`.

#### Visual demos

- **`examples/svg_css_selectors/`** — buttons with sibling/attr selectors
  styling adjacent paths (e.g. `[data-state=active]`, `path + circle`).
- **`examples/svg_css_states/`** — hover/focus driven color + stroke
  changes on icons; demonstrates style-delta path performance.
- **`examples/svg_css_vars/`** — theme switcher via `--primary` token
  with `var(--x, fallback)`; live `calc()` slider for stroke-width.

---

## v0.15.0 — "Reuse + polish"

### 9. `<use href="#id">`

- `gui/svg/xml.go` — build `idIndex map[string]*xmlNode` during initial
  walk. On `<use>`: lookup target, clone subtree, apply attribute
  override + `x,y` translate, splice at use position.
- Recursion guard: cap depth 8, cycle-detect via visited set.
- Reuse existing path/primitive parsers on cloned subtree.

### 10. `<symbol>`

- Treat as defs entry; parse on demand via `<use>` resolution. Honor
  symbol's own viewBox + preserveAspectRatio (uses v0.13.0 work).

### 11. Gradient `spreadMethod`

- `gui/svg/xml_defs.go:~110` parse `spreadMethod`. Additive field on
  `SvgGradientDef`. `interpolateGradient()` (`tessellate.go:1111`) wraps
  `t`: `reflect` = triangle wave, `repeat` = `mod 1`, default `pad` =
  clamp.

### 12. Tessellation tolerance tunable

- Locate `SvgCfg`: `grep -rn "type SvgCfg" gui/`. Add `FlatnessTolerance
float32`. Replace const in `gui/svg/tessellate.go`. Default unchanged
  → fingerprint stable.

### v0.15.0 verification

- `go test ./gui/svg/...`
- Goldens: `testdata/use_*.svg`, `symbol_*.svg`, `spread_*.svg`,
  `flatness_*.svg`. Verify legacy fixtures fingerprint-stable.

#### Visual demos

- **`examples/svg_use_symbol/`** — single `<symbol>` defs block,
  repeated via `<use>` with overrides (color, transform). Side-by-side
  with manually duplicated equivalent.
- **`examples/svg_gradient_spread/`** — same gradient with
  `pad`/`reflect`/`repeat` spread modes, animated stop offsets.
- **`examples/svg_flatness/`** — slider for `FlatnessTolerance`;
  visualize triangle count + visible faceting trade-off on Bezier-heavy
  path (e.g., complex logo).

---

## Cross-cutting

### Breaking-change matrix

| Milestone | Behavior break          | Mitigation                                          |
| --------- | ----------------------- | --------------------------------------------------- |
| v0.13.0   | Additive struct fields  | Keyed literals safe; positional = caller-side fix   |
| v0.14.0   | CSS specificity shift   | None expected unless siblings rely on cascade order |
| v0.15.0   | None expected           | —                                                   |

### Sibling impact scan (run before each release)

```
grep -rn "gui.SvgParsed{"        ../go-glyph ../go-charts ../go-edit ../go-kite
grep -rn "gui.TessellatedPath{"  ../go-glyph ../go-charts ../go-edit ../go-kite
grep -rn "gui.SvgGradientDef{"   ../go-glyph ../go-charts ../go-edit ../go-kite
```

### Critical files

- `gui/svg_parser.go` — `SvgParsed` struct, public API surface
- `gui/svg_types.go` — `TessellatedPath`, `SvgGradientDef`
- `gui/svg/xml.go` — root parser entry, viewBox + new aspect/a11y
- `gui/svg/xml_defs.go` — gradient parser entry
- `gui/svg/tessellate.go` — `resolveGradient`, `pointInTriangle`,
  `interpolateGradient`, flatness const
- `gui/render_svg.go` — viewBox→canvas transform compose point
- `gui/svg_hittest.go` (NEW) — ContainsPoint method
- `gui/css/parse.go` — selector AST extensions

### Documentation updates (per milestone)

- **`README.md`** — feature list, supported elements summary, new examples.
- **`docs/svg-support.md`** — flip status from "unsupported" to
  "supported" for: title/desc/aria (line 66), radialGradient (line 115),
  preserveAspectRatio (line 38), `<use>`/`<symbol>` (lines 61-62),
  spreadMethod, sibling/attr selectors (lines 213-214), `:hover`/`:focus`/
  `:not()` (line 215), `var()` fallback, `calc()`. Add caveats / known
  limitations (e.g. radial focal simplified; calc unit-strict).
- **`CHANGELOG.md`** — entry per milestone with breaking-change
  call-out (preserveAspectRatio default in v0.13.0,
  `LegacyStretchAspect` opt-out).
- Per-example **`README.md`** in each new `examples/svg_*/` dir
  describing what it demonstrates and how to run.

### Out of scope (DEFERRED)

- `<mask>`, advanced filters (feOffset/feColorMatrix/feComposite/
  feDropShadow)
- `<foreignObject>`, `<switch>`, `requiredFeatures`/`systemLanguage`
- SMIL event begin (`begin="click"`), wallclock begin
- CSS `transition`, `animation-play-state`
- Sub-pixel anti-aliasing (backend job, not SVG layer)
- HSL/LAB color functions (parser polish, low priority)
