package gui

import "testing"

func TestRenderSvgNoParser(t *testing.T) {
	w := &Window{}
	shape := &Shape{
		ShapeType: ShapeSVG,
		X:         10,
		Y:         20,
		Width:     100,
		Height:    100,
		Resource:  "<svg></svg>",
	}
	clip := DrawClip{X: 0, Y: 0, Width: 500, Height: 500}
	// Should not panic; emits error placeholder.
	renderSvg(shape, clip, w)

	hasRect := false
	for _, r := range w.renderers {
		if r.Kind == RenderRect && r.Color == Magenta {
			hasRect = true
		}
	}
	if !hasRect {
		t.Fatal("expected magenta error placeholder")
	}
}

func TestRenderSvgOutOfClip(t *testing.T) {
	w := &Window{}
	shape := &Shape{
		ShapeType: ShapeSVG,
		X:         2000,
		Y:         2000,
		Width:     100,
		Height:    100,
	}
	clip := DrawClip{X: 0, Y: 0, Width: 100, Height: 100}
	renderSvg(shape, clip, w)
	if !shape.Disabled {
		t.Fatal("expected disabled when out of clip")
	}
}

func TestRenderSvgWithParser(t *testing.T) {
	w := &Window{}
	w.SetSvgParser(&mockSvgParser{width: 64, height: 64})

	shape := &Shape{
		ShapeType: ShapeSVG,
		X:         0,
		Y:         0,
		Width:     100,
		Height:    100,
		Resource:  "<svg></svg>",
	}
	clip := DrawClip{X: 0, Y: 0, Width: 500, Height: 500}
	renderSvg(shape, clip, w)

	hasSvg := false
	hasClip := false
	for _, r := range w.renderers {
		if r.Kind == RenderSvg {
			hasSvg = true
		}
		if r.Kind == RenderClip {
			hasClip = true
		}
	}
	if !hasSvg {
		t.Fatal("expected RenderSvg command")
	}
	if !hasClip {
		t.Fatal("expected RenderClip for viewBox")
	}
}

func TestEmitSvgPathRendererTint(t *testing.T) {
	w := &Window{}
	path := CachedSvgPath{
		Triangles: []float32{0, 0, 10, 0, 5, 10, 5, 10, 10, 0, 10, 10},
		Color:     Color{0, 0, 0, 255},
	}
	tint := Color{255, 0, 0, 200}
	emitSvgPathRenderer(path, tint, 0, 0, 1.0, w)

	if len(w.renderers) != 1 {
		t.Fatalf("expected 1 renderer, got %d",
			len(w.renderers))
	}
	// Tint should override path color when no vertex colors.
	if w.renderers[0].Color != tint {
		t.Fatalf("expected tint color, got %+v",
			w.renderers[0].Color)
	}
}

func TestEmitSvgPathRendererVertexColors(t *testing.T) {
	w := &Window{}
	path := CachedSvgPath{
		Triangles: []float32{0, 0, 10, 0, 5, 10, 5, 10, 10, 0, 10, 10},
		Color:     Color{0, 0, 0, 255},
		VertexColors: []Color{
			{255, 0, 0, 255},
			{0, 255, 0, 255},
			{0, 0, 255, 255},
			{255, 255, 0, 255},
			{0, 255, 255, 255},
			{255, 0, 255, 255},
		},
	}
	// No tint (A=0) → vertex colors used.
	emitSvgPathRenderer(path, Color{}, 0, 0, 1.0, w)

	if len(w.renderers[0].VertexColors) != 6 {
		t.Fatalf("expected 6 vertex colors, got %d",
			len(w.renderers[0].VertexColors))
	}
}

func TestEmitCachedSvgTextDraw(t *testing.T) {
	w := &Window{}
	draw := CachedSvgTextDraw{
		Text: "hello",
		TextStyle: TextStyle{
			Family: "sans",
			Size:   12,
			Color:  Color{0, 0, 0, 255},
		},
		X: 5,
		Y: 10,
	}
	emitCachedSvgTextDraw(draw, 100, 200, w)

	if len(w.renderers) != 1 {
		t.Fatalf("expected 1 renderer, got %d",
			len(w.renderers))
	}
	r := w.renderers[0]
	if r.Kind != RenderText {
		t.Fatalf("expected RenderText, got %d", r.Kind)
	}
	if r.X != 105 || r.Y != 210 {
		t.Fatalf("expected (105,210), got (%f,%f)", r.X, r.Y)
	}
	if r.Text != "hello" {
		t.Fatalf("expected 'hello', got %q", r.Text)
	}
}

func TestEmitErrorPlaceholder(t *testing.T) {
	w := &Window{}
	emitErrorPlaceholder(10, 20, 50, 30, w)

	if len(w.renderers) != 2 {
		t.Fatalf("expected 2 renderers, got %d",
			len(w.renderers))
	}
	if w.renderers[0].Kind != RenderRect {
		t.Fatal("expected RenderRect")
	}
	if w.renderers[0].Color != Magenta {
		t.Fatal("expected Magenta fill")
	}
	if w.renderers[1].Kind != RenderStrokeRect {
		t.Fatal("expected RenderStrokeRect")
	}
	if w.renderers[1].Color != White {
		t.Fatal("expected White stroke")
	}
}

func TestEmitErrorPlaceholderZeroSize(t *testing.T) {
	w := &Window{}
	emitErrorPlaceholder(10, 20, 0, 30, w)
	if len(w.renderers) != 0 {
		t.Fatal("expected no renderers for zero width")
	}
}

func TestRenderSvgDispatch(t *testing.T) {
	w := &Window{}
	w.SetSvgParser(&mockSvgParser{width: 64, height: 64})

	shape := &Shape{
		ShapeType: ShapeSVG,
		X:         10,
		Y:         10,
		Width:     100,
		Height:    100,
		Resource:  "<svg></svg>",
		Opacity:   1.0,
		Color:     ColorTransparent,
	}
	clip := DrawClip{X: 0, Y: 0, Width: 500, Height: 500}

	// Call through the dispatch.
	renderShapeInner(shape, ColorTransparent, clip, w)

	hasSvg := false
	for _, r := range w.renderers {
		if r.Kind == RenderSvg {
			hasSvg = true
		}
	}
	if !hasSvg {
		t.Fatal("dispatch should call renderSvg")
	}
}

func TestRenderImageDispatch(t *testing.T) {
	w := &Window{}
	shape := &Shape{
		ShapeType: ShapeImage,
		X:         10,
		Y:         10,
		Width:     100,
		Height:    100,
		Resource:  "test.png",
		Opacity:   1.0,
	}
	clip := DrawClip{X: 0, Y: 0, Width: 500, Height: 500}
	renderShapeInner(shape, ColorTransparent, clip, w)

	hasImage := false
	for _, r := range w.renderers {
		if r.Kind == RenderImage {
			hasImage = true
		}
	}
	if !hasImage {
		t.Fatal("dispatch should call renderImage")
	}
}
