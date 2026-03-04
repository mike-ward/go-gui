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
	for i := 0; i < b.N; i++ {
		if _, err := p.ParseSvg(src); err != nil {
			b.Fatalf("parse failed: %v", err)
		}
	}
}

func BenchmarkParserCacheMissInline(b *testing.B) {
	p := New()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		src := `<svg viewBox="0 0 24 24"><rect x="0" y="0" width="24" height="24"/><text x="1" y="` +
			strconv.Itoa(i%100) + `">x</text></svg>`
		if _, err := p.ParseSvg(src); err != nil {
			b.Fatalf("parse failed: %v", err)
		}
	}
}
