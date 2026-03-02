package gui

import "testing"

func TestValidateImageExtensionValid(t *testing.T) {
	for _, ext := range []string{
		"photo.png", "photo.jpg", "photo.jpeg",
		"photo.gif", "photo.bmp", "photo.webp",
	} {
		if err := ValidateImageExtension(ext); err != nil {
			t.Errorf("expected valid: %s, got %v", ext, err)
		}
	}
}

func TestValidateImageExtensionInvalid(t *testing.T) {
	for _, ext := range []string{
		"photo.svg", "photo.tiff", "photo.txt", "photo",
	} {
		if err := ValidateImageExtension(ext); err == nil {
			t.Errorf("expected error for: %s", ext)
		}
	}
}

func TestValidateImageExtensionCaseInsensitive(t *testing.T) {
	if err := ValidateImageExtension("photo.PNG"); err != nil {
		t.Errorf("expected case-insensitive match: %v", err)
	}
}

func TestValidateImagePathRejectsTraversal(t *testing.T) {
	if err := ValidateImagePath("../secret/photo.png"); err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestValidateImagePathValid(t *testing.T) {
	if err := ValidateImagePath("/images/photo.png"); err != nil {
		t.Errorf("expected valid path: %v", err)
	}
}

func TestValidateImagePathBadExtension(t *testing.T) {
	if err := ValidateImagePath("/images/photo.svg"); err == nil {
		t.Fatal("expected error for .svg extension")
	}
}
