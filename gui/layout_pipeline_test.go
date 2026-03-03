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
	w.MouseLock(MouseLockCfg{MouseMove: func(*Layout, *Event, *Window) {}})
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

type stubTextMeasurer struct {
	charWidth  float32
	fontHeight float32
}

func (m *stubTextMeasurer) TextWidth(text string, _ TextStyle) float32 {
	return float32(len(text)) * m.charWidth
}
func (m *stubTextMeasurer) TextHeight(_ string, _ TextStyle) float32 {
	return m.fontHeight
}
func (m *stubTextMeasurer) FontHeight(_ TextStyle) float32 {
	return m.fontHeight
}

func TestLayoutWrapPlainText(t *testing.T) {
	w := &Window{}
	w.textMeasurer = &stubTextMeasurer{charWidth: 10, fontHeight: 20}

	style := TextStyle{Size: 16}
	shape := &Shape{
		ShapeType: ShapeText,
		Width:     100, // fits ~10 chars per line
		TC: &ShapeTextConfig{
			Text:      "hello world this wraps",
			TextStyle: &style,
			TextMode:  TextModeWrap,
		},
	}
	layout := Layout{Shape: shape}
	layoutWrapText(&layout, w)

	// "hello" (50) + " " (10) + "world" (50) = 110 > 100 → 2nd line
	// Each word measured separately; expect multiple lines.
	if shape.Height <= 20 {
		t.Errorf("expected Height > 20 (wrapped), got %f",
			shape.Height)
	}
}

func TestFixedColumnCentering(t *testing.T) {
	w := &Window{}
	w.windowWidth = 300
	w.windowHeight = 300

	// Simulate the get_started Column with centered content
	col := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			Axis:      AxisTopToBottom,
			Width:     300,
			Height:    300,
			Sizing:    FixedFixed,
			MinWidth:  300, MaxWidth: 300,
			MinHeight: 300, MaxHeight: 300,
			HAlign:  HAlignCenter,
			VAlign:  VAlignMiddle,
			Spacing: 10,
			Opacity: 1,
		},
		Children: []Layout{
			{Shape: &Shape{ShapeType: ShapeText, Width: 150, Height: 20, Opacity: 1}},
			{Shape: &Shape{ShapeType: ShapeRectangle, Width: 80, Height: 30, Opacity: 1}},
		},
	}

	layoutPipeline(&col, w)

	// Children should be vertically centered
	// remaining = 300 - 10 (spacing) - 20 - 30 = 240, offset = 120
	text := col.Children[0].Shape
	btn := col.Children[1].Shape

	t.Logf("Text: X=%.1f Y=%.1f W=%.1f H=%.1f", text.X, text.Y, text.Width, text.Height)
	t.Logf("Button: X=%.1f Y=%.1f W=%.1f H=%.1f", btn.X, btn.Y, btn.Width, btn.Height)

	// Vertical: VAlignMiddle
	expectedTextY := float32(120)
	if abs32(text.Y-expectedTextY) > 1 {
		t.Errorf("text Y = %.1f, want ~%.1f", text.Y, expectedTextY)
	}

	// Horizontal: HAlignCenter
	expectedTextX := float32((300 - 150) / 2)
	if abs32(text.X-expectedTextX) > 1 {
		t.Errorf("text X = %.1f, want ~%.1f", text.X, expectedTextX)
	}

	// Now "resize" to 500x500
	w.windowWidth = 500
	w.windowHeight = 500

	col2 := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			Axis:      AxisTopToBottom,
			Width:     500,
			Height:    500,
			Sizing:    FixedFixed,
			MinWidth:  500, MaxWidth: 500,
			MinHeight: 500, MaxHeight: 500,
			HAlign:  HAlignCenter,
			VAlign:  VAlignMiddle,
			Spacing: 10,
			Opacity: 1,
		},
		Children: []Layout{
			{Shape: &Shape{ShapeType: ShapeText, Width: 150, Height: 20, Opacity: 1}},
			{Shape: &Shape{ShapeType: ShapeRectangle, Width: 80, Height: 30, Opacity: 1}},
		},
	}

	layoutPipeline(&col2, w)

	text2 := col2.Children[0].Shape
	btn2 := col2.Children[1].Shape

	t.Logf("After resize - Text: X=%.1f Y=%.1f W=%.1f H=%.1f", text2.X, text2.Y, text2.Width, text2.Height)
	t.Logf("After resize - Button: X=%.1f Y=%.1f W=%.1f H=%.1f", btn2.X, btn2.Y, btn2.Width, btn2.Height)

	// Vertical: remaining = 500 - 10 - 20 - 30 = 440, offset = 220
	expectedTextY2 := float32(220)
	if abs32(text2.Y-expectedTextY2) > 1 {
		t.Errorf("resized text Y = %.1f, want ~%.1f", text2.Y, expectedTextY2)
	}

	// Horizontal: (500 - 150) / 2 = 175
	expectedTextX2 := float32(175)
	if abs32(text2.X-expectedTextX2) > 1 {
		t.Errorf("resized text X = %.1f, want ~%.1f", text2.X, expectedTextX2)
	}
}

func TestTooltipColumnMaxWidthConstrainsText(t *testing.T) {
	w := &Window{}
	w.textMeasurer = &stubTextMeasurer{charWidth: 10, fontHeight: 20}
	w.windowWidth = 800
	w.windowHeight = 600

	longText := "This is a very long tooltip text that should definitely wrap when constrained to max width"
	style := TextStyle{Size: 16}

	// Simulate the tooltip Column with MaxWidth=300 and a
	// FillFit text child (TextModeWrap).
	textShape := &Shape{
		ShapeType: ShapeText,
		Width:     float32(len(longText)) * 10, // ~900px
		Height:    20,
		Sizing:    FillFit,
		Opacity:   1,
		TC: &ShapeTextConfig{
			Text:      longText,
			TextStyle: &style,
			TextMode:  TextModeWrap,
		},
	}
	col := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			Axis:      AxisTopToBottom,
			MaxWidth:  300,
			Opacity:   1,
		},
		Children: []Layout{
			{Shape: textShape},
		},
	}
	col.Children[0].Parent = &col

	layoutPipeline(&col, w)

	if col.Shape.Width > 300 {
		t.Errorf("column width %f exceeds MaxWidth 300",
			col.Shape.Width)
	}
	if textShape.Width > 300 {
		t.Errorf("text width %f exceeds 300", textShape.Width)
	}
	if textShape.Height <= 20 {
		t.Errorf("text height %f not wrapped (expected > 20)",
			textShape.Height)
	}
	t.Logf("col.Width=%.1f text.Width=%.1f text.Height=%.1f",
		col.Shape.Width, textShape.Width, textShape.Height)
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
