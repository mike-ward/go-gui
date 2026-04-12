package gui

import "math"

// renderDrawCanvas renders cached draw-canvas triangle batches.
func renderDrawCanvas(shape *Shape, clip DrawClip, w *Window) {
	if !rectsOverlap(shapeBounds(shape), clip) {
		return
	}
	// Background, border, effects.
	renderContainer(shape, ColorTransparent, clip, w)

	sm := StateMap[string, DrawCanvasCache](w, nsDrawCanvas, capModerate)

	// Content dimensions account for padding.
	cw := shape.Width - shape.PaddingWidth()
	ch := shape.Height - shape.PaddingHeight()

	var cached DrawCanvasCache
	needsDraw := true

	// Skip cache when ID is empty to avoid collisions between
	// multiple ID-less DrawCanvas widgets.
	if shape.ID != "" {
		var ok bool
		cached, ok = sm.Get(shape.ID)
		if ok && cached.Version == shape.Version &&
			cached.TessWidth == cw && cached.TessHeight == ch {
			needsDraw = false
		}
	}

	if needsDraw && shape.Events != nil && shape.Events.OnDraw != nil {
		dc := DrawContext{
			Width:       cw,
			Height:      ch,
			textMeasure: w.textMeasurer,
		}
		shape.Events.OnDraw(&dc)
		cached = DrawCanvasCache{
			Version:    shape.Version,
			TessWidth:  cw,
			TessHeight: ch,
			Batches:    dc.batches,
			Texts:      dc.texts,
		}
		if shape.ID != "" {
			sm.Set(shape.ID, cached)
		}
	}

	if len(cached.Batches) == 0 && len(cached.Texts) == 0 {
		return
	}

	// Content origin accounts for padding.
	ox := shape.X + shape.PaddingLeft()
	oy := shape.Y + shape.PaddingTop()

	// Clip to content area.
	if shape.Clip {
		emitClipCmd(DrawClip{X: ox, Y: oy, Width: cw, Height: ch}, w)
	}

	for _, batch := range cached.Batches {
		emitRenderer(RenderCmd{
			Kind:      RenderSvg,
			Triangles: batch.Triangles,
			Color:     batch.Color,
			X:         ox,
			Y:         oy,
			Scale:     1.0,
		}, w)
	}

	// Emit deferred text commands.
	for i := range cached.Texts {
		t := &cached.Texts[i]
		fontAscent := t.Style.Size * 0.8
		var textWidth float32
		if w.textMeasurer != nil {
			fontAscent = w.textMeasurer.FontAscent(t.Style)
			textWidth = w.textMeasurer.TextWidth(t.Text, t.Style)
		}

		tx := ox + t.X
		ty := oy + t.Y
		rotated := t.Style.RotationRadians != 0

		if rotated {
			deg := t.Style.RotationRadians * (180 / math.Pi)
			emitRenderer(RenderCmd{
				Kind:     RenderRotateBegin,
				RotAngle: deg,
				RotCX:    tx,
				RotCY:    ty,
			}, w)
		}

		emitRenderer(RenderCmd{
			Kind:         RenderText,
			X:            tx,
			Y:            ty,
			Color:        t.Style.Color,
			Text:         t.Text,
			FontName:     t.Style.Family,
			FontSize:     t.Style.Size,
			FontAscent:   fontAscent,
			TextWidth:    textWidth,
			TextStylePtr: w.scratch.renderTextStyles.alloc(t.Style),
			TextGradient: t.Style.Gradient,
		}, w)

		if rotated {
			emitRenderer(RenderCmd{
				Kind: RenderRotateEnd,
			}, w)
		}
	}

	// Restore parent clip.
	if shape.Clip {
		emitClipCmd(clip, w)
	}
}
