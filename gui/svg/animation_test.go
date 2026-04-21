package svg

import (
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

func TestAnimationParseAnimateTransformNonRotate(t *testing.T) {
	elem := `<animateTransform type="scale" from="1" to="2" dur="1s">`
	_, ok := parseAnimateTransformElement(elem, groupStyle{})
	if ok {
		t.Fatalf("expected ok=false for non-rotate")
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
