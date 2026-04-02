package gui

import "strings"

// render_text_helpers.go — text rendering utilities ported from
// V's render_text.v.

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
