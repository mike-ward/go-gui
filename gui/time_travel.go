package gui

import (
	"log"
	"sync"
	"time"
)

// Snapshotter is implemented by user state types that opt into
// time-travel debugging. Snapshot returns a deep-ish copy of the
// receiver; Restore overwrites the receiver from a prior
// Snapshot result. Optional Size reports an approximate byte
// cost for the byte-capped ring buffer; unimplemented types
// use snapshotDefaultSize.
type Snapshotter interface {
	Snapshot() any
	Restore(any)
}

// snapshotSizer is the optional Size() int extension to
// Snapshotter. Types that can cheaply report their heap cost
// implement it to improve byte-cap eviction accuracy.
type snapshotSizer interface {
	Size() int
}

// snapshotDefaultSize is the per-entry byte estimate used when a
// Snapshot value does not implement snapshotSizer. A conservative
// guess; users with fat state should implement Size().
const snapshotDefaultSize = 1024

// snapshotEntry records a single point in history.
type snapshotEntry struct {
	snap  any
	when  time.Time
	cause string
	bytes int
}

// snapshotRing is a byte-capped FIFO ring of Snapshot values.
// Push appends; when total bytes exceed maxBytes, the oldest
// entries are evicted until the buffer fits. Concurrent safe
// via an internal RWMutex — independent of any Window mutex to
// avoid cross-window lock cycles during debug-window reads.
type snapshotRing struct {
	mu         sync.RWMutex
	entries    []snapshotEntry
	totalBytes int
	maxBytes   int
}

// newSnapshotRing returns a ring with the given byte cap. A
// non-positive cap is replaced with defaultHistoryBytes.
func newSnapshotRing(maxBytes int) *snapshotRing {
	if maxBytes <= 0 {
		maxBytes = defaultHistoryBytes
	}
	return &snapshotRing{maxBytes: maxBytes}
}

// defaultHistoryBytes is the default byte cap when
// WindowCfg.HistoryBytes is zero and DebugTimeTravel is on.
const defaultHistoryBytes = 64 << 20 // 64 MiB

// push appends a snapshot captured at when with the given cause
// label. Evicts oldest entries until totalBytes <= maxBytes.
func (r *snapshotRing) push(snap any, when time.Time, cause string) {
	if snap == nil {
		return
	}
	size := snapshotDefaultSize
	if s, ok := snap.(snapshotSizer); ok {
		size = s.Size()
		if size <= 0 {
			size = snapshotDefaultSize
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, snapshotEntry{
		snap:  snap,
		when:  when,
		cause: cause,
		bytes: size,
	})
	r.totalBytes += size
	r.evictLocked()
}

// evictLocked drops oldest entries until totalBytes fits under
// maxBytes. Caller must hold r.mu for writing.
func (r *snapshotRing) evictLocked() {
	if r.totalBytes <= r.maxBytes {
		return
	}
	// Find first index i such that removing [0,i) fits cap.
	i := 0
	for i < len(r.entries) && r.totalBytes > r.maxBytes {
		r.totalBytes -= r.entries[i].bytes
		i++
	}
	if i == 0 {
		return
	}
	// Slide survivors to the front so the underlying array can be
	// reused without unbounded growth. Clear evicted slots so
	// snapshot values become eligible for GC.
	n := copy(r.entries, r.entries[i:])
	for j := n; j < len(r.entries); j++ {
		r.entries[j] = snapshotEntry{}
	}
	r.entries = r.entries[:n]
}

// len returns the number of entries in the ring.
func (r *snapshotRing) len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

// bytes returns the current approximate byte usage.
func (r *snapshotRing) bytes() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.totalBytes
}

// at returns entry idx (0 = oldest), or zero-value and false if
// idx is out of range.
func (r *snapshotRing) at(idx int) (snapshotEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if idx < 0 || idx >= len(r.entries) {
		return snapshotEntry{}, false
	}
	return r.entries[idx], true
}

// last returns the most recent entry, or zero-value and false
// if the ring is empty.
func (r *snapshotRing) last() (snapshotEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := len(r.entries)
	if n == 0 {
		return snapshotEntry{}, false
	}
	return r.entries[n-1], true
}

// shouldSnapshot reports whether an event of this type should
// trigger a post-dispatch snapshot. Plain mouse motion, IME
// composition churn, and focus bookkeeping are skipped —
// they either don't mutate observable state or fire far too
// often to be useful in a scrub timeline.
func shouldSnapshot(t EventType) bool {
	switch t {
	case EventMouseMove,
		EventIMEComposition,
		EventFocused,
		EventUnfocused,
		EventResized:
		return false
	}
	return true
}

// captureSnapshot pushes a post-dispatch snapshot onto the
// window's history if enabled and the user state opts in.
// Safe to call on a nil Window or when history is unset.
func (w *Window) captureSnapshot(e *Event) {
	if w == nil || w.history == nil || e == nil {
		return
	}
	if !shouldSnapshot(e.Type) {
		return
	}
	s, ok := w.state.(Snapshotter)
	if !ok {
		return
	}
	w.history.push(s.Snapshot(), time.Now(), eventCause(e))
}

// eventCause builds a short label describing the event that
// produced a snapshot. Intended for the debug-window timeline;
// the format is informational and may change.
func eventCause(e *Event) string {
	if e == nil {
		return ""
	}
	switch e.Type {
	case EventChar:
		return "char"
	case EventKeyDown:
		return "key"
	case EventMouseDown:
		return "mouse-down"
	case EventMouseUp:
		return "mouse-up"
	case EventMouseScroll:
		return "scroll"
	case EventFileDropped:
		return "file-drop"
	case EventTouchesBegan:
		return "touch-begin"
	case EventTouchesMoved:
		return "touch-move"
	case EventTouchesEnded:
		return "touch-end"
	case EventTouchesCancelled:
		return "touch-cancel"
	}
	return "event"
}

// EnableHistory turns on time-travel snapshot capture with a
// byte cap. A non-positive cap selects defaultHistoryBytes.
// Idempotent — a second call with a different cap updates the
// cap but keeps existing entries. Requires the window's user
// state to implement Snapshotter; otherwise no snapshots are
// ever pushed (but calling the method is still safe).
// Must be called on the main thread (no locking needed because
// history is read/written only from the frame/event path).
func (w *Window) EnableHistory(maxBytes int) {
	if w == nil {
		return
	}
	if maxBytes <= 0 {
		maxBytes = defaultHistoryBytes
	}
	if w.history == nil {
		w.history = newSnapshotRing(maxBytes)
		return
	}
	w.history.mu.Lock()
	w.history.maxBytes = maxBytes
	w.history.evictLocked()
	w.history.mu.Unlock()
}

// HistoryLen returns the number of snapshots currently held.
// Zero when history is disabled. Safe to call from any goroutine.
func (w *Window) HistoryLen() int {
	if w == nil || w.history == nil {
		return 0
	}
	return w.history.len()
}

// OpenDebugWindow queues a secondary Window that hosts the
// time-travel scrubber for this window. Requires the window
// to be part of an App (multi-window mode) and to have
// history enabled; otherwise it logs and returns. Non-blocking:
// the actual window is created on the next App event loop tick.
// Safe to call from any goroutine.
func (w *Window) OpenDebugWindow() {
	if w == nil {
		return
	}
	app := w.App()
	if app == nil {
		log.Println("gui: OpenDebugWindow: no App (single-window mode)")
		return
	}
	if w.history == nil {
		w.EnableHistory(0)
	}
	ctrl := &TimeTravelController{App: w}
	if n := w.history.len(); n > 0 {
		ctrl.Cursor = n - 1
	}
	title := "Time Travel"
	if w.Config.Title != "" {
		title = "Time Travel — " + w.Config.Title
	}
	app.OpenWindow(WindowCfg{
		State:  ctrl,
		Title:  title,
		Width:  480,
		Height: 180,
		OnInit: func(dw *Window) {
			dw.UpdateView(debugWindowView)
		},
	})
}

// debugWindowView is the view generator installed on the
// debug window. Reads the TimeTravelController from its state
// slot and delegates to ctrl.View.
func debugWindowView(dw *Window) View {
	return State[TimeTravelController](dw).View()
}

// Freeze enters time-travel scrub mode. Subsequent user input
// events are dropped by EventFn; the frame loop keeps running
// so the debug window's restore requests take effect and the
// app window continues to repaint. Idempotent. No-op when
// history is disabled. Safe to call from any goroutine.
func (w *Window) Freeze() {
	if w == nil || w.history == nil {
		return
	}
	w.frozen.Store(true)
}

// Resume exits time-travel scrub mode. Clears the virtual
// clock pin so w.Now() returns live time again. Idempotent.
// Safe to call from any goroutine.
func (w *Window) Resume() {
	if w == nil {
		return
	}
	w.frozen.Store(false)
	w.virtualNow.Store(nil)
	w.UpdateWindow()
}

// IsFrozen reports whether the window is currently in a
// read-only time-travel scrub.
func (w *Window) IsFrozen() bool {
	if w == nil {
		return false
	}
	return w.frozen.Load()
}

// PostRestore posts a restore request for history entry idx.
// Intended for cross-goroutine calls from the debug window.
// The request is serialized through the app window's command
// queue so it runs on the main thread under w.mu, avoiding
// torn reads against a concurrent view fn. No-op when history
// is disabled. Safe to call from any goroutine.
func (w *Window) PostRestore(idx int) {
	if w == nil || w.history == nil {
		return
	}
	w.QueueCommand(func(w *Window) {
		w.restoreLocked(idx)
	})
}

// restoreLocked performs the actual Snapshotter.Restore under
// w.mu. Intended to run inside a queued command on the main
// thread. Also pins the virtual clock to the entry's timestamp
// so clock-driven views render consistently with the snapshot.
func (w *Window) restoreLocked(idx int) {
	if w.history == nil {
		return
	}
	entry, ok := w.history.at(idx)
	if !ok {
		return
	}
	s, ok := w.state.(Snapshotter)
	if !ok {
		return
	}
	w.mu.Lock()
	s.Restore(entry.snap)
	w.mu.Unlock()
	when := entry.when
	w.virtualNow.Store(&when)
	w.UpdateWindow()
}

// clear drops all entries.
func (r *snapshotRing) clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.entries {
		r.entries[i] = snapshotEntry{}
	}
	r.entries = r.entries[:0]
	r.totalBytes = 0
}
