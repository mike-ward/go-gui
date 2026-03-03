package gui

// markdown_mermaid.go implements diagram caching and async
// mermaid diagram fetching via the Kroki API.

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mike-ward/go-gui/gui/markdown"
)

const (
	maxConcurrentDiagramFetches = 8
	diagramFetchTimeout         = 30 * time.Second
)

// DiagramState represents the loading state of a diagram.
type DiagramState uint8

const (
	DiagramLoading DiagramState = iota
	DiagramReady
	DiagramError
)

// DiagramCacheEntry stores cached diagram data.
type DiagramCacheEntry struct {
	State     DiagramState
	PNGPath   string  // temp file path
	Error     string
	Width     float32
	Height    float32
	DPI       float32 // DPI used for rendering
	RequestID uint64
}

// BoundedDiagramCache is a FIFO cache for diagram entries.
// Custom cache (not BoundedMap) — needs png_path cleanup
// on evict/overwrite.
type BoundedDiagramCache struct {
	data    map[int64]DiagramCacheEntry
	order   []int64
	maxSize int
}

// NewBoundedDiagramCache creates a diagram cache with the
// given capacity.
func NewBoundedDiagramCache(maxSize int) *BoundedDiagramCache {
	if maxSize < 1 {
		maxSize = 50
	}
	return &BoundedDiagramCache{
		data:    make(map[int64]DiagramCacheEntry),
		maxSize: maxSize,
	}
}

// Get returns a cached diagram entry.
func (c *BoundedDiagramCache) Get(
	key int64,
) (DiagramCacheEntry, bool) {
	e, ok := c.data[key]
	return e, ok
}

// Set adds a diagram entry. Evicts oldest if at capacity.
// Cleans up old temp files on overwrite or eviction.
func (c *BoundedDiagramCache) Set(
	key int64, value DiagramCacheEntry,
) {
	if c.maxSize < 1 {
		return
	}
	// Clean up existing entry's temp file on overwrite.
	if existing, ok := c.data[key]; ok {
		if existing.PNGPath != "" &&
			existing.PNGPath != value.PNGPath {
			os.Remove(existing.PNGPath)
		}
	}
	// If new key, evict oldest if at capacity.
	if _, ok := c.data[key]; !ok {
		if len(c.data) >= c.maxSize && len(c.order) > 0 {
			oldest := c.order[0]
			if oe, ok := c.data[oldest]; ok {
				if oe.PNGPath != "" {
					os.Remove(oe.PNGPath)
				}
			}
			delete(c.data, oldest)
			c.order = c.order[1:]
		}
		c.order = append(c.order, key)
	}
	c.data[key] = value
}

// LoadingCount returns entries in loading state.
func (c *BoundedDiagramCache) LoadingCount() int {
	n := 0
	for _, e := range c.data {
		if e.State == DiagramLoading {
			n++
		}
	}
	return n
}

// Len returns the number of entries.
func (c *BoundedDiagramCache) Len() int {
	return len(c.data)
}

// Clear removes all entries and deletes temp files.
func (c *BoundedDiagramCache) Clear() {
	for _, e := range c.data {
		if e.PNGPath != "" {
			os.Remove(e.PNGPath)
		}
	}
	clear(c.data)
	c.order = c.order[:0]
}

// diagramCacheShouldApplyResult checks if a result should
// be applied (entry still loading with same request ID).
func diagramCacheShouldApplyResult(
	cache *BoundedDiagramCache, hash int64, requestID uint64,
) bool {
	if cache == nil {
		return false
	}
	e, ok := cache.Get(hash)
	if !ok {
		return false
	}
	return e.State == DiagramLoading &&
		e.RequestID == requestID
}

// fetchMermaidAsync fetches a mermaid diagram from Kroki API
// in a background goroutine.
//
// PRIVACY NOTE: Mermaid source is sent to external
// third-party API (kroki.io) for rendering.
func fetchMermaidAsync(
	w *Window, source string, hash int64,
	requestID uint64, bgR, bgG, bgB uint8,
) {
	go func() {
		if len(source) > markdown.MaxMermaidSourceLen {
			w.QueueCommand(func(w *Window) {
				if !diagramCacheShouldApplyResult(
					w.viewState.diagramCache,
					hash, requestID) {
					return
				}
				w.viewState.diagramCache.Set(hash,
					DiagramCacheEntry{
						State:     DiagramError,
						Error:     "Mermaid source too large",
						RequestID: requestID,
					})
				w.UpdateWindow()
			})
			return
		}

		body, err := mermaidHTTPFetch(source)
		if err != nil {
			errMsg := err.Error()
			w.QueueCommand(func(w *Window) {
				if !diagramCacheShouldApplyResult(
					w.viewState.diagramCache,
					hash, requestID) {
					return
				}
				w.viewState.diagramCache.Set(hash,
					DiagramCacheEntry{
						State:     DiagramError,
						Error:     errMsg,
						RequestID: requestID,
					})
				w.UpdateWindow()
			})
			return
		}

		// Decode PNG.
		img, err := png.Decode(bytes.NewReader(body))
		if err != nil {
			errMsg := err.Error()
			w.QueueCommand(func(w *Window) {
				if !diagramCacheShouldApplyResult(
					w.viewState.diagramCache,
					hash, requestID) {
					return
				}
				w.viewState.diagramCache.Set(hash,
					DiagramCacheEntry{
						State:     DiagramError,
						Error:     "PNG decode: " + errMsg,
						RequestID: requestID,
					})
				w.UpdateWindow()
			})
			return
		}

		bounds := img.Bounds()
		finalW := float32(bounds.Dx())
		finalH := float32(bounds.Dy())

		// Write to temp file.
		tmpFile, err := os.CreateTemp("",
			fmt.Sprintf("mermaid_%d_*.png", hash))
		if err != nil {
			return
		}
		tmpPath := tmpFile.Name()
		if err := png.Encode(tmpFile, img); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return
		}
		tmpFile.Close()

		w.QueueCommand(func(w *Window) {
			if !diagramCacheShouldApplyResult(
				w.viewState.diagramCache,
				hash, requestID) {
				os.Remove(tmpPath)
				return
			}
			w.viewState.diagramCache.Set(hash,
				DiagramCacheEntry{
					State:     DiagramReady,
					PNGPath:   tmpPath,
					Width:     finalW,
					Height:    finalH,
					RequestID: requestID,
				})
			w.UpdateWindow()
		})
	}()
	_ = bgR
	_ = bgG
	_ = bgB
}

func mermaidHTTPFetch(source string) ([]byte, error) {
	escaped := strings.ReplaceAll(source, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	escaped = strings.ReplaceAll(escaped, "\r", `\r`)
	escaped = strings.ReplaceAll(escaped, "\t", `\t`)

	jsonData := `{"diagram_source": "` + escaped + `"}`

	client := &http.Client{Timeout: diagramFetchTimeout}
	resp, err := client.Post(
		"https://kroki.io/mermaid/png",
		"application/json",
		strings.NewReader(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return nil, fmt.Errorf(
			"HTTP %d: %s", resp.StatusCode, preview)
	}
	if len(body) > 10*1024*1024 {
		return nil, fmt.Errorf(
			"response too large (%dMB)", len(body)/1024/1024)
	}
	return body, nil
}
