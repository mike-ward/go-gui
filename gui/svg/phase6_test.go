package svg

import (
	"testing"
)

// TestPhase6PulseRingBaseDecomposed — with path.Transform deferred to
// render time, parsing pulse-ring must yield a TessellatedPath that
// carries the decomposed placeholder base (scale=0, translate=12,12)
// while its triangle vertices remain in natural local coords
// (spanning ~20 for r=10 circle). Seeding the anim sandwich with the
// base is what lets replace/additive compose correctly.
func TestPhase6PulseRingBaseDecomposed(t *testing.T) {
	asset := `<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">` +
		`<circle cx="12" cy="12" r="10" ` +
		`transform="translate(12, 12) scale(0)">` +
		`<animateTransform attributeName="transform" type="translate" ` +
		`dur="1s" values="12 12;0 0"/>` +
		`<animateTransform attributeName="transform" type="scale" ` +
		`dur="1s" values="0;1" additive="sum"/>` +
		`</circle></svg>`
	p := New()
	parsed, err := p.ParseSvg(asset)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	tris := p.Tessellate(parsed, 1)
	if len(tris) == 0 {
		t.Fatalf("no tessellated paths")
	}
	tp := tris[0]
	if !tp.HasBaseXform {
		t.Fatalf("expected HasBaseXform=true on animated path")
	}
	if tp.BaseScaleX != 0 || tp.BaseScaleY != 0 {
		t.Fatalf("want BaseScale=(0,0); got (%v,%v)",
			tp.BaseScaleX, tp.BaseScaleY)
	}
	if tp.BaseTransX != 12 || tp.BaseTransY != 12 {
		t.Fatalf("want BaseTrans=(12,12); got (%v,%v)",
			tp.BaseTransX, tp.BaseTransY)
	}
	// Vertices are pre-transform: circle r=10 centered at (12,12)
	// spans roughly (2..22) on each axis.
	minX, maxX := float32(1e30), float32(-1e30)
	for i := 0; i+1 < len(tp.Triangles); i += 2 {
		x := tp.Triangles[i]
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
	}
	if maxX-minX < 10 {
		t.Fatalf("span=%v; vertices leaked baked transform",
			maxX-minX)
	}
}

// TestPhase6StaticTransformDecomposed — a non-animated shape with a
// non-identity transform must still tessellate in local coords and
// carry the decomposed base so the render fallback applies it.
func TestPhase6StaticTransformDecomposed(t *testing.T) {
	asset := `<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">` +
		`<circle cx="0" cy="0" r="5" ` +
		`transform="translate(10, 10) scale(2)"></circle>` +
		`</svg>`
	p := New()
	parsed, err := p.ParseSvg(asset)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	tris := p.Tessellate(parsed, 1)
	if len(tris) == 0 {
		t.Fatalf("no tessellated paths")
	}
	tp := tris[0]
	if !tp.HasBaseXform {
		t.Fatalf("expected HasBaseXform=true for static transform")
	}
	if tp.BaseTransX != 10 || tp.BaseTransY != 10 {
		t.Fatalf("want BaseTrans=(10,10); got (%v,%v)",
			tp.BaseTransX, tp.BaseTransY)
	}
	if !approxEq(tp.BaseScaleX, 2, 1e-5) ||
		!approxEq(tp.BaseScaleY, 2, 1e-5) {
		t.Fatalf("want BaseScale=(2,2); got (%v,%v)",
			tp.BaseScaleX, tp.BaseScaleY)
	}
	// Vertices at local coords: circle r=5 centered at origin →
	// span in x roughly (-5..5). With baked transform they would
	// have been at (0..20).
	minX, maxX := float32(1e30), float32(-1e30)
	for i := 0; i+1 < len(tp.Triangles); i += 2 {
		x := tp.Triangles[i]
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
	}
	if minX < -6 || maxX > 6 {
		t.Fatalf("vertices not in local coords: minX=%v maxX=%v",
			minX, maxX)
	}
}
