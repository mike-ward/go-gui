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
	// attribute overrides are live (phase-2). Falls back to cached
	// triangles when no parser support, no spans, or empty overrides.
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
		// render state. DurSec must also be strictly positive for
		// phase/duration to be meaningful.
		if a.GroupID == "" ||
			!finiteF32(a.DurSec) || a.DurSec <= 0 ||
			!finiteF32(a.BeginSec) ||
			!finiteF32(a.Cycle) ||
			!finiteF32(elapsedSec) ||
			elapsedSec < a.BeginSec {
			continue
		}
		activation := a.BeginSec
		if a.Cycle > 0 {
			n := math.Floor(float64(elapsedSec-a.BeginSec) / float64(a.Cycle))
			activation = a.BeginSec + float32(n)*a.Cycle
		}
		phase := elapsedSec - activation
		var frac float32
		switch {
		case phase < a.DurSec:
			frac = phase / a.DurSec
		case a.Freeze:
			frac = 1
		default:
			continue
		}
		c := animContrib{anim: a, activation: activation}
		switch a.Kind {
		case SvgAnimRotate, SvgAnimOpacity, SvgAnimAttr:
			if len(a.Values) < 2 {
				continue
			}
			c.value = lerpKeyframes(a.Values, a.KeySplines, frac)
		case SvgAnimTranslate, SvgAnimScale:
			if len(a.Values) < 4 {
				continue
			}
			c.valueX, c.valueY = lerpKeyframes2D(
				a.Values, a.KeySplines, frac)
		}
		out = append(out, c)
	}
	return out
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
		st.RotAngle = c.value
		st.RotCX = a.CenterX
		st.RotCY = a.CenterY
	case SvgAnimOpacity:
		switch a.Target {
		case SvgAnimTargetFill:
			st.FillOpacity = c.value
		case SvgAnimTargetStroke:
			st.StrokeOpacity = c.value
		default:
			st.Opacity = c.value
		}
	case SvgAnimAttr:
		applyAttrOverride(&st.AttrOverride, a.AttrName, c.value)
	case SvgAnimTranslate:
		st.TransX = c.valueX
		st.TransY = c.valueY
		st.HasXform = true
	case SvgAnimScale:
		st.ScaleX = c.valueX
		st.ScaleY = c.valueY
		st.HasXform = true
	}
	states[a.GroupID] = st
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
			// re-evaluated in phase-2 (gradient overrides TBD).
		}
		off += span.Count
	}
	return eff
}

// applyAttrOverride sets the override field for attr to val and
// marks the mask bit. Unknown attr names are no-ops.
func applyAttrOverride(o *SvgAnimAttrOverride,
	attr SvgAttrName, val float32) {
	switch attr {
	case SvgAttrCX:
		o.CX = val
		o.Mask |= SvgAnimMaskCX
	case SvgAttrCY:
		o.CY = val
		o.Mask |= SvgAnimMaskCY
	case SvgAttrR:
		o.R = val
		o.Mask |= SvgAnimMaskR
	case SvgAttrRX:
		o.RX = val
		o.Mask |= SvgAnimMaskRX
	case SvgAttrRY:
		o.RY = val
		o.Mask |= SvgAnimMaskRY
	case SvgAttrX:
		o.X = val
		o.Mask |= SvgAnimMaskX
	case SvgAttrY:
		o.Y = val
		o.Mask |= SvgAnimMaskY
	case SvgAttrWidth:
		o.Width = val
		o.Mask |= SvgAnimMaskWidth
	case SvgAttrHeight:
		o.Height = val
		o.Mask |= SvgAnimMaskHeight
	}
}

// lerpKeyframes interpolates evenly-spaced keyframe scalars at
// frac ∈ [0,1]. When splines has the expected 4*(len(vals)-1)
// layout, per-segment fraction is bent via cubic-bezier ease;
// otherwise linear.
func lerpKeyframes(vals, splines []float32, frac float32) float32 {
	n := len(vals)
	if n == 0 {
		return 1
	}
	if n == 1 {
		return vals[0]
	}
	idx, t, atEnd := locateSeg(n, frac, splines)
	if atEnd {
		return vals[n-1]
	}
	return vals[idx] + (vals[idx+1]-vals[idx])*t
}

// lerpKeyframes2D interpolates a paired [x0,y0, x1,y1, ...]
// keyframe stream at frac ∈ [0,1].
func lerpKeyframes2D(vals, splines []float32, frac float32) (float32, float32) {
	n := len(vals) / 2
	if n == 0 {
		return 0, 0
	}
	if n == 1 {
		return vals[0], vals[1]
	}
	idx, t, atEnd := locateSeg(n, frac, splines)
	if atEnd {
		return vals[(n-1)*2], vals[(n-1)*2+1]
	}
	x0, y0 := vals[idx*2], vals[idx*2+1]
	x1, y1 := vals[(idx+1)*2], vals[(idx+1)*2+1]
	return x0 + (x1-x0)*t, y0 + (y1-y0)*t
}

// locateSeg returns the keyframe segment containing frac: the
// lower index, the intra-segment fraction t (bent by splines
// when provided), and atEnd when frac lands on or past the last
// keyframe. frac is clamped; NaN / negative produce idx=0, t=0.
func locateSeg(n int, frac float32, splines []float32) (int, float32, bool) {
	frac = clampFrac(frac)
	seg := frac * float32(n-1)
	idx := max(int(seg), 0)
	if idx >= n-1 {
		return n - 1, 0, true
	}
	t := seg - float32(idx)
	if len(splines) == 4*(n-1) {
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
