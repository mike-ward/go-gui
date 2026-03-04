package svg

import "testing"

func TestParserReleaseParsedRemovesEntry(t *testing.T) {
	p := New()
	parsed, err := p.ParseSvg(`<svg viewBox="0 0 10 10"><rect x="0" y="0" width="10" height="10"/></svg>`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	p.mu.Lock()
	_, ok := p.byParsed[parsed]
	p.mu.Unlock()
	if !ok {
		t.Fatal("expected parser cache entry after parse")
	}
	p.ReleaseParsed(parsed)
	p.mu.Lock()
	_, ok = p.byParsed[parsed]
	p.mu.Unlock()
	if ok {
		t.Fatal("expected parser cache entry to be removed")
	}
}

func TestParserClearSvgParserCacheEmptiesMap(t *testing.T) {
	p := New()
	parsedA, err := p.ParseSvg(`<svg viewBox="0 0 10 10"><rect x="0" y="0" width="10" height="10"/></svg>`)
	if err != nil {
		t.Fatalf("parse A failed: %v", err)
	}
	parsedB, err := p.ParseSvg(`<svg viewBox="0 0 20 20"><rect x="0" y="0" width="20" height="20"/></svg>`)
	if err != nil {
		t.Fatalf("parse B failed: %v", err)
	}
	p.ClearSvgParserCache()
	p.mu.Lock()
	_, okA := p.byParsed[parsedA]
	_, okB := p.byParsed[parsedB]
	p.mu.Unlock()
	if okA {
		t.Fatal("expected A to be removed")
	}
	if okB {
		t.Fatal("expected B to be removed")
	}
}

func TestParserCacheHitReturnsSameParsed(t *testing.T) {
	p := New()
	src := `<svg viewBox="0 0 10 10"><rect x="0" y="0" width="10" height="10"/></svg>`
	a, err := p.ParseSvg(src)
	if err != nil {
		t.Fatalf("first parse failed: %v", err)
	}
	b, err := p.ParseSvg(src)
	if err != nil {
		t.Fatalf("second parse failed: %v", err)
	}
	if a != b {
		t.Fatal("expected parser cache hit to return same parsed pointer")
	}
}

func TestParserInvalidateSvgSourceRemovesOnlyTarget(t *testing.T) {
	p := New()
	srcA := `<svg viewBox="0 0 10 10"><rect x="0" y="0" width="10" height="10"/></svg>`
	srcB := `<svg viewBox="0 0 20 20"><rect x="0" y="0" width="20" height="20"/></svg>`
	a, err := p.ParseSvg(srcA)
	if err != nil {
		t.Fatalf("parse A failed: %v", err)
	}
	b, err := p.ParseSvg(srcB)
	if err != nil {
		t.Fatalf("parse B failed: %v", err)
	}
	p.InvalidateSvgSource(srcA)
	p.mu.Lock()
	_, okA := p.byParsed[a]
	_, okB := p.byParsed[b]
	p.mu.Unlock()
	if okA {
		t.Fatal("expected source A to be removed")
	}
	if !okB {
		t.Fatal("expected source B to remain")
	}
}
