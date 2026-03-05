//go:build darwin

package metal

/*
#include "metal_darwin.h"
*/
import "C"

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

const (
	defaultMaxImageBytes  = int64(16 * 1024 * 1024)
	defaultMaxImagePixels = int64(40_000_000)
)

// metalTexture holds a Metal texture ID and dimensions.
type metalTexture struct {
	id   int32
	w, h int32
}

type metalTexCacheEntry struct {
	tex metalTexture
}

type metalTexCache struct {
	data    map[string]metalTexCacheEntry
	order   []string
	maxSize int
}

func newMetalTexCache(maxSize int) metalTexCache {
	return metalTexCache{
		data:    make(map[string]metalTexCacheEntry, maxSize),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

func (c *metalTexCache) get(
	path string) (metalTexCacheEntry, bool) {
	e, ok := c.data[path]
	if ok {
		c.promote(path)
	}
	return e, ok
}

// promote moves path to the end of the order slice (most
// recently used).
func (c *metalTexCache) promote(path string) {
	for i, k := range c.order {
		if k == path {
			c.order = append(
				c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, path)
			return
		}
	}
}

func (c *metalTexCache) set(
	path string, entry metalTexCacheEntry) {
	if _, exists := c.data[path]; exists {
		c.data[path] = entry
		return
	}
	if len(c.order) >= c.maxSize {
		evict := c.order[0]
		c.order = c.order[1:]
		if old, ok := c.data[evict]; ok {
			if old.tex.id != 0 {
				C.metalDeleteTexture(C.int(old.tex.id))
			}
			delete(c.data, evict)
		}
	}
	c.order = append(c.order, path)
	c.data[path] = entry
}

func (c *metalTexCache) destroyAll() {
	for _, e := range c.data {
		if e.tex.id != 0 {
			C.metalDeleteTexture(C.int(e.tex.id))
		}
	}
	c.data = nil
	c.order = nil
}

func createMetalTexture(w, h int32,
	pixels []byte) metalTexture {
	hasData := C.int(0)
	var ptr unsafe.Pointer
	if len(pixels) > 0 {
		hasData = C.int(1)
		ptr = unsafe.Pointer(&pixels[0])
	}
	id := C.metalCreateTexture(
		C.int(w), C.int(h), ptr, hasData)
	return metalTexture{id: int32(id), w: w, h: h}
}

// --- Image loading ---

func (b *Backend) loadImageTexture(
	path string) (metalTexCacheEntry, error) {
	if len(b.allowedImageRoots) > 0 {
		if err := validatePathAllowed(path,
			b.allowedImageRoots); err != nil {
			return metalTexCacheEntry{}, err
		}
	}
	pre, err := os.Stat(path)
	if err != nil {
		return metalTexCacheEntry{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return metalTexCacheEntry{}, err
	}
	defer func() {
		_ = f.Close()
	}()
	post, err := f.Stat()
	if err != nil {
		return metalTexCacheEntry{}, err
	}
	if !os.SameFile(pre, post) {
		return metalTexCacheEntry{}, fmt.Errorf(
			"image path changed during open: %s", path)
	}
	return b.loadTextureFromFile(path, f)
}

func (b *Backend) loadTextureFromFile(
	path string, f *os.File) (metalTexCacheEntry, error) {
	if err := b.validateImageFile(path, f); err != nil {
		return metalTexCacheEntry{}, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return metalTexCacheEntry{}, err
	}
	src, _, err := image.Decode(f)
	if err != nil {
		return metalTexCacheEntry{}, err
	}

	var nrgba *image.NRGBA
	if existing, ok := src.(*image.NRGBA); ok {
		nrgba = existing
	} else {
		bounds := src.Bounds()
		nrgba = image.NewNRGBA(bounds)
		draw.Draw(nrgba, bounds, src, bounds.Min, draw.Src)
	}
	bounds := nrgba.Bounds()
	w := int32(bounds.Dx())
	h := int32(bounds.Dy())
	tex := createMetalTexture(w, h, nrgba.Pix)
	return metalTexCacheEntry{tex: tex}, nil
}

func (b *Backend) validateImageFile(
	path string, f *os.File) error {
	maxBytes := b.maxImageBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxImageBytes
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() > maxBytes {
		return fmt.Errorf("image file too large: %s", path)
	}
	maxPixels := b.maxImagePixels
	if maxPixels <= 0 {
		maxPixels = defaultMaxImagePixels
	}
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return fmt.Errorf("invalid image dimensions: %s", path)
	}
	if int64(cfg.Width)*int64(cfg.Height) > maxPixels {
		return fmt.Errorf("image dimensions too large: %s", path)
	}
	return nil
}

func (b *Backend) resolveValidatedImagePath(
	src string) (string, error) {
	if strings.ContainsRune(src, 0) {
		return "", fmt.Errorf(
			"invalid image path: contains NUL")
	}
	cleanPath := filepath.Clean(src)
	if cleanPath == "." || cleanPath == "" {
		return "", fmt.Errorf("invalid image path")
	}
	pathAbs, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid image path: %w", err)
	}
	resolvedPath := resolvePathWithParentFallback(pathAbs)
	if len(b.allowedImageRoots) > 0 {
		if err := validatePathAllowed(resolvedPath,
			b.allowedImageRoots); err != nil {
			return "", err
		}
	}
	return resolvedPath, nil
}

func resolvePathWithParentFallback(path string) string {
	if p, err := filepath.EvalSymlinks(path); err == nil {
		return p
	}
	dir := filepath.Dir(path)
	if d, err := filepath.EvalSymlinks(dir); err == nil {
		return filepath.Join(d, filepath.Base(path))
	}
	return path
}

func validatePathAllowed(
	path string, allowedRoots []string) error {
	for i := range allowedRoots {
		root := strings.TrimSpace(allowedRoots[i])
		if root == "" {
			continue
		}
		if pathWithinRoot(path, root) {
			return nil
		}
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		if pathWithinRoot(path,
			resolvePathWithParentFallback(rootAbs)) {
			return nil
		}
	}
	return fmt.Errorf("image path not allowed: %s", path)
}

func normalizeAllowedRoots(
	allowedRoots []string) []string {
	if len(allowedRoots) == 0 {
		return nil
	}
	roots := make([]string, 0, len(allowedRoots))
	for i := range allowedRoots {
		root := strings.TrimSpace(allowedRoots[i])
		if root == "" {
			continue
		}
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		roots = append(roots,
			resolvePathWithParentFallback(rootAbs))
	}
	return roots
}

func pathWithinRoot(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." &&
		!strings.HasPrefix(rel,
			".."+string(filepath.Separator)))
}
