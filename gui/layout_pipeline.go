package gui

// layoutPipeline runs all layout passes in order on a single
// layout tree.
func layoutPipeline(layout *Layout, w *Window) {
	// Width passes.
	layoutWidths(layout)
	layoutFillWidths(layout)
	layoutWrapContainers(layout)
	layoutOverflow(layout, w)
	layoutWrapText(layout, w)

	// Height passes.
	layoutHeights(layout)
	layoutFillHeights(layout)

	// Position passes.
	layoutAdjustScrollOffsets(layout, w)
	floatAttachLayout(layout)
	layoutPositions(layout, 0, 0, w)
	layoutDisables(layout, false)
	layoutScrollContainers(layout, 0)

	// Post-position passes.
	layoutAmend(layout, w)
	applyLayoutTransition(layout, w)
	applyHeroTransition(layout, w)
	layoutSetShapeClips(layout, w.WindowRect())
}

// layoutAmend walks the layout tree children-first, firing
// AmendLayout callbacks. Not for size changes — post-position
// only.
func layoutAmend(layout *Layout, w *Window) {
	for i := range layout.Children {
		layoutAmend(&layout.Children[i], w)
	}
	if layout.Shape.HasEvents() &&
		layout.Shape.Events.AmendLayout != nil {
		layout.Shape.Events.AmendLayout(layout, w)
	}
}

// layoutHover walks the layout tree depth-first, firing OnHover
// callbacks when the mouse is inside a shape. Returns true if
// any hover was handled.
func layoutHover(layout *Layout, w *Window) bool {
	if w.MouseIsLocked() {
		return false
	}
	for i := len(layout.Children) - 1; i >= 0; i-- {
		if layoutHover(&layout.Children[i], w) {
			return true
		}
	}
	shape := layout.Shape
	if shape.Disabled {
		return false
	}
	if !shape.HasEvents() || shape.Events.OnHover == nil {
		return false
	}
	if !shape.PointInShape(w.viewState.mousePosX,
		w.viewState.mousePosY) {
		return false
	}
	if w.dialogCfg.visible &&
		!layoutInDialogLayout(layout) {
		return false
	}
	e := Event{
		MouseX:    w.viewState.mousePosX,
		MouseY:    w.viewState.mousePosY,
		Type:      EventMouseMove,
		MouseButton: MouseInvalid,
	}
	shape.Events.OnHover(layout, &e, w)
	return true
}

// layoutInDialogLayout walks the parent chain checking if any
// ancestor has ID == reservedDialogID.
func layoutInDialogLayout(layout *Layout) bool {
	for p := layout; p != nil; p = p.Parent {
		if p.Shape.ID == reservedDialogID {
			return true
		}
	}
	return false
}

// layoutWrapText is a no-op stub. Text measurement requires
// backend integration not yet available.
func layoutWrapText(_ *Layout, _ *Window) {}
