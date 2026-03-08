package gui

import "testing"

func TestDrawCanvasGenerateLayout(t *testing.T) {
	w := &Window{}
	called := false
	v := DrawCanvas(DrawCanvasCfg{
		ID:      "dc1",
		Version: 1,
		Width:   200,
		Height:  100,
		OnDraw: func(dc *DrawContext) {
			called = true
			dc.FilledRect(0, 0, dc.Width, dc.Height, Red)
		},
	})
	layout := GenerateViewLayout(v, w)
	if !called {
		t.Error("OnDraw not called")
	}
	if layout.Shape.ShapeType != ShapeDrawCanvas {
		t.Errorf("shape type = %d, want ShapeDrawCanvas", layout.Shape.ShapeType)
	}
	if layout.Shape.Width != 200 {
		t.Errorf("width = %f", layout.Shape.Width)
	}
}

func TestDrawCanvasCaching(t *testing.T) {
	w := &Window{}
	callCount := 0
	cfg := DrawCanvasCfg{
		ID:      "dc-cache",
		Version: 1,
		Width:   50,
		Height:  50,
		OnDraw: func(dc *DrawContext) {
			callCount++
			dc.FilledRect(0, 0, 10, 10, Blue)
		},
	}
	// First call: draws.
	v := DrawCanvas(cfg)
	GenerateViewLayout(v, w)
	if callCount != 1 {
		t.Fatalf("first call: count = %d", callCount)
	}
	// Same version: cache hit.
	v = DrawCanvas(cfg)
	GenerateViewLayout(v, w)
	if callCount != 1 {
		t.Errorf("second call with same version: count = %d", callCount)
	}
	// Bump version: redraws.
	cfg.Version = 2
	v = DrawCanvas(cfg)
	GenerateViewLayout(v, w)
	if callCount != 2 {
		t.Errorf("after version bump: count = %d", callCount)
	}
}

func TestDrawCanvasDefaultSizing(t *testing.T) {
	v := DrawCanvas(DrawCanvasCfg{ID: "dc-def"})
	dv := v.(*drawCanvasView)
	if dv.cfg.Sizing != FixedFixed {
		t.Errorf("default sizing = %v, want FixedFixed", dv.cfg.Sizing)
	}
}

func TestDrawCanvasNoOnDraw(t *testing.T) {
	w := &Window{}
	v := DrawCanvas(DrawCanvasCfg{
		ID:     "dc-nil",
		Width:  10,
		Height: 10,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.ShapeType != ShapeDrawCanvas {
		t.Error("expected ShapeDrawCanvas")
	}
}

func TestRenderDrawCanvas(t *testing.T) {
	w := &Window{
		windowWidth: 800, windowHeight: 600,
	}
	v := DrawCanvas(DrawCanvasCfg{
		ID:      "dc-render",
		Version: 1,
		Width:   100,
		Height:  80,
		Clip:    true,
		OnDraw: func(dc *DrawContext) {
			dc.FilledRect(0, 0, 50, 50, Green)
		},
	})
	GenerateViewLayout(v, w)

	shape := &Shape{
		ShapeType: ShapeDrawCanvas,
		ID:        "dc-render",
		X:         10,
		Y:         20,
		Width:     100,
		Height:    80,
		Clip:      true,
	}
	clip := DrawClip{X: 0, Y: 0, Width: 800, Height: 600}
	w.renderers = w.renderers[:0]
	renderDrawCanvas(shape, clip, w)
	// Should have at least: container + clip + svg + restore clip.
	if len(w.renderers) < 2 {
		t.Errorf("renderers = %d, want >= 2", len(w.renderers))
	}
}
