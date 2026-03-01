package gui

import "time"

// heroSnapshot captures element state for hero transitions.
type heroSnapshot struct {
	x, y, width, height float32
}

// HeroTransitionCfg configures hero transition.
type HeroTransitionCfg struct {
	Duration time.Duration
	Easing   EasingFn // nil → EaseOutCubic
}

const heroTransitionID = "__hero_transition__"

// HeroTransition animates elements between views. Only one
// HeroTransition can be active at a time (fixed internal ID).
type HeroTransition struct {
	duration time.Duration
	easing   EasingFn
	OnDone   func(*Window)
	delay    time.Duration
	start    time.Time
	stopped  bool
	outgoing map[string]heroSnapshot
	incoming map[string]heroSnapshot
	progress float32
}

func (h *HeroTransition) ID() string                    { return heroTransitionID }
func (h *HeroTransition) RefreshKind() AnimationRefreshKind { return AnimationRefreshLayout }
func (h *HeroTransition) IsStopped() bool               { return h.stopped }
func (h *HeroTransition) SetStart(now time.Time)        { h.start = now }

// NewHeroTransition creates a HeroTransition with defaults.
func NewHeroTransition(cfg HeroTransitionCfg) *HeroTransition {
	dur := cfg.Duration
	if dur == 0 {
		dur = 300 * time.Millisecond
	}
	eas := cfg.Easing
	if eas == nil {
		eas = EaseOutCubic
	}
	return &HeroTransition{
		duration: dur,
		easing:   eas,
	}
}

func updateHeroTransition(ht *HeroTransition, w *Window, deferred *[]func(*Window)) bool {
	if ht.stopped {
		return false
	}
	elapsed := time.Since(ht.start)
	if elapsed < ht.delay {
		return false
	}
	animElapsed := elapsed - ht.delay
	if animElapsed >= ht.duration {
		ht.progress = 1.0
		ht.stopped = true
		if ht.OnDone != nil {
			*deferred = append(*deferred, ht.OnDone)
		}
		return true
	}
	progress := float32(animElapsed) / float32(ht.duration)
	ht.progress = ht.easing(progress)
	return true
}

// captureHeroSnapshots finds all hero-marked elements.
func captureHeroSnapshots(layout Layout) map[string]heroSnapshot {
	snapshots := make(map[string]heroSnapshot)
	captureHeroesRecursive(layout, snapshots)
	return snapshots
}

func captureHeroesRecursive(layout Layout, snapshots map[string]heroSnapshot) {
	if layout.Shape.Hero && layout.Shape.ID != "" {
		snapshots[layout.Shape.ID] = heroSnapshot{
			x:      layout.Shape.X,
			y:      layout.Shape.Y,
			width:  layout.Shape.Width,
			height: layout.Shape.Height,
		}
	}
	for _, child := range layout.Children {
		captureHeroesRecursive(child, snapshots)
	}
}

// applyHeroTransition modifies layout during render for hero effect.
func applyHeroTransition(layout *Layout, w *Window) {
	a, ok := w.animations[heroTransitionID]
	if !ok {
		return
	}
	ht, ok := a.(*HeroTransition)
	if !ok || ht.stopped {
		return
	}
	applyHeroRecursive(layout, ht.progress, ht.outgoing, ht.incoming)
}

func propagateOpacity(layout *Layout, opacity float32) {
	layout.Shape.Opacity = opacity
	for i := range layout.Children {
		propagateOpacity(&layout.Children[i], opacity)
	}
}

func applyHeroRecursive(layout *Layout, progress float32, outgoing, incoming map[string]heroSnapshot) {
	if layout.Shape.Hero && layout.Shape.ID != "" {
		id := layout.Shape.ID
		morphProgress := f32Min(1, progress*2)
		fadeProgress := f32Max(0, (progress-0.5)*2)

		if out, hasOut := outgoing[id]; hasOut {
			if _, hasIn := incoming[id]; hasIn {
				layout.Shape.X = Lerp(out.x, layout.Shape.X, morphProgress)
				layout.Shape.Y = Lerp(out.y, layout.Shape.Y, morphProgress)
				layout.Shape.Width = Lerp(out.width, layout.Shape.Width, morphProgress)
				layout.Shape.Height = Lerp(out.height, layout.Shape.Height, morphProgress)
			}
		} else {
			propagateOpacity(layout, fadeProgress)
		}
	}
	for i := range layout.Children {
		applyHeroRecursive(&layout.Children[i], progress, outgoing, incoming)
	}
}
