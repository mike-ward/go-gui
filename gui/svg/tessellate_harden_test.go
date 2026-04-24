package svg

import (
	"math"
	"testing"

	"github.com/mike-ward/go-gui/gui"
)

// finiteViewBox rejects NaN and ±Inf on either axis; zero and
// ordinary finite values pass.
func TestFiniteViewBox(t *testing.T) {
	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	cases := []struct {
		x, y float32
		ok   bool
	}{
		{0, 0, true},
		{10, -20, true},
		{nan, 0, false},
		{0, nan, false},
		{inf, 0, false},
		{0, -inf, false},
	}
	for _, c := range cases {
		if got := finiteViewBox(c.x, c.y); got != c.ok {
			t.Errorf("finiteViewBox(%v,%v)=%v want %v",
				c.x, c.y, got, c.ok)
		}
	}
}

// seedFromTransform defers TRS-decomposable non-identity transforms
// onto Base* seed fields; identity returns zero seed with bake=false;
// shear forces bake.
func TestSeedFromTransform_IdentityDefersShearBakes(t *testing.T) {
	// Identity: no base, no bake.
	id := &VectorPath{Transform: identityTransform}
	seed, bake := seedFromTransform(id, nil)
	if seed.HasBaseXform || bake {
		t.Fatalf("identity: want no base/no bake, got %+v bake=%v",
			seed, bake)
	}

	// Pure translate (10,20) — TRS-decomposable.
	tr := &VectorPath{Transform: [6]float32{1, 0, 0, 1, 10, 20}}
	seed, bake = seedFromTransform(tr, nil)
	if bake || !seed.HasBaseXform {
		t.Fatalf("translate: want deferred base, got bake=%v seed=%+v",
			bake, seed)
	}
	if seed.BaseTransX != 10 || seed.BaseTransY != 20 {
		t.Fatalf("translate base wrong: %+v", seed)
	}

	// Shear matrix is not TRS-decomposable — forces bake.
	sh := &VectorPath{Transform: [6]float32{1, 0.5, 0, 1, 0, 0}}
	seed, bake = seedFromTransform(sh, nil)
	if !bake || seed.HasBaseXform {
		t.Fatalf("shear: want bake, got bake=%v seed=%+v", bake, seed)
	}
}

// When the force-bake set contains the path's GroupID, a TRS-decomposable
// transform is still baked so sibling paths with divergent bases all
// render at their authored positions.
func TestSeedFromTransform_ForceBakeOverridesDecompose(t *testing.T) {
	p := &VectorPath{
		GroupID:   "g1",
		Transform: [6]float32{1, 0, 0, 1, 10, 20},
	}
	seed, bake := seedFromTransform(p, map[string]bool{"g1": true})
	if !bake || seed.HasBaseXform {
		t.Fatalf("forceBake: want bake and no base, got bake=%v seed=%+v",
			bake, seed)
	}
}

// buildForceBakeSet returns nil unless a transform-kind animation
// targets a group AND >=2 member paths carry non-identity Transforms.
func TestBuildForceBakeSet_MultiPathTransformAnimGroup(t *testing.T) {
	anims := []gui.SvgAnimation{
		{Kind: gui.SvgAnimRotate, GroupID: "g1"},
	}
	paths := []VectorPath{
		{GroupID: "g1", Transform: [6]float32{1, 0, 0, 1, 5, 0}},
		{GroupID: "g1", Transform: [6]float32{1, 0, 0, 1, 0, 7}},
	}
	got := buildForceBakeSet(paths, anims)
	if !got["g1"] {
		t.Fatalf("expected g1 in force-bake set: %v", got)
	}
}

// Only one path with a non-identity transform in the animated group:
// per-GroupID animation state can still seed it without conflict.
func TestBuildForceBakeSet_SinglePathGroupDeferred(t *testing.T) {
	anims := []gui.SvgAnimation{
		{Kind: gui.SvgAnimTranslate, GroupID: "g1"},
	}
	paths := []VectorPath{
		{GroupID: "g1", Transform: [6]float32{1, 0, 0, 1, 5, 0}},
		{GroupID: "g1", Transform: identityTransform},
		// Sibling in an unrelated group with a transform — doesn't
		// count toward g1.
		{GroupID: "g2", Transform: [6]float32{1, 0, 0, 1, 9, 9}},
	}
	got := buildForceBakeSet(paths, anims)
	if got["g1"] {
		t.Fatalf("expected g1 NOT in force-bake set: %v", got)
	}
}

// No transform-kind animations → nil set. Opacity / attr animations
// don't need force-bake since they do not touch Base*.
func TestBuildForceBakeSet_OpacityAnimNoBake(t *testing.T) {
	anims := []gui.SvgAnimation{
		{Kind: gui.SvgAnimOpacity, GroupID: "g1"},
	}
	paths := []VectorPath{
		{GroupID: "g1", Transform: [6]float32{1, 0, 0, 1, 1, 0}},
		{GroupID: "g1", Transform: [6]float32{1, 0, 0, 1, 2, 0}},
	}
	if got := buildForceBakeSet(paths, anims); got != nil {
		t.Fatalf("opacity-only: expected nil, got %v", got)
	}
}

func TestBuildForceBakeSet_EmptyAnimsReturnsNil(t *testing.T) {
	if got := buildForceBakeSet(nil, nil); got != nil {
		t.Fatalf("want nil, got %v", got)
	}
}

// Tessellation with non-zero viewBox origin must shift every vertex
// into content-from-origin coords. A <rect x=10 y=20 w=5 h=5> in a
// viewBox starting at (10,20) must land near (0,0) after shift.
func TestTessellatePaths_ViewBoxOriginShiftsVertices(t *testing.T) {
	p := New()
	parsed, err := p.ParseSvg(`<svg viewBox="10 20 32 32"
		xmlns="http://www.w3.org/2000/svg">
		<rect x="10" y="20" width="5" height="5" fill="black"/>
	</svg>`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	tris := p.Tessellate(parsed, 1)
	if len(tris) == 0 || len(tris[0].Triangles) == 0 {
		t.Fatal("no triangles")
	}
	var minX, minY, maxX, maxY float32 = 1e30, 1e30, -1e30, -1e30
	for i := 0; i+1 < len(tris[0].Triangles); i += 2 {
		x, y := tris[0].Triangles[i], tris[0].Triangles[i+1]
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}
	// After -10,-20 shift: (0,0)-(5,5).
	if minX < -0.1 || minX > 0.1 || minY < -0.1 || minY > 0.1 {
		t.Fatalf("min post-shift wrong: (%.3f,%.3f)", minX, minY)
	}
	if maxX < 4.9 || maxX > 5.1 || maxY < 4.9 || maxY > 5.1 {
		t.Fatalf("max post-shift wrong: (%.3f,%.3f)", maxX, maxY)
	}
}

// A hostile viewBox with NaN origin must skip the shift so vertices
// retain their authored coords instead of becoming NaN.
func TestTessellatePaths_NaNViewBoxLeavesVerticesFinite(t *testing.T) {
	vg := &VectorGraphic{
		Width: 32, Height: 32,
		ViewBoxX: float32(math.NaN()), ViewBoxY: 0,
		Paths: []VectorPath{{
			Transform: identityTransform,
			FillColor: gui.SvgColor{A: 255},
			Primitive: gui.SvgPrimitive{
				Kind: gui.SvgPrimRect, X: 0, Y: 0, W: 5, H: 5,
			},
			Segments: segmentsForRect(0, 0, 5, 5, 0, 0),
		}},
	}
	tris := vg.tessellatePaths(vg.Paths, 1)
	if len(tris) == 0 || len(tris[0].Triangles) == 0 {
		t.Fatal("no triangles")
	}
	for i, v := range tris[0].Triangles {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			t.Fatalf("non-finite vertex at %d: %v", i, v)
		}
	}
}
