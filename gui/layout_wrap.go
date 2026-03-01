package gui

// wrapRowRange describes a contiguous range of children in one wrap row.
type wrapRowRange struct {
	start, end int
}

// layoutWrapContainers restructures wrap containers into column-of-rows.
func layoutWrapContainers(layout *Layout) {
	for i := range layout.Children {
		layoutWrapContainers(&layout.Children[i])
	}

	if !layout.Shape.Wrap || layout.Shape.Axis != AxisLeftToRight || len(layout.Children) == 0 {
		return
	}

	available := layout.Shape.Width - layout.Shape.PaddingWidth()
	if available <= 0 {
		return
	}

	spacing := layout.Shape.Spacing

	var rows []wrapRowRange
	rowStart := 0
	var rowWidth float32

	for idx := range layout.Children {
		child := &layout.Children[idx]
		if child.Shape.Float || child.Shape.ShapeType == ShapeNone || child.Shape.OverDraw {
			continue
		}
		childW := child.Shape.Width
		var gap float32
		if rowWidth > 0 {
			gap = spacing
		}
		if rowWidth+gap+childW > available && idx > rowStart {
			rows = append(rows, wrapRowRange{start: rowStart, end: idx})
			rowStart = idx
			rowWidth = 0
		}
		if rowWidth > 0 {
			rowWidth += spacing
		}
		rowWidth += childW
	}
	if rowStart < len(layout.Children) {
		rows = append(rows, wrapRowRange{start: rowStart, end: len(layout.Children)})
	}

	if len(rows) <= 1 {
		return
	}

	layout.Shape.Axis = AxisTopToBottom

	newChildren := make([]Layout, 0, len(rows))
	for _, row := range rows {
		rowChildren := make([]Layout, row.end-row.start)
		copy(rowChildren, layout.Children[row.start:row.end])
		newChildren = append(newChildren, Layout{
			Shape: &Shape{
				ShapeType: ShapeRectangle,
				Axis:      AxisLeftToRight,
				Sizing:    FixedFit,
				Width:     available,
				Spacing:   spacing,
				Color:     Color{A: 0},
				HAlign:    layout.Shape.HAlign,
				VAlign:    layout.Shape.VAlign,
				TextDir:   layout.Shape.TextDir,
			},
			Children: rowChildren,
		})
	}

	layout.Children = newChildren
	for i := range layout.Children {
		layoutParents(&layout.Children[i], layout)
	}
}
