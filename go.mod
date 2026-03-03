module github.com/mike-ward/go-gui

go 1.26.0

require (
	github.com/mike-ward/go-glyph v0.0.0
	github.com/mike-ward/go-glyph/backend/sdl2 v0.0.0
	github.com/veandco/go-sdl2 v0.4.40
)

require (
	github.com/yuin/goldmark v1.7.16
	github.com/yuin/goldmark-emoji v1.0.6
)

replace (
	github.com/mike-ward/go-glyph => ../go_glyph
	github.com/mike-ward/go-glyph/backend/sdl2 => ../go_glyph/backend/sdl2
)
