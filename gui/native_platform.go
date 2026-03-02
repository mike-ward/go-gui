package gui

// NativePlatform provides platform-specific native OS operations.
// Set by the backend; nil in tests (operations no-op / return error).
type NativePlatform interface {
	// File dialogs — blocking; call from command queue.
	ShowOpenDialog(title, startDir string, extensions []string, allowMultiple bool) PlatformDialogResult
	ShowSaveDialog(title, startDir, defaultName, defaultExt string, extensions []string, confirmOverwrite bool) PlatformDialogResult
	ShowFolderDialog(title, startDir string) PlatformDialogResult
	ShowMessageDialog(title, body string, level NativeAlertLevel) NativeAlertResult
	ShowConfirmDialog(title, body string, level NativeAlertLevel) NativeAlertResult

	// Notifications.
	SendNotification(title, body string) NativeNotificationResult

	// Print — blocking; call from command queue.
	ShowPrintDialog(cfg NativePrintParams) PrintRunResult

	// File access / security-scoped bookmarks.
	BookmarkLoadAll(appID string) []BookmarkEntry
	BookmarkPersist(appID, path string, data []byte)
	BookmarkStopAccess(data []byte)

	// Accessibility.
	A11yInit(actionCallback func(action, index int))
	A11ySync(nodes []A11yNode, count, focusedIdx int)
	A11yDestroy()
	A11yAnnounce(text string)

	// Window appearance.
	TitlebarDark(dark bool)
}

// PlatformDialogResult is the raw result from native file dialogs.
type PlatformDialogResult struct {
	Status       NativeDialogStatus
	Paths        []PlatformPath
	ErrorCode    string
	ErrorMessage string
}

// PlatformPath pairs a path with optional bookmark data.
type PlatformPath struct {
	Path         string
	BookmarkData []byte
}

// BookmarkEntry is a persisted bookmark loaded at startup.
type BookmarkEntry struct {
	Path string
	Data []byte
}

// NativePrintParams contains bridge-level print dialog parameters.
type NativePrintParams struct {
	Title        string
	JobName      string
	PDFPath      string
	PaperWidth   float32
	PaperHeight  float32
	MarginTop    float32
	MarginRight  float32
	MarginBottom float32
	MarginLeft   float32
	Orientation  int
	Copies       int
	PageRanges   string
	DuplexMode   int
	ColorMode    int
	ScaleMode    int
}

// SetNativePlatform sets the native platform backend.
func (w *Window) SetNativePlatform(np NativePlatform) {
	w.nativePlatform = np
}

// NativePlatformBackend returns the native platform backend (nil in tests).
func (w *Window) NativePlatformBackend() NativePlatform {
	return w.nativePlatform
}
