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

func (t *TweenAnimation) ID() string                        { return t.AnimID }
func (t *TweenAnimation) RefreshKind() AnimationRefreshKind { return AnimationRefreshLayout }
func (t *TweenAnimation) IsStopped() bool                   { return t.stopped }
func (t *TweenAnimation) SetStart(now time.Time)            { t.start = now }
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
	elapsed := time.Since(tw.start)
	if elapsed < tw.delay {
		return false
	}
	animElapsed := elapsed - tw.delay
	if tw.Duration <= 0 {
		animElapsed = tw.Duration
	}
	if animElapsed >= tw.Duration {
		val := tw.To
		*deferred = append(*deferred, queuedCommand{
			kind:    queuedCommandValueFn,
			valueFn: tw.OnValue,
			value:   val,
		})
		if tw.OnDone != nil {
			*deferred = append(*deferred, queuedCommand{
				kind:     queuedCommandWindowFn,
				windowFn: tw.OnDone,
			})
		}
		tw.stopped = true
		return true
	}
	progress := float32(animElapsed) / float32(tw.Duration)
	easing := tw.Easing
	if easing == nil {
		easing = EaseLinear
	}
	eased := easing(progress)
	value := Lerp(tw.From, tw.To, eased)
	*deferred = append(*deferred, queuedCommand{
		kind:    queuedCommandValueFn,
		valueFn: tw.OnValue,
		value:   value,
	})
	return true
}
