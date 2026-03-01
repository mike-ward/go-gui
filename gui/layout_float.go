package gui

// mirrorFloatAttach swaps left/right anchors for RTL mirroring.
func mirrorFloatAttach(a FloatAttach) FloatAttach {
	switch a {
	case FloatTopLeft:
		return FloatTopRight
	case FloatTopRight:
		return FloatTopLeft
	case FloatMiddleLeft:
		return FloatMiddleRight
	case FloatMiddleRight:
		return FloatMiddleLeft
	case FloatBottomLeft:
		return FloatBottomRight
	case FloatBottomRight:
		return FloatBottomLeft
	default:
		return a
	}
}

// floatAttachLayout computes the position of float layouts relative
// to their parent.
func floatAttachLayout(layout *Layout) (float32, float32) {
	if layout.Parent == nil {
		return 0, 0
	}
	parent := layout.Parent.Shape
	isRTL := effectiveTextDir(parent) == TextDirRTL

	anchor := layout.Shape.FloatAnchor
	tieOff := layout.Shape.FloatTieOff
	offsetX := layout.Shape.FloatOffsetX
	if isRTL {
		anchor = mirrorFloatAttach(anchor)
		tieOff = mirrorFloatAttach(tieOff)
		offsetX = -offsetX
	}

	x, y := parent.X, parent.Y
	switch anchor {
	case FloatTopLeft:
	case FloatTopCenter:
		x += parent.Width / 2
	case FloatTopRight:
		x += parent.Width
	case FloatMiddleLeft:
		y += parent.Height / 2
	case FloatMiddleCenter:
		x += parent.Width / 2
		y += parent.Height / 2
	case FloatMiddleRight:
		x += parent.Width
		y += parent.Height / 2
	case FloatBottomLeft:
		y += parent.Height
	case FloatBottomCenter:
		x += parent.Width / 2
		y += parent.Height
	case FloatBottomRight:
		x += parent.Width
		y += parent.Height
	}

	shape := layout.Shape
	switch tieOff {
	case FloatTopLeft:
	case FloatTopCenter:
		x -= shape.Width / 2
	case FloatTopRight:
		x -= shape.Width
	case FloatMiddleLeft:
		y -= shape.Height / 2
	case FloatMiddleCenter:
		x -= shape.Width / 2
		y -= shape.Height / 2
	case FloatMiddleRight:
		x -= shape.Width
		y -= shape.Height / 2
	case FloatBottomLeft:
		y -= shape.Height
	case FloatBottomCenter:
		x -= shape.Width / 2
		y -= shape.Height
	case FloatBottomRight:
		x -= shape.Width
		y -= shape.Height
	}
	x += offsetX
	y += layout.Shape.FloatOffsetY
	return x, y
}

// layoutRemoveFloatingLayouts extracts floating elements from the
// main layout tree, replacing them with placeholders.
func layoutRemoveFloatingLayouts(layout *Layout, layouts *[]*Layout) {
	for i := range layout.Children {
		if layout.Children[i].Shape.Float {
			heapLayout := new(Layout)
			*heapLayout = layout.Children[i]
			for j := range heapLayout.Children {
				heapLayout.Children[j].Parent = heapLayout
			}
			*layouts = append(*layouts, heapLayout)
			layoutRemoveFloatingLayouts(heapLayout, layouts)
			layout.Children[i] = layoutPlaceholder()
		} else {
			layoutRemoveFloatingLayouts(&layout.Children[i], layouts)
		}
	}
}
