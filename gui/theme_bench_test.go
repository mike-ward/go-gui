package gui

import "testing"

func BenchmarkCurrentTheme(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = CurrentTheme()
	}
}
