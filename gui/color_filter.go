package gui

import "math"

// ColorFilter holds a 4x4 column-major color transform matrix.
// Applied as a post-processing pass on container content.
// Operates on premultiplied-alpha pixels from the FBO.
type ColorFilter struct {
	Matrix [16]float32
}

// Pre-computed matrices for constant filters.
var (
	identityMatrix = [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	grayscaleMatrix = [16]float32{
		0.2126, 0.2126, 0.2126, 0,
		0.7152, 0.7152, 0.7152, 0,
		0.0722, 0.0722, 0.0722, 0,
		0, 0, 0, 1,
	}
	sepiaMatrix = [16]float32{
		0.393, 0.349, 0.272, 0,
		0.769, 0.686, 0.534, 0,
		0.189, 0.168, 0.131, 0,
		0, 0, 0, 1,
	}
	invertMatrix = [16]float32{
		-1, 0, 0, 0,
		0, -1, 0, 0,
		0, 0, -1, 0,
		1, 1, 1, 1,
	}
)

// ColorFilterIdentity returns a no-op color filter.
func ColorFilterIdentity() *ColorFilter {
	return &ColorFilter{Matrix: identityMatrix}
}

// ColorFilterGrayscale converts to luminance-weighted grayscale.
func ColorFilterGrayscale() *ColorFilter {
	return &ColorFilter{Matrix: grayscaleMatrix}
}

// ColorFilterSepia applies a warm sepia tone.
func ColorFilterSepia() *ColorFilter {
	return &ColorFilter{Matrix: sepiaMatrix}
}

// ColorFilterSaturate adjusts saturation. 0=grayscale, 1=identity,
// >1=oversaturated.
func ColorFilterSaturate(amount float32) *ColorFilter {
	const lr, lg, lb = 0.2126, 0.7152, 0.0722
	s := amount
	return &ColorFilter{Matrix: [16]float32{
		lr*(1-s) + s, lr * (1 - s), lr * (1 - s), 0,
		lg * (1 - s), lg*(1-s) + s, lg * (1 - s), 0,
		lb * (1 - s), lb * (1 - s), lb*(1-s) + s, 0,
		0, 0, 0, 1,
	}}
}

// ColorFilterBrightness scales RGB channels. 1=identity, <1=dim,
// >1=bright.
func ColorFilterBrightness(amount float32) *ColorFilter {
	return &ColorFilter{Matrix: [16]float32{
		amount, 0, 0, 0,
		0, amount, 0, 0,
		0, 0, amount, 0,
		0, 0, 0, 1,
	}}
}

// ColorFilterContrast scales RGB around 0.5 midpoint. 1=identity,
// 0=all gray, >1=higher contrast. Bias injected via alpha column
// (correct for premultiplied-alpha FBO content).
func ColorFilterContrast(amount float32) *ColorFilter {
	bias := 0.5 * (1 - amount)
	return &ColorFilter{Matrix: [16]float32{
		amount, 0, 0, 0,
		0, amount, 0, 0,
		0, 0, amount, 0,
		bias, bias, bias, 1,
	}}
}

// ColorFilterHueRotate rotates hue by the given angle in degrees.
// Rodrigues rotation around (1,1,1)/sqrt(3) in RGB space.
func ColorFilterHueRotate(degrees float32) *ColorFilter {
	rad := float64(degrees) * math.Pi / 180
	c := float32(math.Cos(rad))
	s := float32(math.Sin(rad))
	const k = 1.0 / 3.0
	sq := float32(1.0 / math.Sqrt(3))
	// Column-major 4x4. Top-left 3x3 is Rodrigues formula.
	return &ColorFilter{Matrix: [16]float32{
		k + c*(1-k), k*(1-c) + s*sq, k*(1-c) - s*sq, 0,
		k*(1-c) - s*sq, k + c*(1-k), k*(1-c) + s*sq, 0,
		k*(1-c) + s*sq, k*(1-c) - s*sq, k + c*(1-k), 0,
		0, 0, 0, 1,
	}}
}

// ColorFilterInvert negates RGB, keeps alpha. Uses the alpha
// column to inject bias (output.rgb = alpha - input.rgb).
func ColorFilterInvert() *ColorFilter {
	return &ColorFilter{Matrix: invertMatrix}
}

// ColorFilterCompose multiplies two color filters (a applied
// first, then b). Returns a new filter representing b*a.
func ColorFilterCompose(a, b *ColorFilter) *ColorFilter {
	var out ColorFilter
	for col := range 4 {
		for row := range 4 {
			var sum float32
			for k := range 4 {
				sum += b.Matrix[k*4+row] * a.Matrix[col*4+k]
			}
			out.Matrix[col*4+row] = sum
		}
	}
	return &out
}
