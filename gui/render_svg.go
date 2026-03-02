package gui

import "log"

// renderSvg renders an SVG shape by loading cached tessellation
// and emitting RenderSvg commands.
func renderSvg(shape *Shape, clip DrawClip, w *Window) {
	dr := DrawClip{
		X: shape.X, Y: shape.Y,
		Width: shape.Width, Height: shape.Height,
	}
	if !rectsOverlap(dr, clip) {
		shape.Disabled = true
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

	// Clip to the scaled viewBox area.
	emitRenderer(RenderCmd{
		Kind: RenderClip,
		X:    sx,
		Y:    sy,
		W:    cached.Width * cached.Scale,
		H:    cached.Height * cached.Scale,
	}, w)

	// Emit tessellated paths.
	for _, path := range cached.RenderPaths {
		emitSvgPathRenderer(path, color, sx, sy, cached.Scale, w)
	}

	// Emit text elements.
	for _, draw := range cached.TextDraws {
		emitCachedSvgTextDraw(draw, sx, sy, w)
	}

	// Emit textPath elements.
	for _, tp := range cached.TextPaths {
		renderSvgTextPath(tp, cached.DefsPaths,
			sx, sy, cached.Scale, w)
	}

	// Emit filtered groups.
	for i, fg := range cached.FilteredGroups {
		emitRenderer(RenderCmd{
			Kind:     RenderFilterBegin,
			GroupIdx: i,
			X:        sx,
			Y:        sy,
			Scale:    cached.Scale,
		}, w)
		for _, path := range fg.RenderPaths {
			emitSvgPathRenderer(path, color,
				sx, sy, cached.Scale, w)
		}
		for _, draw := range fg.TextDraws {
			emitCachedSvgTextDraw(draw, sx, sy, w)
		}
		for _, tp := range fg.TextPaths {
			renderSvgTextPath(tp, cached.DefsPaths,
				sx, sy, cached.Scale, w)
		}
		emitRenderer(RenderCmd{
			Kind: RenderFilterEnd,
		}, w)
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
// the tint overrides the path color.
func emitSvgPathRenderer(path CachedSvgPath, tint Color,
	x, y, scale float32, w *Window) {
	hasVCols := len(path.VertexColors) > 0
	c := path.Color
	if tint.A > 0 && !hasVCols {
		c = tint
	}
	var vcols []Color
	if hasVCols && tint.A == 0 {
		vcols = path.VertexColors
	}
	emitRenderer(RenderCmd{
		Kind:         RenderSvg,
		Triangles:    path.Triangles,
		Color:        c,
		VertexColors: vcols,
		X:            x,
		Y:            y,
		Scale:        scale,
		IsClipMask:   path.IsClipMask,
		ClipGroup:    path.ClipGroup,
	}, w)
}

// emitCachedSvgTextDraw emits a cached SVG text draw as a
// RenderText command.
func emitCachedSvgTextDraw(draw CachedSvgTextDraw,
	shapeX, shapeY float32, w *Window) {
	emitRenderer(RenderCmd{
		Kind:     RenderText,
		Text:     draw.Text,
		X:        shapeX + draw.X,
		Y:        shapeY + draw.Y,
		Color:    draw.TextStyle.Color,
		FontName: draw.TextStyle.Family,
		FontSize: draw.TextStyle.Size,
	}, w)
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
