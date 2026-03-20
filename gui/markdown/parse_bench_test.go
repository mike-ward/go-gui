package markdown

import (
	"strings"
	"testing"
)

func BenchmarkParseSmall(b *testing.B) {
	source := "# Title\n\nSimple paragraph with **bold**, *italic*, and [link](https://example.com)."
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = Parse(source, false)
	}
}

func BenchmarkParseLarge(b *testing.B) {
	var sb strings.Builder
	sb.Grow(128 * 1024)
	sb.WriteString("# Large Document\n\n")
	for i := 0; i < 800; i++ {
		sb.WriteString("Paragraph ")
		sb.WriteString("with [link](https://example.com), ")
		sb.WriteString("`code`, and --- typography ... updates.\n\n")
	}
	source := sb.String()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = Parse(source, false)
	}
}

func BenchmarkParseAbbrFootnotes(b *testing.B) {
	var sb strings.Builder
	sb.Grow(96 * 1024)
	sb.WriteString("*[HTTP]: HyperText Transfer Protocol\n")
	sb.WriteString("*[CPU]: Central Processing Unit\n")
	sb.WriteString("[^1]: First footnote line.\n")
	sb.WriteString("    Continuation line.\n\n")
	for i := 0; i < 1200; i++ {
		sb.WriteString("HTTP on CPU[^1] is common.\n")
	}
	source := sb.String()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = Parse(source, false)
	}
}
