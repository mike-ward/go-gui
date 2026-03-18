//go:build ios

package backend

import (
	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/ios"
)

func Run(w *gui.Window) { ios.Run(w) }
