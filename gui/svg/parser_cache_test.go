package svg

import "testing"

func TestParserReleaseParsedRemovesEntry(t *testing.T) {
	p := New()
	parsed, err := p.ParseSvg(`<svg viewBox="0 0 10 10"><rect x="0" y="0" width="10" height="10"/></svg>`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if _, ok := p.parsed.Load(parsed); !ok {
		t.Fatal("expected parser cache entry after parse")
	}
	p.ReleaseParsed(parsed)
	if _, ok := p.parsed.Load(parsed); ok {
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
	if _, ok := p.parsed.Load(parsedA); ok {
		t.Fatal("expected A to be removed")
	}
	if _, ok := p.parsed.Load(parsedB); ok {
		t.Fatal("expected B to be removed")
	}
}
