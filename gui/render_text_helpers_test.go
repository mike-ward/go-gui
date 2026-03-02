package gui

import "testing"

func TestPasswordMaskKeepNewlinesBasic(t *testing.T) {
	got := passwordMaskKeepNewlines("abc")
	if got != "***" {
		t.Errorf("got %q, want %q", got, "***")
	}
}

func TestPasswordMaskKeepNewlinesEmpty(t *testing.T) {
	got := passwordMaskKeepNewlines("")
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestPasswordMaskKeepNewlinesPreservesNewlines(t *testing.T) {
	got := passwordMaskKeepNewlines("ab\ncd\n")
	if got != "**\n**\n" {
		t.Errorf("got %q, want %q", got, "**\\n**\\n")
	}
}

func TestPasswordMaskKeepNewlinesUnicode(t *testing.T) {
	got := passwordMaskKeepNewlines("🔑key")
	if got != "****" {
		t.Errorf("got %q, want %q", got, "****")
	}
}

func TestPasswordMaskSliceBasic(t *testing.T) {
	mask := "***"
	text := "abc"
	got := passwordMaskSlice(mask, text, 0, 2)
	if got != "**" {
		t.Errorf("got %q, want %q", got, "**")
	}
}

func TestPasswordMaskSliceEmptyMask(t *testing.T) {
	got := passwordMaskSlice("", "abc", 0, 2)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestPasswordMaskSliceInvertedRange(t *testing.T) {
	got := passwordMaskSlice("***", "abc", 2, 1)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestHashCombineU64Deterministic(t *testing.T) {
	a := hashCombineU64(42, 100)
	b := hashCombineU64(42, 100)
	if a != b {
		t.Error("same inputs should produce same output")
	}
}

func TestHashCombineU64DifferentInputs(t *testing.T) {
	a := hashCombineU64(42, 100)
	b := hashCombineU64(42, 101)
	if a == b {
		t.Error("different inputs should produce different outputs")
	}
}

func TestHashCombineU64SeedMatters(t *testing.T) {
	a := hashCombineU64(1, 100)
	b := hashCombineU64(2, 100)
	if a == b {
		t.Error("different seeds should produce different outputs")
	}
}
