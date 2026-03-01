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
	dc.Polyline([]float32{0, 0}, Blue, 2) // too few points
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
