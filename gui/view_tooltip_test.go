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

	if cfg.Color == (Color{}) {
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
	if a.AnimateID != "___tooltip___" {
		t.Errorf("expected ___tooltip___ id, got %q",
			a.AnimateID)
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
