package svg

import (
	"strconv"
	"testing"
)

func BenchmarkParserCacheHitInline(b *testing.B) {
	p := New()
	src := `<svg viewBox="0 0 24 24"><rect x="0" y="0" width="24" height="24"/></svg>`
	if _, err := p.ParseSvg(src); err != nil {
		b.Fatalf("warm parse failed: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := p.ParseSvg(src); err != nil {
			b.Fatalf("parse failed: %v", err)
		}
	}
}

func BenchmarkParserCacheMissInline(b *testing.B) {
	p := New()
	b.ReportAllocs()
	n := 0
	for b.Loop() {
		src := `<svg viewBox="0 0 24 24"><rect x="0" y="0" width="24" height="24"/><text x="1" y="` +
			strconv.Itoa(n%100) + `">x</text></svg>`
		if _, err := p.ParseSvg(src); err != nil {
			b.Fatalf("parse failed: %v", err)
		}
		n++
	}
}
