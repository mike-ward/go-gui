package gui

import (
	"testing"
	"time"
)

func TestSpringTo(t *testing.T) {
	sp := NewSpringAnimation("s", func(float32, *Window) {})
	sp.SpringTo(10, 100)
	if sp.state.position != 10 {
		t.Errorf("position = %f", sp.state.position)
	}
	if sp.state.target != 100 {
		t.Errorf("target = %f", sp.state.target)
	}
}

func TestSpringRetarget(t *testing.T) {
	sp := NewSpringAnimation("s", func(float32, *Window) {})
	sp.SpringTo(0, 50)
	sp.state.velocity = 5
	sp.Retarget(200)
	if sp.state.target != 200 {
		t.Error("target not updated")
	}
	if sp.state.velocity != 5 {
		t.Error("velocity should be preserved")
	}
}

func TestSpringSettles(t *testing.T) {
	var got float32
	sp := NewSpringAnimation("s", func(v float32, _ *Window) { got = v })
	sp.Config = SpringStiff
	sp.SpringTo(100, 100) // already at target
	sp.start = time.Now().Add(-time.Second)
	deferred := make([]queuedCommand, 0, 4)
	updateSpring(sp, 0.016, &deferred)
	runQueuedCommands(deferred)
	if got != 100 {
		t.Errorf("expected 100, got %f", got)
	}
	if !sp.stopped {
		t.Error("should be at rest")
	}
}

func TestSpringStoppedSkips(t *testing.T) {
	sp := NewSpringAnimation("s", func(float32, *Window) {})
	sp.stopped = true
	deferred := make([]queuedCommand, 0, 4)
	ok := updateSpring(sp, 0.016, &deferred)
	if ok {
		t.Error("stopped spring should return false")
	}
}

func TestSpringDelaySkips(t *testing.T) {
	sp := NewSpringAnimation("s", func(float32, *Window) {})
	sp.SpringTo(0, 100)
	sp.delay = time.Hour
	sp.start = time.Now()
	deferred := make([]queuedCommand, 0, 4)
	ok := updateSpring(sp, 0.016, &deferred)
	if ok {
		t.Error("should skip during delay")
	}
}
