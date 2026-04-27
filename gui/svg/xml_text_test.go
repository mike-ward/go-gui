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

// Mixed-content <text>: char data following a <tspan> close must
// survive as its own text run with parent attrs. Pre-fix, the
// trailing "!" was silently dropped because Leading is captured only
// while the parent has no children.
func TestTextMixedContentTrailingTextPreserved(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<text x="0" y="10">Hello <tspan>world</tspan> !</text>` +
		`</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"Hello": false, "world": false, "!": false}
	for _, txt := range vg.Texts {
		if _, ok := want[txt.Text]; ok {
			want[txt.Text] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("missing text run %q in mixed content", k)
		}
	}
}

// Char data between two <tspan> children belongs to the parent <text>
// flow. Earlier impl preserved Leading once and lost the rest.
func TestTextMixedContentInterleaved(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<text x="0" y="10">A <tspan>B</tspan> middle <tspan>C</tspan> end</text>` +
		`</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"A", "B", "middle", "C", "end"}
	got := make(map[string]bool, len(want))
	for _, txt := range vg.Texts {
		got[txt.Text] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing run %q; got runs %+v", w, got)
		}
	}
}

// Ancestor-supplied stroke must survive onto <text>. Earlier impl
// only read stroke from the <text> element itself.
func TestTextInheritsStrokeFromAncestor(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<g stroke="red" stroke-width="3"><text>Stroked</text></g>` +
		`</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	got := vg.Texts[0]
	if got.StrokeColor.R != 255 || got.StrokeColor.G != 0 || got.StrokeColor.B != 0 {
		t.Errorf("StrokeColor = %+v, want red", got.StrokeColor)
	}
	if got.StrokeWidth != 3 {
		t.Errorf("StrokeWidth = %f, want 3", got.StrokeWidth)
	}
}

// stroke="inherit" on <text> must resolve against the cascade, not
// silently force black.
func TestTextStrokeInheritResolvesAgainstCascade(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<g stroke="green"><text stroke="inherit">X</text></g>` +
		`</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("expected text")
	}
	c := vg.Texts[0].StrokeColor
	if c.G == 0 {
		t.Errorf("StrokeColor = %+v, want green from cascade", c)
	}
}

// <tspan> stroke / stroke-width / opacity overrides must apply.
// parseTspan previously copied parent values verbatim with no
// inspection of tspan's own attrs.
func TestTspanStrokeAndOpacityOverrides(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<text stroke="red" stroke-width="2" opacity="1">` +
		`<tspan stroke="blue" stroke-width="5" opacity="0.5">X</tspan>` +
		`</text></svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	var got gui.SvgText
	for _, txt := range vg.Texts {
		if txt.Text == "X" {
			got = txt
		}
	}
	if got.Text != "X" {
		t.Fatal("tspan run not found")
	}
	if got.StrokeColor.B != 255 || got.StrokeColor.R != 0 {
		t.Errorf("tspan StrokeColor = %+v, want blue", got.StrokeColor)
	}
	if got.StrokeWidth != 5 {
		t.Errorf("tspan StrokeWidth = %f, want 5", got.StrokeWidth)
	}
	if got.Opacity > 0.501 || got.Opacity < 0.499 {
		t.Errorf("tspan Opacity = %f, want ~0.5", got.Opacity)
	}
}

// `<text stroke-width="-5">` and "NaN" must clamp to 0 — negatives
// invalid per SVG spec, NaN poisons tessellation.
func TestTextStrokeWidthNaNNegativeClamped(t *testing.T) {
	cases := []string{"-5", "NaN"}
	for _, w := range cases {
		svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
			`<text stroke="red" stroke-width="` + w + `">X</text>` +
			`</svg>`
		vg, err := parseSvg(svg)
		if err != nil {
			t.Fatalf("%s: %v", w, err)
		}
		if len(vg.Texts) == 0 {
			t.Fatalf("%s: no text", w)
		}
		got := vg.Texts[0].StrokeWidth
		// stroke present + width sanitized to 0 → defaults to 1.
		if got != 1 {
			t.Errorf("stroke-width=%q produced StrokeWidth=%f, want 1", w, got)
		}
	}
}

// `<tspan opacity="50%">` must equal 0.5, mirroring the keyframe %
// fix. Earlier impl ran via parseOpacityAttr → parseFloatTrimmed,
// which turned 50 into clamp-to-1 (fully opaque).
func TestTspanOpacityPercentage(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<text opacity="1"><tspan opacity="50%">X</tspan></text>` +
		`</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	var got gui.SvgText
	for _, txt := range vg.Texts {
		if txt.Text == "X" {
			got = txt
		}
	}
	if got.Text != "X" {
		t.Fatal("tspan run not found")
	}
	if got.Opacity > 0.501 || got.Opacity < 0.499 {
		t.Errorf("tspan opacity 50%% = %f, want ~0.5", got.Opacity)
	}
}

// `<text>A <tspan/> B</text>` — self-close tspan still has a Tail
// that must emit "B".
func TestTextBodySelfCloseTspanTailPreserved(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<text x="0" y="10">A <tspan/> B</text></svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"A": false, "B": false}
	for _, txt := range vg.Texts {
		if _, ok := want[txt.Text]; ok {
			want[txt.Text] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("missing run %q after self-close tspan", k)
		}
	}
}

// `<g stroke="red"><text stroke="none">` — text must drop ancestor
// stroke entirely.
func TestTextStrokeNoneOverridesAncestor(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<g stroke="red" stroke-width="3"><text stroke="none">X</text></g>` +
		`</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("no text")
	}
	got := vg.Texts[0]
	if got.StrokeWidth != 0 {
		t.Errorf("StrokeWidth = %f, want 0 (none)", got.StrokeWidth)
	}
	if got.StrokeColor.A != 0 {
		t.Errorf("StrokeColor.A = %d, want 0 (none)", got.StrokeColor.A)
	}
}

// `<text stroke="inherit">` with no ancestor stroke → fall back to
// colorBlack and default width 1 (stroke present case).
func TestTextStrokeInheritNoCascadeFallsBackBlack(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<text stroke="inherit">X</text></svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	if len(vg.Texts) == 0 {
		t.Fatal("no text")
	}
	got := vg.Texts[0]
	if got.StrokeColor.R != 0 || got.StrokeColor.G != 0 || got.StrokeColor.B != 0 {
		t.Errorf("StrokeColor = %+v, want black", got.StrokeColor)
	}
	if got.StrokeWidth != 1 {
		t.Errorf("StrokeWidth = %f, want 1 default", got.StrokeWidth)
	}
}

// `<tspan stroke="none">` must clear stroke even when parent <text>
// supplies one.
func TestTspanStrokeNoneClearsStroke(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">` +
		`<text stroke="red" stroke-width="2">` +
		`<tspan stroke="none">X</tspan></text></svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	var got gui.SvgText
	for _, txt := range vg.Texts {
		if txt.Text == "X" {
			got = txt
		}
	}
	if got.Text != "X" {
		t.Fatal("tspan run not found")
	}
	if got.StrokeColor.A != 0 {
		t.Errorf("StrokeColor.A = %d, want 0 (none)", got.StrokeColor.A)
	}
}
