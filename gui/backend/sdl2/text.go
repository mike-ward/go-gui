package sdl2

import (
	"github.com/mike-ward/go-glyph"
	"github.com/mike-ward/go-gui/gui"
)

// textMeasurer wraps glyph.TextSystem to implement gui.TextMeasurer.
type textMeasurer struct {
	textSys *glyph.TextSystem
}

func (tm *textMeasurer) TextWidth(text string, style gui.TextStyle) float32 {
	cfg := guiStyleToGlyphConfig(style)
	w, err := tm.textSys.TextWidth(text, cfg)
	if err != nil {
		return 0
	}
	return w
}

func (tm *textMeasurer) TextHeight(text string, style gui.TextStyle) float32 {
	cfg := guiStyleToGlyphConfig(style)
	h, err := tm.textSys.TextHeight(text, cfg)
	if err != nil {
		return 0
	}
	return h
}

func (tm *textMeasurer) FontHeight(style gui.TextStyle) float32 {
	cfg := guiStyleToGlyphConfig(style)
	h, err := tm.textSys.FontHeight(cfg)
	if err != nil {
		return style.Size * 1.4
	}
	return h
}

func (tm *textMeasurer) LayoutRichText(
	rt glyph.RichText, cfg glyph.TextConfig,
) (glyph.Layout, error) {
	return tm.textSys.LayoutRichText(rt, cfg)
}

func guiStyleToGlyphConfig(s gui.TextStyle) glyph.TextConfig {
	return glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName:      s.Family,
			Size:          s.Size,
			Color:         glyph.Color(s.Color),
			Typeface:      s.Typeface,
			Underline:     s.Underline,
			Strikethrough: s.Strikethrough,
			LetterSpacing: s.LetterSpacing,
			StrokeWidth:   s.StrokeWidth,
			StrokeColor:   glyph.Color(s.StrokeColor),
		},
		Block: glyph.DefaultBlockStyle(),
	}
}
