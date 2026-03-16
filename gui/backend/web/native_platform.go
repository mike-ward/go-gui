//go:build js && wasm

package web

import (
	"fmt"
	"syscall/js"

	"github.com/mike-ward/go-gui/gui"
)

// nativePlatform implements gui.NativePlatform for wasm.
type nativePlatform struct{}

func (n *nativePlatform) OpenURI(uri string) error {
	w := js.Global().Call("open", uri, "_blank")
	if w.IsNull() || w.IsUndefined() {
		return fmt.Errorf("web: popup blocked for %q", uri)
	}
	return nil
}

// Dialog stubs — not supported in wasm.

func (n *nativePlatform) ShowOpenDialog(_, _ string, _ []string, _ bool) gui.PlatformDialogResult {
	return gui.PlatformDialogResult{Status: gui.DialogError,
		ErrorCode: "unsupported", ErrorMessage: "not available in wasm"}
}

func (n *nativePlatform) ShowSaveDialog(_, _, _, _ string, _ []string, _ bool) gui.PlatformDialogResult {
	return gui.PlatformDialogResult{Status: gui.DialogError,
		ErrorCode: "unsupported", ErrorMessage: "not available in wasm"}
}

func (n *nativePlatform) ShowFolderDialog(_, _ string) gui.PlatformDialogResult {
	return gui.PlatformDialogResult{Status: gui.DialogError,
		ErrorCode: "unsupported", ErrorMessage: "not available in wasm"}
}

func (n *nativePlatform) ShowMessageDialog(_, _ string, _ gui.NativeAlertLevel) gui.NativeAlertResult {
	return gui.NativeAlertResult{Status: gui.DialogError,
		ErrorCode: "unsupported", ErrorMessage: "not available in wasm"}
}

func (n *nativePlatform) ShowConfirmDialog(_, _ string, _ gui.NativeAlertLevel) gui.NativeAlertResult {
	return gui.NativeAlertResult{Status: gui.DialogError,
		ErrorCode: "unsupported", ErrorMessage: "not available in wasm"}
}

func (n *nativePlatform) SendNotification(_, _ string) gui.NativeNotificationResult {
	return gui.NativeNotificationResult{Status: gui.NotificationError,
		ErrorCode: "unsupported", ErrorMessage: "not available in wasm"}
}

func (n *nativePlatform) ShowPrintDialog(_ gui.NativePrintParams) gui.PrintRunResult {
	return gui.PrintRunResult{ErrorCode: "unsupported",
		ErrorMessage: "not available in wasm"}
}

// Bookmark stubs.
func (n *nativePlatform) BookmarkLoadAll(_ string) []gui.BookmarkEntry { return nil }
func (n *nativePlatform) BookmarkPersist(_, _ string, _ []byte)        {}
func (n *nativePlatform) BookmarkStopAccess(_ []byte)                  {}

// Accessibility stubs.
func (n *nativePlatform) A11yInit(_ func(action, index int)) {}
func (n *nativePlatform) A11ySync(_ []gui.A11yNode, _, _ int) {}
func (n *nativePlatform) A11yDestroy()                        {}
func (n *nativePlatform) A11yAnnounce(_ string)               {}

// IME stubs.
func (n *nativePlatform) IMEStart()                     {}
func (n *nativePlatform) IMEStop()                      {}
func (n *nativePlatform) IMESetRect(_, _, _, _ int32)   {}

// Window appearance stub.
func (n *nativePlatform) TitlebarDark(_ bool) {}

// Spell check stubs.
func (n *nativePlatform) SpellCheck(_ string) []gui.SpellRange          { return nil }
func (n *nativePlatform) SpellSuggest(_ string, _, _ int) []string      { return nil }
func (n *nativePlatform) SpellLearn(_ string)                           {}
