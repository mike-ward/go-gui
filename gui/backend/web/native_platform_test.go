//go:build js && wasm

package web

import (
	"syscall/js"
	"testing"
)

// --- dotExtensions ---

func TestDotExtensions(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"png"}, ".png"},
		{[]string{"png", "jpg", "gif"}, ".png,.jpg,.gif"},
		{[]string{"tar+gz"}, ".tar+gz"},
	}
	for _, tt := range tests {
		got := dotExtensions(tt.in)
		if got != tt.want {
			t.Errorf("dotExtensions(%v) = %q, want %q",
				tt.in, got, tt.want)
		}
	}
}

// --- hasPrefixFold ---

func TestHasPrefixFold(t *testing.T) {
	tests := []struct {
		s, prefix string
		want      bool
	}{
		{"http://example.com", "http://", true},
		{"HTTP://EXAMPLE.COM", "http://", true},
		{"https://x", "http://", false},
		{"mailto:a@b", "mailto:", true},
		{"ftp://x", "http://", false},
		{"h", "http://", false},
		{"", "http://", false},
		{"http://x", "HTTP://", true},
	}
	for _, tt := range tests {
		got := hasPrefixFold(tt.s, tt.prefix)
		if got != tt.want {
			t.Errorf("hasPrefixFold(%q, %q) = %v, want %v",
				tt.s, tt.prefix, got, tt.want)
		}
	}
}

// --- jsObject ---

func TestJsObject(t *testing.T) {
	obj := jsObject()
	if obj.Type() != js.TypeObject {
		t.Fatalf("jsObject() type = %v, want Object", obj.Type())
	}
}
