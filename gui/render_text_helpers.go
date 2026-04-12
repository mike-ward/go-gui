package gui

import "strings"

// maskPassword returns a masked version of text, preserving
// newlines when present.
func maskPassword(text string) string {
	if strings.Contains(text, "\n") {
		return passwordMaskKeepNewlines(text)
	}
	return passwordMask(text)
}

// passwordMaskKeepNewlines replaces each rune with '*' but
// preserves '\n' characters.
func passwordMaskKeepNewlines(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if r == '\n' {
			b.WriteByte('\n')
		} else {
			b.WriteByte('*')
		}
	}
	return b.String()
}
