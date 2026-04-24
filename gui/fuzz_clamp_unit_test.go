package gui

import (
	"math"
	"testing"
)

// FuzzClampUnit asserts clampUnit always returns a value in [0, 1]
// regardless of NaN, ±Inf, or out-of-range finite input. NaN must
// fold to 0 — the whole point of the guard. Output must also be
// finite (NaN/Inf would slip through a naive range compare).
func FuzzClampUnit(f *testing.F) {
	f.Add(float32(0))
	f.Add(float32(1))
	f.Add(float32(-0.5))
	f.Add(float32(1.5))
	f.Add(float32(math.NaN()))
	f.Add(float32(math.Inf(1)))
	f.Add(float32(math.Inf(-1)))
	f.Add(float32(math.MaxFloat32))
	f.Add(float32(-math.MaxFloat32))

	f.Fuzz(func(t *testing.T, v float32) {
		got := clampUnit(v)
		if math.IsNaN(float64(got)) || math.IsInf(float64(got), 0) {
			t.Fatalf("clampUnit(%v)=%v not finite", v, got)
		}
		if got < 0 || got > 1 {
			t.Fatalf("clampUnit(%v)=%v out of [0,1]", v, got)
		}
	})
}
