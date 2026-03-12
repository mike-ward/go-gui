package gui

import (
	"math"
	"testing"
)

func TestDrawContextFilledRect(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.FilledRect(10, 20, 30, 40, Red)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// 2 triangles = 12 floats.
	if len(dc.batches[0].Triangles) != 12 {
		t.Errorf("triangles = %d, want 12", len(dc.batches[0].Triangles))
	}
	if dc.batches[0].Color != Red {
		t.Error("color mismatch")
	}
}

func TestDrawContextFilledRectDegenerate(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.FilledRect(0, 0, 0, 10, Red)
	dc.FilledRect(0, 0, 10, 0, Red)
	dc.FilledRect(0, 0, -5, 10, Red)
	if len(dc.batches) != 0 {
		t.Errorf("degenerate rects should produce no batches, got %d", len(dc.batches))
	}
}

func TestDrawContextPolyline(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	// Horizontal line: 2 segments = 2 quads = 4 triangles = 24 floats.
	dc.Polyline([]float32{0, 0, 50, 0, 100, 0}, Blue, 2)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	if len(dc.batches[0].Triangles) != 24 {
		t.Errorf("triangles = %d, want 24", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextPolylineDegenerate(t *testing.T) {
	dc := DrawContext{}
	dc.Polyline([]float32{0, 0}, Blue, 2)       // too few points
	dc.Polyline([]float32{0, 0, 1, 1}, Blue, 0) // zero width
	if len(dc.batches) != 0 {
		t.Errorf("expected 0 batches, got %d", len(dc.batches))
	}
}

func TestDrawContextFilledPolygon(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	// Triangle (3 points = 6 floats) → 1 triangle = 6 floats.
	dc.FilledPolygon([]float32{0, 0, 10, 0, 5, 10}, Green)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d", len(dc.batches))
	}
	if len(dc.batches[0].Triangles) != 6 {
		t.Errorf("triangles = %d, want 6", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextFilledPolygonDegenerate(t *testing.T) {
	dc := DrawContext{}
	dc.FilledPolygon([]float32{0, 0, 1, 1}, Green) // < 6 floats
	if len(dc.batches) != 0 {
		t.Errorf("expected 0 batches, got %d", len(dc.batches))
	}
}

func TestDrawContextBatching(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}

	// Two same-color rects: 1 batch, 24 floats.
	dc.FilledRect(0, 0, 10, 10, Red)
	dc.FilledRect(20, 20, 10, 10, Red)
	if len(dc.batches) != 1 {
		t.Fatalf("consecutive same-color: batches = %d, want 1", len(dc.batches))
	}
	if len(dc.batches[0].Triangles) != 24 {
		t.Errorf("consecutive same-color: triangles = %d, want 24", len(dc.batches[0].Triangles))
	}

	// Different color: new batch.
	dc.FilledRect(0, 0, 10, 10, Blue)
	if len(dc.batches) != 2 {
		t.Fatalf("different color: batches = %d, want 2", len(dc.batches))
	}

	// Same color again: new batch (not consecutive).
	// Actually, my implementation currently only batches *consecutive* calls.
	// Let's verify that.
	dc.FilledRect(40, 40, 10, 10, Red)
	if len(dc.batches) != 3 {
		t.Errorf("non-consecutive same-color: batches = %d, want 3", len(dc.batches))
	}
}

func TestArcToPolyline(t *testing.T) {
	pts := arcToPolyline(50, 50, 25, 25, 0, 2*math.Pi)
	if len(pts) < 8 {
		t.Fatalf("too few arc points: %d", len(pts))
	}
	// First and last points should be close (full circle).
	dx := pts[0] - pts[len(pts)-2]
	dy := pts[1] - pts[len(pts)-1]
	dist := math.Sqrt(float64(dx*dx + dy*dy))
	if dist > 1.0 {
		t.Errorf("full arc not closed: gap = %f", dist)
	}
}

func TestArcToPolylineDegenerate(t *testing.T) {
	pts := arcToPolyline(0, 0, 0, 0, 0, 1)
	if len(pts) != 0 {
		t.Errorf("zero radius should return nil, got %d", len(pts))
	}
}

func TestDrawContextFilledCircle(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.FilledCircle(50, 50, 25, Red)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// Full circle fan: n segments × 6 floats each.
	if len(dc.batches[0].Triangles) < 24 {
		t.Errorf("too few triangle floats: %d", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextCircle(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.Circle(50, 50, 25, Blue, 2)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	if len(dc.batches[0].Triangles) < 24 {
		t.Errorf("too few triangle floats: %d", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextArc(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.Arc(50, 50, 25, 25, 0, math.Pi, Green, 2)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	if len(dc.batches[0].Triangles) < 12 {
		t.Errorf("too few triangle floats: %d", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextFilledArc(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.FilledArc(50, 50, 25, 25, 0, math.Pi, Red)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// Half-circle fan: at least 3 triangles.
	if len(dc.batches[0].Triangles) < 18 {
		t.Errorf("too few triangle floats: %d", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextRect(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.Rect(10, 20, 50, 30, Red, 2)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// 4 edge quads × 2 triangles × 6 floats = 48.
	if len(dc.batches[0].Triangles) != 48 {
		t.Errorf("triangles = %d, want 48", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextRectDegenerate(t *testing.T) {
	dc := DrawContext{}
	dc.Rect(0, 0, 0, 10, Red, 1)
	dc.Rect(0, 0, 10, 0, Red, 1)
	dc.Rect(0, 0, 10, 10, Red, 0)
	if len(dc.batches) != 0 {
		t.Errorf("degenerate rects: batches = %d, want 0", len(dc.batches))
	}
}
