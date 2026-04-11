package gui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSetTitleNilFnNoPanic(t *testing.T) {
	w := &Window{}
	w.SetTitle("hello")
	if w.Config.Title != "hello" {
		t.Errorf("Config.Title = %q, want hello", w.Config.Title)
	}
}

func TestSetTitleCallsBackendFn(t *testing.T) {
	w := &Window{}
	var got string
	w.SetTitleFn(func(s string) { got = s })
	w.SetTitle("world")
	if got != "world" {
		t.Errorf("backend fn got %q, want world", got)
	}
	if w.Config.Title != "world" {
		t.Errorf("Config.Title = %q, want world", w.Config.Title)
	}
}

func TestSetTitleEmpty(t *testing.T) {
	w := &Window{}
	var called bool
	w.SetTitleFn(func(_ string) { called = true })
	w.SetTitle("")
	if !called {
		t.Error("backend fn not called on empty title")
	}
	if w.Config.Title != "" {
		t.Errorf("Config.Title = %q, want empty", w.Config.Title)
	}
}

func TestSanitizeTitleCaps(t *testing.T) {
	in := strings.Repeat("a", maxTitleBytes+500)
	out := sanitizeTitle(in)
	if len(out) > maxTitleBytes {
		t.Errorf("len(out) = %d, want <= %d", len(out), maxTitleBytes)
	}
}

func TestSanitizeTitleTruncatesOnUTF8Boundary(t *testing.T) {
	// Build a string whose raw cut at maxTitleBytes lands inside
	// a 4-byte rune. Prefix with (maxTitleBytes-2) ASCII, then a
	// 4-byte rune (U+1F600 = f0 9f 98 80). Cut index falls on the
	// 3rd byte of the rune; sanitize must back up.
	prefix := strings.Repeat("a", maxTitleBytes-2)
	in := prefix + "\U0001F600" + "tail"
	out := sanitizeTitle(in)
	if !utf8.ValidString(out) {
		t.Errorf("result not valid UTF-8: %q", out)
	}
	if len(out) > maxTitleBytes {
		t.Errorf("len(out) = %d, want <= %d", len(out), maxTitleBytes)
	}
	// The straddling rune must have been dropped entirely.
	if strings.ContainsRune(out, '\U0001F600') {
		t.Error("straddling rune leaked through")
	}
}

func TestSanitizeTitleStripsNUL(t *testing.T) {
	out := sanitizeTitle("foo\x00bar\x00")
	if out != "foobar" {
		t.Errorf("got %q, want foobar", out)
	}
}

func TestSanitizeTitleAllNUL(t *testing.T) {
	out := sanitizeTitle("\x00\x00\x00")
	if out != "" {
		t.Errorf("got %q, want empty", out)
	}
}

func TestSanitizeTitleEmpty(t *testing.T) {
	if out := sanitizeTitle(""); out != "" {
		t.Errorf("got %q, want empty", out)
	}
}

func TestSanitizeTitleDegenerateUTF8(t *testing.T) {
	// All continuation bytes (0x80) past the cut — loop walks back
	// to 0 without panic or infinite spin.
	in := strings.Repeat("\x80", maxTitleBytes+10)
	out := sanitizeTitle(in)
	if len(out) > maxTitleBytes {
		t.Errorf("len(out) = %d, want <= %d", len(out), maxTitleBytes)
	}
}

func TestSanitizeTitleNoAllocFastPath(t *testing.T) {
	in := "short clean title"
	allocs := testing.AllocsPerRun(100, func() {
		_ = sanitizeTitle(in)
	})
	if allocs != 0 {
		t.Errorf("allocs = %v, want 0 on NUL-free short input", allocs)
	}
}
