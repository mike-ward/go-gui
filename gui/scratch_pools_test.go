package gui

import "testing"

func TestTakeFilterRenderersEmpty(t *testing.T) {
	var p scratchPools
	s := p.takeFilterRenderers(0)
	if len(s) != 0 {
		t.Fatalf("want len 0, got %d", len(s))
	}
}

func TestTakeFilterRenderersRequiredCap(t *testing.T) {
	var p scratchPools
	s := p.takeFilterRenderers(100)
	if cap(s) < 100 {
		t.Fatalf("want cap>=100, got %d", cap(s))
	}
}

func TestPutFilterRenderersShrink(t *testing.T) {
	var p scratchPools
	big := make([]RenderCmd, 0, scratchFilterRenderersRetainMax+1)
	p.putFilterRenderers(big)
	if cap(p.filterRenderers) != scratchFilterRenderersShrinkTo {
		t.Fatalf("want cap %d, got %d", scratchFilterRenderersShrinkTo, cap(p.filterRenderers))
	}
}

func TestTakePutFilterRenderersRoundTrip(t *testing.T) {
	var p scratchPools
	s := p.takeFilterRenderers(16)
	s = append(s, RenderCmd{Kind: RenderRect})
	p.putFilterRenderers(s)
	s2 := p.takeFilterRenderers(0)
	if len(s2) != 0 {
		t.Fatalf("want empty after take, got %d", len(s2))
	}
	if cap(s2) < 16 {
		t.Fatal("backing array should be reused")
	}
}

func TestTakeFloatingLayoutsResets(t *testing.T) {
	var p scratchPools
	s := p.takeFloatingLayouts(10)
	if len(s) != 0 || cap(s) < 10 {
		t.Fatal("unexpected size/cap")
	}
	if p.floatingPoolUsed != 0 {
		t.Fatal("floatingPoolUsed should reset")
	}
}

func TestAllocFloatingLayoutNew(t *testing.T) {
	var p scratchPools
	shape := Shape{Width: 42}
	src := Layout{Shape: &shape}
	l := p.allocFloatingLayout(src)
	if l.Shape.Width != 42 {
		t.Fatal("expected width 42")
	}
	if p.floatingPoolUsed != 1 {
		t.Fatal("pool used should be 1")
	}
}

func TestAllocFloatingLayoutReuse(t *testing.T) {
	var p scratchPools
	shape1 := Shape{Width: 1}
	src := Layout{Shape: &shape1}
	first := p.allocFloatingLayout(src)

	// Put back and take again to reset poolUsed.
	p.putFloatingLayouts(p.floatingLayouts)
	_ = p.takeFloatingLayouts(0)

	shape2 := Shape{Width: 2}
	src = Layout{Shape: &shape2}
	second := p.allocFloatingLayout(src)
	if first != second {
		t.Fatal("should reuse same pointer")
	}
	if second.Shape.Width != 2 {
		t.Fatal("expected width 2")
	}
}

func TestTakePutFocusCandidates(t *testing.T) {
	var p scratchPools
	s := p.takeFocusCandidates()
	s = append(s, focusCandidate{})
	p.putFocusCandidates(s)
	s2 := p.takeFocusCandidates()
	if len(s2) != 0 {
		t.Fatal("should be empty")
	}
}

func TestTakePutGradientNormStops(t *testing.T) {
	var p scratchPools
	s := p.takeGradientNormStops(5)
	if cap(s) < 5 {
		t.Fatal("cap too small")
	}
	s = append(s, GradientStop{Pos: 0.5})
	p.putGradientNormStops(s)
	s2 := p.takeGradientNormStops(0)
	if len(s2) != 0 {
		t.Fatal("should be empty after take")
	}
}

func TestTakePutGradientSampleStops(t *testing.T) {
	var p scratchPools
	s := p.takeGradientSampleStops(3)
	s = append(s, GradientStop{})
	p.putGradientSampleStops(s)
	if len(p.gradientSampleStops) != 0 {
		t.Fatal("should be reset")
	}
}

func TestTakePutSvgAnimVals(t *testing.T) {
	var p scratchPools
	s := p.takeSvgAnimVals(8)
	s = append(s, 1.0, 2.0)
	p.putSvgAnimVals(s)
	s2 := p.takeSvgAnimVals(0)
	if len(s2) != 0 {
		t.Fatal("should be empty")
	}
}

func TestTakePutWrapRows(t *testing.T) {
	var p scratchPools
	s := p.takeWrapRows(4)
	s = append(s, wrapRowRange{start: 0, end: 3})
	p.putWrapRows(s)
	s2 := p.takeWrapRows(0)
	if len(s2) != 0 {
		t.Fatal("should be empty")
	}
}

func TestPutWrapRowsShrink(t *testing.T) {
	var p scratchPools
	big := make([]wrapRowRange, 0, scratchWrapRowsRetainMax+1)
	p.putWrapRows(big)
	if cap(p.wrapRows) != scratchWrapRowsShrinkTo {
		t.Fatalf("want cap %d, got %d", scratchWrapRowsShrinkTo, cap(p.wrapRows))
	}
}

func TestTransformSvgTriangles(t *testing.T) {
	var p scratchPools
	tris := []float32{1, 0, 0, 1}
	// Scale by 2.
	m := [6]float32{2, 0, 0, 2, 0, 0}
	out := p.transformSvgTriangles(tris, m)
	if len(out) != 4 || out[0] != 2 || out[1] != 0 || out[2] != 0 || out[3] != 2 {
		t.Fatalf("unexpected: %v", out)
	}
}

func TestTransformSvgTrianglesBatchGrow(t *testing.T) {
	var p scratchPools
	m := [6]float32{1, 0, 0, 1, 0, 0} // identity
	_ = p.transformSvgTriangles([]float32{1, 2}, m)
	_ = p.transformSvgTriangles([]float32{3, 4}, m)
	if p.svgTransformBatchesUsed != 2 {
		t.Fatalf("want 2, got %d", p.svgTransformBatchesUsed)
	}
}

func TestTrimSvgGroupMaps(t *testing.T) {
	var p scratchPools
	p.svgGroupMatrices = make(map[string][6]float32, scratchSvgGroupMatricesRetainMax+1)
	for i := range scratchSvgGroupMatricesRetainMax + 1 {
		p.svgGroupMatrices[string(rune('a'+i%26))+string(rune('0'+i/26))] = [6]float32{}
	}
	p.trimSvgGroupMaps()
	if p.svgGroupMatrices != nil {
		t.Fatal("should be nil after trim")
	}
}

func TestTrimSvgTransformBatches(t *testing.T) {
	var p scratchPools
	big := make([]float32, 0, scratchSvgTrisRetainMax+1)
	p.svgTransformBatches = [][]float32{big}
	p.trimSvgTransformBatches()
	if cap(p.svgTransformBatches[0]) != scratchSvgTrisShrinkTo {
		t.Fatalf("want %d, got %d", scratchSvgTrisShrinkTo, cap(p.svgTransformBatches[0]))
	}
}
