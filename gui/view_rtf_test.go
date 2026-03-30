package gui

import (
	"errors"
	"testing"
	"time"

	"github.com/mike-ward/go-glyph"
)

type rtfStubTextMeasurer struct {
	layout glyph.Layout
}

func (m *rtfStubTextMeasurer) TextWidth(text string, style TextStyle) float32 {
	return float32(len(text)) * style.Size * 0.5
}

func (m *rtfStubTextMeasurer) TextHeight(_ string, style TextStyle) float32 {
	return style.Size * 1.2
}

func (m *rtfStubTextMeasurer) FontAscent(style TextStyle) float32 {
	return style.Size * 0.8
}

func (m *rtfStubTextMeasurer) FontHeight(style TextStyle) float32 {
	return style.Size * 1.2
}

func (m *rtfStubTextMeasurer) LayoutText(
	_ string, style TextStyle, _ float32,
) (glyph.Layout, error) {
	return glyph.Layout{Height: style.Size * 1.2}, nil
}

func (m *rtfStubTextMeasurer) LayoutRichText(
	_ glyph.RichText, _ glyph.TextConfig,
) (glyph.Layout, error) {
	return m.layout, nil
}

func TestRtfHitTestLogic(t *testing.T) {
	item := glyph.Item{
		X: 10, Y: 20, Width: 50,
		Ascent: 12, Descent: 4,
	}
	r := rtfRunRect(item)
	if r.X != 10 || r.Width != 50 {
		t.Fatalf("rect X/W: got %v/%v", r.X, r.Width)
	}
	// Y = run.Y - Ascent = 20 - 12 = 8.
	if r.Y != 8 {
		t.Fatalf("rect Y: expected 8, got %v", r.Y)
	}
	// Height = Ascent + Descent = 16.
	if r.Height != 16 {
		t.Fatalf("rect Height: expected 16, got %v", r.Height)
	}

	// Point inside.
	if !rtfHitTest(item, 30, 15) {
		t.Error("expected hit at (30,15)")
	}
	// Point outside.
	if rtfHitTest(item, 5, 5) {
		t.Error("expected miss at (5,5)")
	}
}

// --- RTF tooltip tests ---

func makeRtfTooltipLayout() (*Layout, *Window) {
	w := &Window{}
	rt := RichText{
		Runs: []RichTextRun{
			{Text: "hello", Tooltip: "tip text"},
		},
	}
	glyphLayout := glyph.Layout{
		Width:  100,
		Height: 20,
		Items: []glyph.Item{
			{
				X: 10, Y: 12, Width: 40,
				Ascent: 12, Descent: 4,
				StartIndex: 0,
			},
		},
	}
	l := &Layout{
		Shape: &Shape{
			ShapeType: ShapeRTF,
			X:         100,
			Y:         200,
			Width:     100,
			Height:    20,
			TC: &ShapeTextConfig{
				RtfLayout: &glyphLayout,
				RtfRuns:   &rt,
			},
		},
	}
	return l, w
}

func TestRtfMouseMoveSetTooltipState(t *testing.T) {
	l, w := makeRtfTooltipLayout()
	e := &Event{MouseX: 20, MouseY: 5}
	rtfMouseMove(l, e, w)

	ts := &w.viewState.tooltip
	if ts.hoverID != "tip text" {
		t.Errorf("hoverID = %q, want 'tip text'", ts.hoverID)
	}
	if ts.text != "tip text" {
		t.Errorf("text = %q, want 'tip text'", ts.text)
	}
	if ts.bounds == (DrawClip{}) {
		t.Error("bounds should be set")
	}
	if ts.floatOffsetX == 0 {
		t.Error("floatOffsetX should be set")
	}
	if ts.blockKey == 0 {
		t.Error("blockKey should be set")
	}
	if !e.IsHandled {
		t.Error("expected event handled")
	}
}

func TestRtfMouseMoveClearsOnNonTooltipRun(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip.hoverID = "old"
	w.viewState.tooltip.text = "old text"
	w.viewState.tooltip.id = "old"

	rt := RichText{
		Runs: []RichTextRun{
			{Text: "plain"},
		},
	}
	glyphLayout := glyph.Layout{
		Width:  100,
		Height: 20,
		Items: []glyph.Item{
			{
				X: 10, Y: 12, Width: 40,
				Ascent: 12, Descent: 4,
				StartIndex: 0,
			},
		},
	}
	l := &Layout{
		Shape: &Shape{
			ShapeType: ShapeRTF,
			X:         0,
			Y:         0,
			Width:     100,
			Height:    20,
			TC: &ShapeTextConfig{
				RtfLayout: &glyphLayout,
				RtfRuns:   &rt,
			},
		},
	}
	e := &Event{MouseX: 20, MouseY: 5}
	rtfMouseMove(l, e, w)

	ts := &w.viewState.tooltip
	if ts.text != "" {
		t.Errorf("text = %q, want empty", ts.text)
	}
	if ts.hoverID != "" {
		t.Errorf("hoverID = %q, want empty", ts.hoverID)
	}
}

func TestRtfMouseMoveUnderlineWithoutLinkDoesNotSetPointingHand(t *testing.T) {
	w := &Window{}
	rt := RichText{
		Runs: []RichTextRun{
			{Text: "underlined", Style: TextStyle{Underline: true}},
		},
	}
	glyphLayout := glyph.Layout{
		Width:  100,
		Height: 20,
		Items: []glyph.Item{
			{
				X: 10, Y: 12, Width: 60,
				Ascent: 12, Descent: 4,
				StartIndex:   0,
				HasUnderline: true,
			},
		},
	}
	l := &Layout{
		Shape: &Shape{
			ShapeType: ShapeRTF,
			Width:     100,
			Height:    20,
			TC: &ShapeTextConfig{
				RtfLayout: &glyphLayout,
				RtfRuns:   &rt,
			},
		},
	}
	e := &Event{MouseX: 20, MouseY: 5}

	rtfMouseMove(l, e, w)

	if got := w.MouseCursorState(); got == CursorPointingHand {
		t.Fatalf("cursor = %v, want non-link cursor", got)
	}
	if e.IsHandled {
		t.Fatal("expected underline-only hover not to consume event")
	}
}

func TestRtfMouseMoveLinkSetsPointingHand(t *testing.T) {
	w := &Window{}
	rt := RichText{
		Runs: []RichTextRun{
			{
				Text:  "link",
				Link:  "https://example.com",
				Style: TextStyle{Underline: true},
			},
		},
	}
	glyphLayout := glyph.Layout{
		Width:  100,
		Height: 20,
		Items: []glyph.Item{
			{
				X: 10, Y: 12, Width: 30,
				Ascent: 12, Descent: 4,
				StartIndex:   0,
				HasUnderline: true,
			},
		},
	}
	l := &Layout{
		Shape: &Shape{
			ShapeType: ShapeRTF,
			Width:     100,
			Height:    20,
			TC: &ShapeTextConfig{
				RtfLayout: &glyphLayout,
				RtfRuns:   &rt,
			},
		},
	}
	e := &Event{MouseX: 20, MouseY: 5}

	rtfMouseMove(l, e, w)

	if got := w.MouseCursorState(); got != CursorPointingHand {
		t.Fatalf("cursor = %v, want %v", got, CursorPointingHand)
	}
	if !e.IsHandled {
		t.Fatal("expected link hover to consume event")
	}
}

func TestRtfGenerateLayoutSuppressesInlineObjectGlyphs(t *testing.T) {
	w := &Window{windowBackend: windowBackend{
		textMeasurer: &rtfStubTextMeasurer{
			layout: glyph.Layout{
				Width:  120,
				Height: 20,
				Items: []glyph.Item{
					{IsObject: true, ObjectID: "math_1", GlyphStart: 4, GlyphCount: 1},
					{GlyphStart: 0, GlyphCount: 4},
				},
			},
		},
	}}

	layout := GenerateViewLayout(RTF(RtfCfg{
		RichText: RichText{
			Runs: []RichTextRun{{
				MathID:    "math_1",
				MathLatex: "x^2",
				Style:     TextStyle{Size: 12},
			}},
		},
	}), w)

	items := layout.Shape.TC.RtfLayout.Items
	if got := items[0].GlyphCount; got != 0 {
		t.Fatalf("object GlyphCount = %d, want 0", got)
	}
	if got := items[1].GlyphCount; got != 4 {
		t.Fatalf("text GlyphCount = %d, want 4", got)
	}
}

func TestLayoutWrapRTFSuppressesInlineObjectGlyphs(t *testing.T) {
	w := &Window{windowBackend: windowBackend{
		textMeasurer: &rtfStubTextMeasurer{
			layout: glyph.Layout{
				Width:  80,
				Height: 24,
				Items: []glyph.Item{
					{IsObject: true, ObjectID: "math_2", GlyphStart: 2, GlyphCount: 1},
				},
			},
		},
	}}

	baseStyle := glyph.TextStyle{Size: 12}
	rt := RichText{
		Runs: []RichTextRun{{
			MathID:    "math_2",
			MathLatex: "y^2",
			Style:     TextStyle{Size: 12},
		}},
	}
	shape := &Shape{
		ShapeType: ShapeRTF,
		Width:     100,
		TC: &ShapeTextConfig{
			TextMode:     TextModeWrap,
			RtfBaseStyle: baseStyle,
			RtfRuns:      &rt,
		},
	}

	layoutWrapRTF(shape, shape.TC, w)

	if shape.TC.RtfLayout == nil {
		t.Fatal("expected wrapped RTF layout")
	}
	if got := shape.TC.RtfLayout.Items[0].GlyphCount; got != 0 {
		t.Fatalf("wrapped object GlyphCount = %d, want 0", got)
	}
}

func TestRtfAmendTooltipClearsOutsideBounds(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip = tooltipState{
		bounds:  DrawClip{X: 100, Y: 200, Width: 40, Height: 16},
		id:      "tip",
		hoverID: "tip",
		text:    "tip text",
	}
	// Mouse outside bounds.
	w.viewState.mousePosX = 300
	w.viewState.mousePosY = 300

	l := &Layout{Shape: &Shape{}}
	rtfAmendTooltip(l, w)

	ts := &w.viewState.tooltip
	if ts.text != "" {
		t.Errorf("text = %q, want empty", ts.text)
	}
	if ts.id != "" {
		t.Errorf("id = %q, want empty", ts.id)
	}
}

func TestRtfAmendTooltipNopWhenNoText(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip = tooltipState{
		hoverID: "widget-tip",
		id:      "widget-tip",
	}
	w.viewState.mousePosX = 300
	w.viewState.mousePosY = 300

	l := &Layout{Shape: &Shape{}}
	rtfAmendTooltip(l, w)

	// WithTooltip state should be preserved.
	ts := &w.viewState.tooltip
	if ts.hoverID != "widget-tip" {
		t.Errorf("hoverID = %q, want 'widget-tip'",
			ts.hoverID)
	}
	if ts.id != "widget-tip" {
		t.Errorf("id = %q, want 'widget-tip'", ts.id)
	}
}

func TestClearTextGuardsNonRtfTooltips(t *testing.T) {
	ts := &tooltipState{
		hoverID:    "widget",
		hoverStart: time.Now(),
		id:         "widget",
		text:       "",
	}
	ts.clearText()

	if ts.hoverID != "widget" {
		t.Errorf("hoverID = %q, want 'widget'", ts.hoverID)
	}
	if ts.id != "widget" {
		t.Errorf("id = %q, want 'widget'", ts.id)
	}
}

func TestRtfRunsKeyStable(t *testing.T) {
	rt := RichText{Runs: []RichTextRun{
		{Text: "hello"},
		{Text: "world"},
	}}
	k1 := rtfRunsKey(&rt)
	k2 := rtfRunsKey(&rt)
	if k1 != k2 {
		t.Error("same content should produce same key")
	}
	if k1 == 0 {
		t.Error("key should be non-zero")
	}
	rt2 := RichText{Runs: []RichTextRun{
		{Text: "different"},
	}}
	if rtfRunsKey(&rt2) == k1 {
		t.Error("different content should produce different key")
	}
}

// --- RTF link context menu tests ---

func TestShowLinkContextMenuSetsState(t *testing.T) {
	w := newTestWindow()
	showLinkContextMenu(w, "https://example.com", 50, 100, 42)

	st := StateReadOr(
		w, nsRtfLinkMenu, nsRtfLinkMenu, rtfLinkMenuState{})
	if !st.Open {
		t.Fatal("expected Open=true")
	}
	if st.Link != "https://example.com" {
		t.Fatalf("expected link=%q got %q",
			"https://example.com", st.Link)
	}
	if st.X != 50 || st.Y != 100 {
		t.Fatalf("expected pos=(50,100) got (%g,%g)",
			st.X, st.Y)
	}
	if st.BlockKey != 42 {
		t.Fatalf("expected BlockKey=42 got %d", st.BlockKey)
	}
	if w.IDFocus() != rtfLinkMenuIDFocus {
		t.Fatalf("expected focus=%d got %d",
			rtfLinkMenuIDFocus, w.IDFocus())
	}
}

func TestRtfLinkMenuDismissClearsState(t *testing.T) {
	w := newTestWindow()
	showLinkContextMenu(w, "https://example.com", 50, 100, 0)
	rtfLinkMenuDismiss(w)

	st := StateReadOr(
		w, nsRtfLinkMenu, nsRtfLinkMenu, rtfLinkMenuState{})
	if st.Open {
		t.Fatal("expected Open=false after dismiss")
	}
	if w.IDFocus() != 0 {
		t.Fatalf("expected focus=0 got %d", w.IDFocus())
	}
}

func TestRtfOnClickRightClickShowsMenu(t *testing.T) {
	w := newTestWindow()
	rt := RichText{
		Runs: []RichTextRun{
			{Text: "click me", Link: "https://safe.example.com"},
		},
	}
	glyphLayout := glyph.Layout{
		Width: 100, Height: 20,
		Items: []glyph.Item{
			{
				X: 10, Y: 12, Width: 60,
				Ascent: 12, Descent: 4,
				StartIndex: 0,
			},
		},
	}
	l := &Layout{
		Shape: &Shape{
			ShapeType: ShapeRTF,
			X:         100, Y: 200,
			Width: 100, Height: 20,
			TC: &ShapeTextConfig{
				RtfLayout: &glyphLayout,
				RtfRuns:   &rt,
			},
		},
	}
	e := &Event{
		MouseX:      20,
		MouseY:      5,
		MouseButton: MouseRight,
	}
	rtfOnClick(l, e, w)

	if !e.IsHandled {
		t.Fatal("expected IsHandled=true")
	}
	st := StateReadOr(
		w, nsRtfLinkMenu, nsRtfLinkMenu, rtfLinkMenuState{})
	if !st.Open {
		t.Fatal("expected context menu Open=true")
	}
	if st.Link != "https://safe.example.com" {
		t.Fatalf("expected link=%q got %q",
			"https://safe.example.com", st.Link)
	}
}

func TestRtfAmendTooltipDismissesMenuOnFocusLoss(t *testing.T) {
	w := newTestWindow()
	showLinkContextMenu(w, "https://example.com", 50, 100, 0)
	// Simulate focus moving away.
	w.SetIDFocus(0)

	l := &Layout{Shape: &Shape{}}
	rtfAmendTooltip(l, w)

	st := StateReadOr(
		w, nsRtfLinkMenu, nsRtfLinkMenu, rtfLinkMenuState{})
	if st.Open {
		t.Fatal("expected menu dismissed on focus loss")
	}
}

func TestRtfGenerateLayoutAddsTooltipChild(t *testing.T) {
	rt := RichText{Runs: []RichTextRun{
		{Text: "abbr", Tooltip: "abbreviation"},
	}}
	w := &Window{}
	w.viewState.tooltip = tooltipState{
		id:           "abbreviation",
		hoverID:      "abbreviation",
		text:         "abbreviation",
		blockKey:     rtfRunsKey(&rt),
		floatOffsetX: 30,
		floatOffsetY: -3,
	}
	v := &rtfView{
		RtfCfg: RtfCfg{RichText: rt},
		sizing: FitFit,
	}
	layout := v.GenerateLayout(w)
	if len(layout.Children) != 1 {
		t.Fatalf("expected 1 child (popup), got %d",
			len(layout.Children))
	}
	child := layout.Children[0]
	if !child.Shape.Float {
		t.Error("popup child should be floating")
	}
}

func TestRtfGenerateLayoutEmptyRichText(t *testing.T) {
	w := &Window{windowBackend: windowBackend{
		textMeasurer: &rtfStubTextMeasurer{
			layout: glyph.Layout{},
		},
	}}
	layout := GenerateViewLayout(RTF(RtfCfg{
		RichText: RichText{},
	}), w)
	if layout.Shape == nil {
		t.Fatal("expected non-nil shape")
	}
	if layout.Shape.ShapeType != ShapeRTF {
		t.Fatalf("type = %v, want ShapeRTF",
			layout.Shape.ShapeType)
	}
}

type rtfErrorMeasurer struct{ rtfStubTextMeasurer }

func (m *rtfErrorMeasurer) LayoutRichText(
	_ glyph.RichText, _ glyph.TextConfig,
) (glyph.Layout, error) {
	return glyph.Layout{}, errors.New("test error")
}

func TestRtfGenerateLayoutHandlesError(t *testing.T) {
	w := &Window{windowBackend: windowBackend{textMeasurer: &rtfErrorMeasurer{}}}
	layout := GenerateViewLayout(RTF(RtfCfg{
		RichText: RichText{
			Runs: []RichTextRun{{Text: "hello"}},
		},
	}), w)
	if layout.Shape == nil {
		t.Fatal("expected non-nil shape")
	}
	// Layout should still produce a shape but with no
	// usable glyph layout dimensions.
	if layout.Shape.Width != 0 || layout.Shape.Height != 0 {
		t.Fatalf("expected zero size, got %gx%g",
			layout.Shape.Width, layout.Shape.Height)
	}
}

func TestRtfOnClickIgnoresUnsafeLink(t *testing.T) {
	w := newTestWindow()
	rt := RichText{
		Runs: []RichTextRun{
			{Text: "evil", Link: "javascript:alert(1)"},
		},
	}
	glyphLayout := glyph.Layout{
		Width: 100, Height: 20,
		Items: []glyph.Item{
			{
				X: 10, Y: 12, Width: 30,
				Ascent: 12, Descent: 4,
				StartIndex: 0,
			},
		},
	}
	l := &Layout{
		Shape: &Shape{
			ShapeType: ShapeRTF,
			Width:     100, Height: 20,
			TC: &ShapeTextConfig{
				RtfLayout: &glyphLayout,
				RtfRuns:   &rt,
			},
		},
	}
	e := &Event{MouseX: 20, MouseY: 5}
	rtfOnClick(l, e, w)
	if e.IsHandled {
		t.Fatal("unsafe link should not be handled")
	}
}

func TestRtfRunsKeyIncludesLinkAndTooltip(t *testing.T) {
	rt1 := RichText{Runs: []RichTextRun{
		{Text: "same", Link: "https://a.com"},
	}}
	rt2 := RichText{Runs: []RichTextRun{
		{Text: "same", Link: "https://b.com"},
	}}
	if rtfRunsKey(&rt1) == rtfRunsKey(&rt2) {
		t.Error("different links should produce different keys")
	}

	rt3 := RichText{Runs: []RichTextRun{
		{Text: "same", Tooltip: "tip A"},
	}}
	rt4 := RichText{Runs: []RichTextRun{
		{Text: "same", Tooltip: "tip B"},
	}}
	if rtfRunsKey(&rt3) == rtfRunsKey(&rt4) {
		t.Error("different tooltips should produce different keys")
	}
}
