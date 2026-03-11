package gui

// renderDrawCanvas renders cached draw-canvas triangle batches.
func renderDrawCanvas(shape *Shape, clip DrawClip, w *Window) {
	dr := DrawClip{
		X: shape.X, Y: shape.Y,
		Width: shape.Width, Height: shape.Height,
	}
	if !rectsOverlap(dr, clip) {
		return
	}
	// Background, border, effects.
	renderContainer(shape, ColorTransparent, clip, w)

	sm := StateMap[string, DrawCanvasCache](w, nsDrawCanvas, capModerate)
	cached, ok := sm.Get(shape.ID)

	// Content dimensions account for padding.
	cw := shape.Width - shape.PaddingWidth()
	ch := shape.Height - shape.PaddingHeight()

	var needsDraw bool
	if !ok || cached.Version != shape.Version || cached.TessWidth != cw || cached.TessHeight != ch {
		needsDraw = true
	}

	if needsDraw && shape.Events != nil && shape.Events.OnDraw != nil {
		dc := DrawContext{
			Width:  cw,
			Height: ch,
		}
		shape.Events.OnDraw(&dc)
		cached = DrawCanvasCache{
			Version:    shape.Version,
			TessWidth:  cw,
			TessHeight: ch,
			Batches:    dc.batches,
		}
		sm.Set(shape.ID, cached)
	}

	if len(cached.Batches) == 0 {
		return
	}

	// Content origin accounts for padding.
	ox := shape.X + shape.PaddingLeft()
	oy := shape.Y + shape.PaddingTop()

	// Clip to content area.
	if shape.Clip {
		emitRenderer(RenderCmd{
			Kind: RenderClip,
			X:    ox,
			Y:    oy,
			W:    cw,
			H:    ch,
		}, w)
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

	// Restore parent clip.
	if shape.Clip {
		emitRenderer(RenderCmd{
			Kind: RenderClip,
			X:    clip.X,
			Y:    clip.Y,
			W:    clip.Width,
			H:    clip.Height,
		}, w)
	}
}
