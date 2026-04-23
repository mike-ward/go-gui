package svg

import "testing"

func TestDefsClipPathRect(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<clipPath id="c1">
				<rect x="0" y="0" width="50" height="50"/>
			</clipPath>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	paths, ok := vg.ClipPaths["c1"]
	if !ok {
		t.Fatal("clipPath c1 not found")
	}
	if len(paths) == 0 {
		t.Error("clipPath c1 has no paths")
	}
}

func TestDefsClipPathNoID(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<clipPath>
				<rect x="0" y="0" width="50" height="50"/>
			</clipPath>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.ClipPaths) != 0 {
		t.Errorf("expected no clip paths, got %d", len(vg.ClipPaths))
	}
}

func TestDefsLinearGradientWithStops(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<linearGradient id="g1" x1="0%" y1="0%" x2="100%" y2="0%">
				<stop offset="0%" stop-color="red"/>
				<stop offset="100%" stop-color="blue"/>
			</linearGradient>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	g, ok := vg.Gradients["g1"]
	if !ok {
		t.Fatal("gradient g1 not found")
	}
	if len(g.Stops) != 2 {
		t.Fatalf("stops = %d, want 2", len(g.Stops))
	}
	if g.Stops[0].Offset != 0 {
		t.Errorf("stop[0].Offset = %f, want 0", g.Stops[0].Offset)
	}
	if g.Stops[1].Offset != 1 {
		t.Errorf("stop[1].Offset = %f, want 1", g.Stops[1].Offset)
	}
}

func TestDefsGradientUserSpace(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<linearGradient id="g2" gradientUnits="userSpaceOnUse"
				x1="10" y1="20" x2="90" y2="80">
				<stop offset="0" stop-color="#000"/>
			</linearGradient>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	g, ok := vg.Gradients["g2"]
	if !ok {
		t.Fatal("gradient g2 not found")
	}
	if g.GradientUnits != "userSpaceOnUse" {
		t.Errorf("gradientUnits = %q, want userSpaceOnUse",
			g.GradientUnits)
	}
	if g.X1 != 10 {
		t.Errorf("X1 = %f, want 10", g.X1)
	}
	if g.Y2 != 80 {
		t.Errorf("Y2 = %f, want 80", g.Y2)
	}
}

func TestDefsGradientSelfClosing(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<linearGradient id="g3" x1="0" y1="0" x2="1" y2="1"/>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	g, ok := vg.Gradients["g3"]
	if !ok {
		t.Fatal("gradient g3 not found")
	}
	if len(g.Stops) != 0 {
		t.Errorf("stops = %d, want 0 for self-closing", len(g.Stops))
	}
}

func TestDefsPath(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<path id="p1" d="M0 0 L10 10"/>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	d, ok := vg.DefsPaths["p1"]
	if !ok {
		t.Fatal("defs path p1 not found")
	}
	if d != "M0 0 L10 10" {
		t.Errorf("d = %q, want M0 0 L10 10", d)
	}
}

func TestDefsFilterGaussianBlur(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<filter id="f1">
				<feGaussianBlur stdDeviation="3"/>
			</filter>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	f, ok := vg.Filters["f1"]
	if !ok {
		t.Fatal("filter f1 not found")
	}
	if f.StdDev != 3 {
		t.Errorf("StdDev = %f, want 3", f.StdDev)
	}
}

func TestDefsFilterMergeNodes(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<filter id="f2">
				<feGaussianBlur stdDeviation="5"/>
				<feMerge>
					<feMergeNode in="blur1"/>
					<feMergeNode in="blur2"/>
					<feMergeNode in="SourceGraphic"/>
				</feMerge>
			</filter>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	f, ok := vg.Filters["f2"]
	if !ok {
		t.Fatal("filter f2 not found")
	}
	if f.BlurLayers != 2 {
		t.Errorf("BlurLayers = %d, want 2", f.BlurLayers)
	}
	if !f.KeepSource {
		t.Error("KeepSource should be true")
	}
}

func TestDefsFilterNoBlurSkipped(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<filter id="f3">
				<feColorMatrix type="saturate" values="0"/>
			</filter>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := vg.Filters["f3"]; ok {
		t.Error("filter without blur should be skipped")
	}
}

func TestDefsGradientStopOpacity(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs>
			<linearGradient id="g4">
				<stop offset="0" stop-color="#ff0000" stop-opacity="0.5"/>
			</linearGradient>
		</defs>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	g, ok := vg.Gradients["g4"]
	if !ok {
		t.Fatal("gradient g4 not found")
	}
	if len(g.Stops) != 1 {
		t.Fatalf("stops = %d, want 1", len(g.Stops))
	}
	if g.Stops[0].Color.A == 255 {
		t.Error("stop alpha should be reduced by stop-opacity")
	}
}

// A stop declared as stop-color="currentColor" must have its
// sentinel marker alpha lifted to opaque before stop-opacity is
// baked in; otherwise a small stop-opacity multiplied by A=2
// collapses to 0 and the gradient sample disappears.
func TestParseGradientStops_CurrentColorStopsPromoteAlpha(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">` +
		`<defs>` +
		`<linearGradient id="gc">` +
		`<stop offset="0" stop-color="currentColor" stop-opacity="0.5"/>` +
		`<stop offset="1" stop-color="currentColor" stop-opacity="1"/>` +
		`</linearGradient>` +
		`</defs></svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	g, ok := vg.Gradients["gc"]
	if !ok {
		t.Fatal("gradient gc not found")
	}
	if len(g.Stops) != 2 {
		t.Fatalf("stops = %d, want 2", len(g.Stops))
	}
	// Sentinel RGB (magenta) must survive so render-time tint can
	// substitute. A must be ~127 (255*0.5) for the first stop and
	// 255 for the second — if sentinel bump failed, A would be 1.
	s := g.Stops[0].Color
	if s.R != 255 || s.B != 255 {
		t.Fatalf("stop 0: expected sentinel magenta RGB, got %+v", s)
	}
	if s.A < 100 || s.A > 160 {
		t.Fatalf("stop 0: expected A~127 (sentinel promoted then "+
			"scaled by 0.5), got %d", s.A)
	}
	if g.Stops[1].Color.A != 255 {
		t.Fatalf("stop 1: expected A=255 (opacity=1), got %d",
			g.Stops[1].Color.A)
	}
}
