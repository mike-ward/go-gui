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

func queueOnDone(deferred *[]queuedCommand, fn func(*Window)) {
	if fn != nil {
		*deferred = append(*deferred, queuedCommand{
			kind: queuedCommandWindowFn, windowFn: fn,
		})
	}
}

func queueOnValue(deferred *[]queuedCommand, fn func(float32, *Window), val float32) {
	*deferred = append(*deferred, queuedCommand{
		kind: queuedCommandValueFn, valueFn: fn, value: val,
	})
}

func commandMarkLayoutRefresh(w *Window) {
	w.markLayoutRefresh()
}

func commandMarkRenderOnlyRefresh(w *Window) {
	w.markRenderOnlyRefresh()
}
