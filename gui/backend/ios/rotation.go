//go:build ios

package ios

import "math"

// applyRotation multiplies mvp by a rotation matrix that
// rotates angleDeg degrees around the point (cx, cy).
func applyRotation(mvp *[16]float32, angleDeg, cx, cy float32) {
	rad := float64(angleDeg) * math.Pi / 180
	cosA := float32(math.Cos(rad))
	sinA := float32(math.Sin(rad))
	tx := cx*(1-cosA) + cy*sinA
	ty := cy*(1-cosA) - cx*sinA
	// Column-major rotation around (cx, cy).
	var rot [16]float32
	rot[0] = cosA
	rot[1] = sinA
	rot[4] = -sinA
	rot[5] = cosA
	rot[10] = 1
	rot[12] = tx
	rot[13] = ty
	rot[15] = 1
	var out [16]float32
	mat4Mul(&out, mvp, &rot)
	*mvp = out
}

// mat4Mul multiplies two 4x4 column-major matrices.
func mat4Mul(out, a, b *[16]float32) {
	for col := range 4 {
		for row := range 4 {
			var sum float32
			for k := range 4 {
				sum += a[k*4+row] * b[col*4+k]
			}
			out[col*4+row] = sum
		}
	}
}
