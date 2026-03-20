package gui

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidImageExtensions lists supported image file extensions.
var ValidImageExtensions = []string{
	".png", ".jpg", ".jpeg",
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
// valid extension. Rejects paths with ".." path components.
// After filepath.Clean, ".." only survives as a leading component
// (e.g. "../foo"), so a prefix check suffices.
func ValidateImagePath(fileName string) error {
	clean := filepath.Clean(fileName)
	if clean == ".." ||
		strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("invalid image path: contains parent reference")
	}
	return ValidateImageExtension(clean)
}
