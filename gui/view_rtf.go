package gui

// view_rtf.go defines the Rich Text Format (RTF) view.
// Renders text with multiple typefaces, sizes, and styles.
// Supports text wrapping, clickable links, and custom runs.

import (
	"math"

	"github.com/mike-ward/go-glyph"

	"github.com/mike-ward/go-gui/gui/markdown"
)

// RtfCfg configures a Rich Text View.
type RtfCfg struct {
	ID             string
	RichText       RichText
	MinWidth       float32
	IDFocus        uint32
	Mode           TextMode
	Invisible      bool
	Clip           bool
	FocusSkip      bool
	Disabled       bool
	HangingIndent  float32 // negative indent for wrapped lines
	BaseTextStyle  *TextStyle
}

type rtfView struct {
	RtfCfg
	sizing Sizing
}

func (v *rtfView) Content() []View { return nil }

func (v *rtfView) GenerateLayout(w *Window) Layout {
	// Convert RichText to glyph.RichText.
	vgRT := v.RichText.toGlyphRichTextWithMath(
		w.viewState.diagramCache)

	// Determine base style.
	var baseStyle glyph.TextStyle
	if v.BaseTextStyle != nil {
		baseStyle = v.BaseTextStyle.ToGlyphStyle()
	} else if len(vgRT.Runs) > 0 {
		baseStyle = vgRT.Runs[0].Style
	}

	cfg := glyph.TextConfig{
		Style: baseStyle,
		Block: glyph.BlockStyle{
			Wrap:   glyph.WrapWord,
			Width:  -1.0,
			Indent: -v.HangingIndent,
		},
	}

	// Layout rich text via text measurer.
	var layout glyph.Layout
	if w.textMeasurer != nil {
		if tm, ok := w.textMeasurer.(interface {
			LayoutRichText(glyph.RichText, glyph.TextConfig) (glyph.Layout, error)
		}); ok {
			if l, err := tm.LayoutRichText(vgRT, cfg); err == nil {
				layout = l
			}
		}
	}

	shape := &Shape{
		ShapeType: ShapeRTF,
		ID:        v.ID,
		IDFocus:   v.IDFocus,
		Width:     layout.Width,
		Height:    layout.Height,
		Clip:      v.Clip,
		FocusSkip: v.FocusSkip,
		Disabled:  v.Disabled,
		MinWidth:  v.MinWidth,
		Sizing:    v.sizing,
		Events: &EventHandlers{
			OnClick:     rtfOnClick,
			OnMouseMove: rtfMouseMove,
		},
		TC: &ShapeTextConfig{
			TextMode:       v.Mode,
			HangingIndent:  v.HangingIndent,
			RtfBaseStyle:   baseStyle,
			RtfLayout:      &layout,
			RtfRuns:        &v.RichText,
		},
	}
	return Layout{Shape: shape}
}

// RTF creates a rich text view.
func RTF(cfg RtfCfg) View {
	if cfg.Invisible {
		return invisibleContainerView()
	}
	sizing := FitFit
	if cfg.Mode == TextModeWrap ||
		cfg.Mode == TextModeWrapKeepSpaces {
		sizing = FillFit
	}
	return &rtfView{RtfCfg: cfg, sizing: sizing}
}

// --- Hit testing ---

const rtfAffineInverseEpsilon = float32(0.000001)

func rtfRunRect(run glyph.Item) DrawClip {
	return DrawClip{
		X:      float32(run.X),
		Y:      float32(run.Y - run.Ascent),
		Width:  float32(run.Width),
		Height: float32(run.Ascent + run.Descent),
	}
}

func rtfAffineInverse(
	t glyph.AffineTransform,
) (glyph.AffineTransform, bool) {
	det := t.XX*t.YY - t.XY*t.YX
	if float32(math.Abs(float64(det))) <=
		rtfAffineInverseEpsilon {
		return glyph.AffineTransform{}, false
	}
	invDet := 1.0 / det
	xx := t.YY * invDet
	xy := -t.XY * invDet
	yx := -t.YX * invDet
	yy := t.XX * invDet
	return glyph.AffineTransform{
		XX: xx, XY: xy, YX: yx, YY: yy,
		X0: -(xx*t.X0 + xy*t.Y0),
		Y0: -(yx*t.X0 + yy*t.Y0),
	}, true
}

func rtfHitTest(
	run glyph.Item, mx, my float32,
	inv *glyph.AffineTransform,
) bool {
	tx, ty := mx, my
	if inv != nil {
		tx, ty = inv.Apply(mx, my)
	}
	r := rtfRunRect(run)
	return tx >= r.X && ty >= r.Y &&
		tx < r.X+r.Width && ty < r.Y+r.Height
}

func rtfFindRunAtIndex(
	l *Layout, startIndex int,
) RichTextRun {
	if l.Shape == nil || l.Shape.TC == nil ||
		l.Shape.TC.RtfRuns == nil {
		return RichTextRun{}
	}
	idx := 0
	for _, r := range l.Shape.TC.RtfRuns.Runs {
		runLen := len(r.Text)
		if startIndex >= idx &&
			startIndex < idx+runLen {
			return r
		}
		idx += runLen
	}
	return RichTextRun{}
}

// --- Event handlers ---

func rtfMouseMove(l *Layout, e *Event, w *Window) {
	if !l.Shape.HasRtfLayout() {
		return
	}
	layout := l.Shape.TC.RtfLayout
	for _, run := range layout.Items {
		if run.IsObject {
			continue
		}
		if rtfHitTest(run, e.MouseX, e.MouseY, nil) {
			found := rtfFindRunAtIndex(l, run.StartIndex)
			if found.Tooltip != "" {
				e.IsHandled = true
				return
			}
			if run.HasUnderline {
				w.SetMouseCursorPointingHand()
				e.IsHandled = true
				return
			}
		}
	}
}

func rtfOnClick(l *Layout, e *Event, w *Window) {
	if !l.Shape.HasRtfLayout() {
		return
	}
	layout := l.Shape.TC.RtfLayout
	for _, run := range layout.Items {
		if run.IsObject {
			continue
		}
		if rtfHitTest(run, e.MouseX, e.MouseY, nil) {
			found := rtfFindRunAtIndex(l, run.StartIndex)
			if found.Link != "" && markdown.IsSafeURL(found.Link) {
				if len(found.Link) > 0 &&
					found.Link[0] == '#' {
					w.ScrollToView(found.Link[1:])
				} else if w.nativePlatform != nil {
					w.nativePlatform.OpenURI(found.Link)
				}
				e.IsHandled = true
			}
			return
		}
	}
}

// showLinkContextMenu is a stub — deferred.
func showLinkContextMenu(
	_ *Window, _, _ string, _, _ float32,
) {
}
