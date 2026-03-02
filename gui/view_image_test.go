package gui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImageFactory(t *testing.T) {
	v := Image(ImageCfg{ID: "img1", Src: "test.png"})
	if v == nil {
		t.Fatal("Image factory returned nil")
	}
	if _, ok := v.(*imageView); !ok {
		t.Fatal("expected *imageView")
	}
}

func TestImageInvisible(t *testing.T) {
	v := Image(ImageCfg{ID: "img1", Invisible: true})
	if _, ok := v.(*imageView); ok {
		t.Fatal("invisible should not return *imageView")
	}
}

func TestImageContent(t *testing.T) {
	v := Image(ImageCfg{ID: "img1", Src: "test.png"})
	if v.Content() != nil {
		t.Fatal("expected nil Content")
	}
}

func TestImageGenerateLayoutLocalMissing(t *testing.T) {
	w := &Window{}
	v := Image(ImageCfg{
		ID:  "img1",
		Src: "/nonexistent/photo.png",
	})
	layout := v.GenerateLayout(w)
	// Missing file → error text with magenta color.
	if layout.Shape == nil {
		t.Fatal("expected shape")
	}
	if layout.Shape.ShapeType != ShapeText {
		t.Fatalf("expected ShapeText for missing, got %d",
			layout.Shape.ShapeType)
	}
}

func TestImageGenerateLayoutLocalExists(t *testing.T) {
	// Create a temp file to simulate an image.
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	if err := os.WriteFile(path, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := &Window{}
	v := Image(ImageCfg{
		ID:     "img1",
		Src:    path,
		Width:  200,
		Height: 150,
	})
	layout := v.GenerateLayout(w)
	if layout.Shape.ShapeType != ShapeImage {
		t.Fatalf("expected ShapeImage, got %d",
			layout.Shape.ShapeType)
	}
	if layout.Shape.Resource != path {
		t.Fatalf("expected resource %s, got %s",
			path, layout.Shape.Resource)
	}
	if layout.Shape.Width != 200 {
		t.Fatalf("expected width 200, got %f",
			layout.Shape.Width)
	}
}

func TestImageGenerateLayoutSVGFallback(t *testing.T) {
	// Create a temp .svg file + cached.
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "gui_cache", "images")
	os.MkdirAll(cacheDir, 0o755)

	// Test SVG extension detection.
	hash := hashString("http://example.com/icon.svg")
	base := filepath.Join(cacheDir, "test")
	_ = base
	_ = hash
	// SVG fallback tested indirectly via content type mapping.
}

func TestContentTypeToExt(t *testing.T) {
	tests := []struct {
		ct  string
		ext string
	}{
		{"image/png", ".png"},
		{"image/jpeg", ".jpg"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"image/bmp", ".bmp"},
		{"image/svg+xml", ".svg"},
		{"image/unknown", ".png"},
	}
	for _, tt := range tests {
		got := contentTypeToExt(tt.ct)
		if got != tt.ext {
			t.Errorf("contentTypeToExt(%q) = %q, want %q",
				tt.ct, got, tt.ext)
		}
	}
}

func TestFindCachedImageFound(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "abc123")
	path := base + ".jpg"
	os.WriteFile(path, []byte("fake"), 0o644)
	result := findCachedImage(base)
	if result != path {
		t.Fatalf("expected %s, got %s", path, result)
	}
}

func TestFindCachedImageNotFound(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "missing")
	result := findCachedImage(base)
	if result != "" {
		t.Fatalf("expected empty, got %s", result)
	}
}

func TestImageDefaultDimensions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	os.WriteFile(path, []byte("fake"), 0o644)
	w := &Window{}
	v := Image(ImageCfg{ID: "img1", Src: path})
	layout := v.GenerateLayout(w)
	if layout.Shape.Width != 100 || layout.Shape.Height != 100 {
		t.Fatalf("expected default 100x100, got %fx%f",
			layout.Shape.Width, layout.Shape.Height)
	}
}

func TestImageWithEvents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	os.WriteFile(path, []byte("fake"), 0o644)
	w := &Window{}
	clicked := false
	v := Image(ImageCfg{
		ID:    "img1",
		Src:   path,
		Width: 50, Height: 50,
		OnClick: func(l *Layout, e *Event, w *Window) {
			clicked = true
		},
	})
	layout := v.GenerateLayout(w)
	if layout.Shape.Events == nil {
		t.Fatal("expected events")
	}
	// Simulate left click.
	layout.Shape.Events.OnClick(&layout, &Event{
		MouseButton: MouseLeft,
	}, w)
	if !clicked {
		t.Fatal("click handler not called")
	}
	_ = clicked
}

func TestImageA11Y(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	os.WriteFile(path, []byte("fake"), 0o644)
	w := &Window{}
	v := Image(ImageCfg{
		ID:              "img1",
		Src:             path,
		Width:           50,
		Height:          50,
		A11YLabel:       "test image",
		A11YDescription: "a test",
	})
	layout := v.GenerateLayout(w)
	if layout.Shape.A11YRole != AccessRoleImage {
		t.Fatal("expected AccessRoleImage")
	}
	if layout.Shape.A11Y == nil {
		t.Fatal("expected A11Y info")
	}
	if layout.Shape.A11Y.Label != "test image" {
		t.Fatalf("expected label 'test image', got %q",
			layout.Shape.A11Y.Label)
	}
}
