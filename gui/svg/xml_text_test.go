package svg

import (
	"testing"

	"github.com/mike-ward/go-gui/gui"
)

func TestTextSimple(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text x="10" y="20">Hello</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected at least 1 text")
	}
	txt := vg.Texts[0]
	if txt.Text != "Hello" {
		t.Errorf("text = %q, want Hello", txt.Text)
	}
	if txt.X != 10 {
		t.Errorf("X = %f, want 10", txt.X)
	}
	if txt.Y != 20 {
		t.Errorf("Y = %f, want 20", txt.Y)
	}
}

func TestTextFontSize(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text font-size="24">Sized</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	if vg.Texts[0].FontSize != 24 {
		t.Errorf("FontSize = %f, want 24", vg.Texts[0].FontSize)
	}
}

func TestTextFontFamily(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text font-family="Helvetica, sans-serif">Family</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	if vg.Texts[0].FontFamily != "Helvetica" {
		t.Errorf("FontFamily = %q, want Helvetica",
			vg.Texts[0].FontFamily)
	}
}

func TestTextBold(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text font-weight="bold">Bold</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	if !vg.Texts[0].IsBold {
		t.Error("expected IsBold = true")
	}
	if vg.Texts[0].FontWeight != 700 {
		t.Errorf("FontWeight = %d, want 700", vg.Texts[0].FontWeight)
	}
}

func TestTextItalic(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text font-style="italic">Italic</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	if !vg.Texts[0].IsItalic {
		t.Error("expected IsItalic = true")
	}
}

func TestTextAnchorMiddle(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text text-anchor="middle">Centered</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	if vg.Texts[0].Anchor != 1 {
		t.Errorf("Anchor = %d, want 1", vg.Texts[0].Anchor)
	}
}

func TestTextAnchorEnd(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text text-anchor="end">End</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	if vg.Texts[0].Anchor != 2 {
		t.Errorf("Anchor = %d, want 2", vg.Texts[0].Anchor)
	}
}

func TestTextFillColor(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text fill="red">Red</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	// SVG "red" = #ff0000
	if vg.Texts[0].Color.R != 255 || vg.Texts[0].Color.G != 0 ||
		vg.Texts[0].Color.B != 0 {
		t.Errorf("Color = %+v, want red (255,0,0)",
			vg.Texts[0].Color)
	}
}

func TestTspanOverrides(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text font-size="16" fill="black">
			<tspan font-size="24" fill="blue">Big</tspan>
		</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text from tspan")
	}
	found := false
	for _, txt := range vg.Texts {
		if txt.Text == "Big" {
			found = true
			if txt.FontSize != 24 {
				t.Errorf("tspan FontSize = %f, want 24",
					txt.FontSize)
			}
			// blue = 0,0,255
			if txt.Color.B != 255 || txt.Color.R != 0 {
				t.Errorf("tspan Color = %+v, want blue",
					txt.Color)
			}
		}
	}
	if !found {
		t.Error("tspan text 'Big' not found")
	}
}

func TestTspanDy(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text x="0" y="10">
			<tspan dy="15">Shifted</tspan>
		</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, txt := range vg.Texts {
		if txt.Text == "Shifted" {
			found = true
			// Parent y=10, dy=15 → Y=25
			if txt.Y != 25 {
				t.Errorf("Y = %f, want 25", txt.Y)
			}
		}
	}
	if !found {
		t.Error("tspan text 'Shifted' not found")
	}
}

func TestTextPathHref(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<defs>
			<path id="curve" d="M10 80 Q95 10 180 80"/>
		</defs>
		<text>
			<textPath href="#curve">Curved</textPath>
		</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.TextPaths) == 0 {
		t.Fatal("expected textPath")
	}
	tp := vg.TextPaths[0]
	if tp.Text != "Curved" {
		t.Errorf("Text = %q, want Curved", tp.Text)
	}
	if tp.PathID != "curve" {
		t.Errorf("PathID = %q, want curve", tp.PathID)
	}
}

func TestTextPathXlinkHref(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"
		xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 200 200">
		<defs>
			<path id="arc" d="M10 80 Q95 10 180 80"/>
		</defs>
		<text>
			<textPath xlink:href="#arc">Legacy</textPath>
		</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.TextPaths) == 0 {
		t.Fatal("expected textPath")
	}
	if vg.TextPaths[0].PathID != "arc" {
		t.Errorf("PathID = %q, want arc", vg.TextPaths[0].PathID)
	}
}

func TestTextPathStartOffset(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<defs>
			<path id="p" d="M0 0 L100 0"/>
		</defs>
		<text>
			<textPath href="#p" startOffset="50%">Mid</textPath>
		</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.TextPaths) == 0 {
		t.Fatal("expected textPath")
	}
	tp := vg.TextPaths[0]
	if tp.StartOffset != 50 {
		t.Errorf("StartOffset = %f, want 50", tp.StartOffset)
	}
	if !tp.IsPercent {
		t.Error("expected IsPercent = true")
	}
}

func TestTextPathNoHrefSkipped(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text>
			<textPath>NoRef</textPath>
		</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.TextPaths) != 0 {
		t.Errorf("textPaths = %d, want 0 (no href)", len(vg.TextPaths))
	}
}

func TestTextPathAnchorOverride(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<defs>
			<path id="p2" d="M0 0 L100 0"/>
		</defs>
		<text text-anchor="start">
			<textPath href="#p2" text-anchor="middle">Centered</textPath>
		</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.TextPaths) == 0 {
		t.Fatal("expected textPath")
	}
	if vg.TextPaths[0].Anchor != 1 {
		t.Errorf("Anchor = %d, want 1 (middle)", vg.TextPaths[0].Anchor)
	}
}

func TestTextDefaultColor(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
		<text>NoFill</text>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	// Default fill = black
	c := vg.Texts[0].Color
	if c != (gui.SvgColor{R: 0, G: 0, B: 0, A: 255}) {
		t.Errorf("default color = %+v, want black", c)
	}
}
