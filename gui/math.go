package gui

import "math"

const f32Tolerance = float32(0.01)

// f32Clamp returns x constrained between lo and hi.
func f32Clamp(x, lo, hi float32) float32 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// intClamp returns x constrained between lo and hi.
func intClamp(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// f32AreClose tests if |a - b| <= f32Tolerance.
func f32AreClose(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= f32Tolerance
}

// f32Abs returns absolute value of x.
func f32Abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// f32Min returns the smaller of a and b.
func f32Min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// f32Max returns the larger of a and b.
func f32Max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// f32IsFinite returns true if value is not NaN or Inf.
func f32IsFinite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

// intMin returns the smaller of a and b.
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// intMax returns the larger of a and b.
func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
