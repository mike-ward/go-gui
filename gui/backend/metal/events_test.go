//go:build darwin

package metal

import (
	"testing"

	"github.com/mike-ward/go-gui/gui"
	"github.com/veandco/go-sdl2/sdl"
)

func TestMapKeyCodeKeypadEnter(t *testing.T) {
	if got := mapKeyCode(sdl.K_KP_ENTER); got != gui.KeyEnter {
		t.Fatalf("mapKeyCode(K_KP_ENTER) = %v, want %v",
			got, gui.KeyEnter)
	}
}
