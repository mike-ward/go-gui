package svg

import (
	"math"
	"testing"
)

// FuzzDecomposeTRS feeds arbitrary 2×3 affine matrices through
// decomposeTRS and asserts:
//   - no panic
//   - ok=true → every returned scalar is finite AND recomposing
//     approximates the input within a slack epsilon
//   - ok=false is permitted (shear, non-decomposable matrix)
//
// Recomposition slack is 1e-3 relative; the decompose epsilon itself
// is 1e-4 absolute, so randomized inputs benefit from a looser
// compare.
func FuzzDecomposeTRS(f *testing.F) {
	f.Add(float32(1), float32(0), float32(0), float32(1),
		float32(0), float32(0)) // identity
	f.Add(float32(2), float32(0), float32(0), float32(3),
		float32(5), float32(7)) // pure scale+trans
	f.Add(float32(0), float32(1), float32(-1), float32(0),
		float32(0), float32(0)) // 90deg rotation
	f.Add(float32(1), float32(1), float32(0), float32(1),
		float32(0), float32(0)) // shear
	f.Add(float32(math.NaN()), float32(0), float32(0), float32(1),
		float32(0), float32(0))
	f.Add(float32(math.Inf(1)), float32(0), float32(0), float32(1),
		float32(0), float32(0))

	f.Fuzz(func(t *testing.T, a, b, c, d, e, fv float32) {
		m := [6]float32{a, b, c, d, e, fv}
		tx, ty, sx, sy, rot, ok := decomposeTRS(m)
		if !ok {
			return
		}
		for i, v := range []float32{tx, ty, sx, sy, rot} {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				t.Fatalf("ok=true but output[%d]=%v non-finite "+
					"(input=%v)", i, v, m)
			}
		}
		// Recompose: m' = T · R · S.
		rad := float64(rot) * math.Pi / 180
		cosT, sinT := float32(math.Cos(rad)), float32(math.Sin(rad))
		ra := cosT * sx
		rb := sinT * sx
		rc := -sinT * sy
		rd := cosT * sy
		re := tx
		rf := ty
		const slack = 1e-3
		relClose := func(want, got float32) bool {
			if math.IsNaN(float64(want)) || math.IsInf(float64(want), 0) {
				return false
			}
			diff := float64(got - want)
			if diff < 0 {
				diff = -diff
			}
			mag := math.Abs(float64(want))
			if mag < 1 {
				mag = 1
			}
			return diff <= slack*mag
		}
		if !relClose(a, ra) || !relClose(b, rb) ||
			!relClose(c, rc) || !relClose(d, rd) ||
			!relClose(e, re) || !relClose(fv, rf) {
			t.Fatalf("recompose mismatch input=%v got=[%v %v %v %v %v %v]",
				m, ra, rb, rc, rd, re, rf)
		}
	})
}
