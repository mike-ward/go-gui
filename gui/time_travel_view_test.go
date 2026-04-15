package gui

import (
	"testing"
	"time"
)

// newFixtureApp returns an app-style window with history filled
// to n entries and cursor at the newest entry.
func newFixtureApp(t *testing.T, n int) (*Window, *testState) {
	t.Helper()
	s := &testState{}
	w := &Window{state: s, focused: true}
	w.history = newSnapshotRing(1 << 20)
	for i := range n {
		s.counter = i + 1
		w.history.push(s.Snapshot(), time.Now(), "e")
	}
	return w, s
}

// TestControllerLenBytes reports ring state.
func TestControllerLenBytes(t *testing.T) {
	w, _ := newFixtureApp(t, 3)
	c := &TimeTravelController{App: w}
	if got := c.Len(); got != 3 {
		t.Fatalf("Len = %d, want 3", got)
	}
	if c.Bytes() == 0 {
		t.Fatal("Bytes = 0, want > 0")
	}
}

// TestControllerJumpClamps clamps negative and overflow idx.
func TestControllerJumpClamps(t *testing.T) {
	w, s := newFixtureApp(t, 3)
	c := &TimeTravelController{App: w}
	c.Jump(-5)
	w.flushCommands()
	if c.Cursor != 0 {
		t.Fatalf("Cursor = %d, want 0 (clamped low)", c.Cursor)
	}
	if s.counter != 1 {
		t.Fatalf("state counter = %d, want 1", s.counter)
	}
	c.Jump(999)
	w.flushCommands()
	if c.Cursor != 2 {
		t.Fatalf("Cursor = %d, want 2 (clamped high)", c.Cursor)
	}
	if s.counter != 3 {
		t.Fatalf("state counter = %d, want 3", s.counter)
	}
}

// TestControllerJumpFreezesApp confirms a jump auto-freezes.
func TestControllerJumpFreezesApp(t *testing.T) {
	w, _ := newFixtureApp(t, 2)
	c := &TimeTravelController{App: w}
	c.Jump(0)
	if !w.IsFrozen() {
		t.Fatal("expected app window frozen after Jump")
	}
}

// TestControllerStepBackForward walks the cursor.
func TestControllerStepBackForward(t *testing.T) {
	w, s := newFixtureApp(t, 3)
	c := &TimeTravelController{App: w, Cursor: 2}
	c.StepBack()
	w.flushCommands()
	if c.Cursor != 1 || s.counter != 2 {
		t.Fatalf("after StepBack: cursor=%d counter=%d", c.Cursor, s.counter)
	}
	c.StepForward()
	w.flushCommands()
	if c.Cursor != 2 || s.counter != 3 {
		t.Fatalf("after StepForward: cursor=%d counter=%d", c.Cursor, s.counter)
	}
}

// TestControllerFirstLast jumps to ends.
func TestControllerFirstLast(t *testing.T) {
	w, s := newFixtureApp(t, 4)
	c := &TimeTravelController{App: w, Cursor: 2}
	c.First()
	w.flushCommands()
	if c.Cursor != 0 || s.counter != 1 {
		t.Fatalf("First: cursor=%d counter=%d", c.Cursor, s.counter)
	}
	c.Last()
	w.flushCommands()
	if c.Cursor != 3 || s.counter != 4 {
		t.Fatalf("Last: cursor=%d counter=%d", c.Cursor, s.counter)
	}
}

// TestControllerResumeLive unfreezes and snaps cursor to newest.
func TestControllerResumeLive(t *testing.T) {
	w, _ := newFixtureApp(t, 3)
	c := &TimeTravelController{App: w}
	c.Jump(0)
	c.ResumeLive()
	if w.IsFrozen() {
		t.Fatal("still frozen after ResumeLive")
	}
	if c.Cursor != 2 {
		t.Fatalf("Cursor = %d, want 2 (newest)", c.Cursor)
	}
	if w.virtualNow.Load() != nil {
		t.Fatal("virtualNow not cleared")
	}
}

// TestControllerToggleFreeze flips freeze on and off.
func TestControllerToggleFreeze(t *testing.T) {
	w, _ := newFixtureApp(t, 2)
	c := &TimeTravelController{App: w, Cursor: 1}
	c.ToggleFreeze()
	if !w.IsFrozen() {
		t.Fatal("ToggleFreeze should freeze")
	}
	c.ToggleFreeze()
	if w.IsFrozen() {
		t.Fatal("second ToggleFreeze should unfreeze")
	}
}

// TestControllerEmptyRing leaves state unchanged.
func TestControllerEmptyRing(t *testing.T) {
	w := &Window{state: &testState{}}
	w.history = newSnapshotRing(1 << 20)
	c := &TimeTravelController{App: w}
	c.Jump(0)
	c.First()
	c.Last()
	if c.Cursor != 0 {
		t.Fatalf("Cursor = %d, want 0 (untouched)", c.Cursor)
	}
	if w.IsFrozen() {
		t.Fatal("empty ring should not auto-freeze")
	}
}

// TestControllerNilSafe tolerates nil receivers and nil app.
func TestControllerNilSafe(t *testing.T) {
	var c *TimeTravelController
	c.Jump(0)
	c.StepBack()
	c.StepForward()
	c.First()
	c.Last()
	c.ResumeLive()
	c.ToggleFreeze()
	if got := c.Len(); got != 0 {
		t.Fatalf("nil Len = %d", got)
	}

	c2 := &TimeTravelController{}
	c2.Jump(0)
	if c2.Cursor != 0 {
		t.Fatalf("nil-App Cursor = %d", c2.Cursor)
	}
}

// TestControllerCauseLabel reports the entry's cause.
func TestControllerCauseLabel(t *testing.T) {
	w := &Window{state: &testState{}}
	w.history = newSnapshotRing(1 << 20)
	w.history.push((&testState{}).Snapshot(), time.Now(), "mouse-down")
	c := &TimeTravelController{App: w, Cursor: 0}
	if got := c.Cause(); got != "mouse-down" {
		t.Fatalf("Cause = %q, want mouse-down", got)
	}
}

// TestControllerView returns a non-nil view for empty and
// populated rings.
func TestControllerView(t *testing.T) {
	c := &TimeTravelController{}
	if c.View() == nil {
		t.Fatal("empty View nil")
	}
	w, _ := newFixtureApp(t, 2)
	c.App = w
	if c.View() == nil {
		t.Fatal("populated View nil")
	}
}
