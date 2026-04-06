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

// arcToPolyline converts an elliptical arc to a flat x,y
// polyline via angular subdivision.
func arcToPolyline(cx, cy, rx, ry, start, sweep float32) []float32 {
	r := rx
	r = max(r, ry)
	if r <= 0 {
		return nil
	}
	n := int(math.Ceil(
		float64(f32Abs(sweep)) / (2 * math.Pi) * 64 *
			math.Sqrt(float64(r)/50+1)))
	n = max(n, 4)
	step := sweep / float32(n)
	pts := make([]float32, 0, (n+1)*2)
	for i := 0; i <= n; i++ {
		a := float64(start + step*float32(i))
		pts = append(pts,
			cx+rx*float32(math.Cos(a)),
			cy+ry*float32(math.Sin(a)))
	}
	return pts
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

// --- FilledRoundedRect ---

func TestDrawContextFilledRoundedRect(t *testing.T) {
	dc := DrawContext{Width: 200, Height: 200}
	dc.FilledRoundedRect(10, 10, 80, 60, 10, Red)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// 3 rects (6 tris) + 4 corners × 8 segs = 32+6 = 38 tris.
	want := 38 * 6
	if len(dc.batches[0].Triangles) != want {
		t.Errorf("triangles = %d, want %d", len(dc.batches[0].Triangles), want)
	}
}

func TestDrawContextFilledRoundedRectZeroRadius(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.FilledRoundedRect(0, 0, 40, 30, 0, Blue)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// Falls back to plain FilledRect: 2 triangles = 12 floats.
	if len(dc.batches[0].Triangles) != 12 {
		t.Errorf("triangles = %d, want 12", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextFilledRoundedRectRadiusClamped(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	// Radius exceeds half the smaller dimension (20/2=10).
	dc.FilledRoundedRect(0, 0, 40, 20, 50, Green)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// Should still produce rounded rect output (not degenerate).
	if len(dc.batches[0].Triangles) == 0 {
		t.Error("expected triangles for clamped radius")
	}
}

func TestDrawContextFilledRoundedRectDegenerate(t *testing.T) {
	dc := DrawContext{}
	dc.FilledRoundedRect(0, 0, 0, 10, 5, Red)
	dc.FilledRoundedRect(0, 0, 10, 0, 5, Red)
	dc.FilledRoundedRect(0, 0, -5, 10, 5, Red)
	if len(dc.batches) != 0 {
		t.Errorf("degenerate: batches = %d, want 0", len(dc.batches))
	}
}

// --- RoundedRect (stroked) ---

func TestDrawContextRoundedRect(t *testing.T) {
	dc := DrawContext{Width: 200, Height: 200}
	dc.RoundedRect(10, 10, 80, 60, 10, Red, 2)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	if len(dc.batches[0].Triangles) < 24 {
		t.Errorf("too few triangles: %d", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextRoundedRectZeroRadius(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.RoundedRect(0, 0, 40, 30, 0, Blue, 1)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// Falls back to plain Rect: 4 quads × 2 tris × 6 = 48.
	if len(dc.batches[0].Triangles) != 48 {
		t.Errorf("triangles = %d, want 48", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextRoundedRectDegenerate(t *testing.T) {
	dc := DrawContext{}
	dc.RoundedRect(0, 0, 0, 10, 5, Red, 1)
	dc.RoundedRect(0, 0, 10, 0, 5, Red, 1)
	dc.RoundedRect(0, 0, 10, 10, 5, Red, 0)
	if len(dc.batches) != 0 {
		t.Errorf("degenerate: batches = %d, want 0", len(dc.batches))
	}
}

// --- DashedLine ---

func TestDrawContextDashedLine(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	// Horizontal line of length 100, dash=20, gap=10.
	// Pattern repeats: 20+10=30, fits 3 full + partial.
	// Dashes: [0,20], [30,50], [60,80], [90,100] = 4 dashes.
	dc.DashedLine(0, 50, 100, 50, Red, 2, 20, 10)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// 4 dash segments × 1 quad × 2 tris × 6 floats = 48.
	if len(dc.batches[0].Triangles) != 48 {
		t.Errorf("triangles = %d, want 48", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextDashedLineFallback(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	// Zero dashLen → falls back to solid line.
	dc.DashedLine(0, 0, 50, 0, Blue, 2, 0, 5)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// Solid line: 1 quad = 12 floats.
	if len(dc.batches[0].Triangles) != 12 {
		t.Errorf("triangles = %d, want 12", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextDashedLineZeroLength(t *testing.T) {
	dc := DrawContext{}
	dc.DashedLine(50, 50, 50, 50, Red, 2, 10, 5)
	if len(dc.batches) != 0 {
		t.Errorf("zero-length: batches = %d, want 0", len(dc.batches))
	}
}

// --- DashedPolyline ---

func TestDrawContextDashedPolyline(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	// L-shape: (0,0)→(50,0)→(50,50), total length = 100.
	dc.DashedPolyline(
		[]float32{0, 0, 50, 0, 50, 50},
		Red, 2, 20, 10,
	)
	if len(dc.batches) == 0 {
		t.Fatal("expected at least 1 batch")
	}
	if len(dc.batches[0].Triangles) < 12 {
		t.Error("too few triangles for dashed polyline")
	}
}

func TestDrawContextDashedPolylineFallback(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.DashedPolyline(
		[]float32{0, 0, 50, 0, 100, 0},
		Blue, 2, 0, 5,
	)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// Falls back to solid Polyline: 2 segments × 12 = 24.
	if len(dc.batches[0].Triangles) != 24 {
		t.Errorf("triangles = %d, want 24", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextDashedPolylineTooFewPoints(t *testing.T) {
	dc := DrawContext{}
	dc.DashedPolyline([]float32{0, 0}, Red, 2, 10, 5)
	if len(dc.batches) != 0 {
		t.Errorf("batches = %d, want 0", len(dc.batches))
	}
}

// --- PolylineJoined ---

func TestDrawContextPolylineJoined(t *testing.T) {
	dc := DrawContext{Width: 200, Height: 200}
	// V-shape: 3 points = 2 segments.
	dc.PolylineJoined(
		[]float32{0, 0, 50, 50, 100, 0},
		Red, 4,
	)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	// 2 segments × 2 triangles × 6 floats = 24.
	if len(dc.batches[0].Triangles) != 24 {
		t.Errorf("triangles = %d, want 24", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextPolylineJoinedStraight(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	// Collinear points: miter should be same as simple polyline.
	dc.PolylineJoined(
		[]float32{0, 50, 50, 50, 100, 50},
		Blue, 2,
	)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	if len(dc.batches[0].Triangles) != 24 {
		t.Errorf("triangles = %d, want 24", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextPolylineJoinedDegenerate(t *testing.T) {
	dc := DrawContext{}
	dc.PolylineJoined([]float32{0, 0}, Red, 2)       // too few
	dc.PolylineJoined([]float32{0, 0, 1, 1}, Red, 0) // zero width
	if len(dc.batches) != 0 {
		t.Errorf("degenerate: batches = %d, want 0", len(dc.batches))
	}
}

func TestDrawContextPolylineJoinedSharpAngle(t *testing.T) {
	dc := DrawContext{Width: 200, Height: 200}
	// Sharp hairpin: should trigger miter limit → bevel.
	dc.PolylineJoined(
		[]float32{0, 0, 50, 0, 0, 1},
		Green, 4,
	)
	if len(dc.batches) == 0 {
		t.Fatal("expected output for sharp-angle polyline")
	}
	if len(dc.batches[0].Triangles) != 24 {
		t.Errorf("triangles = %d, want 24", len(dc.batches[0].Triangles))
	}
}

// --- QuadBezier ---

func TestDrawContextQuadBezier(t *testing.T) {
	dc := DrawContext{Width: 200, Height: 200}
	dc.QuadBezier(0, 0, 50, 100, 100, 0, Red, 2)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	if len(dc.batches[0].Triangles) < 12 {
		t.Errorf("too few triangles: %d", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextQuadBezierNaN(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	dc.QuadBezier(nan, 0, 50, 100, 100, 0, Red, 2)
	dc.QuadBezier(0, 0, inf, 100, 100, 0, Red, 2)
	if len(dc.batches) != 0 {
		t.Errorf("NaN/Inf: batches = %d, want 0", len(dc.batches))
	}
}

func TestDrawContextQuadBezierZeroWidth(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.QuadBezier(0, 0, 50, 100, 100, 0, Red, 0)
	if len(dc.batches) != 0 {
		t.Errorf("zero width: batches = %d, want 0", len(dc.batches))
	}
}

// --- CubicBezier ---

func TestDrawContextCubicBezier(t *testing.T) {
	dc := DrawContext{Width: 200, Height: 200}
	dc.CubicBezier(0, 0, 30, 100, 70, 100, 100, 0, Red, 2)
	if len(dc.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(dc.batches))
	}
	if len(dc.batches[0].Triangles) < 12 {
		t.Errorf("too few triangles: %d", len(dc.batches[0].Triangles))
	}
}

func TestDrawContextCubicBezierNaN(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	nan := float32(math.NaN())
	inf := float32(math.Inf(-1))
	dc.CubicBezier(0, 0, nan, 100, 70, 100, 100, 0, Red, 2)
	dc.CubicBezier(0, 0, 30, 100, 70, inf, 100, 0, Red, 2)
	if len(dc.batches) != 0 {
		t.Errorf("NaN/Inf: batches = %d, want 0", len(dc.batches))
	}
}

func TestDrawContextCubicBezierZeroWidth(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	dc.CubicBezier(0, 0, 30, 100, 70, 100, 100, 0, Red, 0)
	if len(dc.batches) != 0 {
		t.Errorf("zero width: batches = %d, want 0", len(dc.batches))
	}
}

// --- Text ---

func TestDrawContextText(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	style := TextStyle{Size: 12, Family: "sans-serif"}
	dc.Text(10, 20, "hello", style)
	if len(dc.texts) != 1 {
		t.Fatalf("texts = %d, want 1", len(dc.texts))
	}
	if dc.texts[0].Text != "hello" {
		t.Errorf("text = %q, want %q", dc.texts[0].Text, "hello")
	}
	if dc.texts[0].X != 10 || dc.texts[0].Y != 20 {
		t.Errorf("position = (%v,%v), want (10,20)",
			dc.texts[0].X, dc.texts[0].Y)
	}
}

func TestDrawContextTextWidthNoMeasurer(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	w := dc.TextWidth("hello", TextStyle{Size: 12})
	if w != 0 {
		t.Errorf("TextWidth without measurer = %v, want 0", w)
	}
}

func TestDrawContextFontHeightNoMeasurer(t *testing.T) {
	dc := DrawContext{Width: 100, Height: 100}
	h := dc.FontHeight(TextStyle{Size: 14})
	if h != 14 {
		t.Errorf("FontHeight fallback = %v, want 14", h)
	}
}
