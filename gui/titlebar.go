package gui

// TitlebarDark sets the window titlebar to dark or light mode.
// Delegates to the native platform backend (Windows DWM API).
// No-op if no native platform is set.
func (w *Window) TitlebarDark(dark bool) {
	if w.nativePlatform != nil {
		w.nativePlatform.TitlebarDark(dark)
	}
}
