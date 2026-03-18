//go:build js && wasm

package gui

import "encoding/base64"

// storeDiagramPNG base64-encodes PNG bytes and returns a
// data URL. WASM implementation.
func storeDiagramPNG(
	pngBytes []byte, hash int64, prefix string,
) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(pngBytes)
	return "data:image/png;base64," + b64, nil
}

// removeDiagramPNG is a no-op in WASM (data URLs need no
// cleanup).
func removeDiagramPNG(_ string) {}
