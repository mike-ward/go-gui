//go:build darwin

package metal

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

// nativePlatform implements gui.NativePlatform for the Metal
// backend.
type nativePlatform struct{}

func (n *nativePlatform) OpenURI(uri string) error {
	if err := validateOpenURI(uri); err != nil {
		return err
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", uri)
	case "linux":
		cmd = exec.Command("xdg-open", uri)
	case "windows":
		cmd = exec.Command("rundll32",
			"url.dll,FileProtocolHandler", uri)
	default:
		return fmt.Errorf("unsupported platform: %s",
			runtime.GOOS)
	}
	return cmd.Start()
}

func validateOpenURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URI: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "http", "https", "mailto":
		return nil
	default:
		return fmt.Errorf("unsupported URI scheme: %q",
			u.Scheme)
	}
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

func (n *nativePlatform) ShowMessageDialog(title, body string, _ gui.NativeAlertLevel) gui.NativeAlertResult {
	return gui.NativeAlertResult{Status: gui.DialogOK}
}

func (n *nativePlatform) ShowConfirmDialog(title, body string, _ gui.NativeAlertLevel) gui.NativeAlertResult {
	return gui.NativeAlertResult{Status: gui.DialogOK}
}

func (n *nativePlatform) SendNotification(title, body string) gui.NativeNotificationResult {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(
			`display notification %q with title %q`,
			body, title)
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
	return gui.NativeNotificationResult{
		Status: gui.NotificationOK,
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
