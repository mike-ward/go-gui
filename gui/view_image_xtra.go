package gui

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidImageExtensions lists supported image file extensions.
var ValidImageExtensions = []string{
	".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp",
}

// ValidateImageExtension checks that the file has a supported
// image extension.
func ValidateImageExtension(fileName string) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, valid := range ValidImageExtensions {
		if ext == valid {
			return nil
		}
	}
	return fmt.Errorf("unsupported image format: %s", ext)
}

// ValidateImagePath checks that the file path is safe and has a
// valid extension. Rejects paths containing "..".
func ValidateImagePath(fileName string) error {
	if strings.Contains(fileName, "..") {
		return fmt.Errorf("invalid image path: contains ..")
	}
	return ValidateImageExtension(fileName)
}
