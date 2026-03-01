package gui

import (
	"testing"
	"time"
)

func TestAnimationAdd(t *testing.T) {
	w := &Window{}
	tw := NewTweenAnimation("t1", 0, 1, func(float32, *Window) {})
	w.AnimationAdd(tw)
	if !w.HasAnimation("t1") {
		t.Error("animation not found")
	}
}

func TestAnimationRemove(t *testing.T) {
	w := &Window{}
	tw := NewTweenAnimation("t1", 0, 1, func(float32, *Window) {})
	w.AnimationAdd(tw)
	w.AnimationRemove("t1")
	if w.HasAnimation("t1") {
		t.Error("animation should be removed")
	}
}

func TestAnimationReplace(t *testing.T) {
	w := &Window{}
	tw1 := NewTweenAnimation("t1", 0, 50, func(float32, *Window) {})
	tw2 := NewTweenAnimation("t1", 0, 100, func(float32, *Window) {})
	w.AnimationAdd(tw1)
	w.AnimationAdd(tw2)
	a := w.animations["t1"].(*TweenAnimation)
	if a.To != 100 {
		t.Error("should replace with new animation")
	}
}

func TestUpdateAnimate(t *testing.T) {
	called := false
	a := &Animate{
		AnimateID: "a",
		Callback: func(_ *Animate, _ *Window) {
			called = true
		},
		Delay: 0,
		start: time.Now().Add(-time.Second),
	}
	deferred := make([]func(*Window), 0, 4)
	ok := updateAnimate(a, nil, &deferred)
	if !ok {
		t.Error("should return true")
	}
	for _, cb := range deferred {
		cb(nil)
	}
	if !called {
		t.Error("callback not called")
	}
	if !a.stopped {
		t.Error("should be stopped (no repeat)")
	}
}

func TestUpdateBlinkCursor(t *testing.T) {
	b := NewBlinkCursorAnimation()
	b.start = time.Now().Add(-time.Second)
	w := &Window{}
	ok := updateBlinkCursor(b, w)
	if !ok {
		t.Error("should return true after delay")
	}
}
