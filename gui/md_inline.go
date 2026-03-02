package gui

import "strings"

// md_inline.go provides URL safety, image path validation,
// and heading slug generation for the markdown pipeline.

var validImageExts = []string{
	".png", ".jpg", ".jpeg", ".gif", ".svg", ".bmp", ".webp",
}

// isSafeURL checks that a URL does not use dangerous schemes.
func isSafeURL(url string) bool {
	lower := strings.ToLower(strings.TrimSpace(
		decodePercentPrefix(url)))
	if len(lower) == 0 {
		return false
	}
	if strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:") {
		return true
	}
	if !strings.Contains(lower, "://") &&
		!strings.HasPrefix(lower, "javascript:") &&
		!strings.HasPrefix(lower, "data:") &&
		!strings.HasPrefix(lower, "vbscript:") &&
		!strings.HasPrefix(lower, "file:") &&
		!strings.HasPrefix(lower, "blob:") &&
		!strings.HasPrefix(lower, "mhtml:") &&
		!strings.HasPrefix(lower, "ms-help:") &&
		!strings.HasPrefix(lower, "disk:") {
		return true
	}
	return false
}

// decodePercentPrefix decodes leading percent-encoded bytes
// (first 20 chars) for scheme detection.
func decodePercentPrefix(s string) string {
	limit := len(s)
	if limit > 20 {
		limit = 20
	}
	var buf []byte
	i := 0
	for i < limit {
		if s[i] == '%' && i+2 < len(s) {
			hi := hexVal(s[i+1])
			lo := hexVal(s[i+2])
			if hi >= 0 && lo >= 0 {
				buf = append(buf, byte(hi*16+lo))
				i += 3
				continue
			}
		}
		buf = append(buf, s[i])
		i++
	}
	if limit < len(s) {
		buf = append(buf, s[limit:]...)
	}
	return string(buf)
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

// isSafeImagePath validates image paths, blocking traversal
// and absolute paths.
func isSafeImagePath(path string) bool {
	lower := strings.ToLower(
		strings.ReplaceAll(path, "%2e", "."))
	if strings.Contains(lower, "..") {
		return false
	}
	p := strings.TrimSpace(lower)
	if strings.HasPrefix(p, "http://") ||
		strings.HasPrefix(p, "https://") {
		return true
	}
	if !isSafeURL(path) {
		return false
	}
	for _, ext := range validImageExts {
		if strings.HasSuffix(p, ext) {
			return true
		}
	}
	return false
}

// headingSlug converts heading text to a URL-safe slug.
// Lowercase alphanumeric + dashes, no trailing dashes.
func headingSlug(text string) string {
	var buf []byte
	prevDash := false
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch {
		case c >= 'A' && c <= 'Z':
			buf = append(buf, c+32) // lowercase
			prevDash = false
		case (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9'):
			buf = append(buf, c)
			prevDash = false
		default:
			if len(buf) > 0 && !prevDash {
				buf = append(buf, '-')
				prevDash = true
			}
		}
	}
	// Trim trailing dash.
	for len(buf) > 0 && buf[len(buf)-1] == '-' {
		buf = buf[:len(buf)-1]
	}
	return string(buf)
}

// isHTMLTag checks if text between < > looks like an HTML tag.
func isHTMLTag(s string) bool {
	if len(s) == 0 {
		return false
	}
	start := 0
	if s[0] == '/' {
		start = 1
	}
	if start >= len(s) {
		return false
	}
	c := s[start]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
