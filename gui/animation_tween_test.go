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
	deferred := make([]func(*Window), 0, 4)
	updateTween(tw, nil, &deferred)
	for _, cb := range deferred {
		cb(nil)
	}
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
	deferred := make([]func(*Window), 0, 4)
	updateTween(tw, nil, &deferred)
	for _, cb := range deferred {
		cb(nil)
	}
	if !done {
		t.Error("OnDone not called")
	}
}

func TestTweenDelaySkips(t *testing.T) {
	tw := NewTweenAnimation("t", 0, 1, func(float32, *Window) {})
	tw.delay = time.Hour
	tw.start = time.Now()
	deferred := make([]func(*Window), 0, 4)
	ok := updateTween(tw, nil, &deferred)
	if ok {
		t.Error("should not update during delay")
	}
}

func TestTweenStopped(t *testing.T) {
	tw := NewTweenAnimation("t", 0, 1, func(float32, *Window) {})
	tw.stopped = true
	deferred := make([]func(*Window), 0, 4)
	ok := updateTween(tw, nil, &deferred)
	if ok {
		t.Error("stopped tween should return false")
	}
}
