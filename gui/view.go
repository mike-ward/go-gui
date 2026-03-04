package gui

// View is a user-defined view. Views are never displayed
// directly. A Layout is generated from the View. Window does
// not hold a reference to a View. Views should be stateless.
//
// Pipeline: View -> GenerateViewLayout -> Layout ->
// layoutArrange -> renderLayout -> []RenderCmd
type View interface {
	Content() []View
	GenerateLayout(w *Window) Layout
}

// ensureLayoutShape normalizes layout nodes so pipeline passes can
// safely dereference Shape fields.
func ensureLayoutShape(layout *Layout) {
	if layout == nil {
		return
	}
	if layout.Shape == nil {
		layout.Shape = &Shape{ShapeType: ShapeNone}
	}
}

// GenerateViewLayout recursively builds a Layout tree from a
// View tree. Each View produces its own layout, then child
// Views are appended.
func GenerateViewLayout(view View, w *Window) Layout {
	layout := view.GenerateLayout(w)
	ensureLayoutShape(&layout)
	for _, child := range view.Content() {
		layout.Children = append(
			layout.Children,
			GenerateViewLayout(child, w),
		)
	}
	return layout
}
