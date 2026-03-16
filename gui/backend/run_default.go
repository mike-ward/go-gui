//go:build !darwin && !js

package backend

import (
	"github.com/mike-ward/go-gui/gui"
	glbackend "github.com/mike-ward/go-gui/gui/backend/gl"
)

func Run(w *gui.Window) { glbackend.Run(w) }
