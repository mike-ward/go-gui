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

	"github.com/mike-ward/go-gui/gui/markdown"
)

// mathCacheHash computes a cache key for a math expression.
func mathCacheHash(mathID string) int64 {
	h := markdown.MathHash(mathID)
	return int64((h << 32) | uint64(len(mathID)))
}

// blockedLatexCmds lists TeX commands blocked to prevent
// shell escape or file access on remote renderers.
var blockedLatexCmds = []string{
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

// sanitizeLatex strips dangerous TeX commands that could
// enable shell escape or file access on the remote renderer.
func sanitizeLatex(s string) string {
	if len(s) > markdown.MaxLatexSourceLen {
		return ""
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	result := strings.Map(func(r rune) rune {
		switch {
		case r == '\r' || r == '\n' || r == '\t':
			return ' '
		case r < 0x20:
			return -1
		default:
			return r
		}
	}, s)
	result = strings.TrimSpace(result)
	for range 10 {
		prev := result
		for _, cmd := range blockedLatexCmds {
			result = strings.ReplaceAll(result, cmd, "")
		}
		if result == prev {
			break
		}
	}
	return result
}

// queueDiagramError queues a DiagramError cache entry.
func queueDiagramError(
	w *Window, hash int64, requestID uint64, errMsg string,
) {
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
			queueDiagramError(w, hash, requestID, err.Error())
			return
		}
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			queueDiagramError(w, hash, requestID,
				"read body: "+err.Error())
			return
		}

		if resp.StatusCode != 200 {
			preview := truncatePreview(string(body), 200)
			queueDiagramError(w, hash, requestID,
				fmt.Sprintf("HTTP %d: %s",
					resp.StatusCode, preview))
			return
		}

		if len(body) > 10*1024*1024 {
			queueDiagramError(w, hash, requestID,
				"response exceeds 10 MB limit")
			return
		}

		img, err := png.Decode(bytes.NewReader(body))
		if err != nil {
			queueDiagramError(w, hash, requestID,
				"decode PNG: "+err.Error())
			return
		}

		bounds := img.Bounds()
		imgW := float32(bounds.Dx())
		imgH := float32(bounds.Dy())
		imgDPI := float32(dpi)

		tmpFile, err := os.CreateTemp("",
			fmt.Sprintf("math_%d_*.png", hash))
		if err != nil {
			queueDiagramError(w, hash, requestID,
				"create temp file: "+err.Error())
			return
		}
		tmpPath := tmpFile.Name()
		if err := png.Encode(tmpFile, img); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
			queueDiagramError(w, hash, requestID,
				"encode PNG: "+err.Error())
			return
		}
		_ = tmpFile.Close()

		w.QueueCommand(func(w *Window) {
			if !diagramCacheShouldApplyResult(
				w.viewState.diagramCache,
				hash, requestID) {
				_ = os.Remove(tmpPath)
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
