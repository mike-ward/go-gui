package gui

import (
	"math"
	"testing"
)

func makeScrollLayout(idScroll uint32, width, height float32, contentW, contentH float32) (*Layout, *Window) {
	w := &Window{}
	// Build a parent with one oversized child to create scrollable content.
	child := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			Width:     contentW,
			Height:    contentH,
		},
	}
	layout := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			IDScroll:  idScroll,
			Width:     width,
			Height:    height,
			Axis:      AxisTopToBottom,
		},
		Children: []Layout{child},
	}
	w.layout = Layout{
		Shape:    &Shape{ShapeType: ShapeRectangle},
		Children: []Layout{layout},
	}
	return &w.layout.Children[0], w
}

func TestScrollVerticalClampsWithinBounds(t *testing.T) {
	layout, w := makeScrollLayout(1, 100, 100, 100, 300)
	guiTheme.ScrollMultiplier = 1

	// Scroll down (negative delta).
	ok := scrollVertical(layout, -50, w)
	if !ok {
		t.Fatal("expected true")
	}
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	v, _ := sy.Get(uint32(1))
	if v != -50 {
		t.Errorf("expected -50, got %v", v)
	}

	// Over-scroll should clamp.
	scrollVertical(layout, -500, w)
	v, _ = sy.Get(uint32(1))
	// max offset = 100 - 0 - 300 = -200
	if v != -200 {
		t.Errorf("expected -200, got %v", v)
	}
}

func TestScrollHorizontalClampsWithinBounds(t *testing.T) {
	w := &Window{}
	guiTheme.ScrollMultiplier = 1
	child := Layout{
		Shape: &Shape{ShapeType: ShapeRectangle, Width: 400, Height: 50},
	}
	layout := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			IDScroll:  2,
			Width:     100,
			Height:    50,
			Axis:      AxisLeftToRight,
		},
		Children: []Layout{child},
	}
	w.layout = Layout{
		Shape:    &Shape{ShapeType: ShapeRectangle},
		Children: []Layout{layout},
	}

	ok := scrollHorizontal(&w.layout.Children[0], -50, w)
	if !ok {
		t.Fatal("expected true")
	}
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	v, _ := sx.Get(uint32(2))
	if v != -50 {
		t.Errorf("expected -50, got %v", v)
	}

	// Over-scroll clamp: max = 100 - 0 - 400 = -300
	scrollHorizontal(&w.layout.Children[0], -500, w)
	v, _ = sx.Get(uint32(2))
	if v != -300 {
		t.Errorf("expected -300, got %v", v)
	}
}

func TestScrollVerticalNoIDScroll(t *testing.T) {
	layout := &Layout{Shape: &Shape{IDScroll: 0}}
	w := &Window{}
	if scrollVertical(layout, -10, w) {
		t.Error("should return false with IDScroll=0")
	}
}

func TestScrollHorizontalNoIDScroll(t *testing.T) {
	layout := &Layout{Shape: &Shape{IDScroll: 0}}
	w := &Window{}
	if scrollHorizontal(layout, -10, w) {
		t.Error("should return false with IDScroll=0")
	}
}

func TestScrollToView(t *testing.T) {
	w := &Window{}
	// Build layout: scroll container > target child.
	target := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			ID:        "target",
			Y:         150,
			Height:    20,
		},
	}
	scroll := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			IDScroll:  5,
			Y:         10,
			Height:    100,
			Padding:   Padding{Top: 5},
		},
		Children: []Layout{target},
	}
	scroll.Children[0].Parent = &scroll
	w.layout = Layout{
		Shape:    &Shape{ShapeType: ShapeRectangle},
		Children: []Layout{scroll},
	}
	w.layout.Children[0].Parent = &w.layout

	w.ScrollToView("target")
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	v, ok := sy.Get(uint32(5))
	if !ok {
		t.Fatal("scroll offset not set")
	}
	// baseY = 10 + 5 = 15; newScroll = 15 - 150 + 0 = -135
	if v != -135 {
		t.Errorf("expected -135, got %v", v)
	}
}

func TestScrollToViewNotFound(t *testing.T) {
	_ = t
	w := &Window{}
	w.layout = Layout{Shape: &Shape{ShapeType: ShapeRectangle}}
	w.ScrollToView("nonexistent") // should not panic
}

func TestScrollHorizontalByAndTo(t *testing.T) {
	w := &Window{}
	w.ScrollHorizontalBy(10, -30)
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	v, _ := sx.Get(uint32(10))
	if v != -30 {
		t.Errorf("expected -30, got %v", v)
	}

	w.ScrollHorizontalTo(10, -100)
	v, _ = sx.Get(uint32(10))
	if v != -100 {
		t.Errorf("expected -100, got %v", v)
	}
}

func TestScrollVerticalByAndTo(t *testing.T) {
	w := &Window{}
	w.ScrollVerticalBy(10, -20)
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	v, _ := sy.Get(uint32(10))
	if v != -20 {
		t.Errorf("expected -20, got %v", v)
	}

	w.ScrollVerticalTo(10, -50)
	v, _ = sy.Get(uint32(10))
	if v != -50 {
		t.Errorf("expected -50, got %v", v)
	}
}

func TestScrollVerticalToPctAndPct(t *testing.T) {
	_, w := makeScrollLayout(3, 100, 100, 100, 300)

	w.ScrollVerticalToPct(3, 0.5)
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	v, _ := sy.Get(uint32(3))
	// maxOffset = -200, 50% = -100
	if v != -100 {
		t.Errorf("expected -100, got %v", v)
	}

	pct := w.ScrollVerticalPct(3)
	if math.Abs(float64(pct-0.5)) > 0.01 {
		t.Errorf("expected ~0.5, got %v", pct)
	}
}

func TestScrollHorizontalToPctAndPct(t *testing.T) {
	w := &Window{}
	child := Layout{
		Shape: &Shape{ShapeType: ShapeRectangle, Width: 400, Height: 50},
	}
	layout := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			IDScroll:  4,
			Width:     100,
			Height:    50,
			Axis:      AxisLeftToRight,
		},
		Children: []Layout{child},
	}
	w.layout = Layout{
		Shape:    &Shape{ShapeType: ShapeRectangle},
		Children: []Layout{layout},
	}

	w.ScrollHorizontalToPct(4, 1.0)
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	v, _ := sx.Get(uint32(4))
	// maxOffset = 100 - 0 - 400 = -300
	if v != -300 {
		t.Errorf("expected -300, got %v", v)
	}

	pct := w.ScrollHorizontalPct(4)
	if math.Abs(float64(pct-1.0)) > 0.01 {
		t.Errorf("expected ~1.0, got %v", pct)
	}
}

func TestScrollPctNoScrollNeeded(t *testing.T) {
	// Content fits viewport — pct should return 0.
	_, w := makeScrollLayout(6, 200, 200, 100, 100)
	pct := w.ScrollVerticalPct(6)
	if pct != 0 {
		t.Errorf("expected 0, got %v", pct)
	}
	pct = w.ScrollHorizontalPct(6)
	if pct != 0 {
		t.Errorf("expected 0, got %v", pct)
	}
}

func TestScrollVerticalFiresOnScroll(t *testing.T) {
	guiTheme.ScrollMultiplier = 1
	fired := false
	layout, w := makeScrollLayout(7, 100, 100, 100, 300)
	layout.Shape.Events = &EventHandlers{
		OnScroll: func(_ *Layout, _ *Window) { fired = true },
	}
	scrollVertical(layout, -10, w)
	if !fired {
		t.Error("OnScroll callback not fired")
	}
}
