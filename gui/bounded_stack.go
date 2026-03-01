package gui

// BoundedStack is a stack with maximum size. When full, oldest
// entries are dropped (FIFO eviction).
type BoundedStack[T any] struct {
	elements []T
	maxSize  int
}

// NewBoundedStack creates a BoundedStack with the given max size.
func NewBoundedStack[T any](maxSize int) *BoundedStack[T] {
	return &BoundedStack[T]{maxSize: maxSize}
}

// Push adds element to stack. Drops oldest if at capacity.
func (s *BoundedStack[T]) Push(elem T) {
	if len(s.elements) >= s.maxSize {
		// Drop oldest (index 0)
		copy(s.elements, s.elements[1:])
		s.elements[len(s.elements)-1] = elem
		return
	}
	s.elements = append(s.elements, elem)
}

// Pop removes and returns top element. Returns (zero, false) if empty.
func (s *BoundedStack[T]) Pop() (T, bool) {
	if len(s.elements) == 0 {
		var zero T
		return zero, false
	}
	elem := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return elem, true
}

// Len returns number of elements.
func (s *BoundedStack[T]) Len() int {
	return len(s.elements)
}

// IsEmpty returns true if stack has no elements.
func (s *BoundedStack[T]) IsEmpty() bool {
	return len(s.elements) == 0
}

// Clear removes all elements.
func (s *BoundedStack[T]) Clear() {
	s.elements = s.elements[:0]
}
