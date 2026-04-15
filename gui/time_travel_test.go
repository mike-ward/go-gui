package gui

import (
	"sync"
	"testing"
	"time"
)

// testState is a minimal Snapshotter used across ring tests.
type testState struct {
	counter int
	text    string
}

func (s *testState) Snapshot() any {
	return &testState{counter: s.counter, text: s.text}
}

func (s *testState) Restore(v any) {
	src := v.(*testState)
	s.counter = src.counter
	s.text = src.text
}

// sizedState reports its own byte cost to exercise snapshotSizer.
type sizedState struct {
	size int
}

func (s *sizedState) Snapshot() any { return &sizedState{size: s.size} }
func (s *sizedState) Restore(v any) { *s = *v.(*sizedState) }
func (s *sizedState) Size() int     { return s.size }

// TestSnapshotRingPush verifies basic push/len/bytes/at/last.
func TestSnapshotRingPush(t *testing.T) {
	r := newSnapshotRing(1 << 20)
	now := time.Now()
	r.push((&testState{counter: 1}).Snapshot(), now, "evt-1")
	r.push((&testState{counter: 2}).Snapshot(), now.Add(time.Millisecond), "evt-2")
	if got := r.len(); got != 2 {
		t.Fatalf("len = %d, want 2", got)
	}
	if got := r.bytes(); got != 2*snapshotDefaultSize {
		t.Fatalf("bytes = %d, want %d", got, 2*snapshotDefaultSize)
	}
	e, ok := r.at(0)
	if !ok || e.cause != "evt-1" {
		t.Fatalf("at(0) = %+v ok=%v", e, ok)
	}
	e, ok = r.last()
	if !ok || e.cause != "evt-2" {
		t.Fatalf("last = %+v ok=%v", e, ok)
	}
	if _, ok := r.at(5); ok {
		t.Fatal("at(5) should fail")
	}
}

// TestSnapshotRingRestoreRoundTrip ensures Snapshot/Restore
// round-trips value fields through a ring entry.
func TestSnapshotRingRestoreRoundTrip(t *testing.T) {
	s := &testState{counter: 7, text: "hello"}
	r := newSnapshotRing(1 << 20)
	r.push(s.Snapshot(), time.Now(), "init")
	s.counter = 42
	s.text = "changed"
	e, ok := r.at(0)
	if !ok {
		t.Fatal("missing entry")
	}
	s.Restore(e.snap)
	if s.counter != 7 || s.text != "hello" {
		t.Fatalf("restored = %+v, want counter=7 text=hello", s)
	}
}

// TestSnapshotRingByteCap confirms eviction keeps total bytes
// under the cap.
func TestSnapshotRingByteCap(t *testing.T) {
	const cap = 1024
	r := newSnapshotRing(cap)
	for range 10 {
		r.push(&sizedState{size: 400}, time.Now(), "e")
	}
	if got := r.bytes(); got > cap {
		t.Fatalf("bytes = %d, want <= %d", got, cap)
	}
	// 400 * 2 = 800 fits under 1024; 400 * 3 = 1200 does not.
	if got := r.len(); got != 2 {
		t.Fatalf("len = %d, want 2 (400-byte entries under 1024 cap)", got)
	}
}

// TestSnapshotRingEvictsOldest checks that eviction is FIFO:
// the surviving entries are the newest ones.
func TestSnapshotRingEvictsOldest(t *testing.T) {
	r := newSnapshotRing(2 * snapshotDefaultSize)
	r.push((&testState{counter: 1}).Snapshot(), time.Now(), "a")
	r.push((&testState{counter: 2}).Snapshot(), time.Now(), "b")
	r.push((&testState{counter: 3}).Snapshot(), time.Now(), "c")
	if got := r.len(); got != 2 {
		t.Fatalf("len = %d, want 2", got)
	}
	e0, _ := r.at(0)
	e1, _ := r.at(1)
	if e0.cause != "b" || e1.cause != "c" {
		t.Fatalf("survivors = %q, %q, want b, c", e0.cause, e1.cause)
	}
}

// TestSnapshotRingClear drops all entries and zeros totals.
func TestSnapshotRingClear(t *testing.T) {
	r := newSnapshotRing(1 << 20)
	r.push((&testState{}).Snapshot(), time.Now(), "x")
	r.clear()
	if r.len() != 0 || r.bytes() != 0 {
		t.Fatalf("after clear: len=%d bytes=%d", r.len(), r.bytes())
	}
}

// TestSnapshotRingNilSnap ignores nil pushes.
func TestSnapshotRingNilSnap(t *testing.T) {
	r := newSnapshotRing(1 << 20)
	r.push(nil, time.Now(), "nope")
	if r.len() != 0 {
		t.Fatalf("nil push should not add entry")
	}
}

// TestSnapshotRingDefaultCap applies defaultHistoryBytes when
// maxBytes is non-positive.
func TestSnapshotRingDefaultCap(t *testing.T) {
	r := newSnapshotRing(0)
	if r.maxBytes != defaultHistoryBytes {
		t.Fatalf("maxBytes = %d, want %d", r.maxBytes, defaultHistoryBytes)
	}
}

// TestSnapshotRingSizerZeroFallback treats Size()<=0 as the
// default estimate, so a misbehaving Size() cannot starve the
// byte cap.
func TestSnapshotRingSizerZeroFallback(t *testing.T) {
	r := newSnapshotRing(2 * snapshotDefaultSize)
	r.push(&sizedState{size: 0}, time.Now(), "z")
	if got := r.bytes(); got != snapshotDefaultSize {
		t.Fatalf("bytes = %d, want %d", got, snapshotDefaultSize)
	}
}

// TestSnapshotRingConcurrent exercises the RWMutex under -race.
func TestSnapshotRingConcurrent(t *testing.T) {
	r := newSnapshotRing(1 << 20)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := range 1000 {
			r.push((&testState{counter: i}).Snapshot(), time.Now(), "w")
		}
	}()
	go func() {
		defer wg.Done()
		for range 1000 {
			_ = r.len()
			_ = r.bytes()
			_, _ = r.at(0)
			_, _ = r.last()
		}
	}()
	wg.Wait()
}
