package gui

import "testing"

func TestTitlebarDarkNilPlatform(_ *testing.T) {
	w := NewWindow(WindowCfg{})
	defer w.Close()
	// Should not panic with nil nativePlatform.
	w.TitlebarDark(true)
	w.TitlebarDark(false)
}
