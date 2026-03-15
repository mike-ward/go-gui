package gui

import "time"

// TweenAnimation interpolates a value from A to B over a fixed
// duration with easing.
type TweenAnimation struct {
	AnimID   string
	Duration time.Duration
	Easing   EasingFn
	From     float32
	To       float32
	OnValue  func(float32, *Window)
	OnDone   func(*Window)
	start    time.Time
	stopped  bool
}

const tweenDefaultDuration = 300 * time.Millisecond

// ID implements Animation.
func (t *TweenAnimation) ID() string { return t.AnimID }

// RefreshKind implements Animation.
func (t *TweenAnimation) RefreshKind() AnimationRefreshKind { return AnimationRefreshLayout }

// IsStopped implements Animation.
func (t *TweenAnimation) IsStopped() bool { return t.stopped }

// SetStart implements Animation.
func (t *TweenAnimation) SetStart(now time.Time) { t.start = now }

// Update implements Animation.
func (t *TweenAnimation) Update(_ *Window, _ float32, deferred *[]queuedCommand) bool {
	return updateTween(t, deferred)
}

// NewTweenAnimation creates a TweenAnimation with defaults.
func NewTweenAnimation(id string, from, to float32, onValue func(float32, *Window)) *TweenAnimation {
	return &TweenAnimation{
		AnimID:   id,
		Duration: tweenDefaultDuration,
		Easing:   EaseOutCubic,
		From:     from,
		To:       to,
		OnValue:  onValue,
	}
}

func updateTween(tw *TweenAnimation, deferred *[]queuedCommand) bool {
	if tw.stopped {
		return false
	}
	if tw.OnValue == nil {
		tw.stopped = true
		return false
	}
	progress, done := durationProgress(tw.start, tw.Duration)
	if done {
		queueOnValue(deferred, tw.OnValue, tw.To)
		queueOnDone(deferred, tw.OnDone)
		tw.stopped = true
		return true
	}
	easing := tw.Easing
	if easing == nil {
		easing = EaseLinear
	}
	eased := easing(progress)
	queueOnValue(deferred, tw.OnValue, Lerp(tw.From, tw.To, eased))
	return true
}
