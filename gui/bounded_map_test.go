package gui

import "testing"

func TestBoundedMapBasicOperations(t *testing.T) {
	m := NewBoundedMap[string, int](3)
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	if v, ok := m.Get("a"); !ok || v != 1 {
		t.Errorf("a: got %d, %v", v, ok)
	}
	if v, ok := m.Get("b"); !ok || v != 2 {
		t.Errorf("b: got %d, %v", v, ok)
	}
	if v, ok := m.Get("c"); !ok || v != 3 {
		t.Errorf("c: got %d, %v", v, ok)
	}
	if m.Len() != 3 {
		t.Errorf("len: got %d", m.Len())
	}
}

func TestBoundedMapEviction(t *testing.T) {
	m := NewBoundedMap[string, int](3)
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	m.Set("d", 4) // evicts "a"

	if _, ok := m.Get("a"); ok {
		t.Error("a should be evicted")
	}
	if v, ok := m.Get("b"); !ok || v != 2 {
		t.Errorf("b: got %d, %v", v, ok)
	}
	if v, ok := m.Get("c"); !ok || v != 3 {
		t.Errorf("c: got %d, %v", v, ok)
	}
	if v, ok := m.Get("d"); !ok || v != 4 {
		t.Errorf("d: got %d, %v", v, ok)
	}
	if m.Len() != 3 {
		t.Errorf("len: got %d", m.Len())
	}
}

func TestBoundedMapUpdateExisting(t *testing.T) {
	m := NewBoundedMap[string, int](3)
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("a", 10)
	m.Set("c", 3)

	if v, ok := m.Get("a"); !ok || v != 10 {
		t.Errorf("a: got %d, %v", v, ok)
	}
	if v, ok := m.Get("b"); !ok || v != 2 {
		t.Errorf("b: got %d, %v", v, ok)
	}
	if v, ok := m.Get("c"); !ok || v != 3 {
		t.Errorf("c: got %d, %v", v, ok)
	}
	if m.Len() != 3 {
		t.Errorf("len: got %d", m.Len())
	}
}

func TestBoundedMapContains(t *testing.T) {
	m := NewBoundedMap[string, int](3)
	m.Set("a", 1)
	if !m.Contains("a") {
		t.Error("should contain a")
	}
	if m.Contains("b") {
		t.Error("should not contain b")
	}
}

func TestBoundedMapDelete(t *testing.T) {
	m := NewBoundedMap[string, int](3)
	m.Set("a", 1)
	m.Set("b", 2)
	m.Delete("a")

	if _, ok := m.Get("a"); ok {
		t.Error("a should be deleted")
	}
	if v, ok := m.Get("b"); !ok || v != 2 {
		t.Errorf("b: got %d, %v", v, ok)
	}
	if m.Len() != 1 {
		t.Errorf("len: got %d", m.Len())
	}
}

func TestBoundedMapClear(t *testing.T) {
	m := NewBoundedMap[string, int](3)
	m.Set("a", 1)
	m.Set("b", 2)
	m.Clear()

	if m.Len() != 0 {
		t.Errorf("len: got %d", m.Len())
	}
	if _, ok := m.Get("a"); ok {
		t.Error("a should be gone")
	}
}

func TestBoundedMapKeys(t *testing.T) {
	m := NewBoundedMap[string, int](5)
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	keys := m.Keys()
	if len(keys) != 3 {
		t.Fatalf("keys len: got %d", len(keys))
	}
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("keys: got %v", keys)
	}
}

func TestBoundedMapMaxSizeOne(t *testing.T) {
	m := NewBoundedMap[string, int](1)
	m.Set("a", 1)
	if v, ok := m.Get("a"); !ok || v != 1 {
		t.Errorf("a: got %d, %v", v, ok)
	}
	if m.Len() != 1 {
		t.Errorf("len: got %d", m.Len())
	}

	m.Set("b", 2) // evicts "a"
	if _, ok := m.Get("a"); ok {
		t.Error("a should be evicted")
	}
	if v, ok := m.Get("b"); !ok || v != 2 {
		t.Errorf("b: got %d, %v", v, ok)
	}
	if m.Len() != 1 {
		t.Errorf("len: got %d", m.Len())
	}
}

func TestBoundedMapMaxSizeZero(t *testing.T) {
	m := NewBoundedMap[string, int](0)
	m.Set("a", 1)
	if m.Len() != 0 {
		t.Errorf("len: got %d", m.Len())
	}
}

func TestBoundedMapDeleteNonexistent(t *testing.T) {
	m := NewBoundedMap[string, int](3)
	m.Set("a", 1)
	m.Delete("nonexistent") // should not panic
	if m.Len() != 1 {
		t.Errorf("len: got %d", m.Len())
	}
	if v, ok := m.Get("a"); !ok || v != 1 {
		t.Errorf("a: got %d, %v", v, ok)
	}
}

func TestBoundedMapKeysStableAfterDeleteAndInsert(t *testing.T) {
	m := NewBoundedMap[string, int](4)
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	m.Delete("b")
	m.Set("d", 4)
	keys := m.Keys()
	expected := []string{"a", "c", "d"}
	if len(keys) != len(expected) {
		t.Fatalf("keys len: got %d, want %d", len(keys), len(expected))
	}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("key[%d]: got %s, want %s", i, k, expected[i])
		}
	}
}
