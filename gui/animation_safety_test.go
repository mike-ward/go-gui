package gui

import "testing"

func TestUpdateAnimateNilCallbackStops(t *testing.T) {
	a := &Animate{AnimID: "a"}
	deferred := make([]queuedCommand, 0, 1)
	ok := updateAnimate(a, &deferred)
	if ok {
		t.Fatal("expected no update when callback is nil")
	}
	if !a.stopped {
		t.Fatal("expected animate to stop when callback is nil")
	}
}

func TestUpdateTweenNilOnValueStops(t *testing.T) {
	tw := NewTweenAnimation("t", 0, 1, nil)
	deferred := make([]queuedCommand, 0, 1)
	ok := updateTween(tw, &deferred)
	if ok {
		t.Fatal("expected no update when OnValue is nil")
	}
	if !tw.stopped {
		t.Fatal("expected tween to stop when OnValue is nil")
	}
}

func TestUpdateKeyframeNilOnValueStops(t *testing.T) {
	kf := NewKeyframeAnimation("k", []Keyframe{{At: 0, Value: 0}}, nil)
	deferred := make([]queuedCommand, 0, 1)
	ok := updateKeyframe(kf, &deferred)
	if ok {
		t.Fatal("expected no update when OnValue is nil")
	}
	if !kf.stopped {
		t.Fatal("expected keyframe to stop when OnValue is nil")
	}
}

func TestUpdateSpringNilOnValueStops(t *testing.T) {
	sp := NewSpringAnimation("s", nil)
	sp.SpringTo(0, 1)
	deferred := make([]queuedCommand, 0, 1)
	ok := updateSpring(sp, 0.016, &deferred)
	if ok {
		t.Fatal("expected no update when OnValue is nil")
	}
	if !sp.stopped {
		t.Fatal("expected spring to stop when OnValue is nil")
	}
}
