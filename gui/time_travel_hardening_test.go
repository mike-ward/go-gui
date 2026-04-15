package gui

import (
	"math"
	"strings"
	"testing"
	"time"
)

// TestRestoreTypeMismatchRecovers confirms a type-mismatched
// namespace entry is skipped via recover instead of crashing
// the entire restore. Simulates generic-type drift by replacing
// the receiving BoundedMap with one of a different value type
// after capture.
func TestRestoreTypeMismatchRecovers(t *testing.T) {
	const ns = "test.type.drift"
	t.Cleanup(func() {
		snapshotableNamespaces.mu.Lock()
		delete(snapshotableNamespaces.set, ns)
		snapshotableNamespaces.mu.Unlock()
	})
	RegisterNamespaceSnapshot(ns)

	w := &Window{state: &testState{}, focused: true}
	w.history = newSnapshotRing(1 << 20)
	origin := StateMap[string, int](w, ns, capFew)
	origin.Set("k", 1)

	w.EventFn(&Event{Type: EventMouseDown, MouseButton: MouseLeft})

	// Swap the namespace's BoundedMap to a different value type
	// so restoreAny will panic; safeRestoreAny must recover.
	w.viewState.registry.maps[ns] = NewBoundedMap[string, string](capFew)

	w.PostRestore(0)
	w.flushCommands()
	// No panic and the second namespace (untouched) keeps whatever
	// is in it — restore simply skipped this one.
	other, _ := w.viewState.registry.maps[ns].(*BoundedMap[string, string])
	if other == nil {
		t.Fatal("namespace map unexpectedly replaced")
	}
}

// TestSnapshotRingEntrySizeClamped caps a misbehaving Size()
// return so totalBytes cannot overflow.
func TestSnapshotRingEntrySizeClamped(t *testing.T) {
	r := newSnapshotRing(1 << 40)
	r.push((&oversizedState{n: math.MaxInt}).Snapshot(),
		time.Now(), "big")
	e, _ := r.last()
	if e.bytes != snapshotMaxEntryBytes {
		t.Fatalf("bytes = %d, want %d", e.bytes, snapshotMaxEntryBytes)
	}
	if r.bytes() != snapshotMaxEntryBytes {
		t.Fatalf("totalBytes = %d, want %d",
			r.bytes(), snapshotMaxEntryBytes)
	}
}

type oversizedState struct{ n int }

func (s *oversizedState) Snapshot() any { return &oversizedState{n: s.n} }
func (s *oversizedState) Restore(v any) { *s = *v.(*oversizedState) }
func (s *oversizedState) Size() int     { return s.n }

// TestControllerJumpFromSliderNaN rejects NaN and Inf before
// the int conversion, leaving the cursor unchanged.
func TestControllerJumpFromSliderNaN(t *testing.T) {
	w, _ := newFixtureApp(t, 3)
	c := &TimeTravelController{App: w, Cursor: 1}

	c.jumpFromSlider(float32(math.NaN()))
	w.flushCommands()
	if c.Cursor != 1 {
		t.Fatalf("NaN cursor = %d, want 1", c.Cursor)
	}

	c.jumpFromSlider(float32(math.Inf(1)))
	w.flushCommands()
	if c.Cursor != 1 {
		t.Fatalf("+Inf cursor = %d, want 1", c.Cursor)
	}

	c.jumpFromSlider(float32(math.Inf(-1)))
	w.flushCommands()
	if c.Cursor != 1 {
		t.Fatalf("-Inf cursor = %d, want 1", c.Cursor)
	}

	c.jumpFromSlider(2.0)
	w.flushCommands()
	if c.Cursor != 2 {
		t.Fatalf("normal cursor = %d, want 2", c.Cursor)
	}
}

// TestControllerJumpFromSliderNil tolerates a nil receiver.
func TestControllerJumpFromSliderNil(t *testing.T) {
	var c *TimeTravelController
	c.jumpFromSlider(1.0)
}

// TestBoundedMapCloneRestoreRoundTrip round-trips entries
// through cloneAny/restoreAny.
func TestBoundedMapCloneRestoreRoundTrip(t *testing.T) {
	src := NewBoundedMap[string, int](capFew)
	src.Set("a", 1)
	src.Set("b", 2)
	src.Set("c", 3)

	clone := src.cloneAny().(*BoundedMap[string, int])
	src.Set("a", 999)
	src.Delete("b")

	if v, _ := clone.Get("a"); v != 1 {
		t.Fatalf("clone decoupled: a = %d, want 1", v)
	}
	if v, _ := clone.Get("b"); v != 2 {
		t.Fatalf("clone decoupled: b = %d, want 2", v)
	}
	if clone.Len() != 3 {
		t.Fatalf("clone len = %d, want 3", clone.Len())
	}

	dst := NewBoundedMap[string, int](capFew)
	dst.Set("zombie", 42)
	dst.restoreAny(clone)
	if _, ok := dst.Get("zombie"); ok {
		t.Fatal("restore did not clear existing entries")
	}
	if v, _ := dst.Get("a"); v != 1 {
		t.Fatalf("restore a = %d, want 1", v)
	}
}

// TestBoundedMapCloneEmpty clones an empty map to an empty map.
func TestBoundedMapCloneEmpty(t *testing.T) {
	src := NewBoundedMap[int, int](capFew)
	clone := src.cloneAny().(*BoundedMap[int, int])
	if clone.Len() != 0 {
		t.Fatalf("empty clone len = %d, want 0", clone.Len())
	}
}

// TestOpenDebugWindowTruncatesLongTitle caps oversize parent
// titles in the composed debug window title.
func TestOpenDebugWindowTruncatesLongTitle(t *testing.T) {
	app := NewApp()
	w := &Window{state: &testState{}}
	w.app = app
	w.platformID = 1
	w.Config.Title = strings.Repeat("x", 4096)
	w.EnableHistory(0)
	w.OpenDebugWindow()

	select {
	case cfg := <-app.PendingOpen():
		// Composed title is "Time Travel — " + up to 256 bytes
		// of parent. The prefix itself is small; assert total
		// stays well under an unsafe length.
		if len(cfg.Title) > 512 {
			t.Fatalf("title len = %d, want <= 512", len(cfg.Title))
		}
		if !strings.HasPrefix(cfg.Title, "Time Travel — ") {
			t.Fatalf("title prefix wrong: %q", cfg.Title)
		}
	default:
		t.Fatal("no WindowCfg queued")
	}
}

// TestShouldSnapshotSkipsAllBookkeeping covers every event type
// that shouldSnapshot must skip.
func TestShouldSnapshotSkipsAllBookkeeping(t *testing.T) {
	skip := []EventType{
		EventMouseMove, EventIMEComposition,
		EventFocused, EventUnfocused, EventResized,
	}
	for _, et := range skip {
		if shouldSnapshot(et) {
			t.Errorf("shouldSnapshot(%v) = true, want false", et)
		}
	}
	for _, et := range []EventType{
		EventKeyDown, EventMouseDown, EventMouseUp,
		EventMouseScroll, EventChar, EventFileDropped,
	} {
		if !shouldSnapshot(et) {
			t.Errorf("shouldSnapshot(%v) = false, want true", et)
		}
	}
}

// TestEventCauseLabels covers the full cause mapping.
func TestEventCauseLabels(t *testing.T) {
	cases := map[EventType]string{
		EventChar:             "char",
		EventKeyDown:          "key",
		EventMouseDown:        "mouse-down",
		EventMouseUp:          "mouse-up",
		EventMouseScroll:      "scroll",
		EventFileDropped:      "file-drop",
		EventTouchesBegan:     "touch-begin",
		EventTouchesMoved:     "touch-move",
		EventTouchesEnded:     "touch-end",
		EventTouchesCancelled: "touch-cancel",
	}
	for et, want := range cases {
		got := eventCause(&Event{Type: et})
		if got != want {
			t.Errorf("%v = %q, want %q", et, got, want)
		}
	}
	// Nil-safe.
	if got := eventCause(nil); got != "" {
		t.Errorf("nil = %q, want empty", got)
	}
	// Unknown event returns fallback.
	if got := eventCause(&Event{Type: EventType(127)}); got != "event" {
		t.Errorf("unknown fallback = %q, want event", got)
	}
}

// TestFreezeLabel returns distinct strings for frozen / live.
func TestFreezeLabel(t *testing.T) {
	if got := freezeLabel(nil); got != "Freeze" {
		t.Errorf("nil = %q, want Freeze", got)
	}
	w := &Window{}
	w.history = newSnapshotRing(0)
	if got := freezeLabel(w); got != "Freeze" {
		t.Errorf("live = %q, want Freeze", got)
	}
	w.Freeze()
	if got := freezeLabel(w); got != "Frozen" {
		t.Errorf("frozen = %q, want Frozen", got)
	}
}

// TestControllerResumeLiveEmpty on a populated-but-history-len-0
// window leaves Cursor untouched.
func TestControllerResumeLiveEmpty(t *testing.T) {
	w := &Window{state: &testState{}}
	w.history = newSnapshotRing(1 << 20)
	c := &TimeTravelController{App: w, Cursor: 7}
	c.ResumeLive()
	if c.Cursor != 7 {
		t.Fatalf("Cursor = %d, want 7 (empty history leaves cursor)",
			c.Cursor)
	}
}

// TestEnableHistoryShrinkEvicts shrinks the cap and confirms
// oldest entries are evicted to fit.
func TestEnableHistoryShrinkEvicts(t *testing.T) {
	w := &Window{state: &testState{}}
	w.EnableHistory(4 * snapshotDefaultSize)
	for range 4 {
		w.history.push((&testState{}).Snapshot(), time.Now(), "e")
	}
	if w.history.len() != 4 {
		t.Fatalf("pre-shrink len = %d, want 4", w.history.len())
	}
	w.EnableHistory(2 * snapshotDefaultSize)
	if w.history.len() != 2 {
		t.Fatalf("post-shrink len = %d, want 2", w.history.len())
	}
	if w.history.bytes() > 2*snapshotDefaultSize {
		t.Fatalf("bytes = %d over cap", w.history.bytes())
	}
}

// TestTimeTravelScrubIntegration drives a series of events
// through a window with history enabled, scrubs to an earlier
// snapshot, and asserts user state reverts. End-to-end check
// of the capture-post-restore flow without a live backend.
func TestTimeTravelScrubIntegration(t *testing.T) {
	s := &testState{}
	w := &Window{state: s, focused: true}
	w.EnableHistory(0)

	// Drive 5 mouse-down events that each bump the counter.
	for i := 1; i <= 5; i++ {
		s.counter = i
		w.EventFn(&Event{Type: EventMouseDown, MouseButton: MouseLeft})
	}
	if n := w.history.len(); n != 5 {
		t.Fatalf("history len = %d, want 5", n)
	}
	if s.counter != 5 {
		t.Fatalf("live counter = %d, want 5", s.counter)
	}

	// Scrub back to snapshot 1 (counter was 2 when that entry
	// was captured AFTER the second event; snapshot 0 has
	// counter == 1). Jump to 0.
	c := &TimeTravelController{App: w, Cursor: 4}
	c.Jump(0)
	w.flushCommands()
	if s.counter != 1 {
		t.Fatalf("scrubbed counter = %d, want 1", s.counter)
	}
	if !w.IsFrozen() {
		t.Fatal("scrubbing should auto-freeze")
	}

	// ResumeLive restores the newest entry and unfreezes.
	c.ResumeLive()
	w.flushCommands()
	// Resume itself does not replay snapshots; it only unpins
	// time and unfreezes. Cursor snaps to newest entry.
	if c.Cursor != 4 {
		t.Fatalf("Cursor = %d, want 4 after ResumeLive", c.Cursor)
	}
	if w.IsFrozen() {
		t.Fatal("still frozen after ResumeLive")
	}

	// Further events continue to append history.
	s.counter = 6
	w.EventFn(&Event{Type: EventMouseDown, MouseButton: MouseLeft})
	if n := w.history.len(); n != 6 {
		t.Fatalf("post-resume len = %d, want 6", n)
	}
}
