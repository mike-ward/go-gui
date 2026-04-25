# SVG CSS + @keyframes design

Plan for adding general CSS support (`<style>` blocks, inline `style=""`,
selectors, cascade, custom properties) and CSS Animations (`@keyframes`,
`animation-*`) to the SVG parser. Targets faithful rendering of arbitrary
designer-authored SVG-with-CSS assets, not an authoring API.

## Scope

- **T1 paint/animate/transform** — `fill`, `stroke`, `stroke-width`,
  `stroke-dasharray`, `stroke-dashoffset`, `stroke-linecap/linejoin`,
  `opacity`, `fill-opacity`, `stroke-opacity`, `transform`,
  `transform-origin`. `@keyframes` + `animation-*`.
- **T2 visibility** — `display:none`, `visibility:hidden`.
- **T4 media queries** — `@media (prefers-reduced-motion: reduce)` only.
  Other queries never match → rules drop.
- **T6 custom properties** — `var(--x)` with no fallback chain, no `calc()`.

Out of scope: `transition-*`, `:hover`/`:focus`/`:active`, attribute
selectors, sibling combinators, pseudo-elements, `@media
prefers-color-scheme`, CSS `filter:`/`mask:`/`clip-path:`, `em`/`rem` units.

## Pipeline

1. XML parse → SVG node tree (unchanged).
2. **NEW** Extract `<style>` blocks; collect inline `style=""`. Parse via
   `tdewolff/parse/v2/css` tokenizer. Build rule list +
   `KeyframesDef` table.
3. **NEW** Cascade walk: for each element, gather matching CSS rules,
   presentation attrs, inline style, and parent inherited. Sort by
   (origin, !important, specificity, source order). Resolve `var(--x)`
   from element + ancestor maps. Produce `ComputedStyle` embedded in
   `VectorPath`.
4. **NEW** `@keyframes` compiler: for each element with
   `animation-name: foo`, materialize one `SvgAnimation` record per
   animated property; append to anim list. Color tween → new
   `SvgAnimColor` kind (sRGB lerp). Compound transforms split into
   separate `SvgAnimRotate/Translate/Scale` records.
5. SMIL `parseAnimateElement` (unchanged) — appends to same list.
6. `bakePathOpacity` (rewritten) reads `ComputedStyle`, not raw attrs.
7. Tessellate (unchanged).

`mergeGroupStyle` is replaced by the cascade walk. Phase-0 golden tests
gate the refactor.

## Cascade

Full SVG-CSS cascade per spec:

1. Animations (render-time override)
2. Inline `style=""` — specificity (1,0,0,0)
3. Author rules — specificity (id, class, tag), source-order tiebreak,
   `!important` promotes to higher origin
4. Presentation attrs (`fill="red"`) — specificity 0
5. Inherited from parent
6. Initial value

Specificity = `[3]uint16`. Walk matched rules in (origin, important,
specificity, source-order) order; last write wins per property.

## Selectors

- Tag, `#id`, `.class`, compound (`circle.dot#a`)
- Descendant ` `, child `>`, group `,`
- Universal `*`
- `:nth-child(an+b)`
- `:root` (= the `<svg>` element)
- Skipped: attribute, sibling (`+`/`~`), pseudo-elements

Matcher walks right-to-left against element. Ancestor info via stack
passed down during cascade walk (no parent pointers on nodes).

## @keyframes compile

Two stages:

1. Parse `@keyframes name { 0% {…} 100% {…} }` → `KeyframesDef`
   (intermediate per-property timeline).
2. For each element matched by selector with `animation-name: name`,
   clone-and-populate one `SvgAnimation` per animated property.

Mapping:

| CSS                                 | `SvgAnimation`                     |
| ----------------------------------- | ---------------------------------- |
| `animation-duration`                | `DurSec`                           |
| `animation-iteration-count`         | `Cycle`                            |
| `animation-delay`                   | `BeginSec`                         |
| `animation-timing-function: linear` | `CalcMode=linear`                  |
| `cubic-bezier(...)`                 | `CalcMode=spline` + `KeySplines`   |
| `ease`/`ease-in`/`-out`/`-in-out`   | preset cubic-bezier                |
| `steps(n, …)`                       | `CalcMode=discrete`                |
| `animation-direction: reverse`      | reverse Values+KeyTimes at compile |
| `: alternate`/`alternate-reverse`   | new `Alternate bool` flag          |
| `animation-fill-mode: forwards`     | `Freeze=true`                      |
| `: backwards`/`both`                | new `FillBackwards bool` flag      |
| `animation-play-state`              | out of parser scope                |
| multiple `animation: a 1s, b 2s`    | N records, same path id            |

Color tween: new `SvgAnimColor` kind, packed `uint32` RGBA per stop,
sRGB lerp.

Compound transforms: `transform: rotate(X) translate(Y)` split into
separate records. Engine composes in fixed order — accept lossy ordering
for v1; matrix kind is a follow-up.

## Custom properties

Substituted at parse. Resolution map per element (lazy alloc — most
elements have no vars). Inherited from parent. Undefined `var(--x)`
silently drops the property (spec: invalid-at-computed-value-time →
initial). No fallback chain. No `calc()`.

## transform-origin

Pre-tessellation geometric bbox per shape:

- `<circle>`/`<ellipse>`/`<rect>`/`<line>` — direct from attrs.
- `<path>` — control-poly bbox initially (over-estimate, tighten if
  visual diffs appear).
- `<g>` — union of resolved child bboxes.
- `<use>` — referenced element bbox + local transform.
- `<text>` — defer (needs glyph metrics).

Resolve `transform-origin` keyword/% to numeric `RotCX/RotCY` etc. at
compile time.

Defaults differ by source: CSS = center (50% 50%); SMIL = (0,0).

## prefers-reduced-motion

Source: `NativePlatform` gains `PrefersReducedMotion() bool` via
type-assertion adapter — not a direct interface change. Backends opt in
by adding the method:

```go
if rm, ok := np.(interface{ PrefersReducedMotion() bool }); ok {
    reduced = rm.PrefersReducedMotion()
}
```

Snapshot at parse. SvgCache key extended with the bool. Pref flip
invalidates cache.

## API + backward compat

- `ComputedStyle` — internal package only initially.
- New `SvgAnimation` fields (`Alternate`, `FillBackwards`) — additive
  named fields.
- New `SvgAnimColor` const at end of `SvgAnimKind` iota.
- `NativePlatform` interface unchanged — type-assertion adapter.
- All CSS types in new internal `gui/svg/css/` subpackage.
- Survey siblings (go-glyph, go-charts, go-edit, go-kite) for
  `SvgAnimation` literal usage before Phase D ships.
- `currentColor` continues to flow through to render-time `Shape.Color`
  tint — no separate override API.

## Performance

- Hot path (per-frame `ComputedStyle` access) = **zero allocations**.
  Embed struct in `VectorPath`.
- Cold path (parse/cascade/compile) bounded allocation; pre-allocate
  ancestor stack reused across siblings.
- Custom prop map lazy-init only when element defines vars.
- Per-phase benchmarks with `testing.AllocsPerRun(_, _) == 0`
  assertions on hot paths.

## Diagnostics

Deferred. Originally planned `SvgDiag` struct + append-only
`Diagnostics []SvgDiag` on `VectorGraphic` (capped at 100). Not
implemented through phases A-G; all malformed-CSS paths drop silently
via cascade fallthrough today. Revisit when a real consumer needs
authoring feedback.

## Phasing

| Phase | Work                                                                                                                | Days |
| ----- | ------------------------------------------------------------------------------------------------------------------- | ---- |
| 0     | Golden hash-pixel tests for 15 existing SMIL spinners                                                               | 1    |
| A     | `ComputedStyle` refactor; replace `mergeGroupStyle`                                                                 | done |
| B     | tdewolff dep; `<style>` + inline `style=""`; tag/id/class; paint props                                              | done |
| C     | Cascade + selectors complete; custom props (T6)                                                                     | done |
| D     | `@keyframes` compile; `SvgAnimColor`; `Alternate`/`FillBackwards`; animation-\* sub-props; compound transform split | done |
| E     | Pre-tess bbox; `transform-origin` resolution                                                                        | done |
| F     | T2 (display/visibility) + T4 (reduced-motion via NativePlatform); SvgCache key extension                            | done |
| G     | 10-20 CSS spinner golden fixtures; mixed SMIL+CSS                                                                   | done |

Total ~16-22 days. Each phase independently mergeable. No feature flag.

## Testing

- **Unit:** tokenizer wrap, selector parse, selector match against
  synthetic trees, cascade order, custom prop substitution, @keyframes
  compile (each animation-\* sub-prop), color tween endpoints,
  transform-origin per shape kind.
- **Golden:** hash pixel buffer; PNG dumped to tmp on diff. Diff
  threshold ~1% per pixel acceptable.
- **Time control:** deterministic playback at t=0, t=duration/4,
  t=duration/2, etc.
- **CI cost:** ear-clip cap (commit ce75d44) constrains. Cap golden
  assets at <2048 verts. Run in parallel.

## Open follow-ups

1. Color-scheme `@media` — separate ticket if real assets need it.
2. Compound transform matrix kind (`SvgAnimMatrix`) — current split-
   per-function compose is lossy; matrix would fix ordering.
3. Bezier tight bbox vs control-poly bbox — tighten only if visual
   diffs appear in real assets.
4. `animation: a, b` shorthand — only first name retained today
   (`cssAnimSpec` is single-name). Multi-anim needs spec-list.
5. `translateX` / `translateY` — silently dropped (compile loop matches
   `translate` / `rotate` / `scale` only).
6. `SvgDiag` diagnostics — deferred; see Diagnostics section.
