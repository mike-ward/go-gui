package gui

import (
	"strings"
	"testing"
)

func TestSanitizeLatexNormal(t *testing.T) {
	got := sanitizeLatex(`x^2 + y^2 = z^2`)
	if got != `x^2 + y^2 = z^2` {
		t.Errorf("got %q", got)
	}
}

func TestSanitizeLatexBlockedCommands(t *testing.T) {
	tests := []struct {
		input string
		clean string
	}{
		{`\input{file}`, `{file}`},
		{`\write18{cmd}`, `{cmd}`},
		{`\include{f}`, `{f}`},
		{`\def\x{y}`, `\x{y}`},
		{`\immediate\write{x}`, `{x}`},
	}
	for _, tt := range tests {
		got := sanitizeLatex(tt.input)
		if got != tt.clean {
			t.Errorf("sanitizeLatex(%q) = %q, want %q",
				tt.input, got, tt.clean)
		}
	}
}

func TestSanitizeLatexTruncation(t *testing.T) {
	long := strings.Repeat("x", 2001)
	got := sanitizeLatex(long)
	if got != "" {
		t.Error("expected empty string for input exceeding MaxLatexSourceLen")
	}
}

func TestSanitizeLatexAtLimit(t *testing.T) {
	exact := strings.Repeat("x", 2000)
	got := sanitizeLatex(exact)
	if got != exact {
		t.Error("input at exactly MaxLatexSourceLen should pass through")
	}
}

func TestSanitizeLatexControlChars(t *testing.T) {
	// Control chars < 0x20 (except \r, \n, \t) should be stripped.
	input := "a\x01b\x02c"
	got := sanitizeLatex(input)
	if got != "abc" {
		t.Errorf("got %q, want %q", got, "abc")
	}
}

func TestSanitizeLatexWhitespaceNormalization(t *testing.T) {
	// \r\n, \r, \n, \t all become spaces.
	input := "a\r\nb\tc\nd"
	got := sanitizeLatex(input)
	if got != "a b c d" {
		t.Errorf("got %q, want %q", got, "a b c d")
	}
}

func TestSanitizeLatexTrimsWhitespace(t *testing.T) {
	got := sanitizeLatex("  x + y  ")
	if got != "x + y" {
		t.Errorf("got %q, want %q", got, "x + y")
	}
}

func TestSanitizeLatexNestedBlockedCommands(t *testing.T) {
	// Nested: after removing outer \input, \write is exposed.
	input := `\input\write{x}`
	got := sanitizeLatex(input)
	if got != "{x}" {
		t.Errorf("got %q, want %q", got, "{x}")
	}
}

func TestSanitizeLatexEmpty(t *testing.T) {
	got := sanitizeLatex("")
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
