package svg

import "testing"

func TestParseSvg_LowercaseViewbox(t *testing.T) {
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewbox="0 0 24 48">
		<rect width="24" height="48"/>
	</svg>`
	vg := parseSvgT(t, src)
	if vg.Width != 24 || vg.Height != 48 {
		t.Errorf("dims: %v x %v want 24 x 48", vg.Width, vg.Height)
	}
}

func TestParseSvgDimensions_LowercaseViewbox(t *testing.T) {
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewbox="0 0 32 16"/>`
	w, h := parseSvgDimensions(src)
	if w != 32 || h != 16 {
		t.Errorf("dims: %v x %v want 32 x 16", w, h)
	}
}

func TestParseSvgDimensions_CamelCaseStillWorks(t *testing.T) {
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8 4"/>`
	w, h := parseSvgDimensions(src)
	if w != 8 || h != 4 {
		t.Errorf("dims: %v x %v want 8 x 4", w, h)
	}
}
