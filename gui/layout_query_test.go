package gui

import "testing"

func TestCollectFocusCandidatesDedupes(t *testing.T) {
	s1 := &Shape{IDFocus: 9}
	s2 := &Shape{IDFocus: 9}
	root := &Layout{
		Shape: &Shape{},
		Children: []Layout{
			{Shape: s1},
			{Shape: s2},
		},
	}
	var candidates []focusCandidate
	seen := make(map[uint32]bool)
	collectFocusCandidates(root, &candidates, seen)
	if len(candidates) != 1 {
		t.Fatalf("candidates: got %d, want 1", len(candidates))
	}
	if candidates[0].id != 9 {
		t.Errorf("id: got %d, want 9", candidates[0].id)
	}
}

func TestFocusFindNextByID(t *testing.T) {
	s1 := &Shape{IDFocus: 30}
	s2 := &Shape{IDFocus: 10}
	s3 := &Shape{IDFocus: 40}
	candidates := []focusCandidate{
		{id: 30, shape: s1},
		{id: 10, shape: s2},
		{id: 40, shape: s3},
	}
	next, ok := focusFindNext(candidates, 20)
	if !ok {
		t.Fatal("missing next focus")
	}
	if next.IDFocus != 30 {
		t.Errorf("next: got %d, want 30", next.IDFocus)
	}
	fallback, ok := focusFindNext(candidates, 99)
	if !ok {
		t.Fatal("missing fallback")
	}
	if fallback.IDFocus != 10 {
		t.Errorf("fallback: got %d, want 10", fallback.IDFocus)
	}
}

func TestFocusFindPreviousByID(t *testing.T) {
	s1 := &Shape{IDFocus: 30}
	s2 := &Shape{IDFocus: 10}
	s3 := &Shape{IDFocus: 40}
	candidates := []focusCandidate{
		{id: 30, shape: s1},
		{id: 10, shape: s2},
		{id: 40, shape: s3},
	}
	prev, ok := focusFindPrevious(candidates, 35)
	if !ok {
		t.Fatal("missing previous")
	}
	if prev.IDFocus != 30 {
		t.Errorf("prev: got %d, want 30", prev.IDFocus)
	}
	fallback, ok := focusFindPrevious(candidates, 1)
	if !ok {
		t.Fatal("missing fallback")
	}
	if fallback.IDFocus != 40 {
		t.Errorf("fallback: got %d, want 40", fallback.IDFocus)
	}
}
