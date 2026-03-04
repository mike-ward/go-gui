package markdown

import "strings"

// inline.go provides URL safety, image path validation,
// and heading slug generation for the markdown pipeline.

var validImageExts = []string{
	".png", ".jpg", ".jpeg", ".gif", ".svg", ".bmp", ".webp",
}

// IsSafeURL checks that a URL does not use dangerous schemes.
func IsSafeURL(url string) bool {
	trimmed := strings.TrimSpace(url)
	if len(trimmed) == 0 {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(
		decodePercentPrefix(trimmed)))
	if strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:") {
		return true
	}

	// Safe local references and relative paths.
	if strings.HasPrefix(lower, "#") ||
		strings.HasPrefix(lower, "/") ||
		strings.HasPrefix(lower, "./") ||
		strings.HasPrefix(lower, "../") ||
		strings.HasPrefix(lower, "?") {
		return true
	}

	// Any explicit URI scheme outside the allowlist is blocked.
	if hasURIScheme(lower) {
		return false
	}

	// Plain relative paths without an explicit scheme are safe.
	return true
}

// decodePercentPrefix decodes leading percent-encoded bytes
// (first 20 chars) for scheme detection.
func decodePercentPrefix(s string) string {
	limit := len(s)
	if limit > 20 {
		limit = 20
	}
	buf := make([]byte, 0, len(s))
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

func hasURIScheme(s string) bool {
	colon := strings.IndexByte(s, ':')
	if colon <= 0 {
		return false
	}
	for i := 0; i < colon; i++ {
		c := s[i]
		if c == '/' || c == '?' || c == '#' {
			return false
		}
		if i == 0 {
			if c < 'a' || c > 'z' {
				return false
			}
			continue
		}
		if !((c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '+' || c == '-' || c == '.') {
			return false
		}
	}
	return true
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
	if !IsSafeURL(path) {
		return false
	}
	for _, ext := range validImageExts {
		if strings.HasSuffix(p, ext) {
			return true
		}
	}
	return false
}

// HeadingSlug converts heading text to a URL-safe slug.
// Lowercase alphanumeric + dashes, no trailing dashes.
func HeadingSlug(text string) string {
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
