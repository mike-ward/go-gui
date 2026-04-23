package svg

import (
	"os"
	"testing"

	"github.com/mike-ward/go-gui/gui"
)

// --- Small pure functions ---

func TestXmlParseFontWeightBold(t *testing.T) {
	if parseFontWeight("bold") != 700 {
		t.Fatalf("'bold' should be 700")
	}
}

func TestXmlParseFontWeightNormal(t *testing.T) {
	if parseFontWeight("normal") != 400 {
		t.Fatalf("'normal' should be 400")
	}
}

func TestXmlParseFontWeightNumeric(t *testing.T) {
	if parseFontWeight("600") != 600 {
		t.Fatalf("'600' should be 600")
	}
}

func TestXmlParseFontWeightEmpty(t *testing.T) {
	if parseFontWeight("") != 0 {
		t.Fatalf("empty should be 0")
	}
}

func TestXmlParseFontWeightBolder(t *testing.T) {
	if parseFontWeight("bolder") != 700 {
		t.Fatalf("'bolder' should be 700")
	}
}

func TestXmlParseFontWeightLighter(t *testing.T) {
	if parseFontWeight("lighter") != 400 {
		t.Fatalf("'lighter' should be 400")
	}
}

func TestXmlParseFontWeightOutOfRange(t *testing.T) {
	if parseFontWeight("50") != 0 {
		t.Fatalf("'50' out of range should be 0")
	}
}

func TestXmlCleanFontFamilyMultiple(t *testing.T) {
	r := cleanFontFamily("Courier New, monospace")
	if r != "Courier New" {
		t.Fatalf("expected 'Courier New', got %q", r)
	}
}

func TestXmlCleanFontFamilySingle(t *testing.T) {
	r := cleanFontFamily("Arial")
	if r != "Arial" {
		t.Fatalf("expected 'Arial', got %q", r)
	}
}

func TestXmlExtractPlainText(t *testing.T) {
	r := extractPlainText("Hello World")
	if r != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", r)
	}
}

func TestXmlExtractPlainTextBeforeTag(t *testing.T) {
	r := extractPlainText("Hello <tspan>World</tspan>")
	if r != "Hello" {
		t.Fatalf("expected 'Hello', got %q", r)
	}
}

func TestXmlExtractPlainTextEntityDecode(t *testing.T) {
	r := extractPlainText("A &amp; B")
	if r != "A & B" {
		t.Fatalf("expected 'A & B', got %q", r)
	}
}

// --- Tag helpers ---

func TestXmlFindTagNameEnd(t *testing.T) {
	s := "rect width=\"10\">"
	end := findTagNameEnd(s, 0)
	if s[:end] != "rect" {
		t.Fatalf("expected 'rect', got %q", s[:end])
	}
}

func TestXmlFindTagNameEndSlash(t *testing.T) {
	s := "circle/>"
	end := findTagNameEnd(s, 0)
	if s[:end] != "circle" {
		t.Fatalf("expected 'circle', got %q", s[:end])
	}
}

func TestXmlFindClosingTagSimple(t *testing.T) {
	content := "<text>hello</text> rest"
	pos := findClosingTag(content, "text", 6) // after opening tag
	if content[pos:pos+7] != "</text>" {
		t.Fatalf("expected closing tag at pos=%d, got %q", pos, content[pos:])
	}
}

func TestXmlFindClosingTagNested(t *testing.T) {
	content := "<g><g>inner</g></g> rest"
	pos := findClosingTag(content, "g", 3) // after first <g>
	// Should find the outer </g>, not the inner one
	if pos != 15 {
		t.Fatalf("expected outer </g> at 15, got %d", pos)
	}
}

func TestXmlFindClosingTagBoundary(t *testing.T) {
	// "</text" must not match "</textPath"
	content := "<text><textPath>path</textPath>text</text>"
	pos := findClosingTag(content, "text", 6)
	expected := 35 // position of </text>
	if pos != expected {
		t.Fatalf("expected </text> at %d, got %d, substring=%q", expected, pos, content[pos:])
	}
}

// --- Gradient parsing ---

func TestXmlParseGradientCoordOBBPercent(t *testing.T) {
	v := parseGradientCoord("50%", true)
	if f32Abs(v-0.5) > 1e-5 {
		t.Fatalf("expected 0.5, got %f", v)
	}
}

func TestXmlParseGradientCoordOBBBare(t *testing.T) {
	v := parseGradientCoord("0.75", true)
	if f32Abs(v-0.75) > 1e-5 {
		t.Fatalf("expected 0.75, got %f", v)
	}
}

func TestXmlParseGradientCoordUserSpace(t *testing.T) {
	v := parseGradientCoord("100", false)
	if f32Abs(v-100) > 1e-5 {
		t.Fatalf("expected 100, got %f", v)
	}
}

func TestXmlParseGradientStops(t *testing.T) {
	content := `<stop offset="0" stop-color="red"/><stop offset="1" stop-color="blue"/>`
	stops := parseGradientStops(content)
	if len(stops) != 2 {
		t.Fatalf("expected 2 stops, got %d", len(stops))
	}
	if stops[0].Offset != 0 || stops[1].Offset != 1 {
		t.Fatalf("expected offsets 0 and 1, got %f and %f",
			stops[0].Offset, stops[1].Offset)
	}
	// Red
	if stops[0].Color.R != 255 || stops[0].Color.G != 0 {
		t.Fatalf("first stop should be red, got %+v", stops[0].Color)
	}
}

// --- parseSvgDimensions ---

func TestXmlParseSvgDimensionsViewBox(t *testing.T) {
	svg := `<svg viewBox="0 0 100 50">`
	w, h := parseSvgDimensions(svg)
	if w != 100 || h != 50 {
		t.Fatalf("expected (100,50), got (%f,%f)", w, h)
	}
}

func TestXmlParseSvgDimensionsWidthHeight(t *testing.T) {
	svg := `<svg width="200" height="100">`
	w, h := parseSvgDimensions(svg)
	if w != 200 || h != 100 {
		t.Fatalf("expected (200,100), got (%f,%f)", w, h)
	}
}

// --- Integration ---

func TestXmlParseSvgMinimalRect(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100"><rect x="0" y="0" width="50" height="50" fill="red"/></svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vg.Paths) < 1 {
		t.Fatalf("expected at least 1 path, got %d", len(vg.Paths))
	}
}

func TestXmlParseSvgWithGradient(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<defs>
			<linearGradient id="g1" x1="0" y1="0" x2="1" y2="0">
				<stop offset="0" stop-color="red"/>
				<stop offset="1" stop-color="blue"/>
			</linearGradient>
		</defs>
		<rect x="0" y="0" width="100" height="100" fill="url(#g1)"/>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vg.Gradients) == 0 {
		t.Fatalf("expected gradients, got none")
	}
	if _, ok := vg.Gradients["g1"]; !ok {
		t.Fatalf("expected gradient 'g1' in map")
	}
}

func TestXmlParseSvgViewBoxOffset(t *testing.T) {
	svg := `<svg viewBox="10 20 100 100"><rect x="10" y="20" width="50" height="50" fill="red"/></svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vg.ViewBoxX != 10 || vg.ViewBoxY != 20 {
		t.Fatalf("expected viewBox offset (10,20), got (%f,%f)",
			vg.ViewBoxX, vg.ViewBoxY)
	}
}

func TestXmlParseSvgEmpty(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100"></svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vg.Paths) != 0 {
		t.Fatalf("expected 0 paths, got %d", len(vg.Paths))
	}
}

func TestXmlParseSvgFileTooLarge(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/huge.svg"
	// Create file just over the 4MB limit.
	data := make([]byte, maxSvgFileSize+1)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err := parseSvgFile(path)
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
}

// suppress unused import
var _ = gui.SvgColor{}

// --- findRootElementTag ---

func TestFindRootElementTag_LocatesRootSvg(t *testing.T) {
	body := `<svg viewBox="0 0 10 10"><rect/></svg>`
	tag, ok := findRootElementTag(body, "svg")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if tag != `<svg viewBox="0 0 10 10">` {
		t.Fatalf("unexpected tag: %q", tag)
	}
}

func TestFindRootElementTag_RejectsPrefixMatches(t *testing.T) {
	// `<svgfoo` must not match tag "svg": the character after the
	// tag name must be whitespace, '>', or '/'.
	_, ok := findRootElementTag(`<svgfoo width="1"/>`, "svg")
	if ok {
		t.Fatal("prefix-only match must be rejected")
	}
}

func TestFindRootElementTag_MissingTag(t *testing.T) {
	_, ok := findRootElementTag(`<foo/>`, "svg")
	if ok {
		t.Fatal("missing tag must return ok=false")
	}
}

func TestFindRootElementTag_UnterminatedTag(t *testing.T) {
	// `<svg` without a closing `>` is malformed — must not panic
	// or return a stale buffer.
	_, ok := findRootElementTag(`<svg viewBox="0 0 1 1"`, "svg")
	if ok {
		t.Fatal("unterminated tag must return ok=false")
	}
}

// --- scanOpacityAnimTargets ---

func TestScanOpacityAnimTargets_DetectsEachTarget(t *testing.T) {
	body := `<animate attributeName="opacity" values="1;0" dur="1s"/>` +
		`<animate attributeName="fill-opacity" values="1;0" dur="1s"/>` +
		`<animate attributeName="stroke-opacity" values="1;0" dur="1s"/>`
	all, fill, stroke := scanOpacityAnimTargets(body)
	if !all {
		t.Fatal("expected opacity → all=true")
	}
	if !fill {
		t.Fatal("expected fill-opacity → fill=true")
	}
	if !stroke {
		t.Fatal("expected stroke-opacity → stroke=true")
	}
}

func TestScanOpacityAnimTargets_IgnoresAnimateTransform(t *testing.T) {
	// <animateTransform> never targets opacity; must be skipped so
	// it cannot trip the bake-suppression heuristics.
	body := `<animateTransform type="rotate" attributeName="transform" ` +
		`values="0 5 5;360 5 5" dur="1s"/>`
	all, fill, stroke := scanOpacityAnimTargets(body)
	if all || fill || stroke {
		t.Fatalf("animateTransform must not set any flag; "+
			"got all=%v fill=%v stroke=%v", all, fill, stroke)
	}
}

func TestScanOpacityAnimTargets_EmptyBody(t *testing.T) {
	all, fill, stroke := scanOpacityAnimTargets("")
	if all || fill || stroke {
		t.Fatal("empty body must report no targets")
	}
}

func TestScanOpacityAnimTargets_MalformedTagTerminatesSafely(t *testing.T) {
	// `<animate` without `>` should not loop forever and must not
	// flag any target. Exits cleanly when IndexByte('>') fails.
	body := `<animate attributeName="opacity"` // no close
	all, fill, stroke := scanOpacityAnimTargets(body)
	if all || fill || stroke {
		t.Fatal("malformed <animate> must not flag targets")
	}
}

// --- Root-level presentation attribute inheritance ---

// `fill="currentColor"` on the root <svg> element must propagate
// to child shapes so render-time tint can replace the placeholder.
func TestParseSvg_RootFillInheritedByShapes(t *testing.T) {
	svg := `<svg viewBox="0 0 10 10" fill="currentColor" ` +
		`xmlns="http://www.w3.org/2000/svg">` +
		`<rect x="0" y="0" width="10" height="10"/>` +
		`</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(vg.Paths) == 0 {
		t.Fatal("expected a path from the rect")
	}
	// The rect must inherit the currentColor sentinel (magenta RGB)
	// from the root; bakePathOpacity then promotes A to 255.
	// Sentinel RGB survives so render-side tint can substitute.
	if vg.Paths[0].FillColor.R != 255 || vg.Paths[0].FillColor.B != 255 {
		t.Fatalf("expected sentinel magenta RGB on fill, got %+v",
			vg.Paths[0].FillColor)
	}
	if vg.Paths[0].FillColor.A == 0 {
		t.Fatal("fill alpha collapsed to 0 — sentinel bump failed")
	}
}
