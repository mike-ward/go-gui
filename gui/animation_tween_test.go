package gui

import (
	"testing"
	"time"
)

func TestTweenDefaults(t *testing.T) {
	tw := NewTweenAnimation("test", 0, 100, func(float32, *Window) {})
	if tw.Duration != 300*time.Millisecond {
		t.Errorf("default duration = %v", tw.Duration)
	}
	if tw.AnimID != "test" {
		t.Error("id mismatch")
	}
}

func TestTweenCompletesAtTo(t *testing.T) {
	var got float32
	tw := NewTweenAnimation("t", 10, 50, func(v float32, _ *Window) { got = v })
	tw.start = time.Now().Add(-time.Second)
	deferred := make([]queuedCommand, 0, 4)
	ac := newAnimationCommands(&deferred)
	updateTween(tw, &ac)
	runQueuedCommands(deferred)
	if got != 50 {
		t.Errorf("final value = %f, want 50", got)
	}
	if !tw.stopped {
		t.Error("should be stopped")
	}
}

func TestTweenOnDoneCalled(t *testing.T) {
	done := false
	tw := NewTweenAnimation("t", 0, 1, func(float32, *Window) {})
	tw.OnDone = func(*Window) { done = true }
	tw.start = time.Now().Add(-time.Second)
	deferred := make([]queuedCommand, 0, 4)
	ac := newAnimationCommands(&deferred)
	updateTween(tw, &ac)
	runQueuedCommands(deferred)
	if !done {
		t.Error("OnDone not called")
	}
}

func TestTweenRefreshKind(t *testing.T) {
	tw := NewTweenAnimation("t", 0, 1, func(float32, *Window) {})
	if tw.RefreshKind() != AnimationRefreshLayout {
		t.Errorf("RefreshKind = %d, want %d", tw.RefreshKind(), AnimationRefreshLayout)
	}
}

func TestTweenIsStopped(t *testing.T) {
	tw := NewTweenAnimation("t", 0, 1, func(float32, *Window) {})
	if tw.IsStopped() {
		t.Error("new tween should not be stopped")
	}
	tw.stopped = true
	if !tw.IsStopped() {
		t.Error("should report stopped")
	}
}

func TestTweenUpdateInterface(t *testing.T) {
	var got float32
	tw := NewTweenAnimation("t", 10, 50, func(v float32, _ *Window) { got = v })
	tw.start = time.Now().Add(-time.Second)
	deferred := make([]queuedCommand, 0, 4)
	ac := newAnimationCommands(&deferred)
	ok := tw.Update(nil, 0, &ac)
	if !ok {
		t.Error("Update should return true")
	}
	runQueuedCommands(deferred)
	if got != 50 {
		t.Errorf("got %f, want 50", got)
	}
}

func TestTweenStopped(t *testing.T) {
	tw := NewTweenAnimation("t", 0, 1, func(float32, *Window) {})
	tw.stopped = true
	deferred := make([]queuedCommand, 0, 4)
	ac := newAnimationCommands(&deferred)
	ok := updateTween(tw, &ac)
	if ok {
		t.Error("stopped tween should return false")
	}
}
