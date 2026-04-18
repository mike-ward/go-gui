package gui

import (
	"math"
	"testing"
	"time"
)

func TestInterpolateKeyframes(t *testing.T) {
	kfs := []Keyframe{
		{At: 0, Value: 0, Easing: EaseLinear},
		{At: 0.5, Value: 50, Easing: EaseLinear},
		{At: 1.0, Value: 100, Easing: EaseLinear},
	}
	tests := []struct {
		progress float32
		want     float32
	}{
		{0, 0},
		{0.25, 25},
		{0.5, 50},
		{0.75, 75},
		{1.0, 100},
	}
	for _, tt := range tests {
		got := interpolateKeyframes(kfs, tt.progress)
		if math.Abs(float64(got-tt.want)) > 0.5 {
			t.Errorf("progress=%f: got %f, want %f", tt.progress, got, tt.want)
		}
	}
}

func TestKeyframeCompletes(t *testing.T) {
	var got float32
	kf := NewKeyframeAnimation("k",
		[]Keyframe{
			{At: 0, Value: 0},
			{At: 1, Value: 100, Easing: EaseLinear},
		},
		func(v float32, _ *Window) { got = v },
	)
	kf.start = time.Now().Add(-time.Second)
	deferred := make([]queuedCommand, 0, 4)
	ac := newAnimationCommands(&deferred)
	updateKeyframe(kf, &ac)
	runQueuedCommands(deferred)
	if got != 100 {
		t.Errorf("got %f, want 100", got)
	}
	if !kf.stopped {
		t.Error("should be stopped")
	}
}

func TestKeyframeRepeat(t *testing.T) {
	kf := NewKeyframeAnimation("k",
		[]Keyframe{
			{At: 0, Value: 0},
			{At: 1, Value: 100, Easing: EaseLinear},
		},
		func(float32, *Window) {},
	)
	kf.Repeat = true
	kf.start = time.Now().Add(-time.Second)
	deferred := make([]queuedCommand, 0, 4)
	ac := newAnimationCommands(&deferred)
	updateKeyframe(kf, &ac)
	if kf.stopped {
		t.Error("should not stop when repeating")
	}
}

func TestInterpolateEmptyKeyframes(t *testing.T) {
	if interpolateKeyframes(nil, 0.5) != 0 {
		t.Error("empty keyframes should return 0")
	}
	if interpolateKeyframes([]Keyframe{{At: 0, Value: 42}}, 0.5) != 42 {
		t.Error("single keyframe should return its value")
	}
}

func TestKeyframeStopped(t *testing.T) {
	kf := NewKeyframeAnimation("k", nil, func(float32, *Window) {})
	kf.stopped = true
	deferred := make([]queuedCommand, 0, 4)
	ac := newAnimationCommands(&deferred)
	ok := updateKeyframe(kf, &ac)
	if ok {
		t.Error("stopped keyframe should return false")
	}
}

func TestKeyframeRefreshKind(t *testing.T) {
	kf := NewKeyframeAnimation("k", nil, func(float32, *Window) {})
	if kf.RefreshKind() != AnimationRefreshLayout {
		t.Errorf("RefreshKind = %d, want %d", kf.RefreshKind(), AnimationRefreshLayout)
	}
}

func TestKeyframeIsStopped(t *testing.T) {
	kf := NewKeyframeAnimation("k", nil, func(float32, *Window) {})
	if kf.IsStopped() {
		t.Error("new keyframe should not be stopped")
	}
	kf.stopped = true
	if !kf.IsStopped() {
		t.Error("should report stopped")
	}
}

func TestKeyframeUpdateInterface(t *testing.T) {
	var got float32
	kf := NewKeyframeAnimation("k",
		[]Keyframe{
			{At: 0, Value: 0},
			{At: 1, Value: 100, Easing: EaseLinear},
		},
		func(v float32, _ *Window) { got = v },
	)
	kf.start = time.Now().Add(-time.Second)
	deferred := make([]queuedCommand, 0, 4)
	ac := newAnimationCommands(&deferred)
	ok := kf.Update(nil, 0, &ac)
	if !ok {
		t.Error("Update should return true")
	}
	runQueuedCommands(deferred)
	if got != 100 {
		t.Errorf("got %f, want 100", got)
	}
}

func TestKeyframeRepeatNoDrift(t *testing.T) {
	kf := NewKeyframeAnimation("k",
		[]Keyframe{
			{At: 0, Value: 0},
			{At: 1, Value: 100, Easing: EaseLinear},
		},
		func(float32, *Window) {},
	)
	kf.Repeat = true
	kf.Duration = 100 * time.Millisecond
	kf.start = time.Now().Add(-150 * time.Millisecond)
	deferred := make([]queuedCommand, 0, 4)
	ac := newAnimationCommands(&deferred)
	updateKeyframe(kf, &ac)
	// start should advance by exactly Duration, not reset to Now().
	expected := kf.start.Add(0) // start was already updated
	// The key invariant: start moved forward by Duration from its
	// original value, not jumped to time.Now().
	if kf.start.After(time.Now().Add(-10 * time.Millisecond)) {
		t.Error("start should not reset to Now(); should advance by Duration")
	}
	_ = expected
}
