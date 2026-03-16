//go:build js && wasm

package backend

import (
	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/web"
)

// Run starts the GUI event loop using the Canvas2D web backend.
func Run(w *gui.Window) { web.Run(w) }
