package gui

import "time"

// layoutSnapshot captures element position/size.
type layoutSnapshot struct {
	x, y, width, height float32
}

// LayoutTransitionCfg configures layout animation.
type LayoutTransitionCfg struct {
	Duration time.Duration
	Easing   EasingFn // nil → EaseOutCubic
}

const layoutTransitionID = "__layout_transition__"

// LayoutTransition animates layout changes (resize, reorder, add,
// remove) using FLIP-style animation.
type LayoutTransition struct {
	duration  time.Duration
	easing    EasingFn
	OnDone    func(*Window)
	delay     time.Duration
	start     time.Time
	stopped   bool
	snapshots map[string]layoutSnapshot
	progress  float32
}

func (l *LayoutTransition) ID() string                        { return layoutTransitionID }
func (l *LayoutTransition) RefreshKind() AnimationRefreshKind { return AnimationRefreshLayout }
func (l *LayoutTransition) IsStopped() bool                   { return l.stopped }
func (l *LayoutTransition) SetStart(now time.Time)            { l.start = now }
func (l *LayoutTransition) Update(_ *Window, _ float32, deferred *[]queuedCommand) bool {
	return updateLayoutTransition(l, deferred)
}

// AnimateLayout triggers layout transition animation. Call BEFORE
// making layout changes to capture current positions.
func (w *Window) AnimateLayout(cfg LayoutTransitionCfg) {
	dur := cfg.Duration
	if dur == 0 {
		dur = 200 * time.Millisecond
	}
	eas := cfg.Easing
	if eas == nil {
		eas = EaseOutCubic
	}
	lt := &LayoutTransition{
		duration:  dur,
		easing:    eas,
		snapshots: captureLayoutSnapshots(w.layout),
	}
	w.AnimationAdd(lt)
}

func updateLayoutTransition(lt *LayoutTransition, deferred *[]queuedCommand) bool {
	if lt.stopped {
		return false
	}
	elapsed := time.Since(lt.start)
	if elapsed < lt.delay {
		return false
	}
	animElapsed := elapsed - lt.delay
	if animElapsed >= lt.duration {
		lt.progress = 1.0
		lt.stopped = true
		if lt.OnDone != nil {
			*deferred = append(*deferred, queuedCommand{
				kind:     queuedCommandWindowFn,
				windowFn: lt.OnDone,
			})
		}
		return true
	}
	progress := float32(animElapsed) / float32(lt.duration)
	easing := lt.easing
	if easing == nil {
		easing = EaseOutCubic
	}
	lt.progress = easing(progress)
	return true
}

// captureLayoutSnapshots recursively captures all element positions.
func captureLayoutSnapshots(layout Layout) map[string]layoutSnapshot {
	snapshots := make(map[string]layoutSnapshot)
	captureRecursive(&layout, snapshots)
	return snapshots
}

func captureRecursive(layout *Layout, snapshots map[string]layoutSnapshot) {
	if layout.Shape.ID != "" {
		snapshots[layout.Shape.ID] = layoutSnapshot{
			x:      layout.Shape.X,
			y:      layout.Shape.Y,
			width:  layout.Shape.Width,
			height: layout.Shape.Height,
		}
	}
	for i := range layout.Children {
		captureRecursive(&layout.Children[i], snapshots)
	}
}

// getLayoutTransition returns the active layout transition, if any.
func (w *Window) getLayoutTransition() *LayoutTransition {
	a, ok := w.animations[layoutTransitionID]
	if !ok {
		return nil
	}
	lt, ok := a.(*LayoutTransition)
	if !ok {
		return nil
	}
	return lt
}

// applyLayoutTransition interpolates positions during amend phase.
func applyLayoutTransition(layout *Layout, w *Window) {
	lt := w.getLayoutTransition()
	if lt == nil || lt.stopped {
		return
	}
	applyTransitionRecursive(layout, lt)
}

func applyTransitionRecursive(layout *Layout, lt *LayoutTransition) {
	if layout.Shape.ID != "" {
		if old, ok := lt.snapshots[layout.Shape.ID]; ok {
			t := lt.progress
			layout.Shape.X = Lerp(old.x, layout.Shape.X, t)
			layout.Shape.Y = Lerp(old.y, layout.Shape.Y, t)
			layout.Shape.Width = Lerp(old.width, layout.Shape.Width, t)
			layout.Shape.Height = Lerp(old.height, layout.Shape.Height, t)
		}
	}
	for i := range layout.Children {
		applyTransitionRecursive(&layout.Children[i], lt)
	}
}
