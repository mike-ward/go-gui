package imgpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWithinRoot(t *testing.T) {
	tests := []struct {
		path, root string
		want       bool
	}{
		{"/a/b/c", "/a/b", true},
		{"/a/b", "/a/b", true},
		{"/a/b/../c", "/a", true},
		{"/a", "/a/b", false},
		{"/x/y", "/a/b", false},
	}
	for _, tt := range tests {
		got := WithinRoot(tt.path, tt.root)
		if got != tt.want {
			t.Errorf("WithinRoot(%q, %q) = %v, want %v",
				tt.path, tt.root, got, tt.want)
		}
	}
}

func TestValidateAllowed(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "img.png")

	if err := ValidateAllowed(inside, []string{root}); err != nil {
		t.Errorf("expected allowed: %v", err)
	}

	outside := filepath.Join(t.TempDir(), "img.png")
	if err := ValidateAllowed(outside, []string{root}); err == nil {
		t.Error("expected disallowed for path outside root")
	}
}

func TestValidateAllowedSkipsEmpty(t *testing.T) {
	if err := ValidateAllowed("/any/path",
		[]string{"", "  ", ""}); err == nil {
		t.Error("expected error with only empty roots")
	}
}

func TestNormalizeRootsEmpty(t *testing.T) {
	if got := NormalizeRoots(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
	if got := NormalizeRoots([]string{"", "  "}); len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestNormalizeRootsResolves(t *testing.T) {
	root := t.TempDir()
	got := NormalizeRoots([]string{root})
	if len(got) != 1 {
		t.Fatalf("expected 1 root, got %d", len(got))
	}
	if !filepath.IsAbs(got[0]) {
		t.Errorf("expected absolute path, got %q", got[0])
	}
}

func TestResolveWithParentFallback(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ResolveWithParentFallback(f)
	if got == "" {
		t.Fatal("expected non-empty path")
	}
	// Non-existent file in existing dir should resolve parent.
	missing := filepath.Join(dir, "no_such_file.png")
	got = ResolveWithParentFallback(missing)
	if got == "" {
		t.Fatal("expected non-empty path for missing file")
	}
}

func TestSymlinkTraversal(t *testing.T) {
	allowed := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.png")
	if err := os.WriteFile(secret, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(allowed, "link.png")
	if err := os.Symlink(secret, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}
	resolved := ResolveWithParentFallback(link)
	if err := ValidateAllowed(resolved, []string{allowed}); err == nil {
		t.Error("symlink escaping root should be rejected")
	}
}

func TestNULInPath(_ *testing.T) {
	// NUL bytes are not handled by imgpath itself (backends
	// check before calling), but WithinRoot should not panic.
	got := WithinRoot("/a/b\x00/c", "/a")
	_ = got // just ensure no panic
}

func TestWhitespaceRoots(t *testing.T) {
	root := t.TempDir()
	padded := "  " + root + "  "
	inside := filepath.Join(root, "img.png")
	if err := ValidateAllowed(inside, []string{padded}); err != nil {
		t.Errorf("padded root should still match: %v", err)
	}
}
