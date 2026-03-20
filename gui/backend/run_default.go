//go:build !darwin && !js && !android

// Package backend provides platform-specific backend initialization.
package backend

import (
	"github.com/mike-ward/go-gui/gui"
	glbackend "github.com/mike-ward/go-gui/gui/backend/gl"
)

// Run starts the application event loop using the OpenGL backend.
func Run(w *gui.Window) { glbackend.Run(w) }
