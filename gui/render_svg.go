package gui

import (
	"log"
	"math"
	"time"
)

// renderSvg renders an SVG shape by loading cached tessellation
// and emitting RenderSvg commands.
func renderSvg(shape *Shape, clip DrawClip, w *Window) {
	dr := DrawClip{
		X: shape.X, Y: shape.Y,
		Width: shape.Width, Height: shape.Height,
	}
	if !rectsOverlap(dr, clip) {
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
	svgClip, _ := rectIntersection(clip, DrawClip{
		X:      sx,
		Y:      sy,
		Width:  cached.Width * cached.Scale,
		Height: cached.Height * cached.Scale,
	})
	emitRenderer(RenderCmd{
		Kind: RenderClip,
		X:    svgClip.X,
		Y:    svgClip.Y,
		W:    svgClip.Width,
		H:    svgClip.Height,
	}, w)

	// Compute animation state for SMIL animations.
	var animState map[string]svgAnimState
	if cached.HasAnimations && cached.AnimStartNs != 0 {
		animState = w.scratch.takeSvgAnimStates(len(cached.Animations))
		defer w.scratch.putSvgAnimStates(animState)
		nowNs := time.Now().UnixNano()
		// Keep animation alive while SVG is being rendered.
		if cached.AnimHash != "" {
			animSeen := StateMap[string, int64](
				w, nsSvgAnimSeen, capModerate)
			animSeen.Set(cached.AnimHash, nowNs)
		}
		elapsed := float32(nowNs-cached.AnimStartNs) /
			float32(time.Second)
		animState = computeSvgAnimations(
			cached.Animations, elapsed, animState)
	}

	// Emit tessellated paths.
	for _, path := range cached.RenderPaths {
		emitSvgPathRenderer(path, color, sx, sy,
			cached.Scale, animState, w)
	}

	// Emit text elements.
	for i := range cached.TextDraws {
		emitCachedSvgTextDraw(&cached.TextDraws[i], sx, sy, w)
	}

	// Emit textPath elements.
	for i := range cached.TextPathDraws {
		emitCachedSvgTextPathDraw(&cached.TextPathDraws[i], sx, sy, w)
	}

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
		for _, path := range fg.RenderPaths {
			emitSvgPathRenderer(path, color,
				sx, sy, cached.Scale, animState, w)
		}
		for j := range fg.TextDraws {
			emitCachedSvgTextDraw(&fg.TextDraws[j], sx, sy, w)
		}
		for j := range fg.TextPathDraws {
			emitCachedSvgTextPathDraw(&fg.TextPathDraws[j], sx, sy, w)
		}
		emitRenderer(RenderCmd{
			Kind: RenderFilterEnd,
		}, w)

		// KeepSource: re-draw sharp original on top of blur.
		if fg.Filter.KeepSource {
			for _, path := range fg.RenderPaths {
				emitSvgPathRenderer(path, color,
					sx, sy, cached.Scale, animState, w)
			}
			for j := range fg.TextDraws {
				emitCachedSvgTextDraw(&fg.TextDraws[j], sx, sy, w)
			}
			for j := range fg.TextPathDraws {
				emitCachedSvgTextPathDraw(
					&fg.TextPathDraws[j], sx, sy, w)
			}
		}
	}

	// Restore parent clip.
	emitRenderer(RenderCmd{
		Kind: RenderClip,
		X:    clip.X,
		Y:    clip.Y,
		W:    clip.Width,
		H:    clip.Height,
	}, w)
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
	var vAlphaScale float32
	hasVAlpha := false
	if animState != nil && path.GroupID != "" {
		if st, ok := animState[path.GroupID]; ok {
			rotAngle = st.RotAngle
			rotCX = st.RotCX
			rotCY = st.RotCY
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
	Inited   bool
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
			st.Inited = true
		}
		adj := elapsedSec - a.BeginSec
		if adj < 0 {
			adj = 0
		}
		frac := fmod(adj, a.DurSec) / a.DurSec

		switch a.Kind {
		case SvgAnimRotate:
			if len(a.Values) >= 2 {
				from, to := a.Values[0], a.Values[1]
				st.RotAngle = from + (to-from)*frac
				st.RotCX = a.CenterX
				st.RotCY = a.CenterY
			}
		case SvgAnimOpacity:
			if len(a.Values) >= 2 {
				st.Opacity *= lerpKeyframes(a.Values, frac)
			}
		}
		states[a.GroupID] = st
	}
	return states
}

// lerpKeyframes interpolates evenly-spaced keyframe values at
// the given fraction (0..1).
func lerpKeyframes(vals []float32, frac float32) float32 {
	n := len(vals)
	if n == 0 {
		return 1
	}
	if n == 1 {
		return vals[0]
	}
	// Scale fraction to segment index.
	seg := frac * float32(n-1)
	idx := int(seg)
	if idx >= n-1 {
		return vals[n-1]
	}
	t := seg - float32(idx)
	return vals[idx] + (vals[idx+1]-vals[idx])*t
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
