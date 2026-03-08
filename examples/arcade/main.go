// Arcade implements the showcase for the go-gui framework.
package main

import (
	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend"
)

const (
	scrollCatalog uint32 = iota + 1
	scrollDetail
	focusSearch
)

const catalogWidth float32 = 300

func main() {
	w := gui.NewWindow(gui.WindowCfg{
		State:  newArcadeApp(),
		Title:  "Arcade",
		Width:  1280,
		Height: 800,
		OnInit: func(w *gui.Window) {
			w.UpdateView(mainView)
		},
	})
	backend.Run(w)
}

func mainView(w *gui.Window) gui.View {
	ww, wh := w.WindowSize()
	return gui.Row(gui.ContainerCfg{
		Width:   float32(ww),
		Height:  float32(wh),
		Sizing:  gui.FixedFixed,
		Padding: gui.Some(gui.PaddingNone),
		Spacing: gui.Some(float32(0)),
		Content: []gui.View{catalogPanel(w), detailPanel(w)},
	})
}
