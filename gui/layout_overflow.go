package gui

// layoutOverflow hides children that don't fit in an overflow container.
func layoutOverflow(layout *Layout, w *Window) {
	for i := range layout.Children {
		layoutOverflow(&layout.Children[i], w)
	}

	if !layout.Shape.Overflow || layout.Shape.Axis != AxisLeftToRight ||
		len(layout.Children) < 2 || layout.Shape.IDScroll > 0 {
		return
	}

	available := layout.Shape.Width - layout.Shape.PaddingWidth()
	spacing := layout.Shape.Spacing

	// Find the trigger button (last non-float, non-placeholder child).
	triggerIdx := len(layout.Children) - 1
	for triggerIdx > 0 && (layout.Children[triggerIdx].Shape.Float ||
		layout.Children[triggerIdx].Shape.ShapeType == ShapeNone ||
		layout.Children[triggerIdx].Shape.OverDraw) {
		triggerIdx--
	}
	triggerW := layout.Children[triggerIdx].Shape.Width

	var used float32
	visibleCount := 0

	for i := range triggerIdx {
		child := &layout.Children[i]
		if child.Shape.Float || child.Shape.ShapeType == ShapeNone ||
			child.Shape.OverDraw {
			continue
		}
		var gap float32
		if used > 0 {
			gap = spacing
		}
		needed := used + gap + child.Shape.Width
		if needed+spacing+triggerW > available {
			break
		}
		used = needed
		visibleCount++
	}

	if visibleCount >= triggerIdx {
		hideOverflowChild(&layout.Children[triggerIdx])
		visibleCount = triggerIdx
	} else {
		for i := visibleCount; i < triggerIdx; i++ {
			hideOverflowChild(&layout.Children[i])
		}
	}

	om := StateMap[string, int](w, nsOverflow, capModerate)
	old, ok := om.Get(layout.Shape.ID)
	if !ok {
		old = -1
	}
	if old != visibleCount {
		om.Set(layout.Shape.ID, visibleCount)
		ss := StateMap[string, bool](w, nsSelect, capModerate)
		ss.Delete(layout.Shape.ID)
		w.refreshLayout = true
	}
}

func hideOverflowChild(child *Layout) {
	child.Shape.ShapeType = ShapeNone
	child.Shape.Width = 0
	child.Shape.Clip = true
}
