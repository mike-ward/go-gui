package gui

// renderImage renders an image shape by emitting a RenderImage
// command with the shape's resource path and clip radius.
func renderImage(shape *Shape, clip DrawClip, w *Window) {
	dr := DrawClip{
		X: shape.X, Y: shape.Y,
		Width: shape.Width, Height: shape.Height,
	}
	if !rectsOverlap(dr, clip) {
		return
	}

	// Hide Color from renderContainer so it doesn't draw a
	// redundant bg rect; the backend handles the fill itself.
	bgColor := shape.Color
	shape.Color = ColorTransparent
	defer func() { shape.Color = bgColor }()
	renderContainer(shape, ColorTransparent, clip, w)

	emitRenderer(RenderCmd{
		Kind:       RenderImage,
		X:          shape.X,
		Y:          shape.Y,
		W:          shape.Width,
		H:          shape.Height,
		Color:      bgColor,
		Resource:   shape.Resource,
		ClipRadius: w.clipRadius,
	}, w)
}
