package svg

import (
	"strings"
	"testing"

	"github.com/mike-ward/go-gui/gui"
)

// --- Time parsing ---

func TestAnimationParseTimeValueSeconds(t *testing.T) {
	v := parseTimeValue("1.5s")
	if f32Abs(v-1.5) > 1e-5 {
		t.Fatalf("expected 1.5, got %f", v)
	}
}

func TestAnimationParseTimeValueMilliseconds(t *testing.T) {
	v := parseTimeValue("200ms")
	if f32Abs(v-0.2) > 1e-5 {
		t.Fatalf("expected 0.2, got %f", v)
	}
}

func TestAnimationParseTimeValueBare(t *testing.T) {
	v := parseTimeValue("3")
	if f32Abs(v-3) > 1e-5 {
		t.Fatalf("expected 3, got %f", v)
	}
}

// --- Float lists ---

func TestAnimationParseSemicolonFloats(t *testing.T) {
	vals := parseSemicolonFloats("0;0.5;1")
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
	if vals[0] != 0 || f32Abs(vals[1]-0.5) > 1e-5 || vals[2] != 1 {
		t.Fatalf("expected [0,0.5,1], got %v", vals)
	}
}

func TestAnimationParseSemicolonFloatsEmpty(t *testing.T) {
	vals := parseSemicolonFloats("")
	if len(vals) != 0 {
		t.Fatalf("expected empty, got %v", vals)
	}
}

func TestAnimationParseSpaceFloats(t *testing.T) {
	vals := parseSpaceFloats("10 20 30")
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
	if vals[0] != 10 || vals[1] != 20 || vals[2] != 30 {
		t.Fatalf("expected [10,20,30], got %v", vals)
	}
}

func TestAnimationParseSpaceFloatsEmpty(t *testing.T) {
	vals := parseSpaceFloats("")
	if len(vals) != 0 {
		t.Fatalf("expected empty, got %v", vals)
	}
}

// --- parseAnimateElement ---

func TestAnimationParseAnimateElementValid(t *testing.T) {
	elem := `<animate attributeName="opacity" values="1;0;1" dur="2s" begin="0.5s">`
	gs := groupStyle{GroupID: "g1"}
	anim, ok := parseAnimateElement(elem, gs)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if anim.Kind != gui.SvgAnimOpacity {
		t.Fatalf("expected SvgAnimOpacity, got %d", anim.Kind)
	}
	if anim.GroupID != "g1" {
		t.Fatalf("expected GroupID 'g1', got %q", anim.GroupID)
	}
	if len(anim.Values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(anim.Values))
	}
	if f32Abs(anim.DurSec-2) > 1e-5 {
		t.Fatalf("expected dur=2, got %f", anim.DurSec)
	}
	if f32Abs(anim.BeginSec-0.5) > 1e-5 {
		t.Fatalf("expected begin=0.5, got %f", anim.BeginSec)
	}
}

func TestAnimationParseAnimateElementNonOpacity(t *testing.T) {
	elem := `<animate attributeName="fill" values="red;blue" dur="1s">`
	_, ok := parseAnimateElement(elem, groupStyle{})
	if ok {
		t.Fatalf("expected ok=false for non-opacity")
	}
}

func TestAnimationParseAnimateElementNoValues(t *testing.T) {
	elem := `<animate attributeName="opacity" dur="1s">`
	_, ok := parseAnimateElement(elem, groupStyle{})
	if ok {
		t.Fatalf("expected ok=false for missing values")
	}
}

func TestAnimationParseAnimateElementZeroDur(t *testing.T) {
	elem := `<animate attributeName="opacity" values="1;0" dur="0s">`
	_, ok := parseAnimateElement(elem, groupStyle{})
	if ok {
		t.Fatalf("expected ok=false for zero duration")
	}
}

// --- parseAnimateTransformElement ---

func TestAnimationParseAnimateTransformValid(t *testing.T) {
	elem := `<animateTransform type="rotate" from="0 50 50" to="360 50 50" dur="3s">`
	gs := groupStyle{GroupID: "wheel"}
	anim, ok := parseAnimateTransformElement(elem, gs)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if anim.Kind != gui.SvgAnimRotate {
		t.Fatalf("expected SvgAnimRotate, got %d", anim.Kind)
	}
	if f32Abs(anim.Values[0]) > 1e-5 || f32Abs(anim.Values[1]-360) > 1e-5 {
		t.Fatalf("expected from=0 to=360, got %v", anim.Values)
	}
	if f32Abs(anim.CenterX-50) > 1e-5 || f32Abs(anim.CenterY-50) > 1e-5 {
		t.Fatalf("expected center (50,50), got (%f,%f)", anim.CenterX, anim.CenterY)
	}
}

func TestAnimationParseAnimateTransformUnknownType(t *testing.T) {
	elem := `<animateTransform type="skewX" from="0" to="30" dur="1s">`
	_, ok := parseAnimateTransformElement(elem, groupStyle{})
	if ok {
		t.Fatalf("expected ok=false for unsupported transform type")
	}
}

func TestAnimationParseAnimateTransformValuesForm(t *testing.T) {
	elem := `<animateTransform attributeName="transform" type="rotate" ` +
		`dur="0.75s" values="0 12 12;360 12 12" repeatCount="indefinite"/>`
	gs := groupStyle{GroupID: "ring"}
	anim, ok := parseAnimateTransformElement(elem, gs)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if anim.Kind != gui.SvgAnimRotate {
		t.Fatalf("expected SvgAnimRotate, got %d", anim.Kind)
	}
	if len(anim.Values) != 2 {
		t.Fatalf("expected 2 angles, got %d", len(anim.Values))
	}
	if f32Abs(anim.Values[0]) > 1e-5 || f32Abs(anim.Values[1]-360) > 1e-5 {
		t.Fatalf("expected [0,360], got %v", anim.Values)
	}
	if f32Abs(anim.CenterX-12) > 1e-5 || f32Abs(anim.CenterY-12) > 1e-5 {
		t.Fatalf("expected center (12,12), got (%f,%f)",
			anim.CenterX, anim.CenterY)
	}
	if f32Abs(anim.DurSec-0.75) > 1e-5 {
		t.Fatalf("expected dur=0.75, got %f", anim.DurSec)
	}
}

func TestAnimationParseAnimateTransformValuesMulti(t *testing.T) {
	elem := `<animateTransform type="rotate" dur="1s" ` +
		`values="0 5 5;90 5 5;180 5 5;360 5 5"/>`
	anim, ok := parseAnimateTransformElement(elem, groupStyle{GroupID: "g"})
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if len(anim.Values) != 4 {
		t.Fatalf("expected 4 angles, got %d", len(anim.Values))
	}
	if anim.Values[0] != 0 || anim.Values[3] != 360 {
		t.Fatalf("expected [0,...,360], got %v", anim.Values)
	}
}

func TestAnimationParseAnimateTransformValuesSinglePoint(t *testing.T) {
	elem := `<animateTransform type="rotate" dur="1s" values="0 5 5"/>`
	_, ok := parseAnimateTransformElement(elem, groupStyle{})
	if ok {
		t.Fatalf("expected ok=false for single-keyframe values")
	}
}

// --- End-to-end parse of a rotate-via-values asset ---

const rotateValuesAsset = `<svg viewBox="0 0 24 24" ` +
	`xmlns="http://www.w3.org/2000/svg">` +
	`<path id="ring" d="M10,10 L14,10 L14,14 L10,14Z">` +
	`<animateTransform attributeName="transform" type="rotate" ` +
	`dur="0.75s" values="0 12 12;360 12 12" repeatCount="indefinite"/>` +
	`</path></svg>`

func TestAnimationRotateValuesEndToEnd(t *testing.T) {
	parsed, err := New().ParseSvg(rotateValuesAsset)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(parsed.Animations) == 0 {
		t.Fatal("expected at least one parsed animation")
	}
	anim := parsed.Animations[0]
	if anim.Kind != gui.SvgAnimRotate {
		t.Fatalf("expected SvgAnimRotate, got %d", anim.Kind)
	}
	if anim.GroupID != "ring" {
		t.Fatalf("expected GroupID 'ring', got %q", anim.GroupID)
	}
	if len(anim.Values) != 2 ||
		anim.Values[0] != 0 || anim.Values[1] != 360 {
		t.Fatalf("expected [0,360], got %v", anim.Values)
	}
	if anim.CenterX != 12 || anim.CenterY != 12 {
		t.Fatalf("expected center (12,12), got (%f,%f)",
			anim.CenterX, anim.CenterY)
	}
	if len(parsed.Paths) == 0 {
		t.Fatal("expected at least one path")
	}
	if parsed.Paths[0].GroupID != anim.GroupID {
		t.Fatalf("path GroupID %q != anim GroupID %q — animations "+
			"would not bind to the shape",
			parsed.Paths[0].GroupID, anim.GroupID)
	}
}

// rotateNoIDAsset covers the synthetic-ID branch: a shape with
// inline animation but no id= attribute.
const rotateNoIDAsset = `<svg viewBox="0 0 24 24" ` +
	`xmlns="http://www.w3.org/2000/svg">` +
	`<path d="M10,10 L14,10 L14,14 L10,14Z">` +
	`<animateTransform attributeName="transform" type="rotate" ` +
	`dur="0.75s" values="0 12 12;360 12 12" repeatCount="indefinite"/>` +
	`</path></svg>`

// TestAnimationGroupSynthIDBindsSiblings — when a bare <g> wraps
// multiple shapes plus a shared <animateTransform>, the group
// must get a synthesized GroupID so every descendant shape and
// the animation all share it. Previously the animation and
// paths all had GroupID="" and the renderer's non-empty guard
// silently dropped the animation (e.g. 8-dots-rotate asset).
func TestAnimationGroupSynthIDBindsSiblings(t *testing.T) {
	asset := `<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">` +
		`<g>` +
		`<circle cx="3" cy="12" r="2"/>` +
		`<circle cx="21" cy="12" r="2"/>` +
		`<animateTransform attributeName="transform" type="rotate" ` +
		`dur="1.5s" values="0 12 12;360 12 12" repeatCount="indefinite"/>` +
		`</g></svg>`
	parsed, err := New().ParseSvg(asset)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(parsed.Paths) != 2 || len(parsed.Animations) != 1 {
		t.Fatalf("want 2 paths + 1 anim, got %d paths + %d anims",
			len(parsed.Paths), len(parsed.Animations))
	}
	gid := parsed.Animations[0].GroupID
	if gid == "" {
		t.Fatal("animation GroupID must be non-empty")
	}
	for i, p := range parsed.Paths {
		if p.GroupID != gid {
			t.Fatalf("path[%d] GroupID %q != anim GroupID %q",
				i, p.GroupID, gid)
		}
	}
}

func TestAnimationSynthIDCouplesShapeAndAnim(t *testing.T) {
	parsed, err := New().ParseSvg(rotateNoIDAsset)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(parsed.Paths) == 0 || len(parsed.Animations) == 0 {
		t.Fatal("expected at least one path and one animation")
	}
	pID := parsed.Paths[0].GroupID
	aID := parsed.Animations[0].GroupID
	if pID == "" {
		t.Fatal("synthetic id should be non-empty on the path")
	}
	if pID != aID {
		t.Fatalf("path %q != anim %q — animation would miss",
			pID, aID)
	}
}

// --- maxKeyframes DoS caps ---

// TestParseSemicolonFloatsCapsAtMaxKeyframes — a pathological
// input with more than maxKeyframes entries is truncated so
// allocation stays bounded.
func TestParseSemicolonFloatsCapsAtMaxKeyframes(t *testing.T) {
	s := strings.Repeat("1;", maxKeyframes+100)
	got := parseSemicolonFloats(s)
	if len(got) > maxKeyframes {
		t.Fatalf("want len<=%d, got %d", maxKeyframes, len(got))
	}
}

// TestParseRotateValuesCapsAtMaxKeyframes — oversized rotate
// values list is truncated; remaining angles still valid.
func TestParseRotateValuesCapsAtMaxKeyframes(t *testing.T) {
	s := strings.Repeat("0 12 12;", maxKeyframes+50)
	angles, _, _, ok := parseRotateValues(s)
	if !ok {
		t.Fatal("parse failed on truncated-but-valid input")
	}
	if len(angles) > maxKeyframes {
		t.Fatalf("want len<=%d, got %d", maxKeyframes, len(angles))
	}
}

// --- parseFreeze ---

func TestParseFreeze_RecognizesOnlyFreeze(t *testing.T) {
	if !parseFreeze(`<animate fill="freeze">`) {
		t.Fatal("fill=freeze should be recognized")
	}
	if parseFreeze(`<animate fill="remove">`) {
		t.Fatal("fill=remove should be false")
	}
	if parseFreeze(`<animate>`) {
		t.Fatal("missing fill attr should be false")
	}
	if parseFreeze(`<animate fill="FREEZE">`) {
		t.Fatal("case-sensitive: FREEZE must not match")
	}
}

// --- parseRepeatCycle ---

func TestParseRepeatCycle_IndefiniteReturnsDur(t *testing.T) {
	// repeatCount="indefinite" yields dur.
	c := parseRepeatCycle(`<animate repeatCount="indefinite">`, 2.5)
	if f32Abs(c-2.5) > 1e-5 {
		t.Fatalf("expected cycle=2.5, got %f", c)
	}
	// repeatDur="indefinite" yields dur.
	c = parseRepeatCycle(`<animate repeatDur="indefinite">`, 1.0)
	if f32Abs(c-1.0) > 1e-5 {
		t.Fatalf("expected cycle=1.0, got %f", c)
	}
	// No repeat attrs → 0 (play once).
	if parseRepeatCycle(`<animate>`, 1.0) != 0 {
		t.Fatal("no repeat attr should yield cycle=0")
	}
}

func TestParseRepeatCycle_FiniteCountProducesDurTimesN(t *testing.T) {
	c := parseRepeatCycle(`<animate repeatCount="3">`, 2)
	if f32Abs(c-6) > 1e-5 {
		t.Fatalf("expected cycle=6, got %f", c)
	}
}

func TestParseRepeatCycle_RepeatDurProducesExplicitSeconds(t *testing.T) {
	c := parseRepeatCycle(`<animate repeatDur="4s">`, 2)
	if f32Abs(c-4) > 1e-5 {
		t.Fatalf("expected cycle=4, got %f", c)
	}
}

func TestParseRepeatCycle_ClampsHostileCount(t *testing.T) {
	// Hostile repeatCount is clamped so dur*n stays finite and
	// below maxCycleSec. Authoring mistakes or malicious SVGs
	// cannot poison downstream timing math with +Inf.
	c := parseRepeatCycle(`<animate repeatCount="1e30">`, 1)
	if c <= 0 || c > maxCycleSec {
		t.Fatalf("expected clamp into (0, %f], got %f", maxCycleSec, c)
	}
	c = parseRepeatCycle(`<animate repeatDur="1e30s">`, 1)
	if c <= 0 || c > maxCycleSec {
		t.Fatalf("expected clamp into (0, %f], got %f", maxCycleSec, c)
	}
}

// --- clampCycle ---

func TestClampCycle_BoundsAtMaxAndFloor(t *testing.T) {
	if clampCycle(0) != 0 {
		t.Fatal("0 should pass through as 0")
	}
	if clampCycle(-5) != 0 {
		t.Fatal("negative should clamp to 0")
	}
	if clampCycle(maxCycleSec+1) != maxCycleSec {
		t.Fatalf("above max should clamp to %f", maxCycleSec)
	}
	if clampCycle(10) != 10 {
		t.Fatal("in-range should pass through")
	}
}

// --- Opacity sub-target detection ---

func TestParseAnimateElement_FillOpacityTargetsFill(t *testing.T) {
	elem := `<animate attributeName="fill-opacity" values="1;0" dur="1s">`
	a, ok := parseAnimateElement(elem, groupStyle{GroupID: "g"})
	if !ok {
		t.Fatal("expected ok=true for fill-opacity")
	}
	if a.Target != gui.SvgAnimTargetFill {
		t.Fatalf("expected Target=Fill, got %d", a.Target)
	}
	if a.Kind != gui.SvgAnimOpacity {
		t.Fatalf("expected Kind=Opacity, got %d", a.Kind)
	}
}

func TestParseAnimateElement_StrokeOpacityTargetsStroke(t *testing.T) {
	elem := `<animate attributeName="stroke-opacity" values="1;0" dur="1s">`
	a, ok := parseAnimateElement(elem, groupStyle{GroupID: "g"})
	if !ok {
		t.Fatal("expected ok=true for stroke-opacity")
	}
	if a.Target != gui.SvgAnimTargetStroke {
		t.Fatalf("expected Target=Stroke, got %d", a.Target)
	}
}

func TestParseAnimateElement_OpacityTargetsAll(t *testing.T) {
	elem := `<animate attributeName="opacity" values="1;0" dur="1s">`
	a, ok := parseAnimateElement(elem, groupStyle{GroupID: "g"})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if a.Target != gui.SvgAnimTargetAll {
		t.Fatalf("expected Target=All, got %d", a.Target)
	}
}

// --- Rotate center transform baking ---

func TestParseAnimateTransform_RotateCenterUsesInheritedTransform(t *testing.T) {
	// Parent scale(2) translate(10,20) must fold into the rotate
	// pivot so the animated rotation pivots in absolute SVG space.
	inh := groupStyle{
		GroupID:   "g",
		Transform: [6]float32{2, 0, 0, 2, 10, 20},
	}
	elem := `<animateTransform type="rotate" dur="1s" values="0 5 5;360 5 5"/>`
	a, ok := parseAnimateTransformElement(elem, inh)
	if !ok {
		t.Fatal("expected ok=true")
	}
	// applyTransformPt on (5,5) with the matrix yields (20,30).
	if f32Abs(a.CenterX-20) > 1e-5 || f32Abs(a.CenterY-30) > 1e-5 {
		t.Fatalf("expected center (20,30), got (%f,%f)",
			a.CenterX, a.CenterY)
	}
}

// Translate / scale pair values MUST remain in local space.
// Baking the ancestor transform here would apply it twice at
// render time.
func TestPairedTransform_TranslateValuesIgnoreInheritedTransform(t *testing.T) {
	inh := groupStyle{
		GroupID:   "g",
		Transform: [6]float32{2, 0, 0, 2, 10, 20},
	}
	elem := `<animateTransform type="translate" dur="1s" values="0 0;30 40"/>`
	a, ok := parsePairedAnimateTransform(elem, inh, gui.SvgAnimTranslate)
	if !ok {
		t.Fatal("expected ok=true")
	}
	// Values must equal the raw input, not (10,20;70,100).
	want := []float32{0, 0, 30, 40}
	if len(a.Values) != len(want) {
		t.Fatalf("len mismatch: %v", a.Values)
	}
	for i, v := range want {
		if f32Abs(a.Values[i]-v) > 1e-5 {
			t.Fatalf("values[%d]=%f, want %f (inherited transform "+
				"must not be applied)", i, a.Values[i], v)
		}
	}
}

// --- Cycle propagation through resolveBegins ---

// When multiple animations share a syncbase chain, resolveBegins
// must propagate the largest derived cycle to chain participants
// whose own Cycle is 0. Without this, a freeze-chained sequence
// stalls after the first loop.
func TestResolveBegins_PropagatesGlobalCycle(t *testing.T) {
	// Construct a chain of two animations in a group. The first
	// sets BeginSec=0 and repeatCount="indefinite" so it derives a
	// Cycle from parseRepeatCycle. The second has BeginSec=0 but
	// no cycle; it must inherit the global cycle since its begin
	// spec references the first.
	asset := `<svg viewBox="0 0 10 10" xmlns="http://www.w3.org/2000/svg">` +
		`<rect id="box" x="0" y="0" width="10" height="10">` +
		`<animate id="a1" attributeName="opacity" values="1;0" ` +
		`dur="1s" begin="0s" repeatCount="indefinite"/>` +
		`<animate id="a2" attributeName="opacity" values="0;1" ` +
		`dur="1s" begin="a1.end"/>` +
		`</rect></svg>`
	parsed, err := New().ParseSvg(asset)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(parsed.Animations) != 2 {
		t.Fatalf("expected 2 animations, got %d", len(parsed.Animations))
	}
	// Both must carry a positive cycle; second inherits from first.
	for i, a := range parsed.Animations {
		if a.Cycle <= 0 {
			t.Fatalf("anim[%d] expected Cycle>0, got %f", i, a.Cycle)
		}
	}
}
