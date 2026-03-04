package gui

import "time"

const animationCycle = 16 * time.Millisecond

// AnimationAdd registers a new animation. If an animation with the
// same ID exists, it is replaced.
func (w *Window) AnimationAdd(a Animation) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.animationAdd(a)
}

// animationAdd is the lock-free core of AnimationAdd. Callers
// must already hold w.mu (e.g. during Update/GenerateLayout).
func (w *Window) animationAdd(a Animation) {
	a.SetStart(time.Now())
	if w.animations == nil {
		w.animations = make(map[string]Animation)
	}
	w.animations[a.ID()] = a
}

// AnimationRemove stops and removes an animation by ID.
func (w *Window) AnimationRemove(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.animations, id)
}

func (w *Window) hasAnimationLocked(id string) bool {
	_, ok := w.animations[id]
	return ok
}

// HasAnimation returns true if an animation with the given ID is
// currently active.
func (w *Window) HasAnimation(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.hasAnimationLocked(id)
}

// animationLoop runs in a goroutine, updating all animations each
// tick and dispatching deferred callbacks via the command queue.
func (w *Window) animationLoop() {
	if w.animationDone != nil {
		defer close(w.animationDone)
	}
	ticker := time.NewTicker(animationCycle)
	defer ticker.Stop()

	dt := float32(animationCycle) / float32(time.Second)
	deferred := make([]func(*Window), 0, 4)
	stoppedIDs := make([]string, 0, 4)

	for {
		select {
		case <-ticker.C:
		case <-w.animationStop:
			return
		}
		refreshKind := AnimationRefreshNone
		deferred = deferred[:0]
		stoppedIDs = stoppedIDs[:0]

		w.mu.Lock()
		for _, a := range w.animations {
			var updated bool
			switch v := a.(type) {
			case *Animate:
				updated = updateAnimate(v, w, &deferred)
			case *BlinkCursorAnimation:
				updated = updateBlinkCursor(v, w)
			case *TweenAnimation:
				updated = updateTween(v, w, &deferred)
			case *SpringAnimation:
				updated = updateSpring(v, w, dt, &deferred)
			case *KeyframeAnimation:
				updated = updateKeyframe(v, w, &deferred)
			case *HeroTransition:
				updated = updateHeroTransition(v, w, &deferred)
			case *LayoutTransition:
				updated = updateLayoutTransition(v, w, &deferred)
			}
			if updated {
				refreshKind = maxAnimationRefreshKind(refreshKind, a.RefreshKind())
			}
			if a.IsStopped() {
				stoppedIDs = append(stoppedIDs, a.ID())
			}
		}
		for _, id := range stoppedIDs {
			delete(w.animations, id)
		}
		w.mu.Unlock()

		for _, cb := range deferred {
			w.QueueCommand(cb)
		}
		switch refreshKind {
		case AnimationRefreshRenderOnly:
			w.RequestRenderOnly()
		case AnimationRefreshLayout:
			w.UpdateWindow()
		}
	}
}

func (w *Window) stopAnimationLoop() {
	if w.animationStop == nil {
		return
	}
	w.animationStopOnce.Do(func() {
		close(w.animationStop)
		if w.animationDone != nil {
			<-w.animationDone
		}
	})
}

func updateAnimate(a *Animate, w *Window, deferred *[]func(*Window)) bool {
	if a.stopped {
		return false
	}
	if time.Since(a.start) > a.Delay {
		cb := a.Callback
		*deferred = append(*deferred, func(w *Window) {
			cb(a, w)
		})
		if a.Repeat {
			a.start = time.Now()
		} else {
			a.stopped = true
		}
		return true
	}
	return false
}

func updateBlinkCursor(b *BlinkCursorAnimation, w *Window) bool {
	if b.stopped {
		return false
	}
	if time.Since(b.start) > b.delay {
		w.viewState.inputCursorOn = !w.viewState.inputCursorOn
		b.start = time.Now()
		return true
	}
	return false
}
