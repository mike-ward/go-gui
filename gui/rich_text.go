package gui

// rich_text.go defines rich text types for mixed-style
// paragraphs. These wrap glyph.RichText/StyleRun internally
// while providing a gui-native API.

import "github.com/mike-ward/go-glyph"

// RichTextRun is a styled segment within a RichText block.
type RichTextRun struct {
	Text      string
	Style     TextStyle
	Link      string // URL for hyperlinks
	Tooltip   string // tooltip for abbreviations
	MathID    string // cache key for inline math
	MathLatex string // raw LaTeX source
}

// RichText contains runs of styled text for mixed-style
// paragraphs.
type RichText struct {
	Runs []RichTextRun
}

// RichRun creates a styled text run.
func RichRun(text string, style TextStyle) RichTextRun {
	return RichTextRun{Text: text, Style: style}
}

// RichLink creates a hyperlink run with underline styling.
func RichLink(
	text, url string, style TextStyle,
) RichTextRun {
	s := style
	s.Color = guiTheme.ColorSelect
	s.Underline = true
	return RichTextRun{Text: text, Link: url, Style: s}
}

// RichBr creates a line break run.
func RichBr() RichTextRun {
	return RichTextRun{Text: "\n", Style: guiTheme.N3}
}

// RichAbbr creates an abbreviation run with tooltip.
func RichAbbr(
	text, expansion string, style TextStyle,
) RichTextRun {
	s := style
	s.Typeface = glyph.TypefaceBold
	return RichTextRun{
		Text: text, Tooltip: expansion, Style: s,
	}
}

// RichFootnote creates a footnote marker with tooltip.
func RichFootnote(
	id, content string, baseStyle TextStyle,
) RichTextRun {
	s := baseStyle
	s.Size = baseStyle.Size * 0.7
	return RichTextRun{
		Text:    "\u2009[" + id + "]", // thin space
		Tooltip: content,
		Style:   s,
	}
}

// toGlyphRichText converts RichText to glyph.RichText.
func (rt RichText) toGlyphRichText() glyph.RichText {
	return rt.toGlyphRichTextWithMath(nil)
}

// toGlyphRichTextWithMath converts RichText to
// glyph.RichText, emitting InlineObject for math runs
// when the cache has dimensions.
func (rt RichText) toGlyphRichTextWithMath(
	cache *BoundedDiagramCache,
) glyph.RichText {
	runs := make([]glyph.StyleRun, 0, len(rt.Runs))
	for _, run := range rt.Runs {
		if run.MathID != "" && cache != nil {
			hash := mathCacheHash(run.MathID)
			if entry, ok := cache.Get(hash); ok {
				if entry.State == DiagramReady &&
					entry.Width > 0 {
					edpi := entry.DPI
					if edpi <= 0 {
						edpi = 200
					}
					scale := (72.0 / edpi) *
						(run.Style.Size / 12.0)
					gs := run.Style.ToGlyphStyle()
					gs.Object = &glyph.InlineObject{
						ID:     run.MathID,
						Width:  entry.Width * scale,
						Height: entry.Height * scale,
					}
					runs = append(runs, glyph.StyleRun{
						Text:  run.Text,
						Style: gs,
					})
					continue
				}
			}
			// Loading/error: show raw LaTeX as fallback.
			runs = append(runs, glyph.StyleRun{
				Text:  run.MathLatex,
				Style: run.Style.ToGlyphStyle(),
			})
			continue
		}
		runs = append(runs, glyph.StyleRun{
			Text:  run.Text,
			Style: run.Style.ToGlyphStyle(),
		})
	}
	return glyph.RichText{Runs: runs}
}
