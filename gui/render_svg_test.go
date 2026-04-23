package gui

import (
	"math"
	"testing"
)

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
	if len(w.renderers) != 0 {
		t.Fatal("expected no render commands when out of clip")
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
		Color:     Color{0, 0, 0, 255, true},
	}
	tint := Color{255, 0, 0, 200, true}
	emitSvgPathRenderer(path, tint, 0, 0, 1.0, nil, w)

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
		Color:     Color{0, 0, 0, 255, true},
		VertexColors: []Color{
			{255, 0, 0, 255, true},
			{0, 255, 0, 255, true},
			{0, 0, 255, 255, true},
			{255, 255, 0, 255, true},
			{0, 255, 255, 255, true},
			{255, 0, 255, 255, true},
		},
	}
	// No tint (A=0) → vertex colors used.
	emitSvgPathRenderer(path, Color{}, 0, 0, 1.0, nil, w)

	if len(w.renderers[0].VertexColors) != 6 {
		t.Fatalf("expected 6 vertex colors, got %d",
			len(w.renderers[0].VertexColors))
	}
}

func TestEmitSvgPathRendererAnimatedVertexAlphaNoCopy(t *testing.T) {
	w := &Window{}
	path := CachedSvgPath{
		Triangles: []float32{0, 0, 10, 0, 5, 10, 5, 10, 10, 0, 10, 10},
		Color:     Color{0, 0, 0, 255, true},
		VertexColors: []Color{
			{255, 0, 0, 255, true},
			{0, 255, 0, 255, true},
			{0, 0, 255, 255, true},
			{255, 255, 0, 255, true},
			{0, 255, 255, 255, true},
			{255, 0, 255, 255, true},
		},
		GroupID: "g1",
	}
	animState := map[string]svgAnimState{
		"g1": {Opacity: 0.5, FillOpacity: 1, StrokeOpacity: 1, Inited: true},
	}
	emitSvgPathRenderer(path, Color{}, 0, 0, 1.0, animState, w)

	r := w.renderers[0]
	if !r.HasVertexAlpha {
		t.Fatal("expected vertex alpha scaling flag")
	}
	if r.VertexAlphaScale != 0.5 {
		t.Fatalf("expected alpha scale 0.5, got %f", r.VertexAlphaScale)
	}
	if &r.VertexColors[0] != &path.VertexColors[0] {
		t.Fatal("expected vertex colors slice to be reused without copy")
	}
}

func TestEmitCachedSvgTextDraw(t *testing.T) {
	w := &Window{}
	draw := CachedSvgTextDraw{
		Text: "hello",
		TextStyle: TextStyle{
			Family: "sans",
			Size:   12,
			Color:  Color{0, 0, 0, 255, true},
		},
		X: 5,
		Y: 10,
	}
	emitCachedSvgTextDraw(&draw, 100, 200, w)

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
	if r.TextStylePtr == nil {
		t.Fatal("expected TextStylePtr to be set")
	}
	if r.TextStylePtr.Family != "sans" {
		t.Fatalf("expected family 'sans', got %q",
			r.TextStylePtr.Family)
	}
}

func TestEmitCachedSvgTextDrawWithStyle(t *testing.T) {
	w := &Window{}
	draw := CachedSvgTextDraw{
		Text: "styled",
		TextStyle: TextStyle{
			Family:        "serif",
			Size:          24,
			Color:         Color{255, 0, 0, 255, true},
			Underline:     true,
			Strikethrough: true,
			StrokeWidth:   2,
			StrokeColor:   Color{0, 0, 255, 255, true},
		},
		X: 10,
		Y: 20,
	}
	emitCachedSvgTextDraw(&draw, 0, 0, w)

	r := w.renderers[0]
	if r.TextStylePtr == nil {
		t.Fatal("expected TextStylePtr")
	}
	ts := r.TextStylePtr
	if !ts.Underline {
		t.Fatal("expected underline")
	}
	if !ts.Strikethrough {
		t.Fatal("expected strikethrough")
	}
	if ts.StrokeWidth != 2 {
		t.Fatalf("expected stroke width 2, got %f",
			ts.StrokeWidth)
	}
}

func TestEmitCachedSvgTextPathDraw(t *testing.T) {
	w := &Window{}
	draw := CachedSvgTextPathDraw{
		Text: "path text",
		TextStyle: TextStyle{
			Family: "sans",
			Size:   12,
			Color:  Color{10, 20, 30, 255, true},
		},
		Path: TextPathData{
			Polyline: []float32{0, 0, 10, 0},
			Table:    []float32{0, 10},
			TotalLen: 10,
		},
	}
	emitCachedSvgTextPathDraw(&draw, 7, 9, w)
	if len(w.renderers) != 1 {
		t.Fatalf("expected 1 renderer, got %d",
			len(w.renderers))
	}
	r := w.renderers[0]
	if r.Kind != RenderTextPath {
		t.Fatalf("expected RenderTextPath, got %d", r.Kind)
	}
	if r.TextStylePtr == nil || r.TextPath == nil {
		t.Fatal("expected text path pointers")
	}
	if r.TextStylePtr.Family != "sans" {
		t.Fatalf("unexpected family: %q", r.TextStylePtr.Family)
	}
	if r.TextPath.TotalLen != 10 {
		t.Fatalf("unexpected total len: %f", r.TextPath.TotalLen)
	}
}

func TestComputeSvgAnimationsOpacityZeroPreserved(t *testing.T) {
	// Animation A sets opacity=0 on group "g1".
	// Animation B is a rotate on the same group.
	// Opacity must stay 0, not reset to 1.
	anims := []SvgAnimation{
		{
			GroupID: "g1",
			Kind:    SvgAnimOpacity,
			DurSec:  1,
			Values:  []float32{0, 0}, // constant 0
		},
		{
			GroupID: "g1",
			Kind:    SvgAnimRotate,
			DurSec:  2,
			Values:  []float32{0, 360},
			CenterX: 50,
			CenterY: 50,
		},
	}
	states := computeSvgAnimations(anims, 0.5, nil)
	st, ok := states["g1"]
	if !ok {
		t.Fatal("expected state for g1")
	}
	if st.Opacity != 0 {
		t.Fatalf("expected opacity 0, got %f", st.Opacity)
	}
	if st.RotAngle == 0 {
		t.Fatal("expected non-zero rotation")
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

// --- finiteF32 ---

func TestFiniteF32_NaNInfFinite(t *testing.T) {
	if !finiteF32(0) || !finiteF32(1.5) || !finiteF32(-1e20) {
		t.Fatal("finite values should report true")
	}
	if finiteF32(float32(math.NaN())) {
		t.Fatal("NaN must report false")
	}
	if finiteF32(float32(math.Inf(1))) {
		t.Fatal("+Inf must report false")
	}
	if finiteF32(float32(math.Inf(-1))) {
		t.Fatal("-Inf must report false")
	}
}

// --- collectAnimContribs timing guards ---

// collectAnimContribs must reject animations whose timing fields
// are NaN or ±Inf. Otherwise downstream lerp/floor math produces
// NaN values that poison the anim state map.
func TestCollectAnimContribs_RejectsNonFiniteTimings(t *testing.T) {
	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	base := SvgAnimation{
		Kind:    SvgAnimOpacity,
		GroupID: "g",
		Values:  []float32{0, 1},
		DurSec:  1,
	}
	cases := []struct {
		name string
		mut  func(a *SvgAnimation)
	}{
		{"NaN Dur", func(a *SvgAnimation) { a.DurSec = nan }},
		{"Inf Dur", func(a *SvgAnimation) { a.DurSec = inf }},
		{"NaN Begin", func(a *SvgAnimation) { a.BeginSec = nan }},
		{"Inf Begin", func(a *SvgAnimation) { a.BeginSec = inf }},
		{"NaN Cycle", func(a *SvgAnimation) { a.Cycle = nan }},
		{"Inf Cycle", func(a *SvgAnimation) { a.Cycle = inf }},
	}
	for _, tc := range cases {
		a := base
		tc.mut(&a)
		out := collectAnimContribs([]SvgAnimation{a}, 0.5, nil)
		if len(out) != 0 {
			t.Fatalf("%s: expected no contribution, got %d", tc.name, len(out))
		}
	}
	// Sanity: a clean animation still contributes.
	out := collectAnimContribs([]SvgAnimation{base}, 0.5, nil)
	if len(out) != 1 {
		t.Fatalf("clean anim should contribute, got %d", len(out))
	}
}

func TestCollectAnimContribs_RejectsNonFiniteElapsed(t *testing.T) {
	a := SvgAnimation{
		Kind:    SvgAnimOpacity,
		GroupID: "g",
		Values:  []float32{0, 1},
		DurSec:  1,
	}
	out := collectAnimContribs([]SvgAnimation{a}, float32(math.NaN()), nil)
	if len(out) != 0 {
		t.Fatal("NaN elapsed must produce no contributions")
	}
}

// --- Freeze semantics ---

// fill="freeze" must hold the last keyframe value after dur and
// before the next cycle. Without freeze the animation drops out.
func TestComputeSvgAnimations_FreezeHoldsLastValue(t *testing.T) {
	a := SvgAnimation{
		Kind:    SvgAnimOpacity,
		GroupID: "g",
		Values:  []float32{1, 0},
		DurSec:  1,
		Freeze:  true,
	}
	// elapsed 5s, dur 1s, no cycle → past end. Freeze must hold
	// frac=1 → value 0.
	st := computeSvgAnimations([]SvgAnimation{a}, 5, nil)
	if st["g"].Opacity != 0 {
		t.Fatalf("freeze should hold final value 0, got %f", st["g"].Opacity)
	}

	// Without freeze, same animation contributes nothing: state
	// for "g" never gets created.
	a.Freeze = false
	st = computeSvgAnimations([]SvgAnimation{a}, 5, nil)
	if _, ok := st["g"]; ok {
		t.Fatal("non-freeze past-dur must not contribute state")
	}
}

// --- Cycle restart ---

// Cycle>0 must re-fire the animation every cycle seconds. At
// elapsed = Cycle+epsilon the phase is effectively 0 so the lerp
// returns the first keyframe value again.
func TestComputeSvgAnimations_CycleRestart(t *testing.T) {
	a := SvgAnimation{
		Kind:    SvgAnimOpacity,
		GroupID: "g",
		Values:  []float32{1, 0}, // 1→0 over dur
		DurSec:  1,
		Cycle:   2, // repeats every 2s
	}
	// At elapsed=2 (cycle boundary) phase ≈ 0 → opacity ≈ 1.
	st := computeSvgAnimations([]SvgAnimation{a}, 2, nil)
	if math.Abs(float64(st["g"].Opacity-1)) > 1e-3 {
		t.Fatalf("expected opacity ~1 at cycle restart, got %f",
			st["g"].Opacity)
	}
}

// --- Sandwich priority ---

// When two animations target the same attribute on the same
// group, the one whose last activation is later must win under
// SMIL sandwich semantics.
func TestComputeSvgAnimations_SandwichLastActivationWins(t *testing.T) {
	// a1 activates at t=0, a2 at t=1. At elapsed=2 both are past
	// their dur (each is 1s). With Freeze both contribute; a2's
	// activation (1) > a1's (0), so a2's final value wins.
	a1 := SvgAnimation{
		Kind:     SvgAnimOpacity,
		GroupID:  "g",
		Values:   []float32{0, 0.25},
		DurSec:   1,
		BeginSec: 0,
		Freeze:   true,
	}
	a2 := SvgAnimation{
		Kind:     SvgAnimOpacity,
		GroupID:  "g",
		Values:   []float32{0, 0.75},
		DurSec:   1,
		BeginSec: 1,
		Freeze:   true,
	}
	st := computeSvgAnimations([]SvgAnimation{a1, a2}, 2, nil)
	if math.Abs(float64(st["g"].Opacity-0.75)) > 1e-3 {
		t.Fatalf("expected a2's value 0.75 to win, got %f", st["g"].Opacity)
	}
}

// --- Opacity clamp on render ---

// An opacity value outside [0,1] (reached via a hostile anim
// values list) must not drive a negative or >1 multiplier into
// the uint8 alpha cast. clampFrac in emitSvgPathRenderer guards
// this; verify NaN opacity → alpha 0.
func TestEmitSvgPathRenderer_OpacityNaNClampedToZero(t *testing.T) {
	w := &Window{}
	path := CachedSvgPath{
		Triangles: []float32{0, 0, 10, 0, 5, 10, 5, 10, 10, 0, 10, 10},
		Color:     Color{10, 20, 30, 200, true},
		GroupID:   "g1",
	}
	animState := map[string]svgAnimState{
		"g1": {
			Opacity:       float32(math.NaN()),
			FillOpacity:   1,
			StrokeOpacity: 1,
			Inited:        true,
		},
	}
	emitSvgPathRenderer(path, Color{}, 0, 0, 1.0, animState, w)
	if len(w.renderers) != 1 {
		t.Fatalf("expected 1 renderer, got %d", len(w.renderers))
	}
	if w.renderers[0].Color.A != 0 {
		t.Fatalf("NaN opacity must clamp alpha to 0, got %d",
			w.renderers[0].Color.A)
	}
}

// Fill-opacity animation targeting a stroke path must NOT dim
// the stroke (and vice-versa). emitSvgPathRenderer picks
// FillOpacity vs StrokeOpacity based on path.IsStroke.
func TestSvgRender_FillOpacityAnimDoesNotDimStroke(t *testing.T) {
	w := &Window{}
	strokePath := CachedSvgPath{
		Triangles: []float32{0, 0, 10, 0, 5, 10, 5, 10, 10, 0, 10, 10},
		Color:     Color{0, 0, 0, 255, true},
		GroupID:   "g1",
		IsStroke:  true,
	}
	// FillOpacity drops to 0; StrokeOpacity stays 1. The stroke
	// path's alpha must remain 255.
	animState := map[string]svgAnimState{
		"g1": {
			Opacity:       1,
			FillOpacity:   0,
			StrokeOpacity: 1,
			Inited:        true,
		},
	}
	emitSvgPathRenderer(strokePath, Color{}, 0, 0, 1.0, animState, w)
	if w.renderers[0].Color.A != 255 {
		t.Fatalf("stroke should keep alpha 255 when only fill-opacity "+
			"is animated, got %d", w.renderers[0].Color.A)
	}

	// Mirror: fill path under stroke-opacity=0 stays opaque.
	w = &Window{}
	fillPath := strokePath
	fillPath.IsStroke = false
	animState = map[string]svgAnimState{
		"g1": {
			Opacity:       1,
			FillOpacity:   1,
			StrokeOpacity: 0,
			Inited:        true,
		},
	}
	emitSvgPathRenderer(fillPath, Color{}, 0, 0, 1.0, animState, w)
	if w.renderers[0].Color.A != 255 {
		t.Fatalf("fill should keep alpha 255 when only stroke-opacity "+
			"is animated, got %d", w.renderers[0].Color.A)
	}
}
