package gui

import (
	"math"
	"testing"
)

// triangle: (0,0)-(10,0)-(0,10)
func newRightTriPath() *TessellatedPath {
	return &TessellatedPath{
		Triangles: []float32{
			0, 0, 10, 0, 0, 10,
		},
		MinX: 0, MinY: 0, MaxX: 10, MaxY: 10,
	}
}

func TestContainsPointInside(t *testing.T) {
	t.Parallel()
	tp := newRightTriPath()
	if !tp.ContainsPoint(2, 2) {
		t.Error("(2,2) should be inside")
	}
	if !tp.ContainsPoint(0, 0) {
		t.Error("(0,0) vertex should be inside")
	}
}

func TestContainsPointOutside(t *testing.T) {
	t.Parallel()
	tp := newRightTriPath()
	if tp.ContainsPoint(9, 9) {
		t.Error("(9,9) outside hypotenuse — false expected")
	}
	if tp.ContainsPoint(-1, 0) {
		t.Error("(-1,0) negative — false expected")
	}
	if tp.ContainsPoint(20, 20) {
		t.Error("(20,20) outside bbox — false expected")
	}
}

func TestContainsPointStrokeRejected(t *testing.T) {
	t.Parallel()
	tp := newRightTriPath()
	tp.IsStroke = true
	if tp.ContainsPoint(2, 2) {
		t.Error("stroke path should not hit-test")
	}
}

func TestContainsPointEmpty(t *testing.T) {
	t.Parallel()
	var nilTp *TessellatedPath
	if nilTp.ContainsPoint(0, 0) {
		t.Error("nil receiver should return false")
	}
	tp := &TessellatedPath{}
	if tp.ContainsPoint(0, 0) {
		t.Error("empty triangles should return false")
	}
}

func TestContainsPointMultipleTriangles(t *testing.T) {
	t.Parallel()
	// Square covered by two right triangles.
	tp := &TessellatedPath{
		Triangles: []float32{
			0, 0, 10, 0, 0, 10,
			10, 0, 10, 10, 0, 10,
		},
		MinX: 0, MinY: 0, MaxX: 10, MaxY: 10,
	}
	if !tp.ContainsPoint(8, 8) {
		t.Error("(8,8) should hit second triangle")
	}
	if !tp.ContainsPoint(2, 2) {
		t.Error("(2,2) should hit first triangle")
	}
}

func TestContainsPointWithBaseTranslate(t *testing.T) {
	t.Parallel()
	// Local triangle at origin; path translated to (100, 100).
	tp := &TessellatedPath{
		Triangles: []float32{
			0, 0, 10, 0, 0, 10,
		},
		MinX: 0, MinY: 0, MaxX: 10, MaxY: 10,
		HasBaseXform: true,
		BaseTransX:   100,
		BaseTransY:   100,
		BaseScaleX:   1,
		BaseScaleY:   1,
	}
	if !tp.ContainsPoint(102, 102) {
		t.Error("(102,102) should hit translated triangle")
	}
	if tp.ContainsPoint(2, 2) {
		t.Error("(2,2) is in untranslated local space — should miss")
	}
}

func TestContainsPointWithBaseScale(t *testing.T) {
	t.Parallel()
	// Local triangle 0..10; scale 2x → world space 0..20.
	tp := &TessellatedPath{
		Triangles: []float32{
			0, 0, 10, 0, 0, 10,
		},
		MinX: 0, MinY: 0, MaxX: 10, MaxY: 10,
		HasBaseXform: true,
		BaseScaleX:   2,
		BaseScaleY:   2,
	}
	if !tp.ContainsPoint(4, 4) {
		t.Error("(4,4) world (==(2,2) local) should hit")
	}
	if tp.ContainsPoint(15, 15) {
		t.Error("(15,15) world (==(7.5,7.5) local) outside hypotenuse")
	}
}

func TestContainsPointWithBaseRotation(t *testing.T) {
	t.Parallel()
	// Local right triangle (0,0)-(10,0)-(0,10) rotated 90° CCW
	// about pivot (0,0). Forward map: (x,y) → (-y, x). World-space
	// triangle covers (0,0)-(0,0)-(0,10)-(-10,0). Per the
	// seedFromTransform invariant, when rotation is set the matrix
	// translate is absorbed into the pivot and BaseTransX/Y stay 0.
	tp := &TessellatedPath{
		Triangles: []float32{
			0, 0, 10, 0, 0, 10,
		},
		MinX: 0, MinY: 0, MaxX: 10, MaxY: 10,
		HasBaseXform: true,
		BaseScaleX:   1,
		BaseScaleY:   1,
		BaseRotAngle: 90,
		BaseRotCX:    0,
		BaseRotCY:    0,
	}
	// World (-2, 2) maps back to local (2, 2) — inside.
	if !tp.ContainsPoint(-2, 2) {
		t.Error("(-2,2) world should hit rotated triangle")
	}
	// World (2, 2) maps back to local (2, -2) — outside.
	if tp.ContainsPoint(2, 2) {
		t.Error("(2,2) world maps to local (2,-2) — should miss")
	}
	// World (-9, 9) maps back to local (9, 9) — outside hypotenuse.
	if tp.ContainsPoint(-9, 9) {
		t.Error("(-9,9) world (==(9,9) local) outside hypotenuse")
	}
}

func TestContainsPointWithRotationAboutPivot(t *testing.T) {
	t.Parallel()
	// Same triangle, rotated 180° about pivot (5, 5). Forward map:
	// (x,y) → (10-x, 10-y). World vertices: (10,10), (0,10), (10,0).
	tp := &TessellatedPath{
		Triangles: []float32{
			0, 0, 10, 0, 0, 10,
		},
		MinX: 0, MinY: 0, MaxX: 10, MaxY: 10,
		HasBaseXform: true,
		BaseScaleX:   1,
		BaseScaleY:   1,
		BaseRotAngle: 180,
		BaseRotCX:    5,
		BaseRotCY:    5,
	}
	// World (8, 8) maps back to local (2, 2) — inside.
	if !tp.ContainsPoint(8, 8) {
		t.Error("(8,8) world (==(2,2) local) should hit")
	}
	// World (2, 2) maps back to local (8, 8) — outside hypotenuse.
	if tp.ContainsPoint(2, 2) {
		t.Error("(2,2) world (==(8,8) local) outside hypotenuse")
	}
}

func TestContainsPointRejectsNaNInf(t *testing.T) {
	t.Parallel()
	tp := newRightTriPath()
	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	for _, c := range []struct{ px, py float32 }{
		{nan, 0}, {0, nan}, {nan, nan},
		{inf, 0}, {0, inf}, {-inf, -inf},
	} {
		if tp.ContainsPoint(c.px, c.py) {
			t.Errorf("ContainsPoint(%v,%v) should be false", c.px, c.py)
		}
	}
}

func TestContainsPointHandlesNonFiniteBaseXform(t *testing.T) {
	t.Parallel()
	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	// NaN scale → coerced to 1; identity-like transform.
	tp := &TessellatedPath{
		Triangles: []float32{
			0, 0, 10, 0, 0, 10,
		},
		MinX: 0, MinY: 0, MaxX: 10, MaxY: 10,
		HasBaseXform: true,
		BaseScaleX:   nan,
		BaseScaleY:   inf,
		BaseTransX:   nan,
		BaseTransY:   inf,
		BaseRotAngle: nan,
	}
	// Coerced to identity → (2,2) hits, (20,20) misses.
	if !tp.ContainsPoint(2, 2) {
		t.Error("(2,2) should hit after non-finite coercion")
	}
	if tp.ContainsPoint(20, 20) {
		t.Error("(20,20) should miss after non-finite coercion")
	}
}
