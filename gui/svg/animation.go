// Package svg parses and tessellates SVG content.
package svg

import (
	"slices"
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

// parseAnimateElement parses an <animate> element targeting
// opacity (or fill-opacity / stroke-opacity, which scale the same
// rendered alpha channel). Returns the animation and true if valid.
func parseAnimateElement(
	elem string, inherited groupStyle,
) (gui.SvgAnimation, bool) {
	attr, ok := findAttr(elem, "attributeName")
	if !ok {
		return gui.SvgAnimation{}, false
	}
	var target gui.SvgAnimTarget
	switch attr {
	case "opacity":
		target = gui.SvgAnimTargetAll
	case "fill-opacity":
		target = gui.SvgAnimTargetFill
	case "stroke-opacity":
		target = gui.SvgAnimTargetStroke
	default:
		return gui.SvgAnimation{}, false
	}
	valStr, ok := findAttr(elem, "values")
	if !ok || valStr == "" {
		return gui.SvgAnimation{}, false
	}
	vals := parseSemicolonFloats(valStr)
	if len(vals) < 2 {
		return gui.SvgAnimation{}, false
	}
	dur := parseDuration(elem)
	if dur <= 0 {
		return gui.SvgAnimation{}, false
	}
	return gui.SvgAnimation{
		Kind:       gui.SvgAnimOpacity,
		GroupID:    inherited.GroupID,
		Values:     vals,
		KeySplines: parseKeySplinesIfSpline(elem, len(vals)),
		DurSec:     dur,
		BeginSec:   parseBeginLiteral(elem),
		Cycle:      parseRepeatCycle(elem, dur),
		Freeze:     parseFreeze(elem),
		Target:     target,
	}, true
}

// parseAnimateAttributeElement parses an <animate> element
// targeting an animatable primitive attribute (cx, cy, r, x, y,
// width, height, rx, ry). Phase-1 records the animation so shapes
// can be flagged Animated; phase-2 evaluates overrides and
// re-tessellates.
func parseAnimateAttributeElement(
	elem string, inherited groupStyle,
) (gui.SvgAnimation, bool) {
	attr, ok := findAttr(elem, "attributeName")
	if !ok {
		return gui.SvgAnimation{}, false
	}
	name := attrNameFromString(attr)
	if name == gui.SvgAttrNone {
		return gui.SvgAnimation{}, false
	}
	dur := parseDuration(elem)
	if dur <= 0 {
		return gui.SvgAnimation{}, false
	}
	var vals []float32
	if valStr, ok := findAttr(elem, "values"); ok && valStr != "" {
		vals = parseSemicolonFloats(valStr)
	} else {
		fromStr, fromOK := findAttr(elem, "from")
		toStr, toOK := findAttr(elem, "to")
		if fromOK && toOK {
			vals = []float32{parseF32(fromStr), parseF32(toStr)}
		}
	}
	if len(vals) < 2 {
		return gui.SvgAnimation{}, false
	}
	return gui.SvgAnimation{
		Kind:       gui.SvgAnimAttr,
		GroupID:    inherited.GroupID,
		Values:     vals,
		KeySplines: parseKeySplinesIfSpline(elem, len(vals)),
		DurSec:     dur,
		BeginSec:   parseBeginLiteral(elem),
		Cycle:      parseRepeatCycle(elem, dur),
		Freeze:     parseFreeze(elem),
		AttrName:   name,
	}, true
}

func attrNameFromString(s string) gui.SvgAttrName {
	switch s {
	case "cx":
		return gui.SvgAttrCX
	case "cy":
		return gui.SvgAttrCY
	case "r":
		return gui.SvgAttrR
	case "x":
		return gui.SvgAttrX
	case "y":
		return gui.SvgAttrY
	case "width":
		return gui.SvgAttrWidth
	case "height":
		return gui.SvgAttrHeight
	case "rx":
		return gui.SvgAttrRX
	case "ry":
		return gui.SvgAttrRY
	}
	return gui.SvgAttrNone
}

// parseAnimateTransformElement parses an <animateTransform>
// element. Supports type="rotate" (from/to or values), plus
// type="translate" and type="scale" (values form). The
// additive="sum" attribute is not honored in phase 5a —
// animated values replace the base transform. The only corpus
// asset affected is pulse-ring.svg, where the base transform
// is a scale(0) placeholder that the animation fully overrides.
func parseAnimateTransformElement(
	elem string, inherited groupStyle,
) (gui.SvgAnimation, bool) {
	typ, ok := findAttr(elem, "type")
	if !ok {
		return gui.SvgAnimation{}, false
	}
	switch typ {
	case "rotate":
		// fallthrough to original rotate logic below.
	case "translate":
		return parseAnimateTranslateElement(elem, inherited)
	case "scale":
		return parseAnimateScaleElement(elem, inherited)
	default:
		return gui.SvgAnimation{}, false
	}
	dur := parseDuration(elem)
	if dur <= 0 {
		return gui.SvgAnimation{}, false
	}

	if valStr, ok := findAttr(elem, "values"); ok && valStr != "" {
		angles, cx, cy, ok := parseRotateValues(valStr)
		if !ok {
			return gui.SvgAnimation{}, false
		}
		cx, cy = applyInheritedTransformPt(cx, cy, inherited.Transform)
		return gui.SvgAnimation{
			Kind:       gui.SvgAnimRotate,
			GroupID:    inherited.GroupID,
			Values:     angles,
			KeySplines: parseKeySplinesIfSpline(elem, len(angles)),
			CenterX:    cx,
			CenterY:    cy,
			DurSec:     dur,
			BeginSec:   parseBeginLiteral(elem),
			Cycle:      parseRepeatCycle(elem, dur),
			Freeze:     parseFreeze(elem),
		}, true
	}

	fromStr, _ := findAttr(elem, "from")
	toStr, _ := findAttr(elem, "to")
	if fromStr == "" || toStr == "" {
		return gui.SvgAnimation{}, false
	}
	fromParts := parseSpaceFloats(fromStr)
	toParts := parseSpaceFloats(toStr)
	if len(fromParts) < 3 || len(toParts) < 1 {
		return gui.SvgAnimation{}, false
	}
	cx, cy := fromParts[1], fromParts[2]
	cx, cy = applyInheritedTransformPt(cx, cy, inherited.Transform)
	return gui.SvgAnimation{
		Kind:       gui.SvgAnimRotate,
		GroupID:    inherited.GroupID,
		Values:     []float32{fromParts[0], toParts[0]},
		KeySplines: parseKeySplinesIfSpline(elem, 2),
		CenterX:    cx,
		CenterY:    cy,
		DurSec:     dur,
		BeginSec:   parseBeginLiteral(elem),
		Cycle:      parseRepeatCycle(elem, dur),
		Freeze:     parseFreeze(elem),
	}, true
}

// parseAnimateTranslateElement parses <animateTransform
// type="translate"> with values="tx ty;tx ty;..." or from/to.
func parseAnimateTranslateElement(
	elem string, inherited groupStyle,
) (gui.SvgAnimation, bool) {
	return parsePairedAnimateTransform(
		elem, inherited, gui.SvgAnimTranslate)
}

// parseAnimateScaleElement parses <animateTransform type="scale">
// with values="s;s;..." (uniform) or "sx sy;sx sy;..." (non-
// uniform). Uniform entries are normalized to equal sx,sy.
func parseAnimateScaleElement(
	elem string, inherited groupStyle,
) (gui.SvgAnimation, bool) {
	return parsePairedAnimateTransform(
		elem, inherited, gui.SvgAnimScale)
}

// parsePairedAnimateTransform is the shared body for translate
// and scale animateTransform elements. Both produce Values as an
// interleaved [x,y, ...] stream with 2 floats per keyframe.
//
// inherited.Transform is intentionally NOT applied to the pair
// values: translate/scale animateTransform operates in the target
// element's local coordinate space and composes with its inherited
// transform at render time (see emitSvgPathRenderer). Baking the
// ancestor transform into the values here would apply it twice.
// Rotate's CenterX/CenterY are the exception — those are absolute
// SVG-space points used as the pivot, so the ancestor transform
// must be folded in during parse.
func parsePairedAnimateTransform(
	elem string, inherited groupStyle, kind gui.SvgAnimKind,
) (gui.SvgAnimation, bool) {
	dur := parseDuration(elem)
	if dur <= 0 {
		return gui.SvgAnimation{}, false
	}
	var pairs []float32
	if valStr, ok := findAttr(elem, "values"); ok && valStr != "" {
		pairs = parsePairedValues(valStr)
	} else if fromStr, ok := findAttr(elem, "from"); ok {
		toStr, okTo := findAttr(elem, "to")
		if !okTo {
			return gui.SvgAnimation{}, false
		}
		from := parseSpaceFloats(fromStr)
		to := parseSpaceFloats(toStr)
		if len(from) < 1 || len(to) < 1 {
			return gui.SvgAnimation{}, false
		}
		pairs = []float32{
			from[0], pairY(from),
			to[0], pairY(to),
		}
	}
	if len(pairs) < 4 {
		return gui.SvgAnimation{}, false
	}
	return gui.SvgAnimation{
		Kind:       kind,
		GroupID:    inherited.GroupID,
		Values:     pairs,
		KeySplines: parseKeySplinesIfSpline(elem, len(pairs)/2),
		DurSec:     dur,
		BeginSec:   parseBeginLiteral(elem),
		Cycle:      parseRepeatCycle(elem, dur),
		Freeze:     parseFreeze(elem),
	}, true
}

// pairY returns the second component from a parsed space-float
// list. Falls back to the first component (uniform) when only
// one value is present — matches SVG "scale(s)" shorthand.
func pairY(parts []float32) float32 {
	if len(parts) >= 2 {
		return parts[1]
	}
	return parts[0]
}

// parsePairedValues parses a semicolon-separated values= list
// where each entry is "a [b]" (space-separated). Missing second
// component is duplicated (uniform-scale / same-for-y shorthand).
// Returns an interleaved flat slice of 2 floats per entry. Caps
// keyframe count at maxKeyframes.
func parsePairedValues(s string) []float32 {
	parts := strings.Split(s, ";")
	if len(parts) > maxKeyframes {
		parts = parts[:maxKeyframes]
	}
	out := make([]float32, 0, 2*len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		nums := parseSpaceFloats(p)
		if len(nums) == 0 {
			return nil
		}
		x := nums[0]
		y := nums[0]
		if len(nums) >= 2 {
			y = nums[1]
		}
		out = append(out, x, y)
	}
	return out
}

// parseRotateValues parses a semicolon-separated list of rotate
// keyframes like "0 12 12;360 12 12". Each keyframe is "angle
// [cx cy]". Returns angle slice + center from the first keyframe.
// Center must stay constant across keyframes; mismatches are
// accepted but only the first is honored (rare in practice).
func parseRotateValues(s string) ([]float32, float32, float32, bool) {
	parts := strings.Split(s, ";")
	if len(parts) > maxKeyframes {
		parts = parts[:maxKeyframes]
	}
	angles := make([]float32, 0, len(parts))
	var cx, cy float32
	first := true
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		trip := parseSpaceFloats(p)
		if len(trip) == 0 {
			return nil, 0, 0, false
		}
		angles = append(angles, trip[0])
		if first && len(trip) >= 3 {
			cx = trip[1]
			cy = trip[2]
		}
		first = false
	}
	if len(angles) < 2 {
		return nil, 0, 0, false
	}
	return angles, cx, cy, true
}

// parseFreeze reports whether the animation has fill="freeze".
// SMIL fill defaults to "remove"; only "freeze" alters render-time
// behavior in our model.
func parseFreeze(elem string) bool {
	v, ok := findAttr(elem, "fill")
	return ok && v == "freeze"
}

// parseRepeatCycle returns the per-animation cycle period derived
// from repeatCount/repeatDur. repeatCount="indefinite" yields the
// dur (continuous loop). A finite numeric repeatCount yields
// dur*count so the animation re-fires after the full repeat span.
// Returns 0 when the animation should play once (no looping); a
// later resolveBegins pass may still inherit a chain-derived cycle.
// Hostile inputs are clamped: a huge repeatCount is capped at
// maxRepeatCountCycle and the final cycle is never allowed to
// exceed maxCycleSec so downstream comparisons / floor math stay
// finite and bounded.
func parseRepeatCycle(elem string, dur float32) float32 {
	if v, ok := findAttr(elem, "repeatCount"); ok && v != "" {
		if v == "indefinite" {
			return dur
		}
		n := parseF32(v)
		if n > maxRepeatCountCycle {
			n = maxRepeatCountCycle
		}
		if n > 0 {
			return clampCycle(dur * n)
		}
	}
	if v, ok := findAttr(elem, "repeatDur"); ok && v != "" {
		if v == "indefinite" {
			return dur
		}
		t := parseTimeValue(v)
		if t > 0 {
			return clampCycle(t)
		}
	}
	return 0
}

// maxRepeatCountCycle caps repeatCount to bound cycle duration.
// Large finite repeats (e.g. 1e9) are semantically equivalent to
// "indefinite" for any practical session length.
const maxRepeatCountCycle = 1_000_000

// maxCycleSec caps a single cycle period (seconds). Upper bound
// is generous enough for any real asset (hours) while preventing
// +Inf / absurd values from authoring mistakes or hostile SVGs.
const maxCycleSec = float32(3600 * 24)

func clampCycle(v float32) float32 {
	if v <= 0 {
		return 0
	}
	if v > maxCycleSec {
		return maxCycleSec
	}
	return v
}

// parseDuration extracts the "dur" attribute as seconds.
func parseDuration(elem string) float32 {
	s, ok := findAttr(elem, "dur")
	if !ok || s == "" {
		return 0
	}
	return parseTimeValue(s)
}

// parseBeginLiteral returns the first absolute-time entry in a
// semicolon-separated begin list. Syncbase references (entries
// containing ".begin" or ".end") are skipped here and resolved
// in resolveBegins post-pass. Returns 0 when no literal present.
func parseBeginLiteral(elem string) float32 {
	s, ok := findAttr(elem, "begin")
	if !ok || s == "" {
		return 0
	}
	for part := range strings.SplitSeq(s, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, ".begin") ||
			strings.Contains(part, ".end") {
			continue
		}
		return parseTimeValue(part)
	}
	return 0
}

// beginSpec is one activation time for an animation: either an
// absolute offset (targetID=="") or a reference to another
// animation's begin/end plus an offset.
type beginSpec struct {
	targetID string
	isEnd    bool
	offset   float32
}

// parseBeginSpecs parses the "begin" attribute of an <animate>
// element into an ordered spec list. Returns nil when the
// attribute is absent, empty, or contains no syncbase references
// (no post-pass resolution needed). Malformed entries are
// skipped; the caller falls back to parseBeginLiteral. Caps at
// maxKeyframes entries to bound allocation.
func parseBeginSpecs(elem string) []beginSpec {
	s, ok := findAttr(elem, "begin")
	if !ok || s == "" {
		return nil
	}
	if !strings.Contains(s, ".begin") && !strings.Contains(s, ".end") {
		return nil
	}
	parts := strings.Split(s, ";")
	if len(parts) > maxKeyframes {
		parts = parts[:maxKeyframes]
	}
	out := make([]beginSpec, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		sp, ok := parseOneBeginSpec(p)
		if ok {
			out = append(out, sp)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parseOneBeginSpec parses a single begin-list entry. An entry
// is either a time value (literal) or "id.begin[+-]offset" /
// "id.end[+-]offset". Uses LastIndex so ids containing dots
// are preserved intact.
func parseOneBeginSpec(p string) (beginSpec, bool) {
	idxBegin := strings.LastIndex(p, ".begin")
	idxEnd := strings.LastIndex(p, ".end")
	if idxBegin < 0 && idxEnd < 0 {
		return beginSpec{offset: parseTimeValue(p)}, true
	}
	dot := idxBegin
	tokLen := len(".begin")
	isEnd := false
	if idxEnd >= 0 && idxEnd > idxBegin {
		dot = idxEnd
		tokLen = len(".end")
		isEnd = true
	}
	if dot == 0 {
		return beginSpec{}, false
	}
	targetID := strings.TrimSpace(p[:dot])
	if targetID == "" {
		return beginSpec{}, false
	}
	rest := strings.TrimSpace(p[dot+tokLen:])
	var offset float32
	if rest != "" {
		sign := float32(1)
		switch rest[0] {
		case '+':
			rest = rest[1:]
		case '-':
			sign = -1
			rest = rest[1:]
		}
		offset = sign * parseTimeValue(strings.TrimSpace(rest))
	}
	return beginSpec{
		targetID: targetID,
		isEnd:    isEnd,
		offset:   offset,
	}, true
}

// registerAnimation records post-parse bookkeeping for an
// animation just appended to state.animations at position idx:
// self-id → index, plus begin-spec list when syncbase refs are
// present.
func registerAnimation(state *parseState, elem string, idx int) {
	if id, ok := findAttr(elem, "id"); ok && id != "" {
		if state.animIDIndex == nil {
			state.animIDIndex = make(map[string]int)
		}
		state.animIDIndex[id] = idx
	}
	specs := parseBeginSpecs(elem)
	if len(specs) == 0 {
		return
	}
	if state.animBeginSpecs == nil {
		state.animBeginSpecs = make(map[int][]beginSpec)
	}
	state.animBeginSpecs[idx] = specs
}

// resolveBegins walks recorded syncbase specs and writes each
// animation's final BeginSec, plus a per-animation Cycle when the
// begin list defines a chain-restart (multiple resolvable begins
// imply a periodic re-fire). After per-animation resolution, the
// largest derived cycle is propagated to every animation that
// participates in the chain (syncbase begin or BeginSec > 0) so
// freeze-chained sequences re-fire as one global loop. Animations
// with no specs and no repeatCount keep their parse-time defaults.
func resolveBegins(
	anims []gui.SvgAnimation,
	specs map[int][]beginSpec,
	ids map[string]int,
) {
	if len(specs) > 0 {
		resolveBeginsCore(anims, specs, ids)
	}
	propagateGlobalCycle(anims, specs)
}

// resolveBeginsCore resolves first-match BeginSec and derives a
// per-animation Cycle from any second resolvable begin entry. A
// "second begin" indicates the animation re-fires after the first
// activation; the cycle period is its offset from the first.
func resolveBeginsCore(
	anims []gui.SvgAnimation,
	specs map[int][]beginSpec,
	ids map[string]int,
) {
	resolvedFirst := make([]bool, len(anims))
	for i := range anims {
		if _, has := specs[i]; !has {
			resolvedFirst[i] = true
		}
	}
	var resolveFirst func(i int, stack []int) bool
	resolveFirst = func(i int, stack []int) bool {
		if resolvedFirst[i] {
			return true
		}
		if slices.Contains(stack, i) {
			return false
		}
		stack = append(stack, i)
		for _, sp := range specs[i] {
			t, ok := resolveSpec(sp, anims, ids, stack, resolveFirst)
			if !ok {
				continue
			}
			anims[i].BeginSec = t
			resolvedFirst[i] = true
			return true
		}
		resolvedFirst[i] = true
		return false
	}
	for i := range anims {
		if !resolvedFirst[i] {
			resolveFirst(i, nil)
		}
	}
	// Second pass: derive Cycle from a second resolvable begin spec.
	for i, list := range specs {
		if anims[i].Cycle > 0 || len(list) < 2 {
			continue
		}
		seen := false
		var first float32
		for _, sp := range list {
			t, ok := resolveSpec(sp, anims, ids, nil, resolveFirst)
			if !ok {
				continue
			}
			if !seen {
				first = t
				seen = true
				continue
			}
			if t > first {
				anims[i].Cycle = t - first
				break
			}
		}
	}
}

// resolveSpec evaluates a single begin entry to an absolute time.
// stack and recurse may be nil for non-recursive read-only resolves
// (used by the cycle pass after first-pass resolution is complete).
func resolveSpec(
	sp beginSpec, anims []gui.SvgAnimation,
	ids map[string]int, stack []int,
	recurse func(i int, stack []int) bool,
) (float32, bool) {
	if sp.targetID == "" {
		return sp.offset, true
	}
	tgt, ok := ids[sp.targetID]
	if !ok {
		return 0, false
	}
	if recurse != nil && !recurse(tgt, stack) {
		return 0, false
	}
	base := anims[tgt].BeginSec
	if sp.isEnd {
		base += anims[tgt].DurSec
	}
	return base + sp.offset, true
}

// propagateGlobalCycle picks the largest per-animation cycle and
// applies it to chain-participating animations whose own cycle is
// still 0. Chain participation is approximated by "has a syncbase
// begin spec" or "has a non-zero BeginSec" — both indicate the
// animation depends on a chain that must, by design, restart
// periodically. Animations with neither marker (e.g. a one-shot
// fade with no begin and no repeatCount) keep Cycle=0 and play
// once. When no animation has any explicit cycle, this is a no-op.
func propagateGlobalCycle(
	anims []gui.SvgAnimation,
	specs map[int][]beginSpec,
) {
	var global float32
	for i := range anims {
		if anims[i].Cycle > global {
			global = anims[i].Cycle
		}
	}
	if global <= 0 {
		return
	}
	for i := range anims {
		if anims[i].Cycle > 0 {
			continue
		}
		_, hasSpec := specs[i]
		if hasSpec || anims[i].BeginSec > 0 {
			anims[i].Cycle = global
		}
	}
}

// parseTimeValue converts a time string like "1.5s" or "200ms"
// to seconds.
func parseTimeValue(s string) float32 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "ms") {
		return parseF32(s[:len(s)-2]) / 1000
	}
	if strings.HasSuffix(s, "s") {
		return parseF32(s[:len(s)-1])
	}
	// Bare number defaults to seconds per SVG spec.
	return parseF32(s)
}

// parseSemicolonFloats splits a semicolon-separated string into
// float32 values. Caps the result at maxKeyframes entries to
// bound allocation on pathological input.
func parseSemicolonFloats(s string) []float32 {
	parts := strings.Split(s, ";")
	if len(parts) > maxKeyframes {
		parts = parts[:maxKeyframes]
	}
	out := make([]float32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, parseF32(p))
	}
	return out
}

// parseSpaceFloats splits a space-separated string into float32
// values.
func parseSpaceFloats(s string) []float32 {
	fields := strings.Fields(s)
	out := make([]float32, 0, len(fields))
	for _, f := range fields {
		out = append(out, parseF32(f))
	}
	return out
}

// parseKeySplinesIfSpline returns flat 4*(nVals-1) spline control
// points when the element has calcMode="spline" and a matching
// keySplines list. Returns nil otherwise (fast-path linear lerp).
// A segment count mismatch drops splines rather than erroring —
// real-world SVGs sometimes omit the final segment.
func parseKeySplinesIfSpline(elem string, nVals int) []float32 {
	mode, ok := findAttr(elem, "calcMode")
	if !ok || mode != "spline" {
		return nil
	}
	raw, ok := findAttr(elem, "keySplines")
	if !ok || raw == "" {
		return nil
	}
	segs := nVals - 1
	if segs <= 0 || segs > maxKeyframes {
		return nil
	}
	parts := strings.Split(raw, ";")
	if len(parts) > maxKeyframes {
		parts = parts[:maxKeyframes]
	}
	out := make([]float32, 0, 4*segs)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Fields split on comma or whitespace — SVG allows either.
		quads := splitCommaOrSpace(p)
		if len(quads) != 4 {
			return nil
		}
		for _, q := range quads {
			out = append(out, parseF32(q))
		}
	}
	if len(out) != 4*segs {
		return nil
	}
	return out
}

// splitCommaOrSpace splits on runs of commas and/or whitespace.
func splitCommaOrSpace(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
}
