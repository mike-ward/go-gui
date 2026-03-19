package gui

import "testing"

func TestScratchSliceTakeEmpty(t *testing.T) {
	s := scratchSlice[RenderCmd]{retainMax: 1000, shrinkTo: 100}
	b := s.take(0)
	if len(b) != 0 {
		t.Fatalf("want len 0, got %d", len(b))
	}
}

func TestScratchSliceTakeRequiredCap(t *testing.T) {
	s := scratchSlice[RenderCmd]{retainMax: 1000, shrinkTo: 100}
	b := s.take(100)
	if cap(b) < 100 {
		t.Fatalf("want cap>=100, got %d", cap(b))
	}
}

func TestScratchSlicePutShrink(t *testing.T) {
	s := scratchSlice[RenderCmd]{retainMax: 1000, shrinkTo: 100}
	big := make([]RenderCmd, 0, 1001)
	s.put(big)
	if cap(s.buf) != 100 {
		t.Fatalf("want cap 100, got %d", cap(s.buf))
	}
}

func TestScratchSliceRoundTrip(t *testing.T) {
	s := scratchSlice[RenderCmd]{retainMax: 1000, shrinkTo: 100}
	b := s.take(16)
	b = append(b, RenderCmd{Kind: RenderRect})
	s.put(b)
	b2 := s.take(0)
	if len(b2) != 0 {
		t.Fatalf("want empty after take, got %d", len(b2))
	}
	if cap(b2) < 16 {
		t.Fatal("backing array should be reused")
	}
}

func TestScratchMapTakeEmpty(t *testing.T) {
	s := scratchMap[uint32, struct{}]{retainMax: 4096}
	m := s.take(0)
	if len(m) != 0 {
		t.Fatalf("want len 0, got %d", len(m))
	}
}

func TestScratchMapTakeReuse(t *testing.T) {
	s := scratchMap[uint32, struct{}]{retainMax: 4096}
	m := s.take(8)
	m[1] = struct{}{}
	s.put(m)
	m2 := s.take(0)
	if len(m2) != 0 {
		t.Fatal("should be cleared")
	}
}

func TestScratchMapPutDiscard(t *testing.T) {
	s := scratchMap[uint32, struct{}]{retainMax: 2}
	m := make(map[uint32]struct{}, 8)
	m[1] = struct{}{}
	m[2] = struct{}{}
	m[3] = struct{}{}
	s.put(m)
	if s.m != nil {
		t.Fatal("should discard oversized map")
	}
}

func TestTakeFloatingLayoutsResets(t *testing.T) {
	p := newScratchPools()
	s := p.takeFloatingLayouts(10)
	if len(s) != 0 || cap(s) < 10 {
		t.Fatal("unexpected size/cap")
	}
	if p.floatingPoolUsed != 0 {
		t.Fatal("floatingPoolUsed should reset")
	}
}

func TestAllocFloatingLayoutNew(t *testing.T) {
	p := newScratchPools()
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

func TestScratchObjPoolAlloc(t *testing.T) {
	pool := scratchObjPool[TextStyle]{retainMax: 64, shrinkTo: 8}
	s1 := TextStyle{Family: "sans", Size: 14}
	p1 := pool.alloc(s1)
	if p1.Family != "sans" || p1.Size != 14 {
		t.Fatal("alloc should copy value")
	}
	if pool.used != 1 {
		t.Fatal("used should be 1")
	}
}

func TestScratchObjPoolReuse(t *testing.T) {
	pool := scratchObjPool[TextStyle]{retainMax: 64, shrinkTo: 8}
	s1 := TextStyle{Family: "serif", Size: 12}
	first := pool.alloc(s1)

	pool.reset()

	s2 := TextStyle{Family: "mono", Size: 16}
	second := pool.alloc(s2)
	if first != second {
		t.Fatal("should reuse same pointer")
	}
	if second.Family != "mono" || second.Size != 16 {
		t.Fatal("reused slot should have new value")
	}
}

func TestScratchObjPoolResetShrink(t *testing.T) {
	pool := scratchObjPool[TextStyle]{retainMax: 2, shrinkTo: 1}
	for i := range 5 {
		pool.alloc(TextStyle{Size: float32(i)})
	}
	if len(pool.items) != 5 {
		t.Fatalf("want 5 items, got %d", len(pool.items))
	}
	pool.reset()
	if len(pool.items) != 0 || cap(pool.items) != 1 {
		t.Fatalf("want shrunk pool, got len=%d cap=%d",
			len(pool.items), cap(pool.items))
	}
}

func TestAllocFloatingLayoutReuse(t *testing.T) {
	p := newScratchPools()
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
