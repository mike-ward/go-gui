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

// passwordMaskSlice masks a byte-range substring using a
// pre-computed mask string. Operates on rune indices derived
// from byte offsets.
func passwordMaskSlice(mask, text string, startByte, endByte int) string {
	if len(mask) == 0 || endByte <= startByte {
		return ""
	}
	start := byteToRuneIndex(text, startByte)
	end := byteToRuneIndex(text, endByte)
	if start < 0 {
		start = 0
	}
	if start > len(mask) {
		start = len(mask)
	}
	if end < start {
		end = start
	}
	if end > len(mask) {
		end = len(mask)
	}
	return mask[start:end]
}

// hashCombineU64 combines two uint64 values using FNV-style
// multiplication.
func hashCombineU64(seed, value uint64) uint64 {
	return (seed ^ value) * 1099511628211
}
