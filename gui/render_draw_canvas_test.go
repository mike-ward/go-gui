package gui

import "testing"

func TestRenderDrawCanvasOutsideClipSkips(t *testing.T) {
	w := makeWindowWithScratch()
	shape := &Shape{
		ShapeType: ShapeDrawCanvas,
		X:         500, Y: 500,
		Width: 50, Height: 50,
		Color: RGB(100, 100, 100),
	}
	clip := makeClip(0, 0, 100, 100)

	renderDrawCanvas(shape, clip, w)

	if len(w.renderers) != 0 {
		t.Errorf("got %d renderers, want 0 for out-of-clip canvas",
			len(w.renderers))
	}
}

func TestRenderDrawCanvasCallsOnDraw(t *testing.T) {
	w := makeWindowWithScratch()
	called := false
	shape := &Shape{
		ShapeType: ShapeDrawCanvas,
		Width:     100, Height: 100,
		Color: RGB(100, 100, 100),
		Events: &EventHandlers{
			OnDraw: func(dc *DrawContext) {
				called = true
				dc.batches = append(dc.batches, DrawCanvasTriBatch{
					Triangles: []float32{0, 0, 1, 0, 0, 1},
					Color:     RGB(255, 0, 0),
				})
			},
		},
	}
	clip := makeClip(0, 0, 200, 200)

	renderDrawCanvas(shape, clip, w)

	if !called {
		t.Error("OnDraw not called")
	}
	found := false
	for _, r := range w.renderers {
		if r.Kind == RenderSvg {
			found = true
		}
	}
	if !found {
		t.Error("expected RenderSvg for canvas batch")
	}
}

func TestRenderDrawCanvasEmptyBatchesNoOutput(t *testing.T) {
	w := makeWindowWithScratch()
	shape := &Shape{
		ShapeType: ShapeDrawCanvas,
		Width:     100, Height: 100,
		Color: ColorTransparent,
		Events: &EventHandlers{
			OnDraw: func(_ *DrawContext) {
				// produce no batches
			},
		},
	}
	clip := makeClip(0, 0, 200, 200)

	renderDrawCanvas(shape, clip, w)

	for _, r := range w.renderers {
		if r.Kind == RenderSvg {
			t.Error("should not emit RenderSvg for empty batches")
		}
	}
}

func TestRenderDrawCanvasClipBrackets(t *testing.T) {
	w := makeWindowWithScratch()
	shape := &Shape{
		ShapeType: ShapeDrawCanvas,
		Width:     100, Height: 100,
		Color: RGB(100, 100, 100),
		Clip:  true,
		Events: &EventHandlers{
			OnDraw: func(dc *DrawContext) {
				dc.batches = append(dc.batches, DrawCanvasTriBatch{
					Triangles: []float32{0, 0, 1, 0, 0, 1},
					Color:     RGB(255, 0, 0),
				})
			},
		},
	}
	clip := makeClip(0, 0, 200, 200)

	renderDrawCanvas(shape, clip, w)

	clipCount := 0
	for _, r := range w.renderers {
		if r.Kind == RenderClip {
			clipCount++
		}
	}
	if clipCount < 2 {
		t.Errorf("got %d RenderClip, want >= 2 (push + pop)", clipCount)
	}
}

func TestRenderDrawCanvasCachedSkipsOnDraw(t *testing.T) {
	w := makeWindowWithScratch()
	callCount := 0
	shape := &Shape{
		ShapeType: ShapeDrawCanvas,
		ID:        "test-canvas",
		Width:     100, Height: 100,
		Version: 1,
		Color:   RGB(100, 100, 100),
		Events: &EventHandlers{
			OnDraw: func(dc *DrawContext) {
				callCount++
				dc.batches = append(dc.batches, DrawCanvasTriBatch{
					Triangles: []float32{0, 0, 1, 0, 0, 1},
					Color:     RGB(255, 0, 0),
				})
			},
		},
	}
	clip := makeClip(0, 0, 200, 200)

	// First render — should call OnDraw.
	renderDrawCanvas(shape, clip, w)
	if callCount != 1 {
		t.Fatalf("first render: callCount = %d, want 1", callCount)
	}

	// Second render with same version/dimensions — should skip.
	w.renderers = w.renderers[:0]
	renderDrawCanvas(shape, clip, w)
	if callCount != 1 {
		t.Errorf("cached render: callCount = %d, want 1", callCount)
	}
}

func TestRenderDrawCanvasEmptyIDAlwaysRedraws(t *testing.T) {
	w := makeWindowWithScratch()
	callCount := 0
	shape := &Shape{
		ShapeType: ShapeDrawCanvas,
		ID:        "", // empty ID
		Width:     100, Height: 100,
		Version: 1,
		Color:   RGB(100, 100, 100),
		Events: &EventHandlers{
			OnDraw: func(dc *DrawContext) {
				callCount++
				dc.batches = append(dc.batches, DrawCanvasTriBatch{
					Triangles: []float32{0, 0, 1, 0, 0, 1},
					Color:     RGB(255, 0, 0),
				})
			},
		},
	}
	clip := makeClip(0, 0, 200, 200)

	renderDrawCanvas(shape, clip, w)
	renderDrawCanvas(shape, clip, w)

	if callCount != 2 {
		t.Errorf("callCount = %d, want 2 (empty ID always redraws)",
			callCount)
	}
}
