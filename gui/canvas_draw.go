package gui

import "math"

// DrawCanvasCache holds retained tessellation output keyed by
// widget id + version. Cache hit skips OnDraw entirely.
type DrawCanvasCache struct {
	Version uint64
	Batches []DrawCanvasTriBatch
}

// DrawCanvasTriBatch is one flat-color triangle batch.
type DrawCanvasTriBatch struct {
	Triangles []float32
	Color     Color
}

// DrawContext is passed to the OnDraw callback. Drawing methods
// append tessellated triangle batches which are later emitted as
// RenderSvg commands.
type DrawContext struct {
	Width   float32
	Height  float32
	batches []DrawCanvasTriBatch
}

// FilledRect draws a filled rectangle as two triangles.
func (dc *DrawContext) FilledRect(x, y, w, h float32, color Color) {
	if w <= 0 || h <= 0 {
		return
	}
	dc.batches = append(dc.batches, DrawCanvasTriBatch{
		Triangles: []float32{
			x, y,
			x + w, y,
			x + w, y + h,
			x, y,
			x + w, y + h,
			x, y + h,
		},
		Color: color,
	})
}

// Line draws a single line segment.
func (dc *DrawContext) Line(x0, y0, x1, y1 float32, color Color, width float32) {
	dc.Polyline([]float32{x0, y0, x1, y1}, color, width)
}

// Polyline draws a stroked open polyline using simple
// per-segment rectangle expansion (no joins/caps).
func (dc *DrawContext) Polyline(points []float32, color Color, width float32) {
	if len(points) < 4 || width <= 0 {
		return
	}
	hw := width / 2
	tris := make([]float32, 0, ((len(points)/2)-1)*12)
	for i := 0; i+3 < len(points); i += 2 {
		x0, y0 := points[i], points[i+1]
		x1, y1 := points[i+2], points[i+3]
		dx := x1 - x0
		dy := y1 - y0
		ln := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if ln < 1e-6 {
			continue
		}
		// Perpendicular offset.
		nx := -dy / ln * hw
		ny := dx / ln * hw
		// Quad as two triangles.
		tris = append(tris,
			x0+nx, y0+ny,
			x0-nx, y0-ny,
			x1-nx, y1-ny,
			x0+nx, y0+ny,
			x1-nx, y1-ny,
			x1+nx, y1+ny,
		)
	}
	if len(tris) > 0 {
		dc.batches = append(dc.batches, DrawCanvasTriBatch{
			Triangles: tris,
			Color:     color,
		})
	}
}

// Rect draws a stroked rectangle.
func (dc *DrawContext) Rect(x, y, w, h float32, color Color, width float32) {
	if w <= 0 || h <= 0 || width <= 0 {
		return
	}
	pts := []float32{x, y, x + w, y, x + w, y + h, x, y + h, x, y}
	dc.Polyline(pts, color, width)
}

// FilledPolygon draws a filled convex polygon using fan from
// first vertex.
func (dc *DrawContext) FilledPolygon(points []float32, color Color) {
	if len(points) < 6 {
		return
	}
	n := len(points) / 2
	tris := make([]float32, 0, (n-2)*6)
	x0, y0 := points[0], points[1]
	for i := 1; i < n-1; i++ {
		tris = append(tris,
			x0, y0,
			points[i*2], points[i*2+1],
			points[(i+1)*2], points[(i+1)*2+1],
		)
	}
	if len(tris) > 0 {
		dc.batches = append(dc.batches, DrawCanvasTriBatch{
			Triangles: tris,
			Color:     color,
		})
	}
}

// FilledCircle draws a filled circle.
func (dc *DrawContext) FilledCircle(cx, cy, radius float32, color Color) {
	dc.FilledArc(cx, cy, radius, radius, 0, 2*math.Pi, color)
}

// Circle draws a stroked circle.
func (dc *DrawContext) Circle(cx, cy, radius float32, color Color, width float32) {
	dc.Arc(cx, cy, radius, radius, 0, 2*math.Pi, color, width)
}

// Arc draws a stroked elliptical arc.
func (dc *DrawContext) Arc(cx, cy, rx, ry, start, sweep float32, color Color, width float32) {
	if width <= 0 {
		return
	}
	pts := arcToPolyline(cx, cy, rx, ry, start, sweep)
	if len(pts) >= 4 {
		dc.Polyline(pts, color, width)
	}
}

// FilledArc draws a filled elliptical arc (pie slice).
func (dc *DrawContext) FilledArc(cx, cy, rx, ry, start, sweep float32, color Color) {
	pts := arcToPolyline(cx, cy, rx, ry, start, sweep)
	if len(pts) < 4 {
		return
	}
	// Close as pie: center → arc → center.
	poly := make([]float32, 0, len(pts)+4)
	poly = append(poly, cx, cy)
	poly = append(poly, pts...)
	poly = append(poly, cx, cy)
	dc.FilledPolygon(poly, color)
}

// arcToPolyline converts an elliptical arc to a flat x,y
// polyline via angular subdivision.
func arcToPolyline(cx, cy, rx, ry, start, sweep float32) []float32 {
	r := rx
	if ry > r {
		r = ry
	}
	if r <= 0 {
		return nil
	}
	n := int(math.Ceil(
		float64(f32Abs(sweep)) / (2 * math.Pi) * 64 *
			math.Sqrt(float64(r)/50+1)))
	if n < 4 {
		n = 4
	}
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
