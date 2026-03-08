package gui

import "time"

// AnimationRefreshKind indicates what type of refresh an animation
// requires each tick.
type AnimationRefreshKind uint8

// AnimationRefreshKind constants.
const (
	AnimationRefreshNone       AnimationRefreshKind = iota
	AnimationRefreshRenderOnly                      // repaint only
	AnimationRefreshLayout                          // full layout rebuild
)

// maxAnimationRefreshKind returns the higher-priority refresh kind.
func maxAnimationRefreshKind(current, incoming AnimationRefreshKind) AnimationRefreshKind {
	if incoming > current {
		return incoming
	}
	return current
}

// Animation is the interface for all animation types.
type Animation interface {
	ID() string
	RefreshKind() AnimationRefreshKind
	IsStopped() bool
	SetStart(t time.Time)
	Update(*Window, float32, *[]queuedCommand) bool
}

// BlinkCursorAnimation toggles cursor visibility on a timer.
type BlinkCursorAnimation struct {
	id      string
	delay   time.Duration
	start   time.Time
	stopped bool
}

const blinkCursorAnimationID = "___blinky_cursor_animation___"
const blinkCursorAnimationDelay = 600 * time.Millisecond

// NewBlinkCursorAnimation creates a cursor blink animation.
func NewBlinkCursorAnimation() *BlinkCursorAnimation {
	return &BlinkCursorAnimation{
		id:    blinkCursorAnimationID,
		delay: blinkCursorAnimationDelay,
	}
}

// ID implements Animation.
func (a *BlinkCursorAnimation) ID() string { return a.id }

// RefreshKind implements Animation.
func (a *BlinkCursorAnimation) RefreshKind() AnimationRefreshKind { return AnimationRefreshRenderOnly }

// IsStopped implements Animation.
func (a *BlinkCursorAnimation) IsStopped() bool { return a.stopped }

// SetStart implements Animation.
func (a *BlinkCursorAnimation) SetStart(t time.Time) { a.start = t }

// Update implements Animation.
func (a *BlinkCursorAnimation) Update(w *Window, _ float32, _ *[]queuedCommand) bool {
	return updateBlinkCursor(a, w)
}

// Animate waits the specified delay then executes the callback.
type Animate struct {
	AnimateID string
	Callback  func(*Animate, *Window)
	Delay     time.Duration
	Repeat    bool
	Refresh   AnimationRefreshKind // 0 defaults to layout
	start     time.Time
	stopped   bool
}

// ID implements Animation.
func (a *Animate) ID() string { return a.AnimateID }

// RefreshKind implements Animation.
func (a *Animate) RefreshKind() AnimationRefreshKind {
	if a.Refresh != 0 {
		return a.Refresh
	}
	return AnimationRefreshLayout
}

// IsStopped implements Animation.
func (a *Animate) IsStopped() bool { return a.stopped }

// SetStart implements Animation.
func (a *Animate) SetStart(t time.Time) { a.start = t }

// Update implements Animation.
func (a *Animate) Update(_ *Window, _ float32, deferred *[]queuedCommand) bool {
	return updateAnimate(a, deferred)
}
