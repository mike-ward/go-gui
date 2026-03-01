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
	delay    time.Duration
	start    time.Time
	stopped  bool
}

const tweenDefaultDuration = 300 * time.Millisecond

func (t *TweenAnimation) ID() string                    { return t.AnimID }
func (t *TweenAnimation) RefreshKind() AnimationRefreshKind { return AnimationRefreshLayout }
func (t *TweenAnimation) IsStopped() bool               { return t.stopped }
func (t *TweenAnimation) SetStart(now time.Time)        { t.start = now }

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

func updateTween(tw *TweenAnimation, w *Window, deferred *[]func(*Window)) bool {
	if tw.stopped {
		return false
	}
	elapsed := time.Since(tw.start)
	if elapsed < tw.delay {
		return false
	}
	animElapsed := elapsed - tw.delay
	if animElapsed >= tw.Duration {
		val := tw.To
		onValue := tw.OnValue
		*deferred = append(*deferred, func(w *Window) {
			onValue(val, w)
		})
		if tw.OnDone != nil {
			*deferred = append(*deferred, tw.OnDone)
		}
		tw.stopped = true
		return true
	}
	progress := float32(animElapsed) / float32(tw.Duration)
	eased := tw.Easing(progress)
	value := Lerp(tw.From, tw.To, eased)
	onValue := tw.OnValue
	*deferred = append(*deferred, func(w *Window) {
		onValue(value, w)
	})
	return true
}
