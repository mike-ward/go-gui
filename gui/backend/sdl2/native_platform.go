package sdl2

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/mike-ward/go-gui/gui"
)

// nativePlatform implements gui.NativePlatform for SDL2.
// Only OpenURI is functional; other methods are stubs.
type nativePlatform struct{}

func (n *nativePlatform) OpenURI(uri string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", uri)
	case "linux":
		cmd = exec.Command("xdg-open", uri)
	case "windows":
		// "start" is a shell built-in; invoke via cmd.exe.
		cmd = exec.Command("cmd", "/c", "start", "", uri)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func (n *nativePlatform) ShowOpenDialog(_, _ string, _ []string, _ bool) gui.PlatformDialogResult {
	return gui.PlatformDialogResult{}
}

func (n *nativePlatform) ShowSaveDialog(_, _, _, _ string, _ []string, _ bool) gui.PlatformDialogResult {
	return gui.PlatformDialogResult{}
}

func (n *nativePlatform) ShowFolderDialog(_, _ string) gui.PlatformDialogResult {
	return gui.PlatformDialogResult{}
}

func (n *nativePlatform) ShowMessageDialog(_, _ string, _ gui.NativeAlertLevel) gui.NativeAlertResult {
	return gui.NativeAlertResult{}
}

func (n *nativePlatform) ShowConfirmDialog(_, _ string, _ gui.NativeAlertLevel) gui.NativeAlertResult {
	return gui.NativeAlertResult{}
}

func (n *nativePlatform) SendNotification(_, _ string) gui.NativeNotificationResult {
	return gui.NativeNotificationResult{}
}

func (n *nativePlatform) ShowPrintDialog(_ gui.NativePrintParams) gui.PrintRunResult {
	return gui.PrintRunResult{}
}

func (n *nativePlatform) BookmarkLoadAll(_ string) []gui.BookmarkEntry { return nil }
func (n *nativePlatform) BookmarkPersist(_, _ string, _ []byte)        {}
func (n *nativePlatform) BookmarkStopAccess(_ []byte)                  {}

func (n *nativePlatform) A11yInit(_ func(action, index int))  {}
func (n *nativePlatform) A11ySync(_ []gui.A11yNode, _, _ int) {}
func (n *nativePlatform) A11yDestroy()                        {}
func (n *nativePlatform) A11yAnnounce(_ string)               {}

func (n *nativePlatform) TitlebarDark(_ bool) {}
