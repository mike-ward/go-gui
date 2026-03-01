package gui

import "time"

// SpringCfg controls spring physics behavior.
type SpringCfg struct {
	Stiffness float32
	Damping   float32
	Mass      float32
	Threshold float32
}

// Spring presets.
var (
	SpringDefault = SpringCfg{Stiffness: 100, Damping: 10, Mass: 1.0, Threshold: 0.01}
	SpringGentle  = SpringCfg{Stiffness: 50, Damping: 8, Mass: 1.0, Threshold: 0.01}
	SpringBouncy  = SpringCfg{Stiffness: 300, Damping: 15, Mass: 1.0, Threshold: 0.01}
	SpringStiff   = SpringCfg{Stiffness: 500, Damping: 30, Mass: 1.0, Threshold: 0.01}
)

// springState tracks current spring physics.
type springState struct {
	position float32
	velocity float32
	target   float32
	atRest   bool
}

// SpringAnimation uses spring physics for natural motion.
type SpringAnimation struct {
	AnimID  string
	Config  SpringCfg
	OnValue func(float32, *Window)
	OnDone  func(*Window)
	delay   time.Duration
	start   time.Time
	stopped bool
	state   springState
}

func (s *SpringAnimation) ID() string                    { return s.AnimID }
func (s *SpringAnimation) RefreshKind() AnimationRefreshKind { return AnimationRefreshLayout }
func (s *SpringAnimation) IsStopped() bool               { return s.stopped }
func (s *SpringAnimation) SetStart(now time.Time)        { s.start = now }

// NewSpringAnimation creates a SpringAnimation with defaults.
func NewSpringAnimation(id string, onValue func(float32, *Window)) *SpringAnimation {
	return &SpringAnimation{
		AnimID:  id,
		Config:  SpringDefault,
		OnValue: onValue,
	}
}

// SpringTo sets the spring to start at from targeting to.
func (s *SpringAnimation) SpringTo(from, to float32) {
	s.state.position = from
	s.state.velocity = 0
	s.state.target = to
	s.state.atRest = false
	s.stopped = false
}

// Retarget changes the target while preserving position/velocity.
func (s *SpringAnimation) Retarget(to float32) {
	s.state.target = to
	s.state.atRest = false
	s.stopped = false
}

func updateSpring(sp *SpringAnimation, w *Window, dt float32, deferred *[]func(*Window)) bool {
	if sp.stopped || sp.state.atRest {
		return false
	}
	elapsed := time.Since(sp.start)
	if elapsed < sp.delay {
		return false
	}

	cfg := sp.Config
	displacement := sp.state.position - sp.state.target
	springForce := -cfg.Stiffness * displacement
	dampingForce := -cfg.Damping * sp.state.velocity
	acceleration := (springForce + dampingForce) / cfg.Mass

	sp.state.velocity += acceleration * dt
	sp.state.position += sp.state.velocity * dt

	if f32Abs(sp.state.velocity) < cfg.Threshold && f32Abs(displacement) < cfg.Threshold {
		sp.state.position = sp.state.target
		sp.state.velocity = 0
		sp.state.atRest = true
		target := sp.state.target
		onValue := sp.OnValue
		*deferred = append(*deferred, func(w *Window) {
			onValue(target, w)
		})
		if sp.OnDone != nil {
			*deferred = append(*deferred, sp.OnDone)
		}
		sp.stopped = true
		return true
	}

	pos := sp.state.position
	onValue := sp.OnValue
	*deferred = append(*deferred, func(w *Window) {
		onValue(pos, w)
	})
	return true
}
