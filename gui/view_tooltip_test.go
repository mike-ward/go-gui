package gui

import (
	"testing"
	"time"
)

func TestTooltipReturnsView(t *testing.T) {
	v := Tooltip(TooltipCfg{
		ID:      "tip1",
		Content: []View{Text(TextCfg{Text: "hello"})},
	})
	if v == nil {
		t.Fatal("expected non-nil view")
	}
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	if !layout.Shape.Float {
		t.Error("expected floating layout")
	}
}

func TestTooltipCfgDefaults(t *testing.T) {
	cfg := TooltipCfg{ID: "test"}
	applyTooltipDefaults(&cfg)

	if !cfg.Color.IsSet() {
		t.Error("expected non-zero Color")
	}
	if cfg.Delay == 0 {
		t.Error("expected non-zero delay")
	}
	if cfg.Radius == 0 {
		t.Error("expected non-zero radius")
	}
	if cfg.OffsetX == 0 || cfg.OffsetY == 0 {
		t.Error("expected non-zero offsets")
	}
	if cfg.Anchor != FloatBottomCenter {
		t.Errorf("expected BottomCenter anchor, got %d",
			cfg.Anchor)
	}
}

func TestAnimationTooltipReturnsAnimate(t *testing.T) {
	a := AnimationTooltip(TooltipCfg{
		ID:    "tip1",
		Delay: 100 * time.Millisecond,
	})
	if a == nil {
		t.Fatal("expected non-nil Animate")
	}
	if a.AnimID != "___tooltip___" {
		t.Errorf("expected ___tooltip___ id, got %q",
			a.AnimID)
	}
}

func TestAnimationTooltipCallback(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip.bounds = DrawClip{
		X: 10, Y: 10, Width: 50, Height: 50,
	}
	w.viewState.mousePosX = 20
	w.viewState.mousePosY = 20

	a := AnimationTooltip(TooltipCfg{
		ID:    "tip1",
		Delay: 100 * time.Millisecond,
	})
	a.Callback(a, w)

	if w.viewState.tooltip.id != "tip1" {
		t.Errorf("expected tooltip id=tip1, got %q",
			w.viewState.tooltip.id)
	}
}

func TestAnimationTooltipCallbackOutside(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip.bounds = DrawClip{
		X: 10, Y: 10, Width: 50, Height: 50,
	}
	w.viewState.mousePosX = 100
	w.viewState.mousePosY = 100

	a := AnimationTooltip(TooltipCfg{
		ID:    "tip1",
		Delay: 100 * time.Millisecond,
	})
	a.Callback(a, w)

	if w.viewState.tooltip.id != "" {
		t.Error("expected empty tooltip id when outside")
	}
}

func TestWithTooltipReturnsView(t *testing.T) {
	w := &Window{}
	v := WithTooltip(w, WithTooltipCfg{
		ID:      "tip1",
		Text:    "hello",
		Content: []View{Text(TextCfg{Text: "trigger"})},
	})
	if v == nil {
		t.Fatal("expected non-nil view")
	}
}

func TestWithTooltipShowsPopupWhenActive(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip.id = "tip1"
	v := WithTooltip(w, WithTooltipCfg{
		ID:      "tip1",
		Text:    "hello",
		Content: []View{Text(TextCfg{Text: "trigger"})},
	})
	layout := GenerateViewLayout(v, w)
	// Trigger content + tooltip popup.
	if len(layout.Children) < 2 {
		t.Errorf("expected >=2 children, got %d",
			len(layout.Children))
	}
}

func TestWithTooltipHidesPopupWhenInactive(t *testing.T) {
	w := &Window{}
	v := WithTooltip(w, WithTooltipCfg{
		ID:      "tip1",
		Text:    "hello",
		Content: []View{Text(TextCfg{Text: "trigger"})},
	})
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) != 1 {
		t.Errorf("expected 1 child, got %d",
			len(layout.Children))
	}
}

func TestWithTooltipAmendStartsHover(t *testing.T) {
	w := &Window{}
	w.viewState.mousePosX = 50
	w.viewState.mousePosY = 50

	v := WithTooltip(w, WithTooltipCfg{
		ID:      "tip1",
		Text:    "hello",
		Content: []View{Text(TextCfg{Text: "trigger"})},
	})
	layout := GenerateViewLayout(v, w)
	layout.Shape.X = 0
	layout.Shape.Y = 0
	layout.Shape.Width = 100
	layout.Shape.Height = 100

	if layout.Shape.Events == nil ||
		layout.Shape.Events.AmendLayout == nil {
		t.Fatal("expected AmendLayout handler")
	}
	layout.Shape.Events.AmendLayout(&layout, w)

	if w.viewState.tooltip.hoverID != "tip1" {
		t.Errorf("expected hoverID=tip1, got %q",
			w.viewState.tooltip.hoverID)
	}
	if w.viewState.tooltip.hoverStart.IsZero() {
		t.Error("expected non-zero hoverStart")
	}
}

func TestWithTooltipAmendClearsOnLeave(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip.hoverID = "tip1"
	w.viewState.tooltip.hoverStart = time.Now()
	w.viewState.tooltip.id = "tip1"
	w.viewState.mousePosX = 200
	w.viewState.mousePosY = 200

	v := WithTooltip(w, WithTooltipCfg{
		ID:      "tip1",
		Text:    "hello",
		Content: []View{Text(TextCfg{Text: "trigger"})},
	})
	layout := GenerateViewLayout(v, w)
	layout.Shape.X = 0
	layout.Shape.Y = 0
	layout.Shape.Width = 100
	layout.Shape.Height = 100

	layout.Shape.Events.AmendLayout(&layout, w)

	if w.viewState.tooltip.hoverID != "" {
		t.Errorf("expected empty hoverID, got %q",
			w.viewState.tooltip.hoverID)
	}
	if w.viewState.tooltip.id != "" {
		t.Errorf("expected empty tooltip id, got %q",
			w.viewState.tooltip.id)
	}
}

func TestWithTooltipAmendIgnoresOther(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip.hoverID = "other"
	w.viewState.mousePosX = 200
	w.viewState.mousePosY = 200

	v := WithTooltip(w, WithTooltipCfg{
		ID:      "tip1",
		Text:    "hello",
		Content: []View{Text(TextCfg{Text: "trigger"})},
	})
	layout := GenerateViewLayout(v, w)
	layout.Shape.X = 0
	layout.Shape.Y = 0
	layout.Shape.Width = 100
	layout.Shape.Height = 100

	layout.Shape.Events.AmendLayout(&layout, w)

	if w.viewState.tooltip.hoverID != "other" {
		t.Errorf("expected hoverID=other, got %q",
			w.viewState.tooltip.hoverID)
	}
}

func TestWithTooltipAmendSetsIDAfterDelay(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip.hoverID = "tip1"
	w.viewState.tooltip.hoverStart = time.Now().Add(-time.Second)
	w.viewState.mousePosX = 50
	w.viewState.mousePosY = 50

	v := WithTooltip(w, WithTooltipCfg{
		ID:      "tip1",
		Text:    "hello",
		Delay:   500 * time.Millisecond,
		Content: []View{Text(TextCfg{Text: "trigger"})},
	})
	layout := GenerateViewLayout(v, w)
	layout.Shape.X = 0
	layout.Shape.Y = 0
	layout.Shape.Width = 100
	layout.Shape.Height = 100

	layout.Shape.Events.AmendLayout(&layout, w)

	if w.viewState.tooltip.id != "tip1" {
		t.Errorf("expected tooltip id=tip1, got %q",
			w.viewState.tooltip.id)
	}
}

func TestWithTooltipDefaultsID(t *testing.T) {
	w := &Window{}
	w.viewState.tooltip.id = "hello"
	v := WithTooltip(w, WithTooltipCfg{
		Text:    "hello",
		Content: []View{Text(TextCfg{Text: "trigger"})},
	})
	layout := GenerateViewLayout(v, w)
	// ID defaults to Text; tooltip should show.
	if len(layout.Children) < 2 {
		t.Errorf("expected >=2 children, got %d",
			len(layout.Children))
	}
}
