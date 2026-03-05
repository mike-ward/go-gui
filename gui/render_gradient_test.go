package gui

import (
	"math"
	"testing"
)

func TestClampUnit(t *testing.T) {
	tests := []struct{ in, want float32 }{
		{-1, 0}, {0, 0}, {0.5, 0.5}, {1, 1}, {2, 1},
	}
	for _, tc := range tests {
		if got := clampUnit(tc.in); got != tc.want {
			t.Errorf("clampUnit(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestAngleToDirectionCardinal(t *testing.T) {
	const eps = 1e-5
	cases := []struct {
		deg  float32
		wDx  float32
		wDy  float32
		name string
	}{
		{0, 0, -1, "top"},
		{90, 1, 0, "right"},
		{180, 0, 1, "bottom"},
		{270, -1, 0, "left"},
	}
	for _, c := range cases {
		dx, dy := angleToDirection(c.deg)
		if math.Abs(float64(dx-c.wDx)) > eps || math.Abs(float64(dy-c.wDy)) > eps {
			t.Errorf("%s: got (%v,%v), want (%v,%v)", c.name, dx, dy, c.wDx, c.wDy)
		}
	}
}

func TestAngleToDirectionDiagonal(t *testing.T) {
	const eps = 1e-5
	dx, dy := angleToDirection(45)
	expected := float32(math.Sqrt(2) / 2)
	if math.Abs(float64(dx-expected)) > eps || math.Abs(float64(dy+expected)) > eps {
		t.Errorf("45deg: got (%v,%v)", dx, dy)
	}
}

func TestGradientDirectionKeywords(t *testing.T) {
	const eps = 1e-5
	g := &GradientDef{}
	cases := []struct {
		dir GradientDirection
		wDx float32
		wDy float32
	}{
		{GradientToTop, 0, -1},
		{GradientToRight, 1, 0},
		{GradientToBottom, 0, 1},
		{GradientToLeft, -1, 0},
	}
	for _, c := range cases {
		g.Direction = c.dir
		dx, dy := GradientDir(g, 100, 100)
		if math.Abs(float64(dx-c.wDx)) > eps || math.Abs(float64(dy-c.wDy)) > eps {
			t.Errorf("dir %d: got (%v,%v), want (%v,%v)", c.dir, dx, dy, c.wDx, c.wDy)
		}
	}
}

func TestGradientDirectionAngleOverride(t *testing.T) {
	const eps = 1e-5
	g := &GradientDef{HasAngle: true, Angle: 90, Direction: GradientToTop}
	dx, dy := GradientDir(g, 100, 100)
	if math.Abs(float64(dx-1)) > eps || math.Abs(float64(dy)) > eps {
		t.Errorf("angle override: got (%v,%v), want (1,0)", dx, dy)
	}
}

func TestPackRGB(t *testing.T) {
	c := Color{R: 100, G: 150, B: 200, set: true}
	p := PackRGB(c)
	// Unpack: R = p mod 256, G = (p/256) mod 256, B = p/65536
	r := uint8(math.Mod(float64(p), 256))
	g := uint8(math.Mod(float64(p)/256, 256))
	b := uint8(p / 65536)
	if r != 100 || g != 150 || b != 200 {
		t.Errorf("unpack: got (%d,%d,%d), want (100,150,200)", r, g, b)
	}
}

func TestPackAlphaPos(t *testing.T) {
	c := Color{A: 128, set: true}
	p := PackAlphaPos(c, 0.5)
	// Alpha = p mod 256 = 128
	a := uint8(math.Mod(float64(p), 256))
	if a != 128 {
		t.Errorf("alpha: got %d, want 128", a)
	}
}

func TestF32ToU8Saturated(t *testing.T) {
	tests := []struct {
		in   float32
		want uint8
	}{
		{-10, 0}, {0, 0}, {127.5, 128}, {255, 255}, {300, 255},
	}
	for _, tc := range tests {
		if got := f32ToU8Saturated(tc.in); got != tc.want {
			t.Errorf("f32ToU8(%v) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestLerpColorPremultipliedEndpoints(t *testing.T) {
	a := Color{R: 255, G: 0, B: 0, A: 255, set: true}
	b := Color{R: 0, G: 0, B: 255, A: 255, set: true}
	c0 := lerpColorPremultiplied(a, b, 0)
	if c0 != a {
		t.Errorf("t=0: got %v, want %v", c0, a)
	}
	c1 := lerpColorPremultiplied(a, b, 1)
	if c1 != b {
		t.Errorf("t=1: got %v, want %v", c1, b)
	}
}

func TestLerpColorPremultipliedMid(t *testing.T) {
	a := Color{R: 0, G: 0, B: 0, A: 255, set: true}
	b := Color{R: 254, G: 254, B: 254, A: 255, set: true}
	mid := lerpColorPremultiplied(a, b, 0.5)
	if mid.R < 125 || mid.R > 129 {
		t.Errorf("mid R: got %d", mid.R)
	}
}

func TestLerpColorPremultipliedZeroAlpha(t *testing.T) {
	a := Color{R: 255, A: 0, set: true}
	b := Color{R: 0, A: 0, set: true}
	c := lerpColorPremultiplied(a, b, 0.5)
	if c.A != 0 {
		t.Errorf("zero alpha: got A=%d", c.A)
	}
}

func TestSampleGradientStopColorEmpty(t *testing.T) {
	c := SampleGradientStopColor(nil, 0.5)
	if c.A != 0 {
		t.Error("empty stops should return transparent")
	}
}

func TestSampleGradientStopColorSingle(t *testing.T) {
	stops := []GradientStop{{Color: Color{R: 100, A: 255, set: true}, Pos: 0.5}}
	c := SampleGradientStopColor(stops, 0.0)
	if c.R != 100 {
		t.Errorf("single stop: got R=%d", c.R)
	}
}

func TestSampleGradientStopColorTwoStop(t *testing.T) {
	stops := []GradientStop{
		{Color: Color{R: 0, A: 255, set: true}, Pos: 0},
		{Color: Color{R: 254, A: 255, set: true}, Pos: 1},
	}
	mid := SampleGradientStopColor(stops, 0.5)
	if mid.R < 125 || mid.R > 129 {
		t.Errorf("two-stop mid: got R=%d", mid.R)
	}
}

func TestSampleGradientStopColorBoundary(t *testing.T) {
	stops := []GradientStop{
		{Color: Color{R: 10, A: 255, set: true}, Pos: 0},
		{Color: Color{R: 200, A: 255, set: true}, Pos: 1},
	}
	c0 := SampleGradientStopColor(stops, 0)
	if c0.R != 10 {
		t.Errorf("pos=0: got R=%d", c0.R)
	}
	c1 := SampleGradientStopColor(stops, 1)
	if c1.R != 200 {
		t.Errorf("pos=1: got R=%d", c1.R)
	}
}

func TestNormalizeGradientStopsEmpty(t *testing.T) {
	if got := normalizeGradientStops(nil); got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

func TestNormalizeGradientStopsSorted(t *testing.T) {
	stops := []GradientStop{
		{Color: Color{R: 255, A: 255, set: true}, Pos: 0.8},
		{Color: Color{R: 0, A: 255, set: true}, Pos: 0.2},
	}
	norm := normalizeGradientStops(stops)
	if norm[0].Pos > norm[1].Pos {
		t.Error("should be sorted")
	}
}

func TestNormalizeGradientStopsOverLimit(t *testing.T) {
	stops := make([]GradientStop, 10)
	for i := range stops {
		stops[i] = GradientStop{
			Color: Color{R: uint8(i * 25), A: 255, set: true},
			Pos:   float32(i) / 9.0,
		}
	}
	norm := normalizeGradientStops(stops)
	if len(norm) != gradientShaderStopLimit {
		t.Fatalf("want %d, got %d", gradientShaderStopLimit, len(norm))
	}
}

func TestNormalizeGradientStopsIntoReuse(t *testing.T) {
	stops := []GradientStop{
		{Color: Color{R: 0, A: 255, set: true}, Pos: 0},
		{Color: Color{R: 255, A: 255, set: true}, Pos: 1},
	}
	norm := make([]GradientStop, 0, 8)
	sampled := make([]GradientStop, 0, 8)
	result := NormalizeGradientStopsInto(stops, &norm, &sampled)
	if len(result) != 2 {
		t.Fatalf("want 2, got %d", len(result))
	}
}
