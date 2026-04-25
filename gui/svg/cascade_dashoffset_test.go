package svg

import "testing"

func TestCascade_StrokeDashOffsetInheritsFromGroupAttr(t *testing.T) {
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
		<g stroke="black" stroke-dasharray="4" stroke-dashoffset="7">
			<rect width="10" height="10"/>
		</g>
	</svg>`
	vg := parseSvgT(t, src)
	if len(vg.Paths) != 1 {
		t.Fatalf("paths: %d", len(vg.Paths))
	}
	if vg.Paths[0].StrokeDashOffset != 7 {
		t.Errorf("dash-offset: got %v want 7", vg.Paths[0].StrokeDashOffset)
	}
}

func TestCascade_StrokeDashOffsetInheritsFromCSS(t *testing.T) {
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
		<style>g { stroke: black; stroke-dasharray: 4; stroke-dashoffset: 9 }</style>
		<g><rect width="10" height="10"/></g>
	</svg>`
	vg := parseSvgT(t, src)
	if len(vg.Paths) != 1 {
		t.Fatalf("paths: %d", len(vg.Paths))
	}
	if vg.Paths[0].StrokeDashOffset != 9 {
		t.Errorf("dash-offset: got %v want 9", vg.Paths[0].StrokeDashOffset)
	}
}

func TestCascade_StrokeDashOffsetChildOverrides(t *testing.T) {
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
		<g stroke="black" stroke-dasharray="4" stroke-dashoffset="7">
			<rect width="10" height="10" stroke-dashoffset="3"/>
		</g>
	</svg>`
	vg := parseSvgT(t, src)
	if len(vg.Paths) != 1 {
		t.Fatalf("paths: %d", len(vg.Paths))
	}
	if vg.Paths[0].StrokeDashOffset != 3 {
		t.Errorf("dash-offset: got %v want 3", vg.Paths[0].StrokeDashOffset)
	}
}
