//go:build !js

package gui

import (
	"fmt"
	"os"
)

// storeDiagramPNG writes PNG bytes to a temp file and returns
// the file path. Native implementation.
func storeDiagramPNG(
	pngBytes []byte, hash int64, prefix string,
) (string, error) {
	f, err := os.CreateTemp("",
		fmt.Sprintf("%s_%d_*.png", prefix, hash))
	if err != nil {
		return "", err
	}
	path := f.Name()
	if _, err := f.Write(pngBytes); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", err
	}
	_ = f.Close()
	return path, nil
}

// removeDiagramPNG deletes a stored diagram file.
func removeDiagramPNG(path string) {
	_ = os.Remove(path)
}
