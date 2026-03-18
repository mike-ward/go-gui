//go:build darwin && !ios

// Package backend provides platform-specific backend initialization.
package backend

import (
	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/metal"
)

// Run starts the GUI event loop.
func Run(w *gui.Window) { metal.Run(w) }
