package gui

import (
	"log"
	"math"
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
		animState = computeSvgAnimations(
			cached.Animations, elapsed, animState)
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
// the tint overrides the path color. animState applies SMIL
// rotation/opacity per GroupID.
func emitSvgPathRenderer(path CachedSvgPath, tint Color,
	x, y, scale float32,
	animState map[string]svgAnimState, w *Window) {
	hasVCols := len(path.VertexColors) > 0
	c := path.Color
	if tint.A > 0 && !hasVCols {
		c = tint
	}
	var vcols []Color
	if hasVCols && tint.A == 0 {
		vcols = path.VertexColors
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
			if st.Opacity < 1 {
				c.A = uint8(float32(c.A) * st.Opacity)
				if len(vcols) > 0 {
					vAlphaScale = st.Opacity
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
	Opacity  float32 // 0..1
	// TransX/TransY is the animated translate; ScaleX/ScaleY the
	// animated scale. Identity when HasXform is false.
	TransX, TransY float32
	ScaleX, ScaleY float32
	HasXform       bool
	Inited         bool
	AttrOverride   SvgAnimAttrOverride // attribute overrides for re-tessellation
}

// computeSvgAnimations builds a map of per-group animation
// state from parsed SMIL animations and elapsed time.
func computeSvgAnimations(
	anims []SvgAnimation, elapsedSec float32,
	states map[string]svgAnimState,
) map[string]svgAnimState {
	if states == nil {
		states = make(map[string]svgAnimState, len(anims))
	} else {
		clear(states)
	}
	for _, a := range anims {
		if a.GroupID == "" || a.DurSec <= 0 {
			continue
		}
		st := states[a.GroupID]
		if !st.Inited {
			st.Opacity = 1
			st.ScaleX = 1
			st.ScaleY = 1
			st.Inited = true
		}
		adj := elapsedSec - a.BeginSec
		adj = max(adj, 0)
		frac := fmod(adj, a.DurSec) / a.DurSec

		switch a.Kind {
		case SvgAnimRotate:
			if len(a.Values) >= 2 {
				st.RotAngle = lerpKeyframes(a.Values, a.KeySplines, frac)
				st.RotCX = a.CenterX
				st.RotCY = a.CenterY
			}
		case SvgAnimOpacity:
			if len(a.Values) >= 2 {
				st.Opacity *= lerpKeyframes(a.Values, a.KeySplines, frac)
			}
		case SvgAnimAttr:
			if len(a.Values) >= 2 {
				applyAttrOverride(&st.AttrOverride, a.AttrName,
					lerpKeyframes(a.Values, a.KeySplines, frac))
			}
		case SvgAnimTranslate:
			if len(a.Values) >= 4 {
				x, y := lerpKeyframes2D(a.Values, a.KeySplines, frac)
				st.TransX = x
				st.TransY = y
				st.HasXform = true
			}
		case SvgAnimScale:
			if len(a.Values) >= 4 {
				x, y := lerpKeyframes2D(a.Values, a.KeySplines, frac)
				st.ScaleX = x
				st.ScaleY = y
				st.HasXform = true
			}
		}
		states[a.GroupID] = st
	}
	return states
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

// fmod returns x mod y, always positive.
func fmod(x, y float32) float32 {
	r := float32(math.Mod(float64(x), float64(y)))
	if r < 0 {
		r += y
	}
	return r
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
