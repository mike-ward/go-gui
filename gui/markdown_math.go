package gui

// markdown_math.go implements LaTeX math image fetching
// via the codecogs API.

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

// mathCacheHash computes a cache key for a math expression.
func mathCacheHash(mathID string) int64 {
	h := mathHash(mathID)
	return int64((h << 32) | uint64(len(mathID)))
}

// sanitizeLatex strips dangerous TeX commands that could
// enable shell escape or file access on the remote renderer.
func sanitizeLatex(s string) string {
	if len(s) > maxLatexSourceLen {
		return ""
	}
	blocked := []string{
		`\write18`, `\input`, `\include`,
		`\openin`, `\openout`, `\read`, `\write`,
		`\csname`, `\immediate`, `\catcode`,
		`\special`, `\outer`, `\def`, `\edef`,
		`\gdef`, `\xdef`, `\let`, `\futurelet`,
		`\aliasfont`, `\batchmode`, `\copy`,
		`\count`, `\countdef`, `\dimen`, `\dimendef`,
		`\errorstopmode`, `\font`, `\fontdimen`,
		`\halign`, `\hrule`, `\hyphenation`,
		`\if`, `\ifcase`, `\ifcat`, `\ifdim`,
		`\ifeof`, `\iffalse`, `\ifhbox`, `\ifhmode`,
		`\ifinner`, `\ifmmode`, `\ifnum`, `\ifodd`,
		`\iftrue`, `\ifvbox`, `\ifvmode`, `\ifvoid`,
		`\ifx`, `\jobname`, `\kern`, `\long`,
		`\mag`, `\mark`, `\meaning`, `\messages`,
		`\newcount`, `\newdimen`, `\newif`,
		`\newread`, `\newskip`, `\newwrite`,
		`\noexpand`, `\nonstopmode`, `\output`,
		`\pausing`, `\primitive`, `\readline`,
		`\scrollmode`, `\setbox`, `\show`,
		`\showbox`, `\showlists`, `\showthe`,
		`\skip`, `\skipdef`, `\the`, `\toks`,
		`\toksdef`, `\tracingall`, `\tracingcommands`,
		`\tracinglostchars`, `\tracingmacros`,
		`\tracingonline`, `\tracingoutput`,
		`\tracingpages`, `\tracingparagraphs`,
		`\tracingrestores`, `\tracingstats`,
		`\vcenter`, `\valign`, `\vrule`,
	}
	result := s
	for range 10 {
		prev := result
		for _, cmd := range blocked {
			result = strings.ReplaceAll(result, cmd, "")
		}
		if result == prev {
			break
		}
	}
	return result
}

// fetchMathAsync fetches a LaTeX math image from codecogs
// in a background goroutine.
//
// PRIVACY NOTE: LaTeX source is sent to external
// third-party API (latex.codecogs.com) for rendering.
func fetchMathAsync(
	w *Window, latex string, hash int64,
	requestID uint64, dpi int, fgColor Color,
) {
	go func() {
		safeLatex := sanitizeLatex(latex)

		// Build codecogs URL with DPI and optional color.
		lum := 0.299*float64(fgColor.R) +
			0.587*float64(fgColor.G) +
			0.114*float64(fgColor.B)
		colorCmd := ""
		if lum > 128.0 {
			colorCmd = `\color{white}`
		}
		prefix := fmt.Sprintf(`\dpi{%d}%s`, dpi, colorCmd)

		encoded := strings.ReplaceAll(
			prefix+safeLatex, " ", "{}")
		encoded = strings.ReplaceAll(encoded, "#", "%23")
		encoded = strings.ReplaceAll(encoded, "&", "%26")
		url := "https://latex.codecogs.com/png.image?" +
			encoded

		client := &http.Client{Timeout: diagramFetchTimeout}
		resp, err := client.Get(url)
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
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}

		if resp.StatusCode != 200 {
			preview := string(body)
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			errMsg := fmt.Sprintf("HTTP %d: %s",
				resp.StatusCode, preview)
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

		if len(body) > 10*1024*1024 {
			return
		}

		img, err := png.Decode(bytes.NewReader(body))
		if err != nil {
			return
		}

		bounds := img.Bounds()
		imgW := float32(bounds.Dx())
		imgH := float32(bounds.Dy())
		imgDPI := float32(dpi)

		tmpFile, err := os.CreateTemp("",
			fmt.Sprintf("math_%d_*.png", hash))
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
					Width:     imgW,
					Height:    imgH,
					DPI:       imgDPI,
					RequestID: requestID,
				})
			w.UpdateWindow()
		})
	}()
}
