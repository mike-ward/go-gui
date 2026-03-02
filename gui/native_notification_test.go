package gui

import "testing"

func TestNativeNotificationEmptyTitle(t *testing.T) {
	w := &Window{}
	var result NativeNotificationResult
	cfg := NativeNotificationCfg{
		Title:  "",
		OnDone: func(r NativeNotificationResult, _ *Window) { result = r },
	}
	// NativeNotification dispatches synchronously for empty title.
	w.NativeNotification(cfg)
	if result.Status != NotificationError {
		t.Errorf("expected error, got %d", result.Status)
	}
	if result.ErrorCode != "invalid_cfg" {
		t.Errorf("expected 'invalid_cfg', got %q", result.ErrorCode)
	}
}

func TestNativeNotificationNoPlatform(t *testing.T) {
	w := &Window{}
	var result NativeNotificationResult
	cfg := NativeNotificationCfg{
		Title:  "Hello",
		Body:   "World",
		OnDone: func(r NativeNotificationResult, _ *Window) { result = r },
	}
	// Call impl directly (bypasses queue + goroutine).
	nativeNotificationImpl(w, cfg)
	if result.Status != NotificationError {
		t.Errorf("expected error, got %d", result.Status)
	}
	if result.ErrorCode != "unsupported" {
		t.Errorf("expected 'unsupported', got %q", result.ErrorCode)
	}
}

func TestNativeNotificationCfgFields(t *testing.T) {
	cfg := NativeNotificationCfg{Title: "T", Body: "B"}
	if cfg.Title != "T" || cfg.Body != "B" {
		t.Error("fields not set")
	}
}

func TestNativeNotificationStatusValues(t *testing.T) {
	if NotificationOK != 0 || NotificationDenied != 1 || NotificationError != 2 {
		t.Error("unexpected status enum values")
	}
}
