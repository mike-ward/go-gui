package gui

import "testing"

func BenchmarkBoundedStackPushBelowCapacity(b *testing.B) {
	s := NewBoundedStack[int](1024)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if s.Len() == 1024 {
			s.Clear()
		}
		s.Push(i)
	}
}

func BenchmarkBoundedStackPushOverflow(b *testing.B) {
	s := NewBoundedStack[int](1024)
	for i := 0; i < 1024; i++ {
		s.Push(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(i)
	}
}

func BenchmarkBoundedStackPushPopCycle(b *testing.B) {
	s := NewBoundedStack[int](1024)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(i)
		s.Pop()
	}
}
