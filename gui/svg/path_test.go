package svg

import (
	"math"
	"testing"
)

// --- Small pure functions ---

func TestPathIsNumberTokenDigit(t *testing.T) {
	if !isNumberToken("42") {
		t.Fatalf("'42' should be a number token")
	}
}

func TestPathIsNumberTokenNegative(t *testing.T) {
	if !isNumberToken("-3.5") {
		t.Fatalf("'-3.5' should be a number token")
	}
}

func TestPathIsNumberTokenDot(t *testing.T) {
	if !isNumberToken(".5") {
		t.Fatalf("'.5' should be a number token")
	}
}

func TestPathIsNumberTokenCommand(t *testing.T) {
	if isNumberToken("M") {
		t.Fatalf("'M' should not be a number token")
	}
}

func TestPathIsNumberTokenEmpty(t *testing.T) {
	if isNumberToken("") {
		t.Fatalf("empty should not be a number token")
	}
}

func TestPathParseF32Valid(t *testing.T) {
	if f32Abs(parseF32("3.14")-3.14) > 1e-4 {
		t.Fatalf("expected 3.14, got %f", parseF32("3.14"))
	}
}

func TestPathParseF32Negative(t *testing.T) {
	if f32Abs(parseF32("-2.5")+2.5) > 1e-5 {
		t.Fatalf("expected -2.5, got %f", parseF32("-2.5"))
	}
}

func TestPathParseF32Invalid(t *testing.T) {
	if parseF32("abc") != 0 {
		t.Fatalf("invalid should return 0")
	}
}

func TestPathParseF32NonFinite(t *testing.T) {
	if parseF32("NaN") != 0 {
		t.Fatalf("NaN should return 0")
	}
	if parseF32("Inf") != 0 {
		t.Fatalf("Inf should return 0")
	}
}

func TestPathParseNumberList(t *testing.T) {
	nums := parseNumberList("1 2,3 4")
	if len(nums) != 4 {
		t.Fatalf("expected 4 numbers, got %d", len(nums))
	}
	if nums[0] != 1 || nums[1] != 2 || nums[2] != 3 || nums[3] != 4 {
		t.Fatalf("expected [1,2,3,4], got %v", nums)
	}
}

func TestPathParseNumberListEmpty(t *testing.T) {
	nums := parseNumberList("")
	if len(nums) != 0 {
		t.Fatalf("expected empty, got %v", nums)
	}
}

// --- tokenizePath ---

func TestPathTokenizeSimple(t *testing.T) {
	tokens := tokenizePath("M 10 20 L 30 40")
	expected := []string{"M", "10", "20", "L", "30", "40"}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i := range expected {
		if tokens[i] != expected[i] {
			t.Fatalf("token[%d] = %q, want %q", i, tokens[i], expected[i])
		}
	}
}

func TestPathTokenizeCommas(t *testing.T) {
	tokens := tokenizePath("M10,20L30,40")
	expected := []string{"M", "10", "20", "L", "30", "40"}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i := range expected {
		if tokens[i] != expected[i] {
			t.Fatalf("token[%d] = %q, want %q", i, tokens[i], expected[i])
		}
	}
}

func TestPathTokenizeNegativesAndDots(t *testing.T) {
	tokens := tokenizePath("M-5-10L3.5.7")
	// Should split: M, -5, -10, L, 3.5, .7
	if len(tokens) != 6 {
		t.Fatalf("expected 6 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[1] != "-5" || tokens[2] != "-10" {
		t.Fatalf("negative split failed: %v", tokens)
	}
	if tokens[4] != "3.5" || tokens[5] != ".7" {
		t.Fatalf("dot split failed: %v", tokens)
	}
}

func TestPathTokenizeEmpty(t *testing.T) {
	tokens := tokenizePath("")
	if len(tokens) != 0 {
		t.Fatalf("expected empty, got %v", tokens)
	}
}

// --- parsePathD ---

func TestPathParsePathDMoveTo(t *testing.T) {
	segs := parsePathD("M 10 20")
	if len(segs) != 1 || segs[0].Cmd != CmdMoveTo {
		t.Fatalf("expected 1 MoveTo, got %v", segs)
	}
	if segs[0].Points[0] != 10 || segs[0].Points[1] != 20 {
		t.Fatalf("expected (10,20), got %v", segs[0].Points)
	}
}

func TestPathParsePathDLineTo(t *testing.T) {
	segs := parsePathD("M 0 0 L 10 20")
	if len(segs) != 2 || segs[1].Cmd != CmdLineTo {
		t.Fatalf("expected MoveTo+LineTo, got %v", segs)
	}
}

func TestPathParsePathDRelative(t *testing.T) {
	segs := parsePathD("M 10 10 l 5 5")
	if len(segs) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segs))
	}
	// Relative: (10+5, 10+5) = (15, 15)
	if f32Abs(segs[1].Points[0]-15) > 1e-5 || f32Abs(segs[1].Points[1]-15) > 1e-5 {
		t.Fatalf("expected (15,15), got %v", segs[1].Points)
	}
}

func TestPathParsePathDHorizontal(t *testing.T) {
	segs := parsePathD("M 0 5 H 10")
	if len(segs) != 2 || segs[1].Cmd != CmdLineTo {
		t.Fatalf("expected LineTo for H, got %v", segs)
	}
	if segs[1].Points[0] != 10 || segs[1].Points[1] != 5 {
		t.Fatalf("H should keep Y=5, got %v", segs[1].Points)
	}
}

func TestPathParsePathDVertical(t *testing.T) {
	segs := parsePathD("M 5 0 V 10")
	if len(segs) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segs))
	}
	if segs[1].Points[0] != 5 || segs[1].Points[1] != 10 {
		t.Fatalf("V should keep X=5, got %v", segs[1].Points)
	}
}

func TestPathParsePathDClose(t *testing.T) {
	segs := parsePathD("M 0 0 L 10 0 L 10 10 Z")
	if segs[len(segs)-1].Cmd != CmdClose {
		t.Fatalf("last segment should be Close")
	}
}

func TestPathParsePathDCubic(t *testing.T) {
	segs := parsePathD("M 0 0 C 10 0 10 10 0 10")
	if len(segs) != 2 || segs[1].Cmd != CmdCubicTo {
		t.Fatalf("expected CubicTo, got %v", segs)
	}
	if len(segs[1].Points) != 6 {
		t.Fatalf("CubicTo should have 6 args, got %d", len(segs[1].Points))
	}
}

func TestPathParsePathDQuad(t *testing.T) {
	segs := parsePathD("M 0 0 Q 5 10 10 0")
	if len(segs) != 2 || segs[1].Cmd != CmdQuadTo {
		t.Fatalf("expected QuadTo, got %v", segs)
	}
	if len(segs[1].Points) != 4 {
		t.Fatalf("QuadTo should have 4 args, got %d", len(segs[1].Points))
	}
}

func TestPathParsePathDEmpty(t *testing.T) {
	segs := parsePathD("")
	if len(segs) != 0 {
		t.Fatalf("expected empty, got %v", segs)
	}
}

// --- vectorAngle ---

func TestPathVectorAngleParallel(t *testing.T) {
	a := vectorAngle(1, 0, 1, 0)
	if f32Abs(a) > 1e-5 {
		t.Fatalf("parallel vectors should have angle 0, got %f", a)
	}
}

func TestPathVectorAnglePerpendicular(t *testing.T) {
	a := vectorAngle(1, 0, 0, 1)
	if f32Abs(a-math.Pi/2) > 1e-4 {
		t.Fatalf("expected pi/2, got %f", a)
	}
}

func TestPathVectorAngleOpposite(t *testing.T) {
	a := vectorAngle(1, 0, -1, 0)
	if f32Abs(a-math.Pi) > 1e-4 {
		t.Fatalf("expected pi, got %f", a)
	}
}

func TestPathVectorAngleZeroLength(t *testing.T) {
	a := vectorAngle(0, 0, 1, 0)
	if a != 0 {
		t.Fatalf("zero-length should return 0, got %f", a)
	}
}

// --- arcToCubic ---

func TestPathArcToCubicZeroRadius(t *testing.T) {
	segs := arcToCubic(0, 0, 0, 0, 0, false, true, 10, 0)
	if len(segs) != 1 || segs[0].Cmd != CmdLineTo {
		t.Fatalf("zero radius should produce LineTo, got %v", segs)
	}
}

func TestPathArcToCubicQuarterCircle(t *testing.T) {
	segs := arcToCubic(10, 0, 10, 10, 0, false, true, 0, 10)
	if len(segs) < 1 {
		t.Fatalf("expected at least 1 segment, got %d", len(segs))
	}
	for _, seg := range segs {
		if seg.Cmd != CmdCubicTo {
			t.Fatalf("expected CubicTo segments, got cmd=%d", seg.Cmd)
		}
	}
	// Last segment endpoint should be near (0,10)
	last := segs[len(segs)-1]
	ex := last.Points[4]
	ey := last.Points[5]
	if f32Abs(ex) > 0.5 || f32Abs(ey-10) > 0.5 {
		t.Fatalf("endpoint should be near (0,10), got (%f,%f)", ex, ey)
	}
}
