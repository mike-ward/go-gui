//go:build !js

package gl

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // Register JPEG decoder.
	_ "image/png"  // Register PNG decoder.
	"log"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	gogl "github.com/go-gl/gl/v3.3-core/gl"

	"github.com/mike-ward/go-gui/gui/backend/internal/imgpath"
)

const (
	defaultMaxImageBytes  = int64(16 * 1024 * 1024)
	defaultMaxImagePixels = int64(40_000_000)
)

// glTexture holds an OpenGL texture handle and dimensions.
type glTexture struct {
	id   uint32
	w, h int32
}

// glTexCacheEntry holds a cached GL texture. Zero id = negative
// cache (failed load).
type glTexCacheEntry struct {
	tex glTexture
}

type glTexCache struct {
	data    map[string]glTexCacheEntry
	order   []string
	maxSize int
}

func newGLTexCache(maxSize int) glTexCache {
	return glTexCache{
		data:    make(map[string]glTexCacheEntry, maxSize),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

func (c *glTexCache) get(path string) (glTexCacheEntry, bool) {
	e, ok := c.data[path]
	if ok {
		c.promote(path)
	}
	return e, ok
}

// promote moves path to the end of the order slice (most
// recently used).
func (c *glTexCache) promote(path string) {
	for i, k := range c.order {
		if k == path {
			c.order = append(
				c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, path)
			return
		}
	}
}

func (c *glTexCache) set(path string, entry glTexCacheEntry) {
	if _, exists := c.data[path]; exists {
		c.data[path] = entry
		return
	}
	if len(c.order) >= c.maxSize {
		evict := c.order[0]
		c.order = c.order[1:]
		if len(c.order) < cap(c.order)/2 {
			c.order = append([]string(nil), c.order...)
		}
		if old, ok := c.data[evict]; ok {
			if old.tex.id != 0 {
				gogl.DeleteTextures(1, &old.tex.id)
			}
			delete(c.data, evict)
		}
	}
	c.order = append(c.order, path)
	c.data[path] = entry
}

func (c *glTexCache) destroyAll() {
	for _, e := range c.data {
		if e.tex.id != 0 {
			gogl.DeleteTextures(1, &e.tex.id)
		}
	}
	c.data = nil
	c.order = nil
}

// createTexture creates an RGBA8 texture from pixel data.
func createTexture(w, h int32, pixels []byte) glTexture {
	var id uint32
	gogl.GenTextures(1, &id)
	gogl.BindTexture(gogl.TEXTURE_2D, id)
	gogl.TexParameteri(gogl.TEXTURE_2D, gogl.TEXTURE_MIN_FILTER,
		gogl.LINEAR)
	gogl.TexParameteri(gogl.TEXTURE_2D, gogl.TEXTURE_MAG_FILTER,
		gogl.LINEAR)
	gogl.TexParameteri(gogl.TEXTURE_2D, gogl.TEXTURE_WRAP_S,
		gogl.CLAMP_TO_EDGE)
	gogl.TexParameteri(gogl.TEXTURE_2D, gogl.TEXTURE_WRAP_T,
		gogl.CLAMP_TO_EDGE)

	var ptr unsafe.Pointer
	if len(pixels) > 0 {
		ptr = unsafe.Pointer(&pixels[0])
	}
	gogl.TexImage2D(gogl.TEXTURE_2D, 0, gogl.RGBA8,
		w, h, 0, gogl.RGBA, gogl.UNSIGNED_BYTE, ptr)
	gogl.BindTexture(gogl.TEXTURE_2D, 0)
	return glTexture{id: id, w: w, h: h}
}

// createEmptyTexture creates a texture with no initial data
// (for FBO render targets).
func createEmptyTexture(w, h int32) glTexture {
	return createTexture(w, h, nil)
}

// --- FBO management ---

func (b *Backend) ensureFilterFBO(w, h int32) bool {
	if b.filterFBO != 0 && b.filterW == w && b.filterH == h {
		return true
	}
	b.destroyFilterFBO()

	gogl.GenFramebuffers(1, &b.filterFBO)

	texA := createEmptyTexture(w, h)
	texB := createEmptyTexture(w, h)
	b.filterTexA = texA.id
	b.filterTexB = texB.id
	b.filterW = w
	b.filterH = h

	// Attach a stencil renderbuffer so ClipContents works
	// inside filter containers.
	gogl.GenRenderbuffers(1, &b.filterStencil)
	gogl.BindRenderbuffer(gogl.RENDERBUFFER, b.filterStencil)
	gogl.RenderbufferStorage(gogl.RENDERBUFFER,
		gogl.STENCIL_INDEX8, w, h)
	gogl.BindRenderbuffer(gogl.RENDERBUFFER, 0)

	b.bindFBO(b.filterTexA)
	gogl.FramebufferRenderbuffer(gogl.FRAMEBUFFER,
		gogl.STENCIL_ATTACHMENT, gogl.RENDERBUFFER,
		b.filterStencil)
	status := gogl.CheckFramebufferStatus(gogl.FRAMEBUFFER)
	b.unbindFBO()
	if status != gogl.FRAMEBUFFER_COMPLETE {
		log.Printf("gl: incomplete filter framebuffer: 0x%x", status)
		b.destroyFilterFBO()
		return false
	}
	return true
}

func (b *Backend) destroyFilterFBO() {
	if b.filterStencil != 0 {
		gogl.DeleteRenderbuffers(1, &b.filterStencil)
		b.filterStencil = 0
	}
	if b.filterFBO != 0 {
		gogl.DeleteFramebuffers(1, &b.filterFBO)
		b.filterFBO = 0
	}
	if b.filterTexA != 0 {
		gogl.DeleteTextures(1, &b.filterTexA)
		b.filterTexA = 0
	}
	if b.filterTexB != 0 {
		gogl.DeleteTextures(1, &b.filterTexB)
		b.filterTexB = 0
	}
}

func (b *Backend) bindFBO(tex uint32) {
	gogl.BindFramebuffer(gogl.FRAMEBUFFER, b.filterFBO)
	gogl.FramebufferTexture2D(gogl.FRAMEBUFFER,
		gogl.COLOR_ATTACHMENT0, gogl.TEXTURE_2D, tex, 0)
}

func (b *Backend) unbindFBO() {
	gogl.BindFramebuffer(gogl.FRAMEBUFFER, 0)
}

// --- Image loading ---

func (b *Backend) loadImageTexture(path string) (glTexCacheEntry, error) {
	if len(b.allowedImageRoots) > 0 {
		if err := validatePathAllowed(path, b.allowedImageRoots); err != nil {
			return glTexCacheEntry{}, err
		}
	}
	pre, err := os.Stat(path)
	if err != nil {
		return glTexCacheEntry{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return glTexCacheEntry{}, err
	}
	defer func() { _ = f.Close() }()
	post, err := f.Stat()
	if err != nil {
		return glTexCacheEntry{}, err
	}
	if !os.SameFile(pre, post) {
		return glTexCacheEntry{}, fmt.Errorf(
			"image path changed during open: %s", path)
	}
	return b.loadTextureFromFile(path, f)
}

func (b *Backend) loadTextureFromFile(path string, f *os.File) (glTexCacheEntry, error) {
	if err := b.validateImageFile(path, f); err != nil {
		return glTexCacheEntry{}, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return glTexCacheEntry{}, err
	}
	src, _, err := image.Decode(f)
	if err != nil {
		return glTexCacheEntry{}, err
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
	tex := createTexture(w, h, nrgba.Pix)
	return glTexCacheEntry{tex: tex}, nil
}

func (b *Backend) validateImageFile(path string, f *os.File) error {
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

func (b *Backend) resolveValidatedImagePath(src string) (string, error) {
	if strings.ContainsRune(src, 0) {
		return "", fmt.Errorf("invalid image path: contains NUL")
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
		if err := validatePathAllowed(resolvedPath, b.allowedImageRoots); err != nil {
			return "", err
		}
	}
	return resolvedPath, nil
}

// Delegating wrappers — shared implementation in imgpath.

func resolvePathWithParentFallback(path string) string {
	return imgpath.ResolveWithParentFallback(path)
}

func validatePathAllowed(path string, allowedRoots []string) error {
	return imgpath.ValidateAllowed(path, allowedRoots)
}

func normalizeAllowedRoots(allowedRoots []string) []string {
	return imgpath.NormalizeRoots(allowedRoots)
}
