package gui

import (
	"math"
	"testing"

	"github.com/mike-ward/go-glyph"
)

func TestIsSafeURL(t *testing.T) {
	safe := []string{
		"https://example.com",
		"http://example.com",
		"mailto:user@example.com",
		"#anchor",
		"/relative/path",
		"relative/path",
	}
	for _, u := range safe {
		if !isSafeURL(u) {
			t.Errorf("expected safe: %q", u)
		}
	}
}

func TestIsSafeURLBlocked(t *testing.T) {
	blocked := []string{
		"javascript:alert(1)",
		"data:text/html,<h1>",
		"vbscript:msgbox",
		"file:///etc/passwd",
		"blob:http://example.com/x",
	}
	for _, u := range blocked {
		if isSafeURL(u) {
			t.Errorf("expected blocked: %q", u)
		}
	}
}

func TestIsSafeURLEdgeCases(t *testing.T) {
	// Empty.
	if isSafeURL("") {
		t.Error("empty should not be safe")
	}
	// Whitespace prefix.
	if isSafeURL("  javascript:alert(1)") {
		t.Error("whitespace-prefixed javascript: should be blocked")
	}
	// Case insensitive.
	if isSafeURL("JAVASCRIPT:alert(1)") {
		t.Error("JAVASCRIPT: should be blocked")
	}
	// Percent-encoded.
	if isSafeURL("java%73cript:alert(1)") {
		t.Error("percent-encoded javascript: should be blocked")
	}
}

func TestIsSafeImagePath(t *testing.T) {
	safe := []string{
		"https://example.com/image.png",
		"image.jpg",
		"path/to/image.gif",
	}
	for _, p := range safe {
		if !isSafeImagePath(p) {
			t.Errorf("expected safe image: %q", p)
		}
	}

	blocked := []string{
		"../../../etc/passwd.png",
		"image%2e%2e/etc/passwd.png",
	}
	for _, p := range blocked {
		if isSafeImagePath(p) {
			t.Errorf("expected blocked image: %q", p)
		}
	}
}

func TestRtfHitTestLogic(t *testing.T) {
	item := glyph.Item{
		X: 10, Y: 20, Width: 50,
		Ascent: 12, Descent: 4,
	}
	r := rtfRunRect(item)
	if r.X != 10 || r.Width != 50 {
		t.Fatalf("rect X/W: got %v/%v", r.X, r.Width)
	}
	// Y = run.Y - Ascent = 20 - 12 = 8.
	if r.Y != 8 {
		t.Fatalf("rect Y: expected 8, got %v", r.Y)
	}
	// Height = Ascent + Descent = 16.
	if r.Height != 16 {
		t.Fatalf("rect Height: expected 16, got %v", r.Height)
	}

	// Point inside.
	if !rtfHitTest(item, 30, 15, nil) {
		t.Error("expected hit at (30,15)")
	}
	// Point outside.
	if rtfHitTest(item, 5, 5, nil) {
		t.Error("expected miss at (5,5)")
	}
}

func TestRtfAffineInverse(t *testing.T) {
	// Identity matrix.
	id := glyph.AffineTransform{
		XX: 1, XY: 0, YX: 0, YY: 1, X0: 0, Y0: 0,
	}
	inv, ok := rtfAffineInverse(id)
	if !ok {
		t.Fatal("identity should be invertible")
	}
	if math.Abs(float64(inv.XX-1)) > 0.001 ||
		math.Abs(float64(inv.YY-1)) > 0.001 {
		t.Fatalf("identity inverse: XX=%v YY=%v",
			inv.XX, inv.YY)
	}

	// Singular matrix.
	singular := glyph.AffineTransform{
		XX: 0, XY: 0, YX: 0, YY: 0,
	}
	_, ok = rtfAffineInverse(singular)
	if ok {
		t.Fatal("singular matrix should not be invertible")
	}
}

func TestHeadingSlug(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Hello World", "hello-world"},
		{"  Spaces  ", "spaces"},
		{"foo---bar", "foo-bar"},
		{"UPPER", "upper"},
		{"a1b2c3", "a1b2c3"},
		{"", ""},
	}
	for _, tc := range tests {
		got := headingSlug(tc.input)
		if got != tc.want {
			t.Errorf("headingSlug(%q): got %q, want %q",
				tc.input, got, tc.want)
		}
	}
}
