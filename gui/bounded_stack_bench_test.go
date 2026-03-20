package gui

import "testing"

func BenchmarkBoundedStackPushBelowCapacity(b *testing.B) {
	s := NewBoundedStack[int](1024)
	b.ReportAllocs()
	b.ResetTimer()
	n := 0
	for b.Loop() {
		if s.Len() == 1024 {
			s.Clear()
		}
		s.Push(n)
		n++
	}
}

func BenchmarkBoundedStackPushOverflow(b *testing.B) {
	s := NewBoundedStack[int](1024)
	for i := 0; i < 1024; i++ {
		s.Push(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	n := 0
	for b.Loop() {
		s.Push(n)
		n++
	}
}

func BenchmarkBoundedStackPushPopCycle(b *testing.B) {
	s := NewBoundedStack[int](1024)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		s.Push(0)
		s.Pop()
	}
}
