package gui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCachedSvgPathsEmpty(t *testing.T) {
	result := cachedSvgPaths(nil)
	if len(result) != 0 {
		t.Fatal("expected empty result")
	}
}

func TestCachedSvgPathsConversion(t *testing.T) {
	paths := []TessellatedPath{
		{
			Triangles:  []float32{0, 0, 1, 0, 0.5, 1},
			Color:      SvgColor{255, 0, 0, 255},
			IsClipMask: true,
			ClipGroup:  2,
			GroupID:    "g1",
		},
	}
	result := cachedSvgPaths(paths)
	if len(result) != 1 {
		t.Fatalf("expected 1 path, got %d", len(result))
	}
	p := result[0]
	if p.Color != (Color{255, 0, 0, 255}) {
		t.Fatalf("expected red, got %+v", p.Color)
	}
	if !p.IsClipMask {
		t.Fatal("expected clip mask")
	}
	if p.ClipGroup != 2 {
		t.Fatal("expected clip group 2")
	}
	if p.GroupID != "g1" {
		t.Fatal("expected group g1")
	}
}

func TestCachedSvgPathsVertexColors(t *testing.T) {
	paths := []TessellatedPath{
		{
			Triangles: []float32{0, 0, 1, 0, 0.5, 1},
			Color:     SvgColor{0, 0, 0, 255},
			VertexColors: []SvgColor{
				{255, 0, 0, 255},
				{0, 255, 0, 255},
				{0, 0, 255, 255},
			},
		},
	}
	result := cachedSvgPaths(paths)
	if len(result[0].VertexColors) != 3 {
		t.Fatalf("expected 3 vertex colors, got %d",
			len(result[0].VertexColors))
	}
	if result[0].VertexColors[0] != (Color{255, 0, 0, 255}) {
		t.Fatalf("expected red vertex, got %+v",
			result[0].VertexColors[0])
	}
}

func TestComputeTriangleBBoxEmpty(t *testing.T) {
	bbox := computeTriangleBBox(nil)
	if bbox != [4]float32{0, 0, 0, 0} {
		t.Fatalf("expected zero bbox, got %v", bbox)
	}
}

func TestComputeTriangleBBox(t *testing.T) {
	paths := []TessellatedPath{
		{Triangles: []float32{
			10, 20, 30, 40, 50, 60,
		}},
	}
	bbox := computeTriangleBBox(paths)
	if bbox[0] != 10 || bbox[1] != 20 {
		t.Fatalf("expected min (10,20), got (%f,%f)",
			bbox[0], bbox[1])
	}
	if bbox[2] != 40 || bbox[3] != 40 {
		t.Fatalf("expected size (40,40), got (%f,%f)",
			bbox[2], bbox[3])
	}
}

func TestComputeTriangleBBoxMultiplePaths(t *testing.T) {
	paths := []TessellatedPath{
		{Triangles: []float32{0, 0, 10, 10, 5, 5}},
		{Triangles: []float32{-5, -5, 20, 20, 15, 15}},
	}
	bbox := computeTriangleBBox(paths)
	if bbox[0] != -5 || bbox[1] != -5 {
		t.Fatalf("expected min (-5,-5), got (%f,%f)",
			bbox[0], bbox[1])
	}
	if bbox[2] != 25 || bbox[3] != 25 {
		t.Fatalf("expected size (25,25), got (%f,%f)",
			bbox[2], bbox[3])
	}
}

func TestValidateSvgSourceInlineAlwaysValid(t *testing.T) {
	if err := validateSvgSource("<svg></svg>"); err != nil {
		t.Fatalf("inline SVG should be valid: %v", err)
	}
}

func TestValidateSvgSourceRejectsTraversal(t *testing.T) {
	if err := validateSvgSource("../secret/file.svg"); err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestValidateSvgSourceRejectsBadExtension(t *testing.T) {
	if err := validateSvgSource("file.png"); err == nil {
		t.Fatal("expected error for non-svg extension")
	}
}

func TestValidateSvgSourceAcceptsGoodPath(t *testing.T) {
	if err := validateSvgSource("/images/icon.svg"); err != nil {
		t.Fatalf("expected valid: %v", err)
	}
}

func TestCheckSvgSourceSizeInline(t *testing.T) {
	if err := checkSvgSourceSize("<svg></svg>"); err != nil {
		t.Fatalf("small inline should pass: %v", err)
	}
}

func TestCheckSvgSourceSizeFileMissing(t *testing.T) {
	if err := checkSvgSourceSize("/nonexistent.svg"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCheckSvgSourceSizeFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.svg")
	os.WriteFile(path, []byte("<svg></svg>"), 0o644)
	if err := checkSvgSourceSize(path); err != nil {
		t.Fatalf("expected valid: %v", err)
	}
}

func TestLoadSvgNoParser(t *testing.T) {
	w := &Window{}
	_, err := w.LoadSvg("<svg></svg>", 100, 100)
	if err == nil {
		t.Fatal("expected error with no parser")
	}
}

func TestGetSvgDimensionsNoParser(t *testing.T) {
	w := &Window{}
	_, _, err := w.GetSvgDimensions("<svg></svg>")
	if err == nil {
		t.Fatal("expected error with no parser")
	}
}

func TestClearSvgCacheNoOp(t *testing.T) {
	w := &Window{}
	// Should not panic.
	w.ClearSvgCache()
}

func TestRemoveSvgFromCacheNoOp(t *testing.T) {
	w := &Window{}
	// Should not panic.
	w.RemoveSvgFromCache("nonexistent.svg")
}

func TestCachedSvgTextDrawsEmpty(t *testing.T) {
	result := cachedSvgTextDraws(nil, 1.0, &Window{})
	if len(result) != 0 {
		t.Fatal("expected empty result")
	}
}

func TestCachedSvgTextDrawsSkipsEmpty(t *testing.T) {
	texts := []SvgText{{Text: ""}}
	result := cachedSvgTextDraws(texts, 1.0, &Window{})
	if len(result) != 0 {
		t.Fatal("expected empty for blank text")
	}
}

func TestCachedSvgTextDrawsOpacity(t *testing.T) {
	texts := []SvgText{{
		Text:     "hi",
		FontSize: 12,
		Color:    SvgColor{255, 255, 255, 255},
		Opacity:  0.5,
	}}
	result := cachedSvgTextDraws(texts, 1.0, &Window{})
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	if result[0].TextStyle.Color.A != 127 {
		t.Fatalf("expected alpha 127, got %d",
			result[0].TextStyle.Color.A)
	}
}

func TestCachedSvgTextDrawsAnchorMiddle(t *testing.T) {
	texts := []SvgText{{
		Text:     "hello",
		FontSize: 10,
		X:        100,
		Y:        50,
		Anchor:   1,
		Color:    SvgColor{0, 0, 0, 255},
		Opacity:  1.0,
	}}
	// No textMeasurer → tw=0, anchor shift is zero.
	// Just verify the draw is created with correct base X.
	result := cachedSvgTextDraws(texts, 1.0, &Window{})
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	// Without measurer, X = X*scale - 0/2 = 100.
	if result[0].X != 100 {
		t.Fatalf("expected X=100 without measurer, got %f",
			result[0].X)
	}
}

func TestCachedSvgTextDrawsAnchorEnd(t *testing.T) {
	texts := []SvgText{{
		Text:     "hello",
		FontSize: 10,
		X:        100,
		Y:        50,
		Anchor:   2,
		Color:    SvgColor{0, 0, 0, 255},
		Opacity:  1.0,
	}}
	result := cachedSvgTextDraws(texts, 1.0, &Window{})
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	// Without measurer, tw=0, X = 100 - 0 = 100.
	if result[0].X != 100 {
		t.Fatalf("expected X=100 without measurer, got %f",
			result[0].X)
	}
}
