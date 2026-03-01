package gui

import "testing"

func TestLayoutPipelineNoPanic(t *testing.T) {
	w := &Window{}
	shape := &Shape{
		ShapeType: ShapeRectangle,
		Width:     100, Height: 100,
		Sizing: FillFill,
		Opacity: 1,
	}
	layout := Layout{
		Shape: shape,
		Children: []Layout{
			{Shape: &Shape{ShapeType: ShapeRectangle, Width: 50, Height: 50, Opacity: 1}},
			{Shape: &Shape{ShapeType: ShapeRectangle, Width: 30, Height: 30, Opacity: 1}},
		},
	}
	w.windowWidth = 200
	w.windowHeight = 200
	layoutPipeline(&layout, w)
}

func TestLayoutAmendFiresChildrenFirst(t *testing.T) {
	w := &Window{}
	var order []string

	child := &Shape{
		ShapeType: ShapeRectangle,
		Events: &EventHandlers{
			AmendLayout: func(_ *Layout, _ *Window) {
				order = append(order, "child")
			},
		},
		Opacity: 1,
	}
	parent := &Shape{
		ShapeType: ShapeRectangle,
		Events: &EventHandlers{
			AmendLayout: func(_ *Layout, _ *Window) {
				order = append(order, "parent")
			},
		},
		Opacity: 1,
	}
	layout := Layout{
		Shape: parent,
		Children: []Layout{
			{Shape: child},
		},
	}
	layoutAmend(&layout, w)

	if len(order) != 2 {
		t.Fatalf("expected 2 callbacks, got %d", len(order))
	}
	if order[0] != "child" || order[1] != "parent" {
		t.Errorf("expected [child parent], got %v", order)
	}
}

func TestLayoutHoverInsideShape(t *testing.T) {
	w := &Window{}
	w.windowWidth = 200
	w.windowHeight = 200
	w.viewState.mousePosX = 15
	w.viewState.mousePosY = 15

	hovered := false
	shape := &Shape{
		ShapeType: ShapeRectangle,
		ShapeClip: DrawClip{X: 0, Y: 0, Width: 50, Height: 50},
		Events: &EventHandlers{
			OnHover: func(_ *Layout, _ *Event, _ *Window) {
				hovered = true
			},
		},
		Opacity: 1,
	}
	layout := Layout{Shape: shape}
	result := layoutHover(&layout, w)

	if !result {
		t.Error("expected hover to return true")
	}
	if !hovered {
		t.Error("expected OnHover to fire")
	}
}

func TestLayoutHoverOutsideShape(t *testing.T) {
	w := &Window{}
	w.viewState.mousePosX = 100
	w.viewState.mousePosY = 100

	hovered := false
	shape := &Shape{
		ShapeType: ShapeRectangle,
		ShapeClip: DrawClip{X: 0, Y: 0, Width: 50, Height: 50},
		Events: &EventHandlers{
			OnHover: func(_ *Layout, _ *Event, _ *Window) {
				hovered = true
			},
		},
		Opacity: 1,
	}
	layout := Layout{Shape: shape}
	result := layoutHover(&layout, w)

	if result {
		t.Error("expected hover to return false")
	}
	if hovered {
		t.Error("expected OnHover not to fire")
	}
}

func TestLayoutHoverMouseLocked(t *testing.T) {
	w := &Window{}
	w.viewState.mouseLocked = true
	w.viewState.mousePosX = 15
	w.viewState.mousePosY = 15

	shape := &Shape{
		ShapeType: ShapeRectangle,
		ShapeClip: DrawClip{X: 0, Y: 0, Width: 50, Height: 50},
		Events: &EventHandlers{
			OnHover: func(_ *Layout, _ *Event, _ *Window) {},
		},
		Opacity: 1,
	}
	layout := Layout{Shape: shape}
	result := layoutHover(&layout, w)
	if result {
		t.Error("expected false when mouse locked")
	}
}

func TestLayoutHoverBlockedByDialog(t *testing.T) {
	w := &Window{}
	w.viewState.mousePosX = 15
	w.viewState.mousePosY = 15
	w.dialogCfg.visible = true

	hovered := false
	shape := &Shape{
		ShapeType: ShapeRectangle,
		ShapeClip: DrawClip{X: 0, Y: 0, Width: 50, Height: 50},
		Events: &EventHandlers{
			OnHover: func(_ *Layout, _ *Event, _ *Window) {
				hovered = true
			},
		},
		Opacity: 1,
	}
	// Shape NOT inside a dialog layout.
	layout := Layout{Shape: shape}
	result := layoutHover(&layout, w)

	if result {
		t.Error("expected false outside dialog")
	}
	if hovered {
		t.Error("should not hover outside dialog")
	}
}

func TestLayoutInDialogLayoutPositive(t *testing.T) {
	dialog := &Shape{ID: reservedDialogID, ShapeType: ShapeRectangle}
	child := &Shape{ShapeType: ShapeRectangle}
	parent := Layout{Shape: dialog}
	childLayout := Layout{Shape: child, Parent: &parent}

	if !layoutInDialogLayout(&childLayout) {
		t.Error("expected true when inside dialog")
	}
}

func TestLayoutInDialogLayoutNegative(t *testing.T) {
	outer := &Shape{ID: "something_else", ShapeType: ShapeRectangle}
	child := &Shape{ShapeType: ShapeRectangle}
	parent := Layout{Shape: outer}
	childLayout := Layout{Shape: child, Parent: &parent}

	if layoutInDialogLayout(&childLayout) {
		t.Error("expected false outside dialog")
	}
}

func TestLayoutArrangeReturnsLayers(t *testing.T) {
	w := &Window{}
	w.windowWidth = 400
	w.windowHeight = 400
	shape := &Shape{
		ShapeType: ShapeRectangle,
		Width:     400, Height: 400,
		Sizing:  FillFill,
		Opacity: 1,
	}
	layout := Layout{Shape: shape}
	layouts := layoutArrange(&layout, w)

	if len(layouts) < 1 {
		t.Fatal("expected at least main layout")
	}
}

func TestLayoutArrangeWithDialog(t *testing.T) {
	w := &Window{}
	w.windowWidth = 400
	w.windowHeight = 400
	w.Dialog(DialogCfg{
		Title: "Test",
		Body:  "Body",
	})

	shape := &Shape{
		ShapeType: ShapeRectangle,
		Width:     400, Height: 400,
		Sizing:  FillFill,
		Opacity: 1,
	}
	layout := Layout{Shape: shape}
	layouts := layoutArrange(&layout, w)

	if len(layouts) < 2 {
		t.Errorf("expected >= 2 layers with dialog, got %d",
			len(layouts))
	}
}

func TestLayoutArrangeWithToast(t *testing.T) {
	w := &Window{}
	w.windowWidth = 400
	w.windowHeight = 400
	// Manually add a toast (skip animation).
	w.toasts = append(w.toasts, toastNotification{
		id:       1,
		cfg:      ToastCfg{Title: "Hi", Body: "World"},
		animFrac: 1.0,
		phase:    toastVisible,
	})

	shape := &Shape{
		ShapeType: ShapeRectangle,
		Width:     400, Height: 400,
		Sizing:  FillFill,
		Opacity: 1,
	}
	layout := Layout{Shape: shape}
	layouts := layoutArrange(&layout, w)

	if len(layouts) < 2 {
		t.Errorf("expected >= 2 layers with toast, got %d",
			len(layouts))
	}
}

func TestLayoutWrapTextNoop(t *testing.T) {
	// Should not panic.
	layoutWrapText(nil, nil)
}
