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

func (tm *textMeasurer) FontAscent(style gui.TextStyle) float32 {
	cfg := guiStyleToGlyphConfig(style)
	m, err := tm.textSys.FontMetrics(cfg)
	if err != nil {
		return style.Size * 0.8
	}
	return m.Ascender
}

func (tm *textMeasurer) LayoutText(
	text string, style gui.TextStyle, wrapWidth float32,
) (glyph.Layout, error) {
	cfg := guiStyleToGlyphConfig(style)
	if wrapWidth > 0 {
		cfg.Block.Width = wrapWidth
		cfg.Block.Wrap = glyph.WrapWord
	} else if wrapWidth < 0 {
		cfg.Block.Width = -wrapWidth
		cfg.Block.Wrap = glyph.WrapNone
	}
	return tm.textSys.LayoutText(text, cfg)
}

func (tm *textMeasurer) LayoutRichText(
	rt glyph.RichText, cfg glyph.TextConfig,
) (glyph.Layout, error) {
	return tm.textSys.LayoutRichText(rt, cfg)
}

func guiStyleToGlyphConfig(s gui.TextStyle) glyph.TextConfig {
	align := glyph.AlignLeft
	switch s.Align {
	case gui.TextAlignCenter:
		align = glyph.AlignCenter
	case gui.TextAlignRight:
		align = glyph.AlignRight
	}
	return glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName:      s.Family,
			Size:          s.Size,
			Color:         glyph.Color{R: s.Color.R, G: s.Color.G, B: s.Color.B, A: s.Color.A},
			BgColor:       glyph.Color{R: s.BgColor.R, G: s.BgColor.G, B: s.BgColor.B, A: s.BgColor.A},
			Typeface:      s.Typeface,
			Underline:     s.Underline,
			Strikethrough: s.Strikethrough,
			LetterSpacing: s.LetterSpacing,
			StrokeWidth:   s.StrokeWidth,
			StrokeColor:   glyph.Color{R: s.StrokeColor.R, G: s.StrokeColor.G, B: s.StrokeColor.B, A: s.StrokeColor.A},
			Features:      s.Features,
		},
		Block: glyph.BlockStyle{
			Align:       align,
			Wrap:        glyph.WrapWord,
			Width:       -1,
			LineSpacing: s.LineSpacing,
		},
		Gradient: s.Gradient,
	}
}
