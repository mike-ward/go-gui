package gui

import (
	"math"
	"testing"
)

// TestPhase3LerpLinearFastPath — nil splines must yield exact
// linear interpolation (byte-for-byte match with pre-phase-3).
func TestPhase3LerpLinearFastPath(t *testing.T) {
	vals := []float32{0, 10, 0}
	cases := []struct {
		frac, want float32
	}{
		{0, 0}, {0.25, 5}, {0.5, 10}, {0.75, 5}, {1, 0},
	}
	for _, c := range cases {
		got := lerpKeyframes(vals, nil, c.frac)
		if absf(got-c.want) > 1e-5 {
			t.Fatalf("frac=%g: want %g got %g", c.frac, c.want, got)
		}
	}
}

// TestPhase3IdentitySplineMatchesLinear — cubic-bezier
// (0,0,1,1) is the identity curve; ease output must equal linear.
func TestPhase3IdentitySplineMatchesLinear(t *testing.T) {
	vals := []float32{0, 100}
	splines := []float32{0, 0, 1, 1}
	for _, f := range []float32{0.1, 0.3, 0.5, 0.7, 0.9} {
		lin := lerpKeyframes(vals, nil, f)
		eased := lerpKeyframes(vals, splines, f)
		if absf(lin-eased) > 5e-3 {
			t.Fatalf("frac=%g: linear=%g eased=%g", f, lin, eased)
		}
	}
}

// TestPhase3EaseOutBendsAboveLinear — keySplines (.33,.66,.66,1)
// is an ease-out: at t=0.5 the bezier y exceeds 0.5 (faster early,
// slower later). Verify the eased mid-fraction is strictly above
// the linear midpoint.
func TestPhase3EaseOutBendsAboveLinear(t *testing.T) {
	vals := []float32{0, 100}
	splines := []float32{.33, .66, .66, 1}
	lin := lerpKeyframes(vals, nil, 0.5)
	eased := lerpKeyframes(vals, splines, 0.5)
	if eased <= lin {
		t.Fatalf("expected ease-out > linear: lin=%g eased=%g",
			lin, eased)
	}
	if eased < 60 || eased > 85 {
		t.Fatalf("eased mid out of expected band [60,85]: %g", eased)
	}
}

// TestPhase3MismatchedSplinesFallBackToLinear — when splines
// length does not equal 4*(N-1), lerpKeyframes must ignore them.
func TestPhase3MismatchedSplinesFallBackToLinear(t *testing.T) {
	vals := []float32{0, 10, 0}            // 2 segments
	splines := []float32{.33, .66, .66, 1} // only 1 segment
	got := lerpKeyframes(vals, splines, 0.25)
	if absf(got-5) > 1e-5 {
		t.Fatalf("want linear fallback 5, got %g", got)
	}
}

// TestPhase3BezierEdges verifies bezierEase clamps at extremes.
func TestPhase3BezierEdges(t *testing.T) {
	if bezierEase(0, .33, .66, .66, 1) != 0 {
		t.Fatal("bezierEase(0) must be 0")
	}
	if bezierEase(1, .33, .66, .66, 1) != 1 {
		t.Fatal("bezierEase(1) must be 1")
	}
	// Out-of-range x gets clamped by lerpKeyframes, but bezierEase
	// itself should clamp defensively too.
	if bezierEase(-0.1, .4, 0, .6, 1) != 0 {
		t.Fatal("bezierEase(<0) must clamp to 0")
	}
	if bezierEase(1.1, .4, 0, .6, 1) != 1 {
		t.Fatal("bezierEase(>1) must clamp to 1")
	}
}

func absf(f float32) float32 {
	return float32(math.Abs(float64(f)))
}
