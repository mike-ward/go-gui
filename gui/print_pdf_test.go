package gui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mike-ward/go-glyph"
)

func testPrintJob(t *testing.T) PrintJob {
	t.Helper()
	j := NewPrintJob()
	j.OutputPath = filepath.Join(t.TempDir(), "test.pdf")
	return j
}

func TestRenderToPDF_Empty(t *testing.T) {
	j := testPrintJob(t)
	err := renderToPDF(nil, j, 800, 600)
	if err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_Rect(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderRect, X: 10, Y: 10, W: 100, H: 50,
		Color: RGBA(255, 0, 0, 255),
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_RoundedRect(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderRect, X: 10, Y: 10, W: 100, H: 50,
		Color: RGBA(0, 128, 0, 255), Radius: 8,
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_StrokeRect(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderStrokeRect, X: 10, Y: 10, W: 100, H: 50,
		Color: RGBA(0, 0, 255, 255), Thickness: 2,
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_Text(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderText, X: 50, Y: 50,
		Color: RGBA(0, 0, 0, 255),
		Text: "Hello PDF", FontName: "sans", FontSize: 14,
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_MonoFont(t *testing.T) {
	if got := pdfFontName("monospace"); got != "Courier" {
		t.Errorf("monospace -> %q, want Courier", got)
	}
	if got := pdfFontName("serif"); got != "Times" {
		t.Errorf("serif -> %q, want Times", got)
	}
	if got := pdfFontName("sans-serif"); got != "Helvetica" {
		t.Errorf("sans-serif -> %q, want Helvetica", got)
	}
	if got := pdfFontName("Consolas"); got != "Courier" {
		t.Errorf("Consolas -> %q, want Courier", got)
	}
}

func TestRenderToPDF_Line(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderLine, X: 0, Y: 0, OffsetX: 100, OffsetY: 100,
		Color: RGBA(0, 0, 0, 255), Thickness: 1,
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_Circle(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderCircle, X: 50, Y: 50, W: 40, H: 40,
		Color: RGBA(128, 0, 128, 255), Fill: true,
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_ImageMissing(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderImage, X: 0, Y: 0, W: 100, H: 100,
		Resource: "/nonexistent/image.png",
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_ClipBeginEnd(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{
		{Kind: RenderClip, X: 10, Y: 10, W: 200, H: 200},
		{Kind: RenderRect, X: 0, Y: 0, W: 400, H: 400,
			Color: RGBA(255, 0, 0, 255)},
		{Kind: RenderClip}, // end clip
	}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_SkippedKinds(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{
		{Kind: RenderShadow},
		{Kind: RenderBlur},
		{Kind: RenderFilterBegin},
		{Kind: RenderFilterEnd},
		{Kind: RenderCustomShader},
		{Kind: RenderNone},
	}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_Landscape(t *testing.T) {
	j := testPrintJob(t)
	j.Orientation = PrintLandscape
	cmds := []RenderCmd{{
		Kind: RenderRect, X: 10, Y: 10, W: 200, H: 100,
		Color: RGBA(0, 0, 255, 255),
	}}
	if err := renderToPDF(cmds, j, 1024, 768); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_ActualSize(t *testing.T) {
	j := testPrintJob(t)
	j.ScaleMode = PrintScaleActualSize
	cmds := []RenderCmd{{
		Kind: RenderRect, X: 0, Y: 0, W: 100, H: 100,
		Color: RGBA(128, 128, 128, 255),
	}}
	if err := renderToPDF(cmds, j, 400, 300); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_Gradient(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderGradient, X: 0, Y: 0, W: 200, H: 100,
		Gradient: &GradientDef{
			Direction: GradientToRight,
			Stops: []GradientStop{
				{Color: RGBA(255, 0, 0, 255), Pos: 0},
				{Color: RGBA(0, 0, 255, 255), Pos: 1},
			},
		},
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_GradientNilDef(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderGradient, X: 0, Y: 0, W: 200, H: 100,
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_SvgTriangles(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderSvg,
		Triangles: []float32{
			0, 0, 50, 0, 25, 50,
			50, 0, 100, 0, 75, 50,
		},
		Color: RGBA(0, 128, 0, 255),
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_SvgTooFewTriangles(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind:      RenderSvg,
		Triangles: []float32{0, 0, 50},
		Color:     RGBA(0, 128, 0, 255),
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_HeaderFooter(t *testing.T) {
	j := testPrintJob(t)
	j.Title = "Test Document"
	j.JobName = "test-job"
	j.Header = PrintHeaderFooterCfg{
		Enabled: true,
		Left:    "{title}",
		Center:  "Page {page}",
		Right:   "{date}",
	}
	j.Footer = PrintHeaderFooterCfg{
		Enabled: true,
		Center:  "{job}",
	}
	if err := renderToPDF(nil, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_AlphaTransparency(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderRect, X: 10, Y: 10, W: 100, H: 50,
		Color: RGBA(255, 0, 0, 128),
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_GradientBorder(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{
		Kind: RenderGradientBorder, X: 10, Y: 10, W: 100, H: 50,
		Thickness: 2, Radius: 4,
		Gradient: &GradientDef{
			Stops: []GradientStop{
				{Color: RGBA(255, 0, 0, 255)},
				{Color: RGBA(0, 0, 255, 255)},
			},
		},
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func TestRenderToPDF_RTF(t *testing.T) {
	j := testPrintJob(t)
	layout := &glyph.Layout{
		Text: "Hello RTF",
		Items: []glyph.Item{
			{
				StartIndex: 0,
				Length:      9,
				X:          0,
				Y:          20,
				Width:      80,
				Ascent:     12,
				Descent:    4,
				Color:      glyph.Color{R: 0, G: 0, B: 0, A: 255},
			},
		},
		Width:  200,
		Height: 40,
	}
	cmds := []RenderCmd{{
		Kind:      RenderRTF,
		X:         10,
		Y:         10,
		LayoutPtr: layout,
	}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
	// Text is zlib-compressed in the PDF stream, so verify
	// the PDF is larger than a minimal empty page (~910 bytes).
	info, _ := os.Stat(j.OutputPath)
	if info.Size() < 1000 {
		t.Errorf("PDF too small (%d bytes), likely missing text", info.Size())
	}
}

func TestRenderToPDF_RTF_NilLayout(t *testing.T) {
	j := testPrintJob(t)
	cmds := []RenderCmd{{Kind: RenderRTF, X: 10, Y: 10}}
	if err := renderToPDF(cmds, j, 800, 600); err != nil {
		t.Fatal(err)
	}
	assertPDFExists(t, j.OutputPath)
}

func assertPDFExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("PDF not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF file is empty")
	}
}
