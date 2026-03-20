package gui

import (
	"testing"
	"time"
)

// mockNotificationPlatform stubs NativePlatform with a configurable
// SendNotification result.
type mockNotificationPlatform struct {
	result NativeNotificationResult
}

func (m *mockNotificationPlatform) SendNotification(_, _ string) NativeNotificationResult {
	return m.result
}

func (m *mockNotificationPlatform) OpenURI(_ string) error { return nil }
func (m *mockNotificationPlatform) ShowOpenDialog(_, _ string, _ []string, _ bool) PlatformDialogResult {
	return PlatformDialogResult{}
}
func (m *mockNotificationPlatform) ShowSaveDialog(_, _, _, _ string, _ []string, _ bool) PlatformDialogResult {
	return PlatformDialogResult{}
}
func (m *mockNotificationPlatform) ShowFolderDialog(_, _ string) PlatformDialogResult {
	return PlatformDialogResult{}
}
func (m *mockNotificationPlatform) ShowMessageDialog(_, _ string, _ NativeAlertLevel) NativeAlertResult {
	return NativeAlertResult{}
}
func (m *mockNotificationPlatform) ShowConfirmDialog(_, _ string, _ NativeAlertLevel) NativeAlertResult {
	return NativeAlertResult{}
}
func (m *mockNotificationPlatform) ShowPrintDialog(_ NativePrintParams) PrintRunResult {
	return PrintRunResult{}
}
func (m *mockNotificationPlatform) BookmarkLoadAll(_ string) []BookmarkEntry { return nil }
func (m *mockNotificationPlatform) BookmarkPersist(_, _ string, _ []byte)    {}
func (m *mockNotificationPlatform) BookmarkStopAccess(_ []byte)              {}
func (m *mockNotificationPlatform) A11yInit(_ func(action, index int))       {}
func (m *mockNotificationPlatform) A11ySync(_ []A11yNode, _, _ int)          {}
func (m *mockNotificationPlatform) A11yDestroy()                             {}
func (m *mockNotificationPlatform) A11yAnnounce(_ string)                    {}
func (m *mockNotificationPlatform) IMEStart()                                {}
func (m *mockNotificationPlatform) IMEStop()                                 {}
func (m *mockNotificationPlatform) IMESetRect(_, _, _, _ int32)              {}
func (m *mockNotificationPlatform) TitlebarDark(_ bool)                      {}
func (m *mockNotificationPlatform) SpellCheck(_ string) []SpellRange         { return nil }
func (m *mockNotificationPlatform) SpellSuggest(_ string, _, _ int) []string { return nil }
func (m *mockNotificationPlatform) SpellLearn(_ string)                      {}

func TestNativeNotificationEmptyTitle(t *testing.T) {
	w := &Window{}
	var result NativeNotificationResult
	cfg := NativeNotificationCfg{
		Title:  "",
		OnDone: func(r NativeNotificationResult, _ *Window) { result = r },
	}
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
	nativeNotificationImpl(w, cfg)
	if result.Status != NotificationError {
		t.Errorf("expected error, got %d", result.Status)
	}
	if result.ErrorCode != "unsupported" {
		t.Errorf("expected 'unsupported', got %q", result.ErrorCode)
	}
}

func TestNativeNotificationMockPlatform(t *testing.T) {
	w := &Window{}
	w.nativePlatform = &mockNotificationPlatform{
		result: NativeNotificationResult{Status: NotificationOK},
	}
	var result NativeNotificationResult
	cfg := NativeNotificationCfg{
		Title:  "Hello",
		Body:   "World",
		OnDone: func(r NativeNotificationResult, _ *Window) { result = r },
	}
	nativeNotificationImpl(w, cfg)
	// Wait for goroutine to queue the command.
	time.Sleep(5 * time.Millisecond)
	w.flushCommands()
	if result.Status != NotificationOK {
		t.Errorf("expected OK, got %d", result.Status)
	}
}

func TestNativeNotificationMockPlatformError(t *testing.T) {
	w := &Window{}
	w.nativePlatform = &mockNotificationPlatform{
		result: NativeNotificationResult{
			Status:       NotificationError,
			ErrorCode:    "exec_failed",
			ErrorMessage: "mock error",
		},
	}
	var result NativeNotificationResult
	cfg := NativeNotificationCfg{
		Title:  "Hello",
		Body:   "World",
		OnDone: func(r NativeNotificationResult, _ *Window) { result = r },
	}
	nativeNotificationImpl(w, cfg)
	time.Sleep(5 * time.Millisecond)
	w.flushCommands()
	if result.Status != NotificationError {
		t.Errorf("expected error, got %d", result.Status)
	}
	if result.ErrorCode != "exec_failed" {
		t.Errorf("expected 'exec_failed', got %q", result.ErrorCode)
	}
	if result.ErrorMessage != "mock error" {
		t.Errorf("expected 'mock error', got %q", result.ErrorMessage)
	}
}

func TestNativeNotificationNilOnDone(_ *testing.T) {
	w := &Window{}
	cfg := NativeNotificationCfg{Title: "", Body: "B"}
	// Must not panic with nil OnDone.
	w.NativeNotification(cfg)
}

func TestNativeNotificationStatusValues(t *testing.T) {
	if NotificationOK != 0 || NotificationDenied != 1 || NotificationError != 2 {
		t.Error("unexpected status enum values")
	}
}
