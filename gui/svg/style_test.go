package svg

import (
	"math"
	"testing"

	"github.com/mike-ward/go-gui/gui"
)

// --- Color parsing ---

func TestStyleParseSvgColorNamed(t *testing.T) {
	c := parseSvgColor("red")
	if c.R != 255 || c.G != 0 || c.B != 0 || c.A != 255 {
		t.Fatalf("expected red, got %+v", c)
	}
}

func TestStyleParseSvgColorNone(t *testing.T) {
	c := parseSvgColor("none")
	if c != colorTransparent {
		t.Fatalf("expected transparent, got %+v", c)
	}
}

func TestStyleParseSvgColorInherit(t *testing.T) {
	c := parseSvgColor("inherit")
	if c != colorInherit {
		t.Fatalf("expected inherit sentinel, got %+v", c)
	}
}

func TestStyleParseSvgColorCurrentColor(t *testing.T) {
	c := parseSvgColor("currentColor")
	if c != colorInherit {
		t.Fatalf("expected inherit sentinel, got %+v", c)
	}
}

func TestStyleParseSvgColorEmpty(t *testing.T) {
	c := parseSvgColor("")
	if c != colorInherit {
		t.Fatalf("expected inherit sentinel for empty, got %+v", c)
	}
}

func TestStyleParseSvgColorURL(t *testing.T) {
	c := parseSvgColor("url(#grad1)")
	if c != colorTransparent {
		t.Fatalf("expected transparent for url(), got %+v", c)
	}
}

func TestStyleParseSvgColorUnknown(t *testing.T) {
	c := parseSvgColor("notacolor")
	zero := gui.SvgColor{}
	if c != zero {
		t.Fatalf("expected zero color, got %+v", c)
	}
}

func TestStyleParseHexColorRGB(t *testing.T) {
	c := parseHexColor("#f80")
	if c.R != 0xff || c.G != 0x88 || c.B != 0x00 || c.A != 255 {
		t.Fatalf("expected #ff8800, got %+v", c)
	}
}

func TestStyleParseHexColorRGBA(t *testing.T) {
	c := parseHexColor("#f80a")
	if c.R != 0xff || c.G != 0x88 || c.B != 0x00 || c.A != 0xaa {
		t.Fatalf("expected #ff8800aa, got %+v", c)
	}
}

func TestStyleParseHexColorRRGGBB(t *testing.T) {
	c := parseHexColor("#1a2b3c")
	if c.R != 0x1a || c.G != 0x2b || c.B != 0x3c || c.A != 255 {
		t.Fatalf("expected #1a2b3c, got %+v", c)
	}
}

func TestStyleParseHexColorRRGGBBAA(t *testing.T) {
	c := parseHexColor("#1a2b3c80")
	if c.R != 0x1a || c.G != 0x2b || c.B != 0x3c || c.A != 0x80 {
		t.Fatalf("expected #1a2b3c80, got %+v", c)
	}
}

func TestStyleParseHexColorInvalidLength(t *testing.T) {
	c := parseHexColor("#12")
	if c != colorBlack {
		t.Fatalf("expected black for invalid hex, got %+v", c)
	}
}

func TestStyleHexDigit(t *testing.T) {
	if hexDigit('0') != 0 {
		t.Fatalf("'0' should be 0")
	}
	if hexDigit('9') != 9 {
		t.Fatalf("'9' should be 9")
	}
	if hexDigit('a') != 10 {
		t.Fatalf("'a' should be 10")
	}
	if hexDigit('F') != 15 {
		t.Fatalf("'F' should be 15")
	}
	if hexDigit('g') != 0 {
		t.Fatalf("'g' should be 0")
	}
}

func TestStyleParseRGBColor(t *testing.T) {
	c := parseRGBColor("rgb(128,64,32)")
	if c.R != 128 || c.G != 64 || c.B != 32 || c.A != 255 {
		t.Fatalf("expected rgb(128,64,32), got %+v", c)
	}
}

func TestStyleParseRGBAColor(t *testing.T) {
	c := parseRGBColor("rgba(100,200,50,0.5)")
	if c.R != 100 || c.G != 200 || c.B != 50 {
		t.Fatalf("expected rgb(100,200,50), got %+v", c)
	}
	if c.A != 127 {
		t.Fatalf("expected alpha ~127, got %d", c.A)
	}
}

func TestStyleParseRGBColorClamping(t *testing.T) {
	c := parseRGBColor("rgb(300,-10,128)")
	if c.R != 255 || c.G != 0 || c.B != 128 {
		t.Fatalf("expected clamped (255,0,128), got %+v", c)
	}
}

func TestStyleClampByte(t *testing.T) {
	if clampByte(-1) != 0 {
		t.Fatalf("-1 should clamp to 0")
	}
	if clampByte(128) != 128 {
		t.Fatalf("128 should stay 128")
	}
	if clampByte(256) != 255 {
		t.Fatalf("256 should clamp to 255")
	}
}

// --- Attribute extraction ---

func TestStyleFindAttrPresent(t *testing.T) {
	val, ok := findAttr(`<rect width="100" height="50">`, "width")
	if !ok || val != "100" {
		t.Fatalf("expected '100', got %q ok=%v", val, ok)
	}
}

func TestStyleFindAttrMissing(t *testing.T) {
	_, ok := findAttr(`<rect width="100">`, "height")
	if ok {
		t.Fatalf("expected not found")
	}
}

func TestStyleFindAttrSingleQuotes(t *testing.T) {
	val, ok := findAttr(`<rect width='200'>`, "width")
	if !ok || val != "200" {
		t.Fatalf("expected '200', got %q ok=%v", val, ok)
	}
}

func TestStyleFindAttrNoPartialMatch(t *testing.T) {
	// "width" must not match inside "stroke-width"
	elem := `<path stroke-width="3">`
	_, ok := findAttr(elem, "width")
	if ok {
		t.Fatalf("should not match 'width' inside 'stroke-width'")
	}
}

func TestStyleFindStylePropertyPresent(t *testing.T) {
	val, ok := findStyleProperty("fill:red;stroke:blue", "fill")
	if !ok || val != "red" {
		t.Fatalf("expected 'red', got %q ok=%v", val, ok)
	}
}

func TestStyleFindStylePropertyMissing(t *testing.T) {
	_, ok := findStyleProperty("fill:red", "stroke")
	if ok {
		t.Fatalf("expected not found")
	}
}

func TestStyleFindAttrOrStyleStyleWins(t *testing.T) {
	elem := `<rect fill="blue" style="fill:green">`
	val, ok := findAttrOrStyle(elem, "fill")
	if !ok || val != "green" {
		t.Fatalf("style should win, got %q", val)
	}
}

func TestStyleParseFillURLValid(t *testing.T) {
	id, ok := parseFillURL("url(#grad1)")
	if !ok || id != "grad1" {
		t.Fatalf("expected 'grad1', got %q ok=%v", id, ok)
	}
}

func TestStyleParseFillURLNoHash(t *testing.T) {
	_, ok := parseFillURL("url(grad1)")
	if ok {
		t.Fatalf("expected false for url without #")
	}
}

func TestStyleParseFillURLNotURL(t *testing.T) {
	_, ok := parseFillURL("red")
	if ok {
		t.Fatalf("expected false for non-url")
	}
}

func TestStyleParseClipPathURL(t *testing.T) {
	id, ok := parseClipPathURL(`<g clip-path="url(#clip1)">`)
	if !ok || id != "clip1" {
		t.Fatalf("expected 'clip1', got %q ok=%v", id, ok)
	}
}

func TestStyleParseFilterURL(t *testing.T) {
	id, ok := parseFilterURL(`<g filter="url(#blur1)">`)
	if !ok || id != "blur1" {
		t.Fatalf("expected 'blur1', got %q ok=%v", id, ok)
	}
}

// --- Transform parsing ---

func TestStyleMatrixMultiplyIdentity(t *testing.T) {
	id := identityTransform
	r := matrixMultiply(id, id)
	if r != id {
		t.Fatalf("identity * identity should be identity, got %v", r)
	}
}

func TestStyleMatrixMultiplyTranslate(t *testing.T) {
	id := identityTransform
	tr := [6]float32{1, 0, 0, 1, 10, 20}
	r := matrixMultiply(id, tr)
	if r != tr {
		t.Fatalf("identity * translate should be translate, got %v", r)
	}
}

func TestStyleApplyTransformPtIdentity(t *testing.T) {
	rx, ry := applyTransformPt(5, 7, identityTransform)
	if rx != 5 || ry != 7 {
		t.Fatalf("identity transform should not change point, got (%f,%f)", rx, ry)
	}
}

func TestStyleApplyTransformPtScale(t *testing.T) {
	m := [6]float32{2, 0, 0, 3, 0, 0}
	rx, ry := applyTransformPt(4, 5, m)
	if rx != 8 || ry != 15 {
		t.Fatalf("scale(2,3) on (4,5) should be (8,15), got (%f,%f)", rx, ry)
	}
}

func TestStyleParseTransformTranslate(t *testing.T) {
	m := parseTransform("translate(10,20)")
	if f32Abs(m[4]-10) > 1e-5 || f32Abs(m[5]-20) > 1e-5 {
		t.Fatalf("expected translate(10,20), got %v", m)
	}
}

func TestStyleParseTransformScale(t *testing.T) {
	m := parseTransform("scale(2,3)")
	if f32Abs(m[0]-2) > 1e-5 || f32Abs(m[3]-3) > 1e-5 {
		t.Fatalf("expected scale(2,3), got %v", m)
	}
}

func TestStyleParseTransformRotate90(t *testing.T) {
	m := parseTransform("rotate(90)")
	// cos(90°)=0, sin(90°)=1
	if f32Abs(m[0]) > 1e-5 || f32Abs(m[1]-1) > 1e-5 {
		t.Fatalf("expected rotate(90), got %v", m)
	}
}

func TestStyleParseTransformRotateWithCenter(t *testing.T) {
	m := parseTransform("rotate(90,50,50)")
	// Transform point (50,50) through rotation about itself → (50,50)
	rx, ry := applyTransformPt(50, 50, m)
	if f32Abs(rx-50) > 1e-3 || f32Abs(ry-50) > 1e-3 {
		t.Fatalf("center should be invariant, got (%f,%f)", rx, ry)
	}
}

func TestStyleParseTransformMatrix(t *testing.T) {
	m := parseTransform("matrix(1,0,0,1,5,10)")
	if m[4] != 5 || m[5] != 10 {
		t.Fatalf("expected matrix translate (5,10), got %v", m)
	}
}

func TestStyleParseTransformChained(t *testing.T) {
	m := parseTransform("translate(10,0) scale(2)")
	// translate then scale: point (0,0) → (10,0) → (20,0)?
	// No: compose left-to-right means result = I * translate * scale
	// So [2,0,0,2,10,0] → applying to (1,0): 2*1+10=12, 0
	rx, ry := applyTransformPt(1, 0, m)
	if f32Abs(rx-12) > 1e-4 || f32Abs(ry) > 1e-4 {
		t.Fatalf("chained transform on (1,0) should be (12,0), got (%f,%f)", rx, ry)
	}
}

func TestStyleParseTransformEmpty(t *testing.T) {
	m := parseTransform("")
	if m != identityTransform {
		t.Fatalf("empty transform should be identity, got %v", m)
	}
}

func TestStyleIsIdentityTransformTrue(t *testing.T) {
	if !isIdentityTransform(identityTransform) {
		t.Fatalf("identity should be identity")
	}
}

func TestStyleIsIdentityTransformFalse(t *testing.T) {
	m := [6]float32{2, 0, 0, 1, 0, 0}
	if isIdentityTransform(m) {
		t.Fatalf("scaled matrix should not be identity")
	}
}

// --- Misc helpers ---

func TestStyleParseLength(t *testing.T) {
	if f32Abs(parseLength("12.5px")-12.5) > 1e-5 {
		t.Fatalf("expected 12.5, got %f", parseLength("12.5px"))
	}
}

func TestStyleParseLengthEmpty(t *testing.T) {
	if parseLength("") != 0 {
		t.Fatalf("empty should be 0")
	}
}

func TestStyleClampViewBoxDimNegative(t *testing.T) {
	if clampViewBoxDim(-5) != 0 {
		t.Fatalf("negative should clamp to 0")
	}
}

func TestStyleClampViewBoxDimInRange(t *testing.T) {
	if clampViewBoxDim(500) != 500 {
		t.Fatalf("in-range should be unchanged")
	}
}

func TestStyleClampViewBoxDimOverMax(t *testing.T) {
	if clampViewBoxDim(20000) != maxViewBoxDim {
		t.Fatalf("over max should clamp to %f", maxViewBoxDim)
	}
}

func TestStyleApplyOpacityFull(t *testing.T) {
	c := gui.SvgColor{R: 100, G: 100, B: 100, A: 200}
	r := applyOpacity(c, 1.0)
	if r != c {
		t.Fatalf("opacity 1.0 should not change color")
	}
}

func TestStyleApplyOpacityHalf(t *testing.T) {
	c := gui.SvgColor{R: 100, G: 100, B: 100, A: 200}
	r := applyOpacity(c, 0.5)
	if r.A != 100 {
		t.Fatalf("expected alpha 100, got %d", r.A)
	}
}

// --- Stroke attribute extraction ---

func TestStyleGetStrokeLinecapRound(t *testing.T) {
	cap := getStrokeLinecap(`<path stroke-linecap="round">`)
	if cap != gui.RoundCap {
		t.Fatalf("expected RoundCap, got %d", cap)
	}
}

func TestStyleGetStrokeLinejoinBevel(t *testing.T) {
	join := getStrokeLinejoin(`<path stroke-linejoin="bevel">`)
	if join != gui.BevelJoin {
		t.Fatalf("expected BevelJoin, got %d", join)
	}
}

func TestStyleGetStrokeWidth(t *testing.T) {
	w := getStrokeWidth(`<path stroke-width="3">`)
	if f32Abs(w-3) > 1e-5 {
		t.Fatalf("expected 3, got %f", w)
	}
}

func TestStyleGetStrokeDasharrayValid(t *testing.T) {
	da := getStrokeDasharray(`<path stroke-dasharray="5 3">`)
	if len(da) != 2 || da[0] != 5 || da[1] != 3 {
		t.Fatalf("expected [5,3], got %v", da)
	}
}

func TestStyleGetStrokeDasharrayNone(t *testing.T) {
	da := getStrokeDasharray(`<path stroke-dasharray="none">`)
	if da != nil {
		t.Fatalf("expected nil for 'none', got %v", da)
	}
}

func TestStyleGetStrokeDasharrayOddDoubled(t *testing.T) {
	da := getStrokeDasharray(`<path stroke-dasharray="5 3 2">`)
	if len(da) != 6 {
		t.Fatalf("odd-count should double, got len=%d", len(da))
	}
	if da[0] != 5 || da[3] != 5 {
		t.Fatalf("doubled array mismatch: %v", da)
	}
}

// --- parseTransform: skewX ---

func TestStyleParseTransformSkewX(t *testing.T) {
	m := parseTransform("skewX(45)")
	expected := float32(math.Tan(45.0 * math.Pi / 180.0))
	if f32Abs(m[2]-expected) > 1e-4 {
		t.Fatalf("expected m[2]=%f, got %f", expected, m[2])
	}
}
