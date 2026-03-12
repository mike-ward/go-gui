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
	if p.Color != (Color{255, 0, 0, 255, true}) {
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
	if result[0].VertexColors[0] != (Color{255, 0, 0, 255, true}) {
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

func TestValidateSvgSourceRejectsNul(t *testing.T) {
	if err := validateSvgSource("file.svg\x00"); err == nil {
		t.Fatal("expected error for NUL byte")
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

func TestBuildDefsPathDataCache(t *testing.T) {
	defs := map[string]string{
		"p1": "M 0 0 L 10 0",
		"p2": "M 0 0 L 0 0",
	}
	textPaths := []SvgTextPath{
		{PathID: "p1"},
		{PathID: "missing"},
	}
	cache := buildDefsPathDataCache(textPaths, nil, defs, 2.0)
	if len(cache) != 1 {
		t.Fatalf("expected 1 cached path, got %d", len(cache))
	}
	entry, ok := cache["p1"]
	if !ok {
		t.Fatal("expected p1 in cache")
	}
	if len(entry.polyline) < 4 || len(entry.table) < 2 || entry.totalLen <= 0 {
		t.Fatal("expected non-empty cached polyline/table/length")
	}
}

func TestBuildDefsPathDataCacheIncludesFilteredTextPaths(t *testing.T) {
	defs := map[string]string{
		"fg1": "M 0 0 L 20 0",
	}
	filtered := []SvgParsedFilteredGroup{
		{
			TextPaths: []SvgTextPath{
				{PathID: "fg1"},
			},
		},
	}
	cache := buildDefsPathDataCache(nil, filtered, defs, 1.0)
	if len(cache) != 1 {
		t.Fatalf("expected 1 cached path from filtered group, got %d", len(cache))
	}
}

func TestValidateSvgSourceWithRootsAcceptsWithinRoot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "icon.svg")
	if err := validateSvgSourceWithRoots(path, []string{root}); err != nil {
		t.Fatalf("expected path in root to be allowed: %v", err)
	}
}

func TestValidateSvgSourceWithRootsRejectsOutsideRoot(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	path := filepath.Join(other, "icon.svg")
	if err := validateSvgSourceWithRoots(path, []string{root}); err == nil {
		t.Fatal("expected path outside root to be rejected")
	}
}

func TestValidateSvgSourceWithRootsRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	target := filepath.Join(other, "outside.svg")
	if err := os.WriteFile(target, []byte("<svg/>"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	link := filepath.Join(root, "link.svg")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := validateSvgSourceWithRoots(link, []string{root}); err == nil {
		t.Fatal("expected symlink escape to be rejected")
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

func TestLoadSvgRespectsAllowedRoots(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "ok.svg")
	if err := os.WriteFile(inside, []byte("<svg viewBox=\"0 0 10 10\"></svg>"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	w := NewWindow(WindowCfg{AllowedSvgRoots: []string{root}})
	w.SetSvgParser(&mockSvgParser{width: 10, height: 10})
	if _, err := w.LoadSvg(inside, 10, 10); err != nil {
		t.Fatalf("expected in-root load to pass: %v", err)
	}
	outsideDir := t.TempDir()
	outside := filepath.Join(outsideDir, "nope.svg")
	if err := os.WriteFile(outside, []byte("<svg viewBox=\"0 0 10 10\"></svg>"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := w.LoadSvg(outside, 10, 10); err == nil {
		t.Fatal("expected out-of-root load to fail")
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
	result := cachedSvgTextDraws(nil, 1.0, nil, &Window{})
	if len(result) != 0 {
		t.Fatal("expected empty result")
	}
}

func TestCachedSvgTextDrawsSkipsEmpty(t *testing.T) {
	texts := []SvgText{{Text: ""}}
	result := cachedSvgTextDraws(texts, 1.0, nil, &Window{})
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
	result := cachedSvgTextDraws(texts, 1.0, nil, &Window{})
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	if result[0].TextStyle.Color.A != 128 {
		t.Fatalf("expected alpha 128, got %d",
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
	result := cachedSvgTextDraws(texts, 1.0, nil, &Window{})
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	// Without measurer, X = X*scale - 0/2 = 100.
	if result[0].X != 100 {
		t.Fatalf("expected X=100 without measurer, got %f",
			result[0].X)
	}
}

func TestCachedSvgTextDrawsIsBoldNoWeight(t *testing.T) {
	texts := []SvgText{{
		Text:     "bold",
		FontSize: 12,
		IsBold:   true,
		Color:    SvgColor{0, 0, 0, 255},
		Opacity:  1.0,
	}}
	result := cachedSvgTextDraws(texts, 1.0, nil, &Window{})
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	if result[0].TextStyle.Family != " Bold" {
		t.Fatalf("expected family ' Bold', got %q",
			result[0].TextStyle.Family)
	}
}

func TestCachedSvgTextDrawsIsBoldWithWeight700(t *testing.T) {
	texts := []SvgText{{
		Text:       "bold",
		FontSize:   12,
		FontWeight: 700,
		IsBold:     true,
		Color:      SvgColor{0, 0, 0, 255},
		Opacity:    1.0,
	}}
	result := cachedSvgTextDraws(texts, 1.0, nil, &Window{})
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	// FontWeight 700 → pangoWeightName returns "Bold", so IsBold
	// fallback should NOT append a second "Bold".
	if result[0].TextStyle.Family != " Bold" {
		t.Fatalf("expected family ' Bold', got %q",
			result[0].TextStyle.Family)
	}
}

func TestCachedSvgTextPathDrawsIsBoldNoWeight(t *testing.T) {
	defs := map[string]string{"p1": "M 0 0 L 100 0"}
	cache := buildDefsPathDataCache(
		[]SvgTextPath{{PathID: "p1", IsBold: true,
			FontSize: 12, Color: SvgColor{0, 0, 0, 255},
			Opacity: 1.0, Text: "bold"}},
		nil, defs, 1.0)
	result := cachedSvgTextPathDraws(
		[]SvgTextPath{{PathID: "p1", IsBold: true,
			FontSize: 12, Color: SvgColor{0, 0, 0, 255},
			Opacity: 1.0, Text: "bold"}},
		cache, 1.0)
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	if result[0].TextStyle.Family != " Bold" {
		t.Fatalf("expected family ' Bold', got %q",
			result[0].TextStyle.Family)
	}
}

func TestCachedSvgTextDrawsOpacityRounding(t *testing.T) {
	texts := []SvgText{{
		Text:     "hi",
		FontSize: 12,
		Color:    SvgColor{255, 255, 255, 255},
		Opacity:  0.999,
	}}
	result := cachedSvgTextDraws(texts, 1.0, nil, &Window{})
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	// 255 * 0.999 = 254.745 → should round to 255, not truncate to 254.
	if result[0].TextStyle.Color.A != 255 {
		t.Fatalf("expected alpha 255 (rounded), got %d",
			result[0].TextStyle.Color.A)
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
	result := cachedSvgTextDraws(texts, 1.0, nil, &Window{})
	if len(result) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(result))
	}
	// Without measurer, tw=0, X = 100 - 0 = 100.
	if result[0].X != 100 {
		t.Fatalf("expected X=100 without measurer, got %f",
			result[0].X)
	}
}
