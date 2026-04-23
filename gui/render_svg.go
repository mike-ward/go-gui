package gui

import (
	"cmp"
	"log"
	"math"
	"slices"
	"time"
)

// renderSvg renders an SVG shape by loading cached tessellation
// and emitting RenderSvg commands.
func renderSvg(shape *Shape, clip DrawClip, w *Window) {
	if !rectsOverlap(shapeBounds(shape), clip) {
		return
	}

	cached, err := w.LoadSvg(shape.Resource, shape.Width, shape.Height)
	if err != nil {
		log.Printf("renderSvg: %v", err)
		emitErrorPlaceholder(shape.X, shape.Y,
			shape.Width, shape.Height, w)
		return
	}

	color := shape.Color
	if shape.Disabled {
		color = dimAlpha(color)
	}

	// Center SVG content within container (aspect-preserving
	// scale may leave unused space in one dimension).
	sx := shape.X + (shape.Width-cached.Width*cached.Scale)/2
	sy := shape.Y + (shape.Height-cached.Height*cached.Scale)/2

	// Clip to intersection of parent clip and scaled viewBox.
	svgClip, ok := rectIntersection(clip, DrawClip{
		X:      sx,
		Y:      sy,
		Width:  cached.Width * cached.Scale,
		Height: cached.Height * cached.Scale,
	})
	if !ok {
		return
	}
	emitClipCmd(svgClip, w)

	// Compute animation state for SMIL animations.
	var animState map[string]svgAnimState
	if cached.HasAnimations && cached.AnimStartNs != 0 {
		animState = w.scratch.svgAnimStates.take(len(cached.Animations))
		defer w.scratch.svgAnimStates.put(animState)
		nowNs := time.Now().UnixNano()
		// Keep animation alive while SVG is being rendered.
		if cached.AnimHash != "" {
			animSeen := StateMap[string, int64](
				w, nsSvgAnimSeen, capImageCache)
			animSeen.Set(cached.AnimHash, nowNs)
		}
		elapsed := float32(nowNs-cached.AnimStartNs) /
			float32(time.Second)
		contribScratch := w.scratch.svgAnimContribs.take(
			len(cached.Animations))
		animState = computeSvgAnimationsReuse(
			cached.Animations, elapsed, animState, contribScratch)
		w.scratch.svgAnimContribs.put(contribScratch)
	}

	// Substitute fresh triangles for animated primitive shapes when
	// attribute overrides are live. Falls back to cached triangles
	// when no parser support, no spans, or empty overrides.
	renderPaths := cached.RenderPaths
	var animTris []TessellatedPath
	if cached.HasAttrAnim && len(cached.AnimatedSpans) > 0 {
		overrides := extractAttrOverrides(w, animState)
		if len(overrides) > 0 {
			if ap, ok := w.svgParser.(AnimatedSvgParser); ok {
				reuse := w.scratch.svgAnimTriangles.take(0)
				animTris = ap.TessellateAnimated(
					cached.Parsed, cached.Scale, overrides, reuse)
				defer w.scratch.svgAnimTriangles.put(animTris)
			}
			w.scratch.svgAnimOverrides.put(overrides)
		}
		if len(animTris) > 0 {
			renderPaths = buildEffectiveRenderPaths(w, cached, animTris)
			defer w.scratch.effectiveSvgPaths.put(renderPaths)
		}
	}

	// Emit main paths, text, and textPath elements.
	emitSvgGroup(renderPaths, cached.TextDraws,
		cached.TextPathDraws, color, sx, sy,
		cached.Scale, animState, w)

	// Emit filtered groups.
	for i, fg := range cached.FilteredGroups {
		emitRenderer(RenderCmd{
			Kind:       RenderFilterBegin,
			GroupIdx:   i,
			X:          sx,
			Y:          sy,
			W:          fg.BBox[2] * cached.Scale,
			H:          fg.BBox[3] * cached.Scale,
			Scale:      cached.Scale,
			BlurRadius: fg.Filter.StdDev * cached.Scale,
			Layers:     fg.Filter.BlurLayers,
		}, w)
		emitSvgGroup(fg.RenderPaths, fg.TextDraws,
			fg.TextPathDraws, color, sx, sy,
			cached.Scale, animState, w)
		emitRenderer(RenderCmd{
			Kind: RenderFilterEnd,
		}, w)

		// KeepSource: re-draw sharp original on top of blur.
		if fg.Filter.KeepSource {
			emitSvgGroup(fg.RenderPaths, fg.TextDraws,
				fg.TextPathDraws, color, sx, sy,
				cached.Scale, animState, w)
		}
	}

	// Restore parent clip.
	emitClipCmd(clip, w)
}

// emitSvgGroup emits paths, text draws, and text path draws.
func emitSvgGroup(
	paths []CachedSvgPath, textDraws []CachedSvgTextDraw,
	textPathDraws []CachedSvgTextPathDraw,
	color Color, sx, sy, scale float32,
	animState map[string]svgAnimState, w *Window,
) {
	for _, path := range paths {
		emitSvgPathRenderer(path, color, sx, sy, scale, animState, w)
	}
	for i := range textDraws {
		emitCachedSvgTextDraw(&textDraws[i], sx, sy, w)
	}
	for i := range textPathDraws {
		emitCachedSvgTextPathDraw(&textPathDraws[i], sx, sy, w)
	}
}

// emitSvgPathRenderer emits a single SVG path as a RenderSvg
// command. If tint has alpha>0 and path has no vertex colors,
// the tint overrides the path color; the path's own alpha is
// modulated in so per-element opacity (baked into path.Color.A
// during parsing) survives the override. animState applies SMIL
// rotation/opacity per GroupID.
func emitSvgPathRenderer(path CachedSvgPath, tint Color,
	x, y, scale float32,
	animState map[string]svgAnimState, w *Window) {
	hasVCols := len(path.VertexColors) > 0
	c := path.Color
	if tint.A > 0 && !hasVCols {
		c = tint
		c.A = blendAlpha(tint.A, path.Color.A)
	}
	var vcols []Color
	if hasVCols {
		if tint.A == 0 {
			vcols = path.VertexColors
		} else {
			// Tint active on a gradient path: replace each vertex
			// RGB with tint RGB while modulating its alpha so the
			// gradient's alpha shape (e.g. fade-in tail of tail-
			// spin) survives. vcols is allocated from a frame-
			// scoped arena so repeated renders of tinted gradients
			// avoid a per-frame heap allocation.
			vcols = w.scratch.takeVColors(len(path.VertexColors))
			for i, vc := range path.VertexColors {
				t := tint
				t.A = blendAlpha(tint.A, vc.A)
				vcols[i] = t
			}
			c = tint
		}
	}

	var rotAngle, rotCX, rotCY float32
	var transX, transY, scaleX, scaleY float32
	hasXform := false
	var vAlphaScale float32
	hasVAlpha := false
	if animState != nil && path.GroupID != "" {
		if st, ok := animState[path.GroupID]; ok {
			rotAngle = st.RotAngle
			rotCX = st.RotCX
			rotCY = st.RotCY
			if st.HasXform {
				transX = st.TransX
				transY = st.TransY
				scaleX = st.ScaleX
				scaleY = st.ScaleY
				hasXform = true
			}
			opa := st.Opacity
			if path.IsStroke {
				opa *= st.StrokeOpacity
			} else {
				opa *= st.FillOpacity
			}
			// Clamp to [0,1] so hostile or out-of-range animation
			// values cannot drive a negative or oversized alpha
			// through the uint8 cast (undefined conversion).
			opa = clampFrac(opa)
			if opa < 1 {
				c.A = uint8(float32(c.A) * opa)
				if len(vcols) > 0 {
					vAlphaScale = opa
					hasVAlpha = true
				}
			}
		}
	}

	emitRenderer(RenderCmd{
		Kind:             RenderSvg,
		Triangles:        path.Triangles,
		Color:            c,
		VertexColors:     vcols,
		VertexAlphaScale: vAlphaScale,
		HasVertexAlpha:   hasVAlpha,
		X:                x,
		Y:                y,
		Scale:            scale,
		IsClipMask:       path.IsClipMask,
		ClipGroup:        path.ClipGroup,
		RotAngle:         rotAngle,
		RotCX:            rotCX,
		RotCY:            rotCY,
		TransX:           transX,
		TransY:           transY,
		ScaleX:           scaleX,
		ScaleY:           scaleY,
		HasXform:         hasXform,
	}, w)
}

// emitCachedSvgTextDraw emits a cached SVG text draw as a
// RenderText command. Takes pointer into CachedSvg.TextDraws
// slice so TextStylePtr remains stable.
func emitCachedSvgTextDraw(draw *CachedSvgTextDraw,
	shapeX, shapeY float32, w *Window) {
	emitRenderer(RenderCmd{
		Kind:         RenderText,
		Text:         draw.Text,
		X:            shapeX + draw.X,
		Y:            shapeY + draw.Y,
		Color:        draw.TextStyle.Color,
		FontName:     draw.TextStyle.Family,
		FontSize:     draw.TextStyle.Size,
		TextWidth:    draw.TextWidth,
		TextStylePtr: &draw.TextStyle,
		TextGradient: draw.Gradient,
	}, w)
}

func emitCachedSvgTextPathDraw(draw *CachedSvgTextPathDraw,
	shapeX, shapeY float32, w *Window) {
	emitRenderer(RenderCmd{
		Kind:         RenderTextPath,
		Text:         draw.Text,
		X:            shapeX,
		Y:            shapeY,
		TextStylePtr: &draw.TextStyle,
		TextPath:     &draw.Path,
	}, w)
}

// svgAnimState holds computed per-group animation state.
type svgAnimState struct {
	RotAngle float32 // rotation degrees
	RotCX    float32 // rotation center X (SVG space)
	RotCY    float32 // rotation center Y (SVG space)
	Opacity  float32 // 0..1; <animate attributeName="opacity">
	// FillOpacity / StrokeOpacity track the per-paint opacity
	// animations. They scale only the matching path role at render
	// time so a fill-opacity animation does not dim the stroke.
	FillOpacity   float32
	StrokeOpacity float32
	// TransX/TransY is the animated translate; ScaleX/ScaleY the
	// animated scale. Identity when HasXform is false.
	TransX, TransY float32
	ScaleX, ScaleY float32
	HasXform       bool
	Inited         bool
	AttrOverride   SvgAnimAttrOverride // attribute overrides for re-tessellation
}

// computeSvgAnimations builds a map of per-group animation state
// from parsed SMIL animations and elapsed time. Implements SMIL
// "sandwich" semantics: each animation's last activation time is
// computed (BeginSec + n*Cycle for the largest n with that <=
// elapsed); contributions are sorted by activation ascending and
// applied last-write-wins per attribute. fill="freeze" lets a past
// animation continue to contribute its last keyframe value until
// its cycle restarts or a later-activated animation overrides.
func computeSvgAnimations(
	anims []SvgAnimation, elapsedSec float32,
	states map[string]svgAnimState,
) map[string]svgAnimState {
	return computeSvgAnimationsReuse(anims, elapsedSec, states, nil)
}

// computeSvgAnimationsReuse is the render-path variant that
// accepts a scratch []animContrib to avoid per-frame allocation.
func computeSvgAnimationsReuse(
	anims []SvgAnimation, elapsedSec float32,
	states map[string]svgAnimState,
	contribScratch []animContrib,
) map[string]svgAnimState {
	if states == nil {
		states = make(map[string]svgAnimState, len(anims))
	} else {
		clear(states)
	}
	contribs := collectAnimContribs(anims, elapsedSec, contribScratch)
	if len(contribs) == 0 {
		return states
	}
	slices.SortStableFunc(contribs, cmpAnimContrib)
	for i := range contribs {
		applyAnimContrib(&contribs[i], states)
	}
	return states
}

// cmpAnimContrib orders contributions by ascending activation
// time so sandwich priority applies last-write-wins.
func cmpAnimContrib(a, b animContrib) int {
	return cmp.Compare(a.activation, b.activation)
}

// animContrib is one animation's evaluated contribution at the
// current frame: the resolved value(s), the last activation time
// used for sandwich-priority ordering, and a back-pointer to the
// animation for kind/group dispatch.
type animContrib struct {
	anim       *SvgAnimation
	value      float32
	valueX     float32
	valueY     float32
	activation float32
}

// collectAnimContribs evaluates every animation's phase against
// elapsed and returns the subset that contributes this frame
// (active or frozen). Skipped animations: missing GroupID/dur,
// not-yet-activated (elapsed < BeginSec), past dur with !Freeze.
// reuse may be a scratch slice whose backing array will be
// reused in place; pass nil to allocate a fresh slice.
func collectAnimContribs(
	anims []SvgAnimation, elapsedSec float32,
	reuse []animContrib,
) []animContrib {
	out := reuse[:0]
	if cap(out) < len(anims) {
		out = make([]animContrib, 0, len(anims))
	}
	for i := range anims {
		a := &anims[i]
		// Reject non-finite timing fields so downstream lerp / floor
		// math cannot produce NaN values that would propagate into
		// render state. DurSec must be strictly positive for normal
		// animations; <set> is zero-duration and bypasses this check.
		if a.GroupID == "" ||
			!finiteF32(a.DurSec) ||
			(!a.IsSet && a.DurSec <= 0) ||
			!finiteF32(a.BeginSec) ||
			!finiteF32(a.Cycle) ||
			!finiteF32(elapsedSec) ||
			elapsedSec < a.BeginSec {
			continue
		}
		activation := a.BeginSec
		if a.Cycle > 0 && a.Restart != SvgAnimRestartNever {
			n := math.Floor(float64(elapsedSec-a.BeginSec) / float64(a.Cycle))
			if a.Restart == SvgAnimRestartWhenNotActive && n > 0 {
				prev := a.BeginSec + float32(n-1)*a.Cycle
				if elapsedSec-prev < a.DurSec {
					// Previous activation still within its active
					// duration — suppress the re-trigger.
					n--
				}
			}
			activation = a.BeginSec + float32(n)*a.Cycle
		}
		var frac float32
		if a.IsSet {
			// Zero-duration: always contribute to-value once active.
			// Freeze=false is a parse flag only — the value still
			// contributes, but sandwich ordering lets later activations
			// replace it (matching SMIL instantaneous semantics).
			frac = 1
		} else {
			phase := elapsedSec - activation
			switch {
			case phase < a.DurSec:
				frac = phase / a.DurSec
			case a.Freeze:
				frac = 1
			default:
				continue
			}
		}
		c := animContrib{anim: a, activation: activation}
		switch a.Kind {
		case SvgAnimRotate, SvgAnimOpacity, SvgAnimAttr:
			if len(a.Values) < 2 {
				continue
			}
			c.value = lerpKeyframes(
				a.Values, a.KeySplines, a.KeyTimes, a.CalcMode, frac)
			if a.Accumulate {
				c.value += accumOffset(a, activation) *
					(a.Values[len(a.Values)-1] - a.Values[0])
			}
		case SvgAnimTranslate, SvgAnimScale:
			if len(a.Values) < 4 {
				continue
			}
			c.valueX, c.valueY = lerpKeyframes2D(
				a.Values, a.KeySplines, a.KeyTimes, a.CalcMode, frac)
			if a.Accumulate {
				n := accumOffset(a, activation)
				last := len(a.Values) - 2
				c.valueX += n * (a.Values[last] - a.Values[0])
				c.valueY += n * (a.Values[last+1] - a.Values[1])
			}
		case SvgAnimMotion:
			if len(a.MotionPath) < 4 || len(a.MotionLengths) < 2 {
				continue
			}
			c.valueX, c.valueY, c.value = motionSample(a, frac)
		}
		out = append(out, c)
	}
	return out
}

// motionSample interpolates along an animateMotion's flattened path
// by arc length. frac ∈ [0,1] scales to [0, totalLen]; returns the
// (x,y) point and — when MotionRotate==auto — the tangent angle in
// degrees. Returns zeros when lens/poly are inconsistent or the
// total length is non-finite — the caller treats that as the
// identity contribution.
func motionSample(a *SvgAnimation, frac float32) (float32, float32, float32) {
	lens := a.MotionLengths
	poly := a.MotionPath
	n := len(lens)
	if n < 2 || len(poly) < 2*n {
		return 0, 0, 0
	}
	total := lens[n-1]
	if !finiteF32(total) || total < 0 {
		return 0, 0, 0
	}
	target := clampFrac(frac) * total
	idx := 0
	for i := 1; i < n; i++ {
		if lens[i] >= target {
			idx = i - 1
			break
		}
		idx = i
	}
	if idx >= n-1 {
		idx = n - 2
	}
	span := lens[idx+1] - lens[idx]
	var t float32
	if span > 0 {
		t = (target - lens[idx]) / span
	}
	x0, y0 := poly[idx*2], poly[idx*2+1]
	x1, y1 := poly[(idx+1)*2], poly[(idx+1)*2+1]
	x := x0 + (x1-x0)*t
	y := y0 + (y1-y0)*t
	var angle float32
	if a.MotionRotate != SvgAnimMotionRotateNone {
		dx := x1 - x0
		dy := y1 - y0
		angle = float32(math.Atan2(float64(dy), float64(dx))) *
			(180 / math.Pi)
		if a.MotionRotate == SvgAnimMotionRotateAutoReverse {
			angle += 180
		}
	}
	return x, y, angle
}

// accumOffset returns the repeat-count offset for accumulate=sum.
// The count is floor((activation - BeginSec) / Cycle) — the number
// of completed prior cycles. Cycle must be >0.
func accumOffset(a *SvgAnimation, activation float32) float32 {
	if a.Cycle <= 0 {
		return 0
	}
	return float32(math.Floor(
		float64(activation-a.BeginSec) / float64(a.Cycle)))
}

// applyAnimContrib writes one contribution into the state map,
// initializing the per-group state on first touch. Replace
// semantics (not multiply) — sandwich priority is enforced by the
// sort ordering before this call.
func applyAnimContrib(c *animContrib, states map[string]svgAnimState) {
	a := c.anim
	st := states[a.GroupID]
	if !st.Inited {
		st.Opacity = 1
		st.FillOpacity = 1
		st.StrokeOpacity = 1
		st.ScaleX = 1
		st.ScaleY = 1
		st.Inited = true
	}
	switch a.Kind {
	case SvgAnimRotate:
		if a.Additive {
			st.RotAngle += c.value
		} else {
			st.RotAngle = c.value
		}
		st.RotCX = a.CenterX
		st.RotCY = a.CenterY
	case SvgAnimOpacity:
		applyOpacityContrib(&st, c.value, a.Target, a.Additive)
	case SvgAnimAttr:
		applyAttrOverride(&st.AttrOverride, a.AttrName,
			c.value, a.Additive)
	case SvgAnimTranslate:
		if a.Additive {
			st.TransX += c.valueX
			st.TransY += c.valueY
		} else {
			st.TransX = c.valueX
			st.TransY = c.valueY
		}
		st.HasXform = true
	case SvgAnimScale:
		if a.Additive {
			st.ScaleX += c.valueX
			st.ScaleY += c.valueY
		} else {
			st.ScaleX = c.valueX
			st.ScaleY = c.valueY
		}
		st.HasXform = true
	case SvgAnimMotion:
		if a.Additive {
			st.TransX += c.valueX
			st.TransY += c.valueY
		} else {
			st.TransX = c.valueX
			st.TransY = c.valueY
		}
		if a.MotionRotate != SvgAnimMotionRotateNone {
			st.RotAngle = c.value
		}
		st.HasXform = true
	}
	states[a.GroupID] = st
}

// applyOpacityContrib dispatches the opacity value to the correct
// sub-channel, honoring additive=sum. Additive adds to the existing
// channel (init 1 on first touch per sandwich); non-additive
// replaces. Clamping to [0,1] is deferred to render time.
func applyOpacityContrib(st *svgAnimState, v float32,
	target SvgAnimTarget, additive bool) {
	switch target {
	case SvgAnimTargetFill:
		if additive {
			st.FillOpacity += v
		} else {
			st.FillOpacity = v
		}
	case SvgAnimTargetStroke:
		if additive {
			st.StrokeOpacity += v
		} else {
			st.StrokeOpacity = v
		}
	default:
		if additive {
			st.Opacity += v
		} else {
			st.Opacity = v
		}
	}
}

// extractAttrOverrides pulls the AttrOverride from each svgAnimState
// with a non-zero mask into a scratch-backed map keyed by GroupID.
// Returns an empty map when no overrides are live.
func extractAttrOverrides(w *Window,
	states map[string]svgAnimState,
) map[string]SvgAnimAttrOverride {
	overrides := w.scratch.svgAnimOverrides.take(len(states))
	for gid, st := range states {
		if st.AttrOverride.Mask != 0 {
			overrides[gid] = st.AttrOverride
		}
	}
	return overrides
}

// buildEffectiveRenderPaths returns a scratch-backed copy of
// cached.RenderPaths with animated spans overlaid by fresh triangles
// from animTris. animTris is indexed by span in document order:
// spans[i] maps to animTris[offset ..+spans[i].Count].
func buildEffectiveRenderPaths(w *Window, cached *CachedSvg,
	animTris []TessellatedPath,
) []CachedSvgPath {
	eff := w.scratch.effectiveSvgPaths.take(len(cached.RenderPaths))
	eff = append(eff, cached.RenderPaths...)
	off := 0
	for _, span := range cached.AnimatedSpans {
		if off+span.Count > len(animTris) {
			// Defensive: parser returned fewer than expected — keep
			// cached for the unfulfilled tail.
			break
		}
		for k := range span.Count {
			t := &animTris[off+k]
			base := &eff[span.Start+k]
			base.Triangles = t.Triangles
			// Keep cached color / group / clip metadata; triangles
			// are the only dynamic field. Vertex colors are not
			// re-evaluated here (gradient overrides TBD).
		}
		off += span.Count
	}
	return eff
}

// applyAttrOverride writes val into the override field for attr.
// Sandwich semantics with additive:
//   - non-additive: replace the field, clear AdditiveMask bit.
//   - additive, bit unset: set field=val, mark both Mask and
//     AdditiveMask; delta will be applied to the primitive base at
//     re-tessellate time.
//   - additive, bit set: sum val into the existing field; leave
//     AdditiveMask unchanged (pre-existing non-additive value stays
//     pre-resolved, pre-existing additive value stays a delta).
//
// Unknown attr names are no-ops.
func applyAttrOverride(o *SvgAnimAttrOverride,
	attr SvgAttrName, val float32, additive bool) {
	bit := attrMaskBit(attr)
	if bit == 0 {
		return
	}
	f := attrFieldPtr(o, attr)
	if f == nil {
		return
	}
	if additive {
		if o.Mask&bit == 0 {
			*f = val
			o.AdditiveMask |= bit
		} else {
			*f += val
		}
		o.Mask |= bit
		return
	}
	*f = val
	o.Mask |= bit
	o.AdditiveMask &^= bit
}

// attrMaskBit maps a SvgAttrName to its SvgAnimAttrMask bit.
func attrMaskBit(attr SvgAttrName) SvgAnimAttrMask {
	switch attr {
	case SvgAttrCX:
		return SvgAnimMaskCX
	case SvgAttrCY:
		return SvgAnimMaskCY
	case SvgAttrR:
		return SvgAnimMaskR
	case SvgAttrRX:
		return SvgAnimMaskRX
	case SvgAttrRY:
		return SvgAnimMaskRY
	case SvgAttrX:
		return SvgAnimMaskX
	case SvgAttrY:
		return SvgAnimMaskY
	case SvgAttrWidth:
		return SvgAnimMaskWidth
	case SvgAttrHeight:
		return SvgAnimMaskHeight
	}
	return 0
}

// attrFieldPtr returns a pointer to the override field for attr.
func attrFieldPtr(o *SvgAnimAttrOverride,
	attr SvgAttrName) *float32 {
	switch attr {
	case SvgAttrCX:
		return &o.CX
	case SvgAttrCY:
		return &o.CY
	case SvgAttrR:
		return &o.R
	case SvgAttrRX:
		return &o.RX
	case SvgAttrRY:
		return &o.RY
	case SvgAttrX:
		return &o.X
	case SvgAttrY:
		return &o.Y
	case SvgAttrWidth:
		return &o.Width
	case SvgAttrHeight:
		return &o.Height
	}
	return nil
}

// lerpKeyframes interpolates keyframe scalars at frac ∈ [0,1].
// Linear by default; spline bends per-segment t via cubic-bezier;
// discrete returns the covering keyframe. keyTimes, when non-nil,
// overrides uniform i/(n-1) spacing.
func lerpKeyframes(
	vals, splines, keyTimes []float32,
	mode SvgAnimCalcMode, frac float32,
) float32 {
	n := len(vals)
	if n == 0 {
		return 1
	}
	if n == 1 {
		return vals[0]
	}
	idx, t, atEnd := locateSeg(n, frac, splines, keyTimes, mode)
	if atEnd {
		return vals[n-1]
	}
	if mode == SvgAnimCalcDiscrete {
		return vals[idx]
	}
	return vals[idx] + (vals[idx+1]-vals[idx])*t
}

// lerpKeyframes2D interpolates a paired [x0,y0, x1,y1, ...]
// keyframe stream at frac ∈ [0,1].
func lerpKeyframes2D(
	vals, splines, keyTimes []float32,
	mode SvgAnimCalcMode, frac float32,
) (float32, float32) {
	n := len(vals) / 2
	if n == 0 {
		return 0, 0
	}
	if n == 1 {
		return vals[0], vals[1]
	}
	idx, t, atEnd := locateSeg(n, frac, splines, keyTimes, mode)
	if atEnd {
		return vals[(n-1)*2], vals[(n-1)*2+1]
	}
	if mode == SvgAnimCalcDiscrete {
		return vals[idx*2], vals[idx*2+1]
	}
	x0, y0 := vals[idx*2], vals[idx*2+1]
	x1, y1 := vals[(idx+1)*2], vals[(idx+1)*2+1]
	return x0 + (x1-x0)*t, y0 + (y1-y0)*t
}

// locateSeg returns the keyframe segment containing frac: the
// lower index, the intra-segment fraction t (bent by splines when
// mode==Spline), and atEnd when frac lands on or past the last
// keyframe. Segment boundaries come from keyTimes when supplied,
// else uniform i/(n-1) for linear/spline and i/n for discrete.
// frac is clamped; NaN / negative → idx=0, t=0.
func locateSeg(
	n int, frac float32, splines, keyTimes []float32,
	mode SvgAnimCalcMode,
) (int, float32, bool) {
	frac = clampFrac(frac)
	if len(keyTimes) == n {
		return locateSegKeyTimes(n, frac, splines, keyTimes, mode)
	}
	if mode == SvgAnimCalcDiscrete {
		if frac >= 1 {
			return n - 1, 0, true
		}
		idx := int(frac * float32(n))
		if idx >= n {
			idx = n - 1
		}
		return idx, 0, false
	}
	seg := frac * float32(n-1)
	idx := max(int(seg), 0)
	if idx >= n-1 {
		return n - 1, 0, true
	}
	t := seg - float32(idx)
	if mode == SvgAnimCalcSpline && len(splines) == 4*(n-1) {
		off := idx * 4
		t = bezierCalc(t, splines[off], splines[off+1],
			splines[off+2], splines[off+3])
	}
	return idx, t, false
}

// locateSegKeyTimes walks keyTimes to find the segment covering
// frac. Discrete: keyTimes[i] starts keyframe i. Linear/spline:
// intra-segment t is (frac-keyTimes[i]) / (keyTimes[i+1]-
// keyTimes[i]); zero-width segments (duplicate keyTimes) yield
// t=0 (jump to upper keyframe — matches SMIL discrete-boundary).
func locateSegKeyTimes(
	n int, frac float32, splines, keyTimes []float32,
	mode SvgAnimCalcMode,
) (int, float32, bool) {
	if frac >= keyTimes[n-1] {
		return n - 1, 0, true
	}
	idx := 0
	for i := 0; i < n-1; i++ {
		if frac >= keyTimes[i] && frac < keyTimes[i+1] {
			idx = i
			break
		}
	}
	if mode == SvgAnimCalcDiscrete {
		return idx, 0, false
	}
	span := keyTimes[idx+1] - keyTimes[idx]
	var t float32
	if span > 0 {
		t = (frac - keyTimes[idx]) / span
	}
	if mode == SvgAnimCalcSpline && len(splines) == 4*(n-1) {
		off := idx * 4
		t = bezierCalc(t, splines[off], splines[off+1],
			splines[off+2], splines[off+3])
	}
	return idx, t, false
}

// clampFrac maps NaN / <0 → 0, >1 → 1; otherwise identity.
func clampFrac(f float32) float32 {
	if f != f {
		return 0
	}
	return clampUnit(f)
}

// finiteF32 reports whether f is finite (not NaN, not ±Inf).
// Used by animation gating so hostile parsed timing values cannot
// poison downstream render state with NaN.
func finiteF32(f float32) bool {
	return !math.IsNaN(float64(f)) && !math.IsInf(float64(f), 0)
}

// blendAlpha multiplies two 0..255 alpha channels as if they were
// in [0,1], rounding toward zero. Used to fold a tint alpha into
// the path's baked alpha without losing per-element opacity.
func blendAlpha(a, b uint8) uint8 {
	return uint8(uint16(a) * uint16(b) / 255)
}

// emitErrorPlaceholder draws a magenta rectangle placeholder.
func emitErrorPlaceholder(x, y, w, h float32, win *Window) {
	if w <= 0 || h <= 0 {
		return
	}
	emitRenderer(RenderCmd{
		Kind:  RenderRect,
		X:     x,
		Y:     y,
		W:     w,
		H:     h,
		Color: Magenta,
		Fill:  true,
	}, win)
	emitRenderer(RenderCmd{
		Kind:      RenderStrokeRect,
		X:         x,
		Y:         y,
		W:         w,
		H:         h,
		Color:     White,
		Thickness: 1,
	}, win)
}
