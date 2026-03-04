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

	sm := StateMapRead[string, DrawCanvasCache](w, nsDrawCanvas)
	if sm == nil {
		return
	}
	cached, ok := sm.Get(shape.ID)
	if !ok {
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
			W:    shape.Width - shape.PaddingWidth(),
			H:    shape.Height - shape.PaddingHeight(),
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
