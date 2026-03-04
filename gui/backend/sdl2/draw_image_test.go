package sdl2

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveValidatedImagePathAllowedRoots(t *testing.T) {
	root := t.TempDir()
	imgPath := filepath.Join(root, "img.png")
	if err := os.WriteFile(imgPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write image: %v", err)
	}

	b := &Backend{allowedImageRoots: []string{root}}
	got, err := b.resolveValidatedImagePath(imgPath)
	if err != nil {
		t.Fatalf("resolve path: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty resolved path")
	}
}

func TestResolveValidatedImagePathRejectsOutsideRoots(t *testing.T) {
	root := t.TempDir()
	outsideDir := t.TempDir()
	imgPath := filepath.Join(outsideDir, "img.png")
	if err := os.WriteFile(imgPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write image: %v", err)
	}

	b := &Backend{allowedImageRoots: []string{root}}
	if _, err := b.resolveValidatedImagePath(imgPath); err == nil {
		t.Fatal("expected disallowed image path")
	}
}

func TestValidateImageFileRejectsByBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "big.bin")
	if err := os.WriteFile(path, make([]byte, 1024), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer f.Close()

	b := &Backend{maxImageBytes: 32}
	if err := b.validateImageFile(path, f); err == nil {
		t.Fatal("expected image byte limit error")
	}
}

func TestValidateImageFileRejectsByPixels(t *testing.T) {
	path := filepath.Join(t.TempDir(), "img.png")
	if err := writePNG(path, 100, 100); err != nil {
		t.Fatalf("write png: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer f.Close()

	b := &Backend{maxImageBytes: 1 << 20, maxImagePixels: 1000}
	if err := b.validateImageFile(path, f); err == nil {
		t.Fatal("expected image pixel limit error")
	}
}

func writePNG(path string, w, h int) error {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
