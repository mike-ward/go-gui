package gui

import "testing"

func TestDrawCanvasGenerateLayout(t *testing.T) {
	w := &Window{}
	v := DrawCanvas(DrawCanvasCfg{
		ID:      "dc1",
		Version: 1,
		Width:   200,
		Height:  100,
		OnDraw: func(dc *DrawContext) {
			dc.FilledRect(0, 0, dc.Width, dc.Height, Red)
		},
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.ShapeType != ShapeDrawCanvas {
		t.Errorf("shape type = %d, want ShapeDrawCanvas", layout.Shape.ShapeType)
	}
	if layout.Shape.Width != 200 {
		t.Errorf("width = %f", layout.Shape.Width)
	}
	if layout.Shape.Events.OnDraw == nil {
		t.Error("OnDraw not set on shape")
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
	clip := DrawClip{Width: 100, Height: 100}

	// First call: draws.
	v := DrawCanvas(cfg)
	layout := GenerateViewLayout(v, w)
	renderDrawCanvas(layout.Shape, clip, w)
	if callCount != 1 {
		t.Fatalf("first call: count = %d", callCount)
	}

	// Same version: cache hit.
	v = DrawCanvas(cfg)
	layout = GenerateViewLayout(v, w)
	renderDrawCanvas(layout.Shape, clip, w)
	if callCount != 1 {
		t.Errorf("second call with same version: count = %d", callCount)
	}

	// Bump version: redraws.
	cfg.Version = 2
	v = DrawCanvas(cfg)
	layout = GenerateViewLayout(v, w)
	renderDrawCanvas(layout.Shape, clip, w)
	if callCount != 2 {
		t.Errorf("after version bump: count = %d", callCount)
	}
}

func TestDrawCanvasResizeRedraw(t *testing.T) {
	w := &Window{}
	callCount := 0
	lastWidth := float32(0)
	cfg := DrawCanvasCfg{
		ID:      "dc-resize",
		Version: 1,
		Width:   50,
		Height:  50,
		OnDraw: func(dc *DrawContext) {
			callCount++
			lastWidth = dc.Width
		},
	}
	clip := DrawClip{Width: 100, Height: 100}

	// First draw.
	v := DrawCanvas(cfg)
	layout := GenerateViewLayout(v, w)
	renderDrawCanvas(layout.Shape, clip, w)
	if callCount != 1 || lastWidth != 50 {
		t.Fatalf("first draw: count=%d, width=%f", callCount, lastWidth)
	}

	// Change width (simulate layout engine change).
	layout.Shape.Width = 80
	renderDrawCanvas(layout.Shape, clip, w)
	if callCount != 2 || lastWidth != 80 {
		t.Errorf("after resize: count=%d, width=%f", callCount, lastWidth)
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
	layout := GenerateViewLayout(v, w)

	clip := DrawClip{X: 0, Y: 0, Width: 800, Height: 600}
	w.renderers = w.renderers[:0]
	renderDrawCanvas(layout.Shape, clip, w)
	// Should have: container + clip + svg + restore clip.
	if len(w.renderers) < 3 {
		t.Errorf("renderers = %d, want >= 3", len(w.renderers))
	}
}
