//go:build android

package backend

import "github.com/mike-ward/go-gui/gui"

func Run(w *gui.Window) {
	panic("android: use android.SetWindow/Start/Render pattern")
}
