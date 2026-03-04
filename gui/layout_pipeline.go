package gui

import "github.com/mike-ward/go-glyph"

// layoutPipeline runs all layout passes in order on a single
// layout tree.
func layoutPipeline(layout *Layout, w *Window) {
	// Width passes.
	layoutWidths(layout)
	layoutFillWidths(layout)
	layoutWrapContainers(layout, w)
	layoutOverflow(layout, w)
	layoutWrapText(layout, w)

	// Height passes.
	layoutHeights(layout)
	layoutFillHeights(layout)

	// Position passes.
	layoutAdjustScrollOffsets(layout, w)
	fx, fy := floatAttachLayout(layout, w.WindowRect())
	layoutPositions(layout, fx, fy, w)
	layoutDisables(layout, false)
	layoutScrollContainers(layout, 0)

	// Post-position passes.
	layoutAmend(layout, w)
	applyLayoutTransition(layout, w)
	applyHeroTransition(layout, w)
	layoutSetShapeClips(layout, w.WindowRect())
}

// layoutAmend walks the layout tree children-first, firing
// AmendLayout callbacks. Not for size changes — post-position
// only.
func layoutAmend(layout *Layout, w *Window) {
	for i := range layout.Children {
		layoutAmend(&layout.Children[i], w)
	}
	if layout.Shape.HasEvents() &&
		layout.Shape.Events.AmendLayout != nil {
		layout.Shape.Events.AmendLayout(layout, w)
	}
}

// layoutHover walks the layout tree depth-first, firing OnHover
// callbacks when the mouse is inside a shape. Returns true if
// any hover was handled.
func layoutHover(layout *Layout, w *Window) bool {
	if w.MouseIsLocked() {
		return false
	}
	for i := len(layout.Children) - 1; i >= 0; i-- {
		if layoutHover(&layout.Children[i], w) {
			return true
		}
	}
	shape := layout.Shape
	if shape.Disabled {
		return false
	}
	if !shape.HasEvents() || shape.Events.OnHover == nil {
		return false
	}
	if !shape.PointInShape(w.viewState.mousePosX,
		w.viewState.mousePosY) {
		return false
	}
	if w.dialogCfg.visible &&
		!layoutInDialogLayout(layout) {
		return false
	}
	e := Event{
		MouseX:      w.viewState.mousePosX,
		MouseY:      w.viewState.mousePosY,
		Type:        EventMouseMove,
		MouseButton: MouseInvalid,
	}
	shape.Events.OnHover(layout, &e, w)
	return true
}

// layoutInDialogLayout walks the parent chain checking if any
// ancestor has ID == reservedDialogID.
func layoutInDialogLayout(layout *Layout) bool {
	for p := layout; p != nil; p = p.Parent {
		if p.Shape.ID == reservedDialogID {
			return true
		}
	}
	return false
}

// layoutWrapText re-layouts text and RTF shapes that use text
// wrapping. Called after fill-widths so actual widths are known.
func layoutWrapText(layout *Layout, w *Window) {
	if layout == nil || w == nil {
		return
	}
	layoutWrapTextWalk(layout, w)
}

func layoutWrapTextWalk(layout *Layout, w *Window) {
	for i := range layout.Children {
		layoutWrapTextWalk(&layout.Children[i], w)
	}
	shape := layout.Shape
	tc := shape.TC
	if tc == nil {
		return
	}
	if tc.TextMode != TextModeWrap &&
		tc.TextMode != TextModeWrapKeepSpaces {
		return
	}
	if shape.Width <= 0 {
		return
	}
	switch shape.ShapeType {
	case ShapeRTF:
		layoutWrapRTF(shape, tc, w)
	case ShapeText:
		layoutWrapPlainText(shape, tc, w)
	}
}

func layoutWrapRTF(shape *Shape, tc *ShapeTextConfig, w *Window) {
	if tc.RtfRuns == nil {
		return
	}
	tm, ok := w.textMeasurer.(interface {
		LayoutRichText(glyph.RichText, glyph.TextConfig) (glyph.Layout, error)
	})
	if !ok {
		return
	}
	vgRT := tc.RtfRuns.toGlyphRichTextWithMath(
		w.viewState.diagramCache)
	cfg := glyph.TextConfig{
		Style: tc.RtfBaseStyle,
		Block: glyph.BlockStyle{
			Wrap:   glyph.WrapWord,
			Width:  shape.Width,
			Indent: -tc.HangingIndent,
		},
	}
	l, err := tm.LayoutRichText(vgRT, cfg)
	if err != nil {
		return
	}
	tc.RtfLayout = &l
	shape.Height = l.Height
}

func layoutWrapPlainText(shape *Shape, tc *ShapeTextConfig,
	w *Window,
) {
	if w.textMeasurer == nil || tc.TextStyle == nil {
		return
	}
	if len(tc.Text) == 0 {
		return
	}
	l, err := w.textMeasurer.LayoutText(
		tc.Text, *tc.TextStyle, shape.Width)
	if err != nil {
		return
	}
	shape.Height = l.Height
}
