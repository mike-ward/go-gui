package gui

// renderLayout walks the layout tree and emits RenderCmd entries
// into window.renderers. Clip rectangles bracket clipped children.
func renderLayout(layout *Layout, bgColor Color, clip DrawClip, w *Window) {
	renderShape(layout.Shape, bgColor, clip, w)

	shapeClip := clip
	if layout.Shape.OverDraw {
		shapeClip = layout.Shape.ShapeClip
		if layout.Shape.ScrollbarOrientation == ScrollbarVertical {
			shapeClip.Y = clip.Y
			shapeClip.Height = clip.Height
		}
		if layout.Shape.ScrollbarOrientation == ScrollbarHorizontal {
			shapeClip.X = clip.X
			shapeClip.Width = clip.Width
		}
		emitRenderer(RenderCmd{
			Kind: RenderClip,
			X:    shapeClip.X,
			Y:    shapeClip.Y,
			W:    shapeClip.Width,
			H:    shapeClip.Height,
		}, w)
	} else if layout.Shape.Clip {
		sc := layout.Shape.ShapeClip
		isRTL := effectiveTextDir(layout.Shape) == TextDirRTL
		var padX float32
		if isRTL {
			padX = layout.Shape.Padding.Right + layout.Shape.SizeBorder
		} else {
			padX = layout.Shape.PaddingLeft()
		}
		shapeClip = DrawClip{
			X:      sc.X + padX,
			Y:      sc.Y + layout.Shape.PaddingTop(),
			Width:  f32Max(0, sc.Width-layout.Shape.PaddingWidth()),
			Height: f32Max(0, sc.Height-layout.Shape.PaddingHeight()),
		}
		emitRenderer(RenderCmd{
			Kind: RenderClip,
			X:    shapeClip.X,
			Y:    shapeClip.Y,
			W:    shapeClip.Width,
			H:    shapeClip.Height,
		}, w)
	}

	// Propagate rounded clip radius to child images.
	savedClipRadius := w.clipRadius
	w.clipRadius = resolveClipRadius(savedClipRadius, layout.Shape)

	color := bgColor
	if layout.Shape.Color != ColorTransparent {
		color = layout.Shape.Color
	}
	for i := range layout.Children {
		renderLayout(&layout.Children[i], color, shapeClip, w)
	}

	w.clipRadius = savedClipRadius
	if layout.Shape.Clip || layout.Shape.OverDraw {
		emitRenderer(RenderCmd{
			Kind: RenderClip,
			X:    clip.X,
			Y:    clip.Y,
			W:    clip.Width,
			H:    clip.Height,
		}, w)
	}
}

// renderShape dispatches to the type-specific renderer, applying
// opacity when needed.
func renderShape(shape *Shape, parentColor Color, clip DrawClip, w *Window) {
	// Degrade safely if a text-like shape is missing text config.
	if (shape.ShapeType == ShapeText || shape.ShapeType == ShapeRTF) &&
		shape.TC == nil {
		return
	}

	if shape.Opacity < 1.0 {
		origColor := shape.Color
		origBorder := shape.ColorBorder
		shape.Color = shape.Color.WithOpacity(shape.Opacity)
		shape.ColorBorder = shape.ColorBorder.WithOpacity(shape.Opacity)
		renderShapeInner(shape, parentColor, clip, w)
		shape.Color = origColor
		shape.ColorBorder = origBorder
	} else {
		renderShapeInner(shape, parentColor, clip, w)
	}
}

// renderShapeInner dispatches to the type-specific renderer after
// visibility checks.
func renderShapeInner(shape *Shape, parentColor Color, clip DrawClip, w *Window) {
	hasBorder := shape.SizeBorder > 0 && shape.ColorBorder != ColorTransparent
	hasText := shape.ShapeType == ShapeText && shape.TC != nil
	isSvg := shape.ShapeType == ShapeSVG
	isCanvas := shape.ShapeType == ShapeDrawCanvas
	hasFX := shape.FX != nil && (shape.FX.Gradient != nil ||
		shape.FX.BorderGradient != nil)

	if shape.Color == ColorTransparent && !hasFX && !hasBorder &&
		!hasText && !isSvg && !isCanvas {
		return
	}

	switch shape.ShapeType {
	case ShapeRectangle:
		renderContainer(shape, parentColor, clip, w)
	case ShapeText:
		// TODO: renderText — Phase 3
	case ShapeImage:
		// TODO: renderImage — Phase 3
	case ShapeCircle:
		renderCircle(shape, clip, w)
	case ShapeRTF:
		// TODO: renderRtf — Phase 7
	case ShapeSVG:
		// TODO: renderSvg — Phase 7
	case ShapeDrawCanvas:
		// TODO: renderDrawCanvas — Phase 7
	case ShapeNone:
		// no-op
	}
}

// renderContainer draws a rectangle (possibly with shadow, gradient,
// blur, or border).
func renderContainer(shape *Shape, parentColor Color, clip DrawClip, w *Window) {
	fx := shape.FX
	hasFX := fx != nil

	// Shadow
	if hasFX && fx.Shadow != nil &&
		fx.Shadow.Color.A > 0 && fx.Shadow.BlurRadius > 0 {
		emitRenderer(RenderCmd{
			Kind:       RenderShadow,
			X:          shape.X + fx.Shadow.OffsetX,
			Y:          shape.Y + fx.Shadow.OffsetY,
			W:          shape.Width,
			H:          shape.Height,
			Radius:     shape.Radius,
			BlurRadius: fx.Shadow.BlurRadius,
			Color:      fx.Shadow.Color,
			OffsetX:    fx.Shadow.OffsetX,
			OffsetY:    fx.Shadow.OffsetY,
		}, w)
	}

	// Gradient fill
	if hasFX && fx.Gradient != nil {
		emitRenderer(RenderCmd{
			Kind:     RenderGradient,
			X:        shape.X,
			Y:        shape.Y,
			W:        shape.Width,
			H:        shape.Height,
			Radius:   shape.Radius,
			Gradient: fx.Gradient,
		}, w)
	} else if hasFX && fx.BlurRadius > 0 && shape.Color.A > 0 {
		// Blur
		c := shape.Color
		if shape.Disabled {
			c = dimAlpha(c)
		}
		emitRenderer(RenderCmd{
			Kind:       RenderBlur,
			X:          shape.X,
			Y:          shape.Y,
			W:          shape.Width,
			H:          shape.Height,
			Radius:     shape.Radius,
			BlurRadius: fx.BlurRadius,
			Color:      c,
		}, w)
	} else {
		// Border gradient or plain rectangle
		if hasFX && fx.BorderGradient != nil {
			emitRenderer(RenderCmd{
				Kind:      RenderGradientBorder,
				X:         shape.X,
				Y:         shape.Y,
				W:         shape.Width,
				H:         shape.Height,
				Radius:    shape.Radius,
				Thickness: shape.SizeBorder,
				Gradient:  fx.BorderGradient,
			}, w)
		} else {
			renderRectangle(shape, clip, w)
		}
	}
}

// renderRectangle draws a shape as a filled rectangle with optional
// stroke border.
func renderRectangle(shape *Shape, clip DrawClip, w *Window) {
	dr := DrawClip{
		X: shape.X, Y: shape.Y,
		Width: shape.Width, Height: shape.Height,
	}
	c := shape.Color
	if shape.Disabled {
		c = dimAlpha(c)
	}

	if rectsOverlap(dr, clip) {
		// Fill
		if c.A > 0 {
			emitRenderer(RenderCmd{
				Kind:   RenderRect,
				X:      dr.X,
				Y:      dr.Y,
				W:      dr.Width,
				H:      dr.Height,
				Color:  c,
				Fill:   true,
				Radius: shape.Radius,
			}, w)
		}
		// Border
		if shape.SizeBorder > 0 {
			cb := shape.ColorBorder
			if shape.Disabled {
				cb = dimAlpha(cb)
			}
			if cb.A > 0 {
				emitRenderer(RenderCmd{
					Kind:      RenderStrokeRect,
					X:         dr.X,
					Y:         dr.Y,
					W:         dr.Width,
					H:         dr.Height,
					Color:     cb,
					Radius:    shape.Radius,
					Thickness: shape.SizeBorder,
				}, w)
			}
		}
	} else {
		shape.Disabled = true
	}
}

// renderCircle draws a shape as a circle in the middle of the
// shape's rectangular region.
func renderCircle(shape *Shape, clip DrawClip, w *Window) {
	dr := DrawClip{
		X: shape.X, Y: shape.Y,
		Width: shape.Width, Height: shape.Height,
	}
	c := shape.Color
	if shape.Disabled {
		c = dimAlpha(c)
	}

	if rectsOverlap(dr, clip) {
		radius := f32Min(shape.Width, shape.Height) / 2
		cx := shape.X + shape.Width/2
		cy := shape.Y + shape.Height/2

		if c.A > 0 {
			emitRenderer(RenderCmd{
				Kind:   RenderCircle,
				X:      cx,
				Y:      cy,
				Radius: radius,
				Fill:   true,
				Color:  c,
			}, w)
		}

		// Border
		fx := shape.FX
		if fx != nil && fx.BorderGradient != nil && shape.SizeBorder > 0 {
			emitRenderer(RenderCmd{
				Kind:      RenderGradientBorder,
				X:         dr.X,
				Y:         dr.Y,
				W:         dr.Width,
				H:         dr.Height,
				Radius:    radius,
				Thickness: shape.SizeBorder,
				Gradient:  fx.BorderGradient,
			}, w)
		} else if shape.SizeBorder > 0 {
			cb := shape.ColorBorder
			if shape.Disabled {
				cb = dimAlpha(cb)
			}
			if cb.A > 0 {
				emitRenderer(RenderCmd{
					Kind:      RenderStrokeRect,
					X:         dr.X,
					Y:         dr.Y,
					W:         dr.Width,
					H:         dr.Height,
					Color:     cb,
					Radius:    radius,
					Thickness: shape.SizeBorder,
				}, w)
			}
		}
	} else {
		shape.Disabled = true
	}
}
