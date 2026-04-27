# SVG Completeness Roadmap — v0.13.0 → v0.15.0

## Progress

| Milestone                | Status        | Date       | PR               |
| ------------------------ | ------------- | ---------- | ---------------- |
| v0.13.0 SVG completeness | ✅ code+tests | 2026-04-26 | (pending commit) |
| v0.14.0 CSS phase 2      | ✅ code+tests | 2026-04-26 | (pending commit) |
| v0.15.0 Reuse + polish   | ✅ code+tests | 2026-04-26 | (pending commit) |

**Per-task checklist (v0.13.0):**

- [x] A11y metadata
- [x] radialGradient
- [x] preserveAspectRatio (spec-correct default; `LegacyStretchAspect` flag dropped — no actual stretch was happening; existing default already `xMidYMid meet`)
- [x] ContainsPoint API
- [x] Visual examples: svg_a11y, svg_radial, svg_aspect, svg_hittest
- [x] Docs: README.md, docs/svg-support.md status flips, CHANGELOG.md

**Per-task checklist (v0.14.0):**

- [x] Sibling + attr selectors
- [x] :hover / :focus / :not() (parser + matcher; runtime auto-toggle deferred to v0.15.0)
- [x] var() fallback
- [x] calc()
- [x] Visual examples: svg_css_selectors, svg_css_vars (svg_css_states moved to v0.15.0)
- [x] Docs: README.md, docs/svg-support.md, CHANGELOG.md, per-example READMEs

**Per-task checklist (v0.15.0):**

- [x] `<use>` element (href / xlink:href, x/y → translate, attribute
      pass-through, depth-8 cycle guard, id stripped on clone)
- [x] `<symbol>` element (children inlined when targeted by `<use>`;
      symbol viewBox honor deferred — minor polish, low ROI)
- [x] Gradient spreadMethod (pad / reflect / repeat for both linear
      and radial; subdivision still pad-based for stop boundaries)
- [x] Tessellation tolerance tunable (`SvgCfg.FlatnessTolerance`,
      plumbed via `SvgParseOpts`; default 0 keeps the historic 0.15
      floor and existing fingerprints stable)
- [x] :hover / :focus state plumbing (`SvgCfg.HoveredElementID` /
      `FocusedElementID` → `SvgParseOpts` → cascade `MatchState`;
      cache invalidates per-state). **Automatic mouse-driven hover
      detection on the Svg widget itself remains deferred** — apps
      drive the IDs by hit-testing `TessellatedPath.ContainsPoint`.
      Tracked as a v0.16.0 followup.
- [x] Visual examples: svg_use_symbol, svg_gradient_spread,
      svg_flatness, svg_css_states
- [x] Docs: README.md, docs/svg-support.md, CHANGELOG.md, per-example READMEs

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

> **Path note:** CSS package lives at `gui/svg/css/`, not `gui/css/`.
> Recon (2026-04-26) found existing scaffolding: `Combinator`
> enum (`CombStart`/`CombDescendant`/`CombChild`) at
> `gui/svg/css/types.go:78-84`, `Specificity [3]uint16` triplet at
> `types.go:22-43`, match engine at `gui/svg/css/match.go`.
> `Decl.CustomProp bool` already flags `--foo` declarations
> (`parse.go:223-230`); `var()` resolution is the missing piece.

### 5. Sibling + attribute selectors

- **Combinators:** extend enum at `gui/svg/css/types.go:78-84` with
  `CombAdjacent` (`+`) and `CombGeneralSibling` (`~`).
- **Parser:** `parseComplexSelector()` (`parse.go:456`). Add `+`/`~`
  branches alongside existing `>` handler.
- **Match engine:** `ComplexSelector.Matches()` (`match.go:30`) needs
  prior siblings to evaluate `+`/`~`. Extend signature:
  ```go
  Matches(el ElementInfo, ancestors, siblings []ElementInfo) bool
  ```
  Update `Match()` (`match.go:80`) callers. Sole internal call site:
  `gui/svg/phaseD_test.go`.
- **Attr selectors:** add `Attr` field to `Compound` (`types.go:65-100`):
  ```go
  type AttrSel struct { Name, Value string; Op uint8 }
  Attrs []AttrSel
  ```
  Ops: `=` `~=` `|=` `^=` `$=` `*=`. Parse `[name op value]` token group
  in `parseCompound()` (`parse.go:558`). Specificity: `c.Spec[1]++` per
  attr selector.

### 6. CSS `:hover` / `:focus` / `:not()`

- **Parse pseudo-classes:** extend `parseCompound()` (`parse.go:558`)
  alongside existing `:nth-child` handler. Add `HoverPseudo`,
  `FocusPseudo bool` to `Compound`. Specificity `c.Spec[1]++` each.
- **`:not()`:** scope = single inner compound (no list, no nesting —
  covers 95% real-world; full Selectors L4 list deferred). Field:
  `Not *Compound` on `Compound`. Specificity contributes inner's spec
  (CSS Selectors L4).
- **Match-time state:** add `MatchState struct { Hover, Focus bool }` to
  `ElementInfo` (`types.go`). `Compound.Matches` consults state when
  `HoverPseudo`/`FocusPseudo` set. Cascade callsite in
  `gui/svg/style.go:computeStyle` reads `info.State` straight from
  parser.
- **Runtime mouse-event toggle: DEFERRED to v0.15.0.** Touches the
  `gui/svg` ↔ `gui` ↔ backend interface boundary (state must reach the
  parser via `SvgParseOpts`, with cache-key invalidation per-element).
  Lands cleanly alongside `<use>`/`<symbol>` dynamic-cascade work in
  v0.15.0. v0.14.0 ships parser + matcher so apps can drive state via a
  pre-set `ElementInfo.State` (build-time), with the auto-toggle from
  mouse hover/focus following in v0.15.0.

### 7. `var(--x, fallback)`

- **Value AST:** extend `Decl` (`types.go`) with parsed value:
  ```go
  type ValueExpr struct {
      Kind uint8 // literal | varref | concat
      Text string
      VarName string
      Fallback *ValueExpr
      Parts []ValueExpr // concat
  }
  ValueAST ValueExpr
  ```
  Retain legacy `Value string` for non-`var()` paths to keep ParseFull
  callers green.
- **Tokenize at parse:** extend `joinTokens` path (`parse.go:329-335`).
  When `FunctionToken` Data == "var", split into `VarRef`. Recursive for
  nested `var()` in fallback.
- **Resolve at cascade:** new step in `Match()` (`match.go:80`) — walk
  `ancestors` slice collecting `Decl.CustomProp` declarations into
  `map[string]string`. Resolve `VarRef` against map at emit time. Cycle
  guard depth 8.

### 8. `calc()` (basic)

- **Strategy:** strings throughout — calc emits numeric-string at parse
  time (e.g. `calc(10px + 4px)` → `"14px"`). No `CssLength` typed AST
  refactor (deferred — not blocking).
- **Tokenizer + RPN:** new `gui/svg/css/calc.go`. Supported ops:
  `+ - * /`, units `px`, unitless. Mixed-unit reject (spec-strict —
  `calc(10px + 50%)` errors).
- **Resolution point:** at parse time during `joinTokens`. `FunctionToken`
  with Data == "calc" intercepts, evaluates, substitutes literal.
- **Errors:** parse error logged via existing `ParseOptions` error sink;
  declaration dropped (CSS spec: invalid `calc()` invalidates property).

### v0.14.0 verification

- `go test ./gui/svg/css/...`
- `go test ./gui/svg/...` (cascade integration intact)
- `go test ./gui/...`
- `go vet ./...` + `golangci-lint run ./gui/...`
- Golden cascade fixtures: `gui/svg/css/testdata/*.css` + assertion JSON
  (no testdata dir today — invent lightweight harness, table-driven
  loader extending existing `parse_test.go` style).
- Sibling re-check (recon: zero hits today, but confirm pre-merge):
  ```
  grep -rn "css\.\(Match\|ParseFull\|ParseStylesheet\|ElementInfo\|Compound\|ComplexSelector\)" \
    ../go-glyph ../go-charts ../go-edit ../go-kite
  ```

#### Visual demos

- **`examples/svg_css_selectors/`** — buttons with sibling/attr selectors
  styling adjacent paths (e.g. `[data-state=active]`, `path + circle`).
- **`examples/svg_css_vars/`** — theme switcher via `--primary` token
  with `var(--x, fallback)`; live `calc()` slider for stroke-width.
- (Moved to v0.15.0) **`examples/svg_css_states/`** — needs runtime
  mouse-event auto-toggle, lands alongside the v0.15.0 dynamic-cascade
  work.

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
| v0.14.0   | `ComplexSelector.Matches` adds siblings param | Sole internal call site; no external CSS pkg consumers (recon) |
| v0.14.0   | Additive `Compound`/`ElementInfo`/`Decl` fields | Keyed literals safe; zero-value MatchState neutral |
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
- `gui/svg/css/types.go` — Compound, ComplexSelector, Specificity, Combinator enum
- `gui/svg/css/parse.go` — selector parser, value tokenizer (var/calc)
- `gui/svg/css/match.go` — match engine, cascade
- `gui/svg/css/calc.go` (NEW) — calc() RPN evaluator

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
