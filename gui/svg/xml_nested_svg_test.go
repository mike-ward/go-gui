package svg

import (
	"math"
	"strings"
	"testing"

	"github.com/mike-ward/go-gui/gui"
)

const epsilon32 = 1e-4

// firstShapePath returns the first VectorPath whose segments are
// non-empty (i.e. a real shape and not a clip-mask placeholder).
func firstShapePath(t *testing.T, vg *VectorGraphic) VectorPath {
	t.Helper()
	for _, p := range vg.Paths {
		if len(p.Segments) > 0 {
			return p
		}
	}
	t.Fatalf("no shape paths in parsed SVG")
	return VectorPath{}
}

func nearEq32(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= epsilon32
}

func assertTransform(t *testing.T, got [6]float32, want [6]float32) {
	t.Helper()
	for i := range 6 {
		if !nearEq32(got[i], want[i]) {
			t.Fatalf("transform[%d] = %v, want %v (full got=%v want=%v)",
				i, got[i], want[i], got, want)
		}
	}
}

// Regression: prior to nested-<svg> support, descendants of an inner
// <svg> were silently dropped. The inner circle must now render.
func TestNestedSvg_InnerContentRenders(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg x="10" y="10" width="50" height="50" viewBox="0 0 10 10">
			<circle cx="5" cy="5" r="5" fill="red"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(vg.Paths) == 0 {
		t.Fatal("expected inner circle to produce a path; got 0")
	}
}

// Translate-only nested viewport (no viewBox): transform reduces to
// translate(x, y). Multiple paths inherit the same composed matrix.
func TestNestedSvg_TranslateOnly(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg x="10" y="20">
			<rect x="0" y="0" width="5" height="5"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{1, 0, 0, 1, 10, 20})
}

// viewBox on inner svg establishes user-space scaling. inner viewBox
// 0 0 10 10 mapped onto a 50x50 viewport at (10,10) -> uniform scale
// 5, translate (10,10).
func TestNestedSvg_ViewBoxScale(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg x="10" y="10" width="50" height="50" viewBox="0 0 10 10">
			<rect x="0" y="0" width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{5, 0, 0, 5, 10, 10})
}

// preserveAspectRatio="xMidYMid meet" picks min(sx,sy) and centers
// the residual in the major axis. Viewport 50x100, inner viewBox
// 0 0 10 10 -> scale 5, y residual 50, half = 25.
func TestNestedSvg_PreserveAspectMeet(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg x="0" y="0" width="50" height="100" viewBox="0 0 10 10"
			preserveAspectRatio="xMidYMid meet">
			<rect width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{5, 0, 0, 5, 0, 25})
}

// preserveAspectRatio="none" keeps independent scales.
func TestNestedSvg_PreserveAspectNone(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg x="0" y="0" width="50" height="100" viewBox="0 0 10 10"
			preserveAspectRatio="none">
			<rect width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{5, 0, 0, 10, 0, 0})
}

// SVG2: `transform=` on a nested <svg> applies AFTER the viewport
// mapping. Composed = parent × elemTransform × viewportTx.
// elemTx = translate(7,3); viewportTx = translate(10,20) (no viewBox).
// Final translate = (17,23); scale untouched.
func TestNestedSvg_TransformAttrComposition(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg x="10" y="20" transform="translate(7,3)">
			<rect x="0" y="0" width="5" height="5"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{1, 0, 0, 1, 17, 23})
}

// Missing width/height on inner svg defaults to 100% of the parent
// viewport. Parent 100x100, inner viewBox 0 0 50 50 -> scale 2.
func TestNestedSvg_DefaultsToParentViewport(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg viewBox="0 0 50 50">
			<rect width="50" height="50"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{2, 0, 0, 2, 0, 0})
}

// Deep-but-legal nesting: parseSvgContent bails at maxGroupDepth and
// never recurses unbounded. Stay below the decoder's +8 slack so the
// XML tree itself decodes cleanly; the assertion is that parsing
// returns without panic, with the rect dropped past the depth cap.
func TestNestedSvg_DepthCapBounded(t *testing.T) {
	depth := maxGroupDepth + 4
	var b strings.Builder
	b.WriteString(`<svg viewBox="0 0 100 100">`)
	for range depth {
		b.WriteString(`<svg>`)
	}
	b.WriteString(`<rect x="0" y="0" width="1" height="1"/>`)
	for range depth {
		b.WriteString(`</svg>`)
	}
	b.WriteString(`</svg>`)
	vg, err := parseSvg(b.String())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(vg.Paths) != 0 {
		t.Fatalf("expected depth cap to drop deeply nested rect; got %d paths",
			len(vg.Paths))
	}
}

// preserveAspectRatio="xMidYMid slice" picks max(sx,sy) and centers
// the residual in the major axis. Viewport 50x100, inner viewBox
// 0 0 10 10 -> scale 10 (uniform max), x residual -50, half = -25.
func TestNestedSvg_PreserveAspectSlice(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg x="0" y="0" width="50" height="100" viewBox="0 0 10 10"
			preserveAspectRatio="xMidYMid slice">
			<rect width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{10, 0, 0, 10, -25, 0})
}

// Negative viewBox origin: vbX=-5, vbY=-5 shifts user-space so (0,0)
// in descendant coords lands at +5,+5 in outer. Viewport 100x100,
// scale 10. tx = 0 + 0*xFrac - 10*(-5) = 50.
func TestNestedSvg_NegativeViewBoxOrigin(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg x="0" y="0" width="100" height="100"
			viewBox="-5 -5 10 10" preserveAspectRatio="none">
			<rect width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{10, 0, 0, 10, 50, 50})
}

// Sibling nested <svg> elements at the same depth must each see the
// outer viewport, not the previous sibling's. Test: two siblings,
// each width="50%" of parent 100. Both should scale identically.
func TestNestedSvg_SiblingViewportRestored(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg x="0" y="0" width="50%" height="50%" viewBox="0 0 10 10"
			preserveAspectRatio="none">
			<rect width="10" height="10"/>
		</svg>
		<svg x="50" y="50" width="50%" height="50%" viewBox="0 0 10 10"
			preserveAspectRatio="none">
			<rect width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var shapes []VectorPath
	for _, p := range vg.Paths {
		if len(p.Segments) > 0 {
			shapes = append(shapes, p)
		}
	}
	if len(shapes) != 2 {
		t.Fatalf("expected 2 shape paths; got %d", len(shapes))
	}
	assertTransform(t, shapes[0].Transform, [6]float32{5, 0, 0, 5, 0, 0})
	assertTransform(t, shapes[1].Transform, [6]float32{5, 0, 0, 5, 50, 50})
}

// Adversarial: NaN/Inf/giant attribute values must not poison the
// transform. computeNestedSvgViewport returns a finite matrix.
func TestNestedSvg_HardenNaNInfInputs(t *testing.T) {
	cases := []map[string]string{
		{"x": "NaN", "y": "Inf", "width": "50", "height": "50"},
		{"width": "-Inf", "height": "NaN", "viewBox": "0 0 10 10"},
		{"x": "1e30%", "y": "1e30%", "width": "100", "height": "100"},
		{"viewBox": "NaN NaN 10 10", "width": "100", "height": "100"},
		{"viewBox": "0 0 Inf Inf", "width": "100", "height": "100"},
		{"viewBox": "0 0 0 0", "width": "100", "height": "100"},
		{"viewBox": "0 0 -5 -5", "width": "100", "height": "100"},
	}
	parent := viewportRect{X: 0, Y: 0, W: 100, H: 100}
	for i, attrs := range cases {
		_, m := computeNestedSvgViewport(attrs, parent)
		for j, v := range m {
			f := float64(v)
			if math.IsNaN(f) || math.IsInf(f, 0) {
				t.Fatalf("case %d: m[%d] non-finite: %v (full=%v)",
					i, j, v, m)
			}
		}
	}
}

// Adversarial: poisoned parent viewport must be sanitized before use.
func TestNestedSvg_HardenPoisonedParent(t *testing.T) {
	parent := viewportRect{
		X: float32(math.NaN()),
		Y: float32(math.Inf(1)),
		W: float32(math.NaN()),
		H: float32(math.Inf(-1)),
	}
	_, m := computeNestedSvgViewport(map[string]string{
		"width": "50%", "height": "50%",
	}, parent)
	for i, v := range m {
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			t.Fatalf("m[%d] non-finite: %v (full=%v)", i, v, m)
		}
	}
}

// Defaults & percent: width="100%", height="50%" on inner svg should
// resolve against the parent viewport's user-space dimensions
// (100x100 here -> 100x50).
func TestNestedSvg_PercentWidthHeight(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<svg width="100%" height="50%" viewBox="0 0 10 10"
			preserveAspectRatio="none">
			<rect width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{10, 0, 0, 5, 0, 0})
}

// All 9 align modes (excluding "none") drive xFrac/yFrac via
// nestedAlignFractions. Inner viewBox is square (10x10) but outer
// viewport is 100x50; meet picks scale 5, leaves 50px x-residual.
// xFrac dictates where the residual lands.
func TestNestedSvg_AlignFractionsAllModes(t *testing.T) {
	cases := []struct {
		mode   string
		wantTx float32
		wantTy float32
	}{
		{"xMinYMin meet", 0, 0},
		{"xMidYMin meet", 25, 0},
		{"xMaxYMin meet", 50, 0},
		{"xMinYMid meet", 0, 0},
		{"xMidYMid meet", 25, 0},
		{"xMaxYMid meet", 50, 0},
		{"xMinYMax meet", 0, 0},
		{"xMidYMax meet", 25, 0},
		{"xMaxYMax meet", 50, 0},
	}
	for _, tc := range cases {
		svg := `<svg viewBox="0 0 200 100">
			<svg width="100" height="50" viewBox="0 0 10 10"
				preserveAspectRatio="` + tc.mode + `">
				<rect width="10" height="10"/>
			</svg>
		</svg>`
		vg, err := parseSvg(svg)
		if err != nil {
			t.Fatalf("%s: parse: %v", tc.mode, err)
		}
		p := firstShapePath(t, vg)
		want := [6]float32{5, 0, 0, 5, tc.wantTx, tc.wantTy}
		if !nearEq32(p.Transform[4], want[4]) ||
			!nearEq32(p.Transform[5], want[5]) {
			t.Fatalf("%s: tx,ty = (%v,%v); want (%v,%v); full=%v",
				tc.mode, p.Transform[4], p.Transform[5],
				want[4], want[5], p.Transform)
		}
	}
}

// Vertical residual coverage: y-axis residual lands per yFrac. Inner
// viewport 50x100, viewBox 10x10 -> meet picks min(5,10)=5; y residual
// = 100 - 5*10 = 50. yFrac × 50 yields the ty.
func TestNestedSvg_AlignFractionsYAxis(t *testing.T) {
	cases := []struct {
		mode   string
		wantTy float32
	}{
		{"xMinYMin meet", 0},
		{"xMinYMid meet", 25},
		{"xMinYMax meet", 50},
	}
	for _, tc := range cases {
		svg := `<svg viewBox="0 0 100 200">
			<svg width="50" height="100" viewBox="0 0 10 10"
				preserveAspectRatio="` + tc.mode + `">
				<rect width="10" height="10"/>
			</svg>
		</svg>`
		vg, err := parseSvg(svg)
		if err != nil {
			t.Fatalf("%s: parse: %v", tc.mode, err)
		}
		p := firstShapePath(t, vg)
		if !nearEq32(p.Transform[5], tc.wantTy) {
			t.Fatalf("%s: ty=%v want %v (full=%v)",
				tc.mode, p.Transform[5], tc.wantTy, p.Transform)
		}
	}
}

// display:none on a nested <svg> drops the entire subtree.
func TestNestedSvg_DisplayNoneSkipsSubtree(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<rect width="10" height="10"/>
		<svg display="none">
			<rect width="20" height="20"/>
			<circle r="5"/>
		</svg>
		<rect width="30" height="30"/>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var shapes int
	for _, p := range vg.Paths {
		if len(p.Segments) > 0 {
			shapes++
		}
	}
	if shapes != 2 {
		t.Fatalf("expected 2 shape paths (siblings of hidden svg); got %d",
			shapes)
	}
}

// transform= on nested <svg> composes BEFORE viewport mapping:
// final = parent · elemTransform · viewportTx. elemTx scale(2) and
// viewportTx scale(5) (10x10 viewBox into 50x50 rect) → diag scale 10,
// translate from rect origin (10,20).
func TestNestedSvg_TransformWithViewBoxCompose(t *testing.T) {
	svg := `<svg viewBox="0 0 200 200">
		<svg x="10" y="20" width="50" height="50" viewBox="0 0 10 10"
			transform="scale(2)" preserveAspectRatio="none">
			<rect width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	// scale(2) applied first; then viewportTx places the 10x10 viewBox
	// into the 50x50 rect at (10,20). Compose: outer point =
	// scale(2) · translate(10,20)·scale(5) · innerPoint
	//        = [10,0,0,10, 20, 40] · innerPoint.
	assertTransform(t, p.Transform, [6]float32{10, 0, 0, 10, 20, 40})
}

// Three levels of nested <svg>, each scaling 2x via viewBox, compose
// into a final scale of 8. Verifies state.curViewport tracks across
// recursion correctly.
func TestNestedSvg_ThreeLevelComposition(t *testing.T) {
	svg := `<svg viewBox="0 0 80 80">
		<svg width="80" height="80" viewBox="0 0 40 40">
			<svg width="40" height="40" viewBox="0 0 20 20">
				<svg width="20" height="20" viewBox="0 0 10 10">
					<rect width="10" height="10"/>
				</svg>
			</svg>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	assertTransform(t, p.Transform, [6]float32{8, 0, 0, 8, 0, 0})
}

// fill on outer <svg> must inherit through a nested <svg> wrapper to
// descendant shapes. Regression: nested-svg cascade should propagate
// presentation attrs the same as <g>.
func TestNestedSvg_FillInheritsThroughViewport(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100" fill="red">
		<svg width="50" height="50" viewBox="0 0 10 10">
			<rect width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	want := gui.SvgColor{R: 255, G: 0, B: 0, A: 255}
	if p.FillColor != want {
		t.Fatalf("FillColor=%+v; want %+v", p.FillColor, want)
	}
}

// Negative percent on width/height: SVG spec says length must be
// non-negative. resolveViewportLength yields a negative pre-clamp;
// clampViewBoxDim then floors to 0, giving a degenerate inner
// viewport. Inner viewBox path must be skipped (vbW > 0 guards it),
// falling through to translate-only (1,0,0,1,x,y) with a 0-dim rect.
// Document the actual contract via assertion.
func TestResolveViewportLength_NegativePercent(t *testing.T) {
	got := resolveViewportLength("-50%", 100, 0)
	want := float32(-50)
	if !nearEq32(got, want) {
		t.Fatalf("resolveViewportLength(-50%%,100,0)=%v want %v", got, want)
	}
	// End-to-end: negative width zeroes the viewport. Inner has a
	// viewBox; nums[2]>0 still passes (10) but w=0 → sx=0. Inner
	// content collapses to a point at (0,0).
	svg := `<svg viewBox="0 0 100 100">
		<svg width="-50%" height="50" viewBox="0 0 10 10"
			preserveAspectRatio="none">
			<rect width="10" height="10"/>
		</svg>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := firstShapePath(t, vg)
	if !nearEq32(p.Transform[0], 0) {
		t.Fatalf("expected sx=0 from clamped negative width; got %v",
			p.Transform[0])
	}
}
