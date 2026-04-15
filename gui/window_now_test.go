package gui

import (
	"testing"
	"time"
)

// TestWindowNowLive confirms Now() returns a live clock value
// when no virtual instant is pinned.
func TestWindowNowLive(t *testing.T) {
	w := &Window{}
	before := time.Now()
	got := w.Now()
	after := time.Now()
	if got.Before(before) || got.After(after) {
		t.Fatalf("Now() = %v, want in [%v, %v]", got, before, after)
	}
}

// TestWindowNowPinned confirms Now() returns the pinned instant
// when a virtual clock is set, then returns to live on clear.
func TestWindowNowPinned(t *testing.T) {
	w := &Window{}
	pinned := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	w.setVirtualNow(&pinned)
	if got := w.Now(); !got.Equal(pinned) {
		t.Fatalf("pinned Now() = %v, want %v", got, pinned)
	}
	w.setVirtualNow(nil)
	got := w.Now()
	if got.Equal(pinned) {
		t.Fatalf("cleared Now() still equal to pinned %v", pinned)
	}
	if time.Since(got) > time.Second {
		t.Fatalf("cleared Now() = %v, not live", got)
	}
}

// TestWindowNowNil confirms Now() is safe on a nil receiver and
// returns live time.
func TestWindowNowNil(t *testing.T) {
	var w *Window
	before := time.Now()
	got := w.Now()
	after := time.Now()
	if got.Before(before) || got.After(after) {
		t.Fatalf("nil Window Now() = %v, want in [%v, %v]", got, before, after)
	}
}
