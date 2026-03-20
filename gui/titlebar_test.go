package gui

import "testing"

func TestTitlebarDarkNilPlatform(t *testing.T) {
	w := NewWindow(WindowCfg{})
	defer w.Close()
	// Should not panic with nil nativePlatform.
	w.TitlebarDark(true)
	w.TitlebarDark(false)
}
