package tempfont

import (
	"os"
	"testing"
)

func TestWriteCreatesUniqueFile(t *testing.T) {
	path1, err := Write("go-gui-test", []byte("one"))
	if err != nil {
		t.Fatalf("Write(path1): %v", err)
	}
	defer func() { _ = os.Remove(path1) }()

	path2, err := Write("go-gui-test", []byte("two"))
	if err != nil {
		t.Fatalf("Write(path2): %v", err)
	}
	defer func() { _ = os.Remove(path2) }()

	if path1 == path2 {
		t.Fatal("expected unique temp file names")
	}

	got1, err := os.ReadFile(path1)
	if err != nil {
		t.Fatalf("ReadFile(path1): %v", err)
	}
	if string(got1) != "one" {
		t.Fatalf("path1 contents = %q", got1)
	}

	got2, err := os.ReadFile(path2)
	if err != nil {
		t.Fatalf("ReadFile(path2): %v", err)
	}
	if string(got2) != "two" {
		t.Fatalf("path2 contents = %q", got2)
	}
}
