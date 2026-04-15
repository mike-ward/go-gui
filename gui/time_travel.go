package gui

import (
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
