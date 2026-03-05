//go:build darwin

package backend

import (
	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/metal"
)

func Run(w *gui.Window) { metal.Run(w) }
