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
		got := lerpKeyframes(vals, nil, nil, SvgAnimCalcLinear, c.frac)
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
		lin := lerpKeyframes(vals, nil, nil, SvgAnimCalcLinear, f)
		eased := lerpKeyframes(vals, splines, nil, SvgAnimCalcSpline, f)
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
	lin := lerpKeyframes(vals, nil, nil, SvgAnimCalcLinear, 0.5)
	eased := lerpKeyframes(vals, splines, nil, SvgAnimCalcSpline, 0.5)
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
	got := lerpKeyframes(vals, splines, nil, SvgAnimCalcSpline, 0.25)
	if absf(got-5) > 1e-5 {
		t.Fatalf("want linear fallback 5, got %g", got)
	}
}

// TestLerpKeyframesNegativeFracReturnsFirst — negative frac
// clamps to 0; no panic on negative slice index.
func TestLerpKeyframesNegativeFracReturnsFirst(t *testing.T) {
	vals := []float32{10, 20, 30}
	got := lerpKeyframes(vals, nil, nil, SvgAnimCalcLinear, -5)
	if got != 10 {
		t.Fatalf("negative frac should clamp to first, got %g", got)
	}
}

// TestLerpKeyframesNaNFracReturnsFirst — NaN frac maps to 0.
func TestLerpKeyframesNaNFracReturnsFirst(t *testing.T) {
	vals := []float32{10, 20, 30}
	got := lerpKeyframes(vals, nil, nil, SvgAnimCalcLinear, float32(math.NaN()))
	if got != 10 {
		t.Fatalf("NaN frac should clamp to first, got %g", got)
	}
}

// TestLerpKeyframesDiscreteHoldsKeyframe — with calcMode=discrete,
// value jumps to the keyframe covering frac's sub-interval rather
// than interpolating. Spec: [i/n, (i+1)/n) selects keyframe i.
func TestLerpKeyframesDiscreteHoldsKeyframe(t *testing.T) {
	vals := []float32{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120}
	// n=12 keyframes → each segment is 1/12 wide. frac=0.45 lands
	// in bucket idx = floor(0.45 * 12) = 5 → values[5] = 60.
	got := lerpKeyframes(vals, nil, nil, SvgAnimCalcDiscrete, 0.45)
	if got != 60 {
		t.Fatalf("discrete frac=0.45 want 60, got %g", got)
	}
	// Boundary: frac at exact segment edge lands in the next bucket.
	got = lerpKeyframes(vals, nil, nil, SvgAnimCalcDiscrete, 1.0/12)
	if got != 20 {
		t.Fatalf("discrete frac=1/12 want 20, got %g", got)
	}
	// End-of-range: frac=1 holds last keyframe.
	got = lerpKeyframes(vals, nil, nil, SvgAnimCalcDiscrete, 1.0)
	if got != 120 {
		t.Fatalf("discrete frac=1 want 120, got %g", got)
	}
}

// TestLerpKeyframesKeyTimesNonUniform — keyTimes="0;.2;1" makes the
// middle keyframe land at frac=0.2 rather than 0.5; frac=0.1 is
// halfway through the first segment, frac=0.6 is halfway through
// the second.
func TestLerpKeyframesKeyTimesNonUniform(t *testing.T) {
	vals := []float32{0, 10, 0}
	keyTimes := []float32{0, 0.2, 1}
	// Halfway through first segment [0, 0.2] → vals lerps 0→10 at
	// t=0.5 → 5.
	got := lerpKeyframes(vals, nil, keyTimes, SvgAnimCalcLinear, 0.1)
	if absf(got-5) > 1e-5 {
		t.Fatalf("frac=0.1 want 5, got %g", got)
	}
	// Halfway through second segment [0.2, 1] (span 0.8; 0.6-0.2=
	// 0.4; t=0.5) → vals lerps 10→0 at t=0.5 → 5.
	got = lerpKeyframes(vals, nil, keyTimes, SvgAnimCalcLinear, 0.6)
	if absf(got-5) > 1e-5 {
		t.Fatalf("frac=0.6 want 5, got %g", got)
	}
	// At keyTime boundary = 0.2 exactly → start of second segment,
	// t=0 → value = 10.
	got = lerpKeyframes(vals, nil, keyTimes, SvgAnimCalcLinear, 0.2)
	if absf(got-10) > 1e-5 {
		t.Fatalf("frac=0.2 want 10, got %g", got)
	}
}

// TestLerpKeyframesKeyTimesDiscrete — discrete + keyTimes holds
// keyframe[i] across [keyTimes[i], keyTimes[i+1]).
func TestLerpKeyframesKeyTimesDiscrete(t *testing.T) {
	vals := []float32{10, 20, 30}
	keyTimes := []float32{0, 0.2, 1}
	got := lerpKeyframes(vals, nil, keyTimes, SvgAnimCalcDiscrete, 0.1)
	if got != 10 {
		t.Fatalf("discrete frac=0.1 want 10, got %g", got)
	}
	got = lerpKeyframes(vals, nil, keyTimes, SvgAnimCalcDiscrete, 0.5)
	if got != 20 {
		t.Fatalf("discrete frac=0.5 want 20, got %g", got)
	}
	got = lerpKeyframes(vals, nil, keyTimes, SvgAnimCalcDiscrete, 1.0)
	if got != 30 {
		t.Fatalf("discrete frac=1 want 30, got %g", got)
	}
}

func absf(f float32) float32 {
	return float32(math.Abs(float64(f)))
}
