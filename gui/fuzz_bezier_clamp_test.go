package gui

import (
	"math"
	"testing"
)

// FuzzBezierClamp feeds (t, cp1x, cp1y, cp2x, cp2y) through the same
// clampUnit(bezierCalc(...)) wrapper used by locateSeg when SMIL
// keySplines drive progress. Asserts the clamped output is always a
// finite value in [0, 1] regardless of overshoot, NaN, or degenerate
// control points that would ordinarily make bezierCalc return NaN or
// a value outside the unit interval.
func FuzzBezierClamp(f *testing.F) {
	f.Add(float32(0.5), float32(0.25), float32(0.1),
		float32(0.25), float32(1.0))
	f.Add(float32(0), float32(0), float32(0), float32(1), float32(1))
	f.Add(float32(1), float32(0), float32(0), float32(1), float32(1))
	f.Add(float32(-0.5), float32(-2), float32(-2),
		float32(2), float32(2))
	f.Add(float32(1.5), float32(0), float32(1),
		float32(1), float32(0))
	f.Add(float32(math.NaN()), float32(0.5), float32(0.5),
		float32(0.5), float32(0.5))

	f.Fuzz(func(t *testing.T, tv, x1, y1, x2, y2 float32) {
		got := clampUnit(bezierCalc(tv, x1, y1, x2, y2))
		if math.IsNaN(float64(got)) || math.IsInf(float64(got), 0) {
			t.Fatalf("non-finite: t=%v cp=(%v,%v,%v,%v) got=%v",
				tv, x1, y1, x2, y2, got)
		}
		if got < 0 || got > 1 {
			t.Fatalf("out of range: t=%v cp=(%v,%v,%v,%v) got=%v",
				tv, x1, y1, x2, y2, got)
		}
	})
}
