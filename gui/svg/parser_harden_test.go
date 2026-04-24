package svg

import (
	"math"
	"testing"

	"github.com/mike-ward/go-gui/gui"
)

// nonNegF32 folds NaN and negative inputs to 0 so oversized / reversed-
// winding primitives from pathological spline overshoot never reach
// segmentsForRect / segmentsForEllipse.
func TestNonNegF32_NaNAndNegativeReturnZero(t *testing.T) {
	cases := []struct {
		in, want float32
	}{
		{0, 0},
		{1, 1},
		{-1, 0},
		{float32(math.NaN()), 0},
		{float32(math.Inf(-1)), 0},
		{float32(math.Inf(1)), float32(math.Inf(1))},
	}
	for _, c := range cases {
		got := nonNegF32(c.in)
		// NaN != NaN — compare bit-pattern via math.IsNaN.
		if math.IsNaN(float64(got)) != math.IsNaN(float64(c.want)) ||
			(!math.IsNaN(float64(got)) && got != c.want) {
			t.Errorf("nonNegF32(%v)=%v want %v", c.in, got, c.want)
		}
	}
}

// overrideScalar replaces non-finite override values with the parsed
// base so NaN/Inf keyframes cannot contaminate primitive geometry.
func TestOverrideScalar_NaNFallsBackToBase(t *testing.T) {
	ov := gui.SvgAnimAttrOverride{Mask: gui.SvgAnimMaskCX}
	nan := float32(math.NaN())
	if got := overrideScalar(5, nan, &ov, gui.SvgAnimMaskCX); got != 5 {
		t.Fatalf("NaN v: want base=5, got %v", got)
	}
	inf := float32(math.Inf(1))
	if got := overrideScalar(5, inf, &ov, gui.SvgAnimMaskCX); got != 5 {
		t.Fatalf("+Inf v: want base=5, got %v", got)
	}
	// Finite replace still wins.
	if got := overrideScalar(5, 9, &ov, gui.SvgAnimMaskCX); got != 9 {
		t.Fatalf("finite replace: want 9, got %v", got)
	}
	// Unset mask bit returns base unconditionally.
	ovEmpty := gui.SvgAnimAttrOverride{}
	if got := overrideScalar(5, 9, &ovEmpty, gui.SvgAnimMaskCX); got != 5 {
		t.Fatalf("unset mask: want 5, got %v", got)
	}
}

// buildParsed must propagate the viewBox origin onto SvgParsed so the
// render path can apply it as an outer translate. Animation fields
// and path coords stay in raw viewBox space so SMIL animateTransform
// in replace mode cannot erase the mapping.
func TestBuildParsed_PropagatesViewBoxOrigin(t *testing.T) {
	p := New()
	vg := &VectorGraphic{
		Width: 32, Height: 32,
		ViewBoxX: 10, ViewBoxY: 20,
		Animations: []gui.SvgAnimation{
			{
				Kind:       gui.SvgAnimMotion,
				GroupID:    "g1",
				MotionPath: []float32{100, 200, 110, 210},
			},
			{
				Kind:    gui.SvgAnimRotate,
				GroupID: "g1",
				CenterX: 50, CenterY: 60,
			},
		},
	}
	parsed := p.buildParsed(1, vg, 1)
	if parsed.ViewBoxX != 10 || parsed.ViewBoxY != 20 {
		t.Fatalf("viewBox origin not propagated: x=%v y=%v",
			parsed.ViewBoxX, parsed.ViewBoxY)
	}
	// MotionPath and rotate center must stay in authored viewBox
	// coords — render applies the outer shift.
	if parsed.Animations[0].MotionPath[0] != 100 ||
		parsed.Animations[0].MotionPath[1] != 200 {
		t.Fatalf("motion path shifted at parse: %v",
			parsed.Animations[0].MotionPath[:2])
	}
	if parsed.Animations[1].CenterX != 50 ||
		parsed.Animations[1].CenterY != 60 {
		t.Fatalf("rotate center shifted at parse: cx=%v cy=%v",
			parsed.Animations[1].CenterX, parsed.Animations[1].CenterY)
	}
}
