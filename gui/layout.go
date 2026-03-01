package gui

// Layout defines a tree of layouts. Views generate Layouts.
type Layout struct {
	Shape    *Shape
	Parent   *Layout
	Children []Layout
}

// layoutParents sets the parent pointer of all nodes.
func layoutParents(layout *Layout, parent *Layout) {
	layout.Parent = parent
	for i := range layout.Children {
		layoutParents(&layout.Children[i], layout)
	}
}

// layoutDisables walks the Layout and disables children that
// have a disabled ancestor.
func layoutDisables(layout *Layout, disabled bool) {
	isDisabled := disabled || layout.Shape.Disabled
	layout.Shape.Disabled = isDisabled
	for i := range layout.Children {
		layoutDisables(&layout.Children[i], isDisabled)
	}
}

// layoutPlaceholder returns an empty placeholder Layout.
func layoutPlaceholder() Layout {
	return Layout{
		Shape: &Shape{ShapeType: ShapeNone},
	}
}

// spacing does the fence-post calculation for spacings.
func (layout *Layout) spacing() float32 {
	count := 0
	for i := range layout.Children {
		c := &layout.Children[i]
		if c.Shape.Float || c.Shape.ShapeType == ShapeNone || c.Shape.OverDraw {
			continue
		}
		count++
	}
	return float32(intMax(0, count-1)) * layout.Shape.Spacing
}

// contentWidth returns total content width.
func contentWidth(layout *Layout) float32 {
	var width float32
	if layout.Shape.Axis == AxisLeftToRight {
		width += layout.spacing()
		for i := range layout.Children {
			if layout.Children[i].Shape.OverDraw {
				continue
			}
			width += layout.Children[i].Shape.Width
		}
	} else {
		for i := range layout.Children {
			if layout.Children[i].Shape.OverDraw {
				continue
			}
			width = f32Max(width, layout.Children[i].Shape.Width)
		}
	}
	return width
}

// contentHeight returns total content height.
func contentHeight(layout *Layout) float32 {
	var height float32
	if layout.Shape.Axis == AxisTopToBottom {
		height += layout.spacing()
		for i := range layout.Children {
			if layout.Children[i].Shape.OverDraw {
				continue
			}
			height += layout.Children[i].Shape.Height
		}
	} else {
		for i := range layout.Children {
			if layout.Children[i].Shape.OverDraw {
				continue
			}
			height = f32Max(height, layout.Children[i].Shape.Height)
		}
	}
	return height
}
