//go:build !js

package gui

import (
	"os"
	"testing"
)

func TestStoreDiagramPNGWritesFile(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', 1, 2, 3}
	path, err := storeDiagramPNG(data, 42, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(data) {
		t.Errorf("file size %d, want %d", len(got), len(data))
	}
	for i := range data {
		if got[i] != data[i] {
			t.Fatalf("byte %d: got %d, want %d", i, got[i], data[i])
		}
	}
}

func TestRemoveDiagramPNGDeletesFile(t *testing.T) {
	f, err := os.CreateTemp("", "diagram_test_*.png")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	_ = f.Close()

	removeDiagramPNG(path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file should be deleted: %v", err)
	}
}

func TestRemoveDiagramPNGNonexistent(_ *testing.T) {
	// Should not panic on missing file.
	removeDiagramPNG("/tmp/nonexistent_diagram_test_12345.png")
}
