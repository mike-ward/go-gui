# Nested `<svg>` Support — Implementation Plan

Status: v1 + clip-to-viewport landed. All steps complete. Steps 1–4
(transform composition, viewBox / preserveAspectRatio, depth cap) and
step 5 (synthesized rectangle clip via `vg.ClipPaths`) implemented in
`xml_nested_svg.go` and `xml.go::parseSvgContent`. Step 6 audit:
`findAllByName` already recurses into nested-`<svg>` defs; regression
tests in `xml_nested_svg_test.go` lock the behavior in.

## Problem

`gui/svg/xml.go::parseSvgContent` only recurses into `g`/`a`. Every
other unknown element falls through to the default-ignore branch.
Nested `<svg>` is a standard SVG container (used for embedded
viewports, imported icon fragments, foreignObject-adjacent layouts)
and its descendants are silently dropped today.

Repro:

```xml
<svg viewBox="0 0 100 100">
  <svg x="10" y="10" width="50" height="50" viewBox="0 0 10 10">
    <circle cx="5" cy="5" r="5" fill="red"/>
  </svg>
</svg>
```

The inner circle never renders.

## Goal

Full SVG 1.1 / SVG 2 nested-viewport semantics, not a `<g>`-shortcut:

- `x`, `y`, `width`, `height` attributes on the inner `<svg>`
  establish a viewport rectangle in the parent's coordinate system.
- Inner `viewBox` (if present) maps onto that viewport via the
  `preserveAspectRatio` rules already implemented for the root.
- Coordinates inside the inner `<svg>` are in inner-viewBox space.
- Content outside the inner viewport rectangle is clipped (overflow
  defaults to `hidden` for `<svg>`).

## Architecture

The current pipeline is one cascade walk that emits absolute-space
`VectorPath`s. Nested `<svg>` adds a transform stack and per-viewport
clip:

1. **Transform stack.** Each nested viewport pushes a 3x3 affine
   that maps inner-viewBox space to outer space. Composes with any
   `transform=` on the element. Already half-present: `parseState`
   carries an inherited transform via `ComputedStyle.Transform`.
   Verify it composes correctly when the inner element supplies
   both `viewBox` and `preserveAspectRatio`.
2. **Clip rectangle.** Synthesize an SVG `clipPath` covering the
   inner viewport rectangle in *outer* coordinates and apply it to
   every descendant path. Equivalent to wrapping the subtree in a
   `<g clip-path="url(#__nestedN)">`. Reuse the existing
   `ClipPaths` machinery (`gui/svg/xml_defs.go::parseDefsClipPaths`)
   rather than introducing a parallel clip pipeline.
3. **A11y / titles.** Nested `<svg>` may carry its own `<title>` /
   `<desc>`. Decide whether to merge into root `vg.A11y` or expose
   per-subtree. Recommendation: ignore for v1 (root-only).

## Code Touchpoints

- `gui/svg/xml.go::parseSvgContent` — add `case "svg":` ahead of the
  default branch. Mirror the `case "g"` structure: compute style,
  push ancestors, recurse. Two extras: viewport-transform composition
  and synthesized clip emission.
- `gui/svg/xml.go` — extract the root viewport math from
  `parseSvgWith` (the `viewBox` / `width` / `height` /
  `preserveAspectRatio` block) into a helper that both root and
  nested cases call. Returns viewport rect + inner-to-outer transform.
- `gui/svg/types.go::ComputedStyle` — likely no change; `Transform`
  already accumulates. Confirm `ClipPathID` slots compose with the
  synthesized nested-viewport clip.
- `gui/svg/xml_defs.go` — add a code path that registers a
  rectangle-only clip path under a synthetic id (e.g.
  `__nested_svg_N`) when nested viewports have a non-trivial overflow
  rectangle. Avoid emitting when the nested `<svg>` lacks `width` /
  `height` or its rect already contains all descendants (cheap bbox
  check; skipping the clip saves render-time work).
- `gui/svg/xml_text.go` — should "just work" via the cascade;
  re-verify a `<text>` inside a nested `<svg>` honors the new
  transform.
- New test file: `gui/svg/xml_nested_svg_test.go` covering:
  - inner content present (regression for current bug),
  - viewBox scaling,
  - `preserveAspectRatio` slice/meet behavior,
  - clip-to-viewport (geometry outside nested rect dropped),
  - transform composition with `transform=` on the nested element,
  - depth-cap respected (use existing `maxGroupDepth`).

## Edge Cases

- **Missing `width` / `height`.** Per SVG, defaults to 100% of the
  outer viewport, which we approximate as the parent's
  `width`/`height`. Single source of truth: pass parent viewport
  dims into the nested handler.
- **Negative or NaN dimensions.** Reuse `clampViewBoxDim`.
- **Self-referential / pathologically nested.** Existing
  `maxGroupDepth` check applies; ensure the new branch increments
  depth like `case "g"`.
- **`overflow: visible`.** SVG 2 lets authors disable the clip via
  `overflow:visible` on the inner `<svg>`. Out-of-scope for v1 —
  document and skip.
- **`<symbol>` resolution via `<use>`.** Already expanded by
  `expandUseElements` before this walk; nested-svg from
  symbol-instantiation should fall out for free, but worth a test.
- **Defs inside nested `<svg>`.** The defs pre-pass currently walks
  the whole tree (`collectStyleBlocks` recurses unconditionally).
  Confirm `parseDefsClipPaths` / `parseDefsGradients` /
  `parseDefsFilters` likewise reach nested-svg children. If not,
  promote them or scope ids per-viewport — the latter is more work
  and probably not worth it for v1.

## Test-First Order

1. Write the failing repro test (inner circle missing). ✅
2. Add `case "svg":` recursing as a plain group — circle now renders
   in *outer* coordinates. Several tests fail (positioning, scaling). ✅
3. Add viewport-rect transform composition. Positioning tests pass. ✅
4. Add `viewBox` + `preserveAspectRatio` mapping. Scaling tests pass. ✅
5. Add synthesized clip. Out-of-bounds-content tests pass. ✅
6. Audited; `findAllByName` already recurses into nested-`<svg>`
   defs. Regression tests added; no code change needed. ✅

Coverage in `xml_nested_svg_test.go`: regression (inner content
present), translate-only viewport, viewBox uniform scale, meet
alignment residual, `preserveAspectRatio="none"` independent scales,
SVG2 transform-attr × viewport composition order, default
width/height inherit from parent viewport, depth cap (descendants
past `maxGroupDepth` are dropped without panic), percentage
width/height resolved against parent. Plus clip-to-viewport coverage:
outside-viewport descendants inherit synth clip id, inside-viewport
ditto, empty `<svg/>` skips emission, sibling viewports get distinct
ids, doubly-nested innermost clip wins, defs reachability for
`<clipPath>` / `<linearGradient>` / `<filter>` inside nested svg.

## v1 Limitations

- Author `clip-path=` on the inner `<svg>` element itself is
  overwritten by the synthesized viewport clip. SVG specifies
  intersection composition; not implemented.
- Descendants with their own `clip-path=` override the inherited
  viewport clip on themselves (cascade behavior) — author clip wins,
  no intersection.
- `overflow:visible` opt-out on the inner `<svg>` is not honored
  (already noted under Edge Cases).

## Open Questions

- Should nested `<svg>` participate in `<use>`/`<symbol>` resolution
  loops, or is the existing `expandUseElements` pass sufficient?
- Per-viewport id namespacing for `<defs>`: skip for v1, or reject
  duplicates with a warning?
- Root-only A11y is acceptable, or do consumers expect descendant
  `<title>` text?
