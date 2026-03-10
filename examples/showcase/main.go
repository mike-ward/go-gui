// Package main implements a faithful showcase port for the go-gui framework.
package main

import (
	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend"
)

const (
	scrollCatalog uint32 = iota + 1
	scrollDetail
	scrollPlaceholderA
	scrollPlaceholderB
	focusSearch
)

const catalogWidth float32 = 300

func main() {
	gui.SetTheme(gui.ThemeDarkBordered)
	gui.SetMarkdownExternalAPIsEnabled(true)

	w := gui.NewWindow(gui.WindowCfg{
		State:  newShowcaseApp(),
		Title:  "Gui Showcase",
		Width:  950,
		Height: 700,
		OnInit: func(w *gui.Window) {
			loadEmbeddedLocales()
			app := gui.State[ShowcaseApp](w)
			syncThemeGenFromCfg(app, gui.CurrentTheme().Cfg)
			w.UpdateView(mainView)
			w.AnimationAdd(&gui.Animate{
				AnimateID: "shader_tick",
				Repeat:    true,
				Callback:  func(_ *gui.Animate, _ *gui.Window) {},
			})
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
		Padding: gui.NoPadding,
		Spacing: gui.NoSpacing,
		Content: []gui.View{catalogPanel(w), detailPanel(w)},
	})
}
