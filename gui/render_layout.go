package gui

import "strings"

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
	isImage := shape.ShapeType == ShapeImage
	isSvg := shape.ShapeType == ShapeSVG
	isCanvas := shape.ShapeType == ShapeDrawCanvas
	hasFX := shape.FX != nil && (shape.FX.Gradient != nil ||
		shape.FX.BorderGradient != nil)

	isRTF := shape.ShapeType == ShapeRTF

	if shape.Color == ColorTransparent && !hasFX && !hasBorder &&
		!hasText && !isImage && !isSvg && !isCanvas && !isRTF {
		return
	}

	switch shape.ShapeType {
	case ShapeRectangle:
		renderContainer(shape, parentColor, clip, w)
	case ShapeText:
		renderText(shape, clip, w)
	case ShapeImage:
		renderImage(shape, clip, w)
	case ShapeCircle:
		renderCircle(shape, clip, w)
	case ShapeRTF:
		renderRtf(shape, clip, w)
	case ShapeSVG:
		renderSvg(shape, clip, w)
	case ShapeDrawCanvas:
		renderDrawCanvas(shape, clip, w)
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

// renderText emits a RenderText command for a text shape.
func renderText(shape *Shape, clip DrawClip, w *Window) {
	tc := shape.TC
	if tc == nil || len(tc.Text) == 0 {
		return
	}
	dr := DrawClip{
		X: shape.X, Y: shape.Y,
		Width: shape.Width, Height: shape.Height,
	}
	if !rectsOverlap(dr, clip) {
		return
	}
	c := tc.TextStyle.Color
	if shape.Opacity < 1.0 {
		c = c.WithOpacity(shape.Opacity)
	}
	if shape.Disabled {
		c = dimAlpha(c)
	}

	text := tc.Text
	if tc.TextIsPassword {
		if strings.Contains(tc.Text, "\n") {
			text = passwordMaskKeepNewlines(tc.Text)
		} else {
			text = passwordMask(tc.Text)
		}
	}

	cmd := RenderCmd{
		Kind:     RenderText,
		X:        shape.X + shape.PaddingLeft(),
		Y:        shape.Y + shape.PaddingTop(),
		Color:    c,
		Text:     text,
		FontName: tc.TextStyle.Family,
		FontSize: tc.TextStyle.Size,
	}
	if tc.TextMode == TextModeWrap ||
		tc.TextMode == TextModeWrapKeepSpaces {
		cmd.W = shape.Width
	}
	emitRenderer(cmd, w)
}

// renderRtf emits a RenderRTF command for pre-shaped rich text.
func renderRtf(shape *Shape, clip DrawClip, w *Window) {
	if !shape.HasRtfLayout() {
		return
	}
	dr := DrawClip{
		X: shape.X, Y: shape.Y,
		Width: shape.Width, Height: shape.Height,
	}
	if !rectsOverlap(dr, clip) {
		return
	}
	baseX := shape.X + shape.PaddingLeft()
	baseY := shape.Y + shape.PaddingTop()
	emitRenderer(RenderCmd{
		Kind:      RenderRTF,
		X:         baseX,
		Y:         baseY,
		LayoutPtr: shape.TC.RtfLayout,
	}, w)

	// Emit RenderImage for inline math objects.
	cache := w.viewState.diagramCache
	if cache == nil {
		return
	}
	for i := range shape.TC.RtfLayout.Items {
		item := &shape.TC.RtfLayout.Items[i]
		if !item.IsObject {
			continue
		}
		hash := mathCacheHash(item.ObjectID)
		entry, ok := cache.Get(hash)
		if !ok || entry.State != DiagramReady {
			continue
		}
		h := float32(item.Ascent + item.Descent)
		emitRenderer(RenderCmd{
			Kind:     RenderImage,
			X:        baseX + float32(item.X),
			Y:        baseY + float32(item.Y-item.Ascent),
			W:        float32(item.Width),
			H:        h,
			Resource: entry.PNGPath,
		}, w)
	}
}
