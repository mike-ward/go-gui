package gui

type queuedCommandKind uint8

const (
	queuedCommandWindowFn queuedCommandKind = iota
	queuedCommandValueFn
	queuedCommandAnimateFn
)

type queuedCommand struct {
	kind queuedCommandKind

	windowFn  func(*Window)
	valueFn   func(float32, *Window)
	animateFn func(*Animate, *Window)

	value   float32
	animate *Animate
}

func commandMarkLayoutRefresh(w *Window) {
	w.markLayoutRefresh()
}

func commandMarkRenderOnlyRefresh(w *Window) {
	w.markRenderOnlyRefresh()
}
