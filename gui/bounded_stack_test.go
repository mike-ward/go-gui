package gui

import "testing"

func TestBoundedStackPushPop(t *testing.T) {
	s := NewBoundedStack[int](3)
	s.Push(1)
	s.Push(2)
	s.Push(3)

	if s.Len() != 3 {
		t.Errorf("len: got %d", s.Len())
	}
	if v, ok := s.Pop(); !ok || v != 3 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 2 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 1 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if !s.IsEmpty() {
		t.Error("should be empty")
	}
}

func TestBoundedStackOverflowEvictsOldest(t *testing.T) {
	s := NewBoundedStack[int](3)
	s.Push(1)
	s.Push(2)
	s.Push(3)
	s.Push(4) // evicts 1

	if s.Len() != 3 {
		t.Errorf("len: got %d", s.Len())
	}
	if v, ok := s.Pop(); !ok || v != 4 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 3 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 2 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if _, ok := s.Pop(); ok {
		t.Error("should be empty")
	}
}

func TestBoundedStackPopEmpty(t *testing.T) {
	s := NewBoundedStack[int](3)
	if _, ok := s.Pop(); ok {
		t.Error("pop on empty should return false")
	}
	if !s.IsEmpty() {
		t.Error("should be empty")
	}
}

func TestBoundedStackClear(t *testing.T) {
	s := NewBoundedStack[int](3)
	s.Push(1)
	s.Push(2)
	s.Clear()
	if !s.IsEmpty() {
		t.Error("should be empty after clear")
	}
	if s.Len() != 0 {
		t.Errorf("len: got %d", s.Len())
	}
}

func TestBoundedStackDefaultSize(t *testing.T) {
	s := &BoundedStack[int]{maxSize: 50}
	if s.maxSize != 50 {
		t.Errorf("maxSize: got %d", s.maxSize)
	}
}

func TestBoundedStackManyPushes(t *testing.T) {
	s := NewBoundedStack[int](5)
	for i := range 100 {
		s.Push(i)
	}
	if s.Len() != 5 {
		t.Errorf("len: got %d", s.Len())
	}
	// should have 95,96,97,98,99
	if v, ok := s.Pop(); !ok || v != 99 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 98 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 97 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 96 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 95 {
		t.Errorf("pop: got %d, %v", v, ok)
	}
}
