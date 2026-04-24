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

// cloneAnimationsForShift must deep-copy every MotionPath backing
// array. A shallow clone (outer slice only) would still alias the
// source's MotionPath, and a second shift pass on cache rebuild would
// accumulate drift.
func TestCloneAnimationsForShift_DeepCopiesMotionPath(t *testing.T) {
	src := []gui.SvgAnimation{
		{Kind: gui.SvgAnimMotion, MotionPath: []float32{1, 2, 3, 4}},
		{Kind: gui.SvgAnimRotate, CenterX: 10, CenterY: 20},
	}
	out := cloneAnimationsForShift(src)
	if len(out) != len(src) {
		t.Fatalf("len mismatch: got %d want %d", len(out), len(src))
	}
	// Motion path must have its own backing array.
	if &out[0].MotionPath[0] == &src[0].MotionPath[0] {
		t.Fatal("MotionPath aliases source backing array")
	}
	// Mutate clone; source must be untouched.
	out[0].MotionPath[0] = 999
	if src[0].MotionPath[0] != 1 {
		t.Fatalf("source mutated via clone: got %v want 1",
			src[0].MotionPath[0])
	}
}

// cloneAnimationsForShift handles an empty Motion path without
// crashing. Nil slices are legitimate for non-motion kinds.
func TestCloneAnimationsForShift_EmptyMotionPathSafe(t *testing.T) {
	src := []gui.SvgAnimation{
		{Kind: gui.SvgAnimMotion, MotionPath: nil},
		{Kind: gui.SvgAnimRotate},
	}
	out := cloneAnimationsForShift(src)
	if len(out) != 2 || out[0].MotionPath != nil {
		t.Fatalf("unexpected clone: %+v", out)
	}
}

// Re-running buildParsed on the same VectorGraphic must not accumulate
// viewBox shift on MotionPath. Regression: a shallow slice clone
// would alias and re-subtract the origin each call.
func TestBuildParsed_ViewBoxShiftIsIdempotent(t *testing.T) {
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
	originalMotion := append([]float32(nil), vg.Animations[0].MotionPath...)
	originalCX := vg.Animations[1].CenterX
	originalCY := vg.Animations[1].CenterY

	first := p.buildParsed(1, vg, 1)
	// First call: shifted by (-10,-20).
	if first.Animations[0].MotionPath[0] != 90 ||
		first.Animations[0].MotionPath[1] != 180 {
		t.Fatalf("first shift wrong: %v",
			first.Animations[0].MotionPath[:2])
	}
	if first.Animations[1].CenterX != 40 ||
		first.Animations[1].CenterY != 40 {
		t.Fatalf("rotate center wrong: cx=%v cy=%v",
			first.Animations[1].CenterX, first.Animations[1].CenterY)
	}
	// Source must be unchanged so the next rebuild starts from
	// absolute viewBox coords again.
	for i, v := range vg.Animations[0].MotionPath {
		if v != originalMotion[i] {
			t.Fatalf("source MotionPath mutated at %d: got %v want %v",
				i, v, originalMotion[i])
		}
	}
	if vg.Animations[1].CenterX != originalCX ||
		vg.Animations[1].CenterY != originalCY {
		t.Fatalf("source rotate center mutated")
	}
	// Second call must produce the same shifted output, not
	// double-shifted values.
	second := p.buildParsed(2, vg, 1)
	if second.Animations[0].MotionPath[0] != 90 ||
		second.Animations[0].MotionPath[1] != 180 {
		t.Fatalf("second shift drifted: %v",
			second.Animations[0].MotionPath[:2])
	}
	if second.Animations[1].CenterX != 40 {
		t.Fatalf("second rotate center drifted: %v",
			second.Animations[1].CenterX)
	}
}

// Non-finite viewBox origin must skip the shift entirely so geometry
// and animations still render with authored coords instead of being
// splattered with NaN/Inf.
func TestBuildParsed_NaNViewBoxSkipsAnimShift(t *testing.T) {
	p := New()
	vg := &VectorGraphic{
		Width: 32, Height: 32,
		ViewBoxX: float32(math.NaN()), ViewBoxY: 0,
		Animations: []gui.SvgAnimation{
			{
				Kind:    gui.SvgAnimRotate,
				GroupID: "g1",
				CenterX: 5, CenterY: 7,
			},
		},
	}
	out := p.buildParsed(99, vg, 1)
	if out.Animations[0].CenterX != 5 || out.Animations[0].CenterY != 7 {
		t.Fatalf("NaN viewBox leaked into centers: cx=%v cy=%v",
			out.Animations[0].CenterX, out.Animations[0].CenterY)
	}
}
