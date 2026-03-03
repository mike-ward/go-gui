package main

import (
	_ "embed"

	"github.com/mike-ward/go-gui/gui"
	sdl2 "github.com/mike-ward/go-gui/gui/backend/sdl2"
)

//go:embed markdown_source.md
var markdownSource string

type App struct{}

func main() {
	gui.SetTheme(gui.ThemeDarkBordered)

	w := gui.NewWindow(gui.WindowCfg{
		State:  &App{},
		Width:  600,
		Height: 600,
		Title:  "Markdown View",
		OnInit: func(w *gui.Window) {
			w.UpdateView(mainView)
		},
	})

	sdl2.Run(w)
}

func mainView(w *gui.Window) gui.View {
	ww, wh := w.WindowSize()
	theme := gui.CurrentTheme()

	style := gui.DefaultMarkdownStyle()
	style.CodeBlockBG = gui.RGB(40, 44, 52)

	return gui.Column(gui.ContainerCfg{
		Width:    float32(ww),
		Height:   float32(wh),
		Sizing:   gui.FixedFixed,
		Padding:  gui.Some(theme.PaddingLarge),
		IDFocus:  1,
		IDScroll: 1,
		Content: []gui.View{
			w.Markdown(gui.MarkdownCfg{
				Source:     markdownSource,
				Style:      style,
				Mode:       gui.TextModeWrap,
				Color:      theme.ColorPanel,
				SizeBorder: 1,
				Radius:     theme.RadiusMedium,
				Padding:    theme.PaddingMedium,
			}),
		},
	})
}
