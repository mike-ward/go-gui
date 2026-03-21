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
	deferred := make([]queuedCommand, 0, 4)
	ok := updateTransition(&ht.transitionBase, &deferred)
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
	outgoing := map[string]posSnapshot{
		"h": {x: 0, y: 0, width: 100, height: 100},
	}
	incoming := map[string]posSnapshot{
		"h": {x: 100, y: 100, width: 200, height: 200},
	}
	// progress=0 → morphProgress=0 → should be at outgoing position
	applyHeroRecursive(&layout, 0, outgoing, incoming)
	if layout.Shape.X != 0 {
		t.Errorf("X = %f, want 0", layout.Shape.X)
	}
}

func TestHeroTransitionID(t *testing.T) {
	ht := NewHeroTransition(HeroTransitionCfg{})
	if ht.ID() != heroTransitionID {
		t.Errorf("ID = %q, want %q", ht.ID(), heroTransitionID)
	}
}

func TestHeroTransitionRefreshKind(t *testing.T) {
	ht := NewHeroTransition(HeroTransitionCfg{})
	if ht.RefreshKind() != AnimationRefreshLayout {
		t.Errorf("RefreshKind = %d, want %d", ht.RefreshKind(), AnimationRefreshLayout)
	}
}

func TestHeroTransitionUpdateInterface(t *testing.T) {
	ht := NewHeroTransition(HeroTransitionCfg{})
	ht.start = time.Now().Add(-time.Second)
	deferred := make([]queuedCommand, 0, 4)
	ok := ht.Update(nil, 0, &deferred)
	if !ok {
		t.Error("Update should return true on completion")
	}
	if !ht.stopped {
		t.Error("should be stopped")
	}
}

func TestPropagateOpacity(t *testing.T) {
	layout := Layout{
		Shape: &Shape{Opacity: 1},
		Children: []Layout{
			{Shape: &Shape{Opacity: 1}},
		},
	}
	propagateOpacity(&layout, 0.5)
	if layout.Shape.Opacity != 0.5 {
		t.Errorf("parent opacity = %f, want 0.5", layout.Shape.Opacity)
	}
	if layout.Children[0].Shape.Opacity != 0.5 {
		t.Errorf("child opacity = %f, want 0.5", layout.Children[0].Shape.Opacity)
	}
}

func TestHeroTransitionOnDone(t *testing.T) {
	done := false
	ht := NewHeroTransition(HeroTransitionCfg{
		OnDone: func(*Window) { done = true },
	})
	ht.start = time.Now().Add(-time.Second)
	deferred := make([]queuedCommand, 0, 4)
	updateTransition(&ht.transitionBase, &deferred)
	runQueuedCommands(deferred)
	if !done {
		t.Error("OnDone not called")
	}
}
