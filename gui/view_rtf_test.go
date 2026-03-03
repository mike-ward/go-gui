package gui

import (
	"math"
	"testing"
	"time"

	"github.com/mike-ward/go-glyph"
)

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
	if !rtfHitTest(item, 30, 15, nil) {
		t.Error("expected hit at (30,15)")
	}
	// Point outside.
	if rtfHitTest(item, 5, 5, nil) {
		t.Error("expected miss at (5,5)")
	}
}

func TestRtfAffineInverse(t *testing.T) {
	// Identity matrix.
	id := glyph.AffineTransform{
		XX: 1, XY: 0, YX: 0, YY: 1, X0: 0, Y0: 0,
	}
	inv, ok := rtfAffineInverse(id)
	if !ok {
		t.Fatal("identity should be invertible")
	}
	if math.Abs(float64(inv.XX-1)) > 0.001 ||
		math.Abs(float64(inv.YY-1)) > 0.001 {
		t.Fatalf("identity inverse: XX=%v YY=%v",
			inv.XX, inv.YY)
	}

	// Singular matrix.
	singular := glyph.AffineTransform{
		XX: 0, XY: 0, YX: 0, YY: 0,
	}
	_, ok = rtfAffineInverse(singular)
	if ok {
		t.Fatal("singular matrix should not be invertible")
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
