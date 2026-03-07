package sdl2

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/filedialog"
	"github.com/veandco/go-sdl2/sdl"
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

func (n *nativePlatform) ShowOpenDialog(title, startDir string, extensions []string, allowMultiple bool) gui.PlatformDialogResult {
	return filedialog.ShowOpenDialog(title, startDir, extensions, allowMultiple)
}

func (n *nativePlatform) ShowSaveDialog(title, startDir, defaultName, defaultExt string, extensions []string, confirmOverwrite bool) gui.PlatformDialogResult {
	return filedialog.ShowSaveDialog(title, startDir, defaultName, defaultExt, extensions, confirmOverwrite)
}

func (n *nativePlatform) ShowFolderDialog(title, startDir string) gui.PlatformDialogResult {
	return filedialog.ShowFolderDialog(title, startDir)
}

func (n *nativePlatform) ShowMessageDialog(title, body string, level gui.NativeAlertLevel) gui.NativeAlertResult {
	flags := alertLevelToSDLFlags(level)
	if err := sdl.ShowSimpleMessageBox(flags, title, body, nil); err != nil {
		return gui.NativeAlertResult{
			Status:       gui.DialogError,
			ErrorMessage: err.Error(),
		}
	}
	return gui.NativeAlertResult{Status: gui.DialogOK}
}

func (n *nativePlatform) ShowConfirmDialog(title, body string, level gui.NativeAlertLevel) gui.NativeAlertResult {
	data := &sdl.MessageBoxData{
		Flags:   alertLevelToSDLFlags(level),
		Title:   title,
		Message: body,
		Buttons: []sdl.MessageBoxButtonData{
			{ButtonID: 1, Text: "OK"},
			{ButtonID: 0, Text: "Cancel"},
		},
	}
	buttonID, err := sdl.ShowMessageBox(data)
	if err != nil {
		return gui.NativeAlertResult{
			Status:       gui.DialogError,
			ErrorMessage: err.Error(),
		}
	}
	if buttonID == 1 {
		return gui.NativeAlertResult{Status: gui.DialogOK}
	}
	return gui.NativeAlertResult{Status: gui.DialogCancel}
}

func (n *nativePlatform) SendNotification(title, body string) gui.NativeNotificationResult {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(
			`display notification %q with title %q`, body, title)
		cmd = exec.Command("osascript", "-e", script)
	case "linux":
		cmd = exec.Command("notify-send", title, body)
	default:
		return gui.NativeNotificationResult{
			Status:       gui.NotificationError,
			ErrorMessage: "unsupported platform: " + runtime.GOOS,
		}
	}
	if err := cmd.Start(); err != nil {
		return gui.NativeNotificationResult{
			Status:       gui.NotificationError,
			ErrorMessage: err.Error(),
		}
	}
	return gui.NativeNotificationResult{Status: gui.NotificationOK}
}

func alertLevelToSDLFlags(level gui.NativeAlertLevel) uint32 {
	switch level {
	case gui.AlertWarning:
		return sdl.MESSAGEBOX_WARNING
	case gui.AlertCritical:
		return sdl.MESSAGEBOX_ERROR
	default:
		return sdl.MESSAGEBOX_INFORMATION
	}
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
