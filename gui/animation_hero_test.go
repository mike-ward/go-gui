package gui

import (
	"testing"
	"time"
)

func TestCaptureHeroSnapshots(t *testing.T) {
	layout := Layout{
		Shape: &Shape{ID: "root"},
		Children: []Layout{
			{Shape: &Shape{ID: "hero1", Hero: true, X: 10, Y: 20, Width: 100, Height: 50}},
			{Shape: &Shape{ID: "nothero", X: 5, Y: 5, Width: 10, Height: 10}},
			{Shape: &Shape{ID: "hero2", Hero: true, X: 30, Y: 40, Width: 200, Height: 100}},
		},
	}
	snaps := captureHeroSnapshots(layout)
	if len(snaps) != 2 {
		t.Errorf("got %d snapshots, want 2", len(snaps))
	}
	if s, ok := snaps["hero1"]; !ok || s.x != 10 {
		t.Error("hero1 snapshot wrong")
	}
}

func TestHeroTransitionUpdate(t *testing.T) {
	ht := NewHeroTransition(HeroTransitionCfg{})
	ht.start = time.Now().Add(-time.Second)
	deferred := make([]func(*Window), 0, 4)
	ok := updateHeroTransition(ht, nil, &deferred)
	if !ok {
		t.Error("should update")
	}
	if !ht.stopped {
		t.Error("should be stopped after duration")
	}
	if ht.progress != 1.0 {
		t.Errorf("progress = %f, want 1.0", ht.progress)
	}
}

func TestApplyHeroRecursive(t *testing.T) {
	layout := Layout{
		Shape: &Shape{ID: "h", Hero: true, X: 100, Y: 100, Width: 200, Height: 200, Opacity: 1},
	}
	outgoing := map[string]heroSnapshot{
		"h": {x: 0, y: 0, width: 100, height: 100},
	}
	incoming := map[string]heroSnapshot{
		"h": {x: 100, y: 100, width: 200, height: 200},
	}
	// progress=0 → morphProgress=0 → should be at outgoing position
	applyHeroRecursive(&layout, 0, outgoing, incoming)
	if layout.Shape.X != 0 {
		t.Errorf("X = %f, want 0", layout.Shape.X)
	}
}
