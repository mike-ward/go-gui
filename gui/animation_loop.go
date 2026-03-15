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
	deferred := make([]queuedCommand, 0, 8)
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
			updated := a.Update(w, dt, &deferred)
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

		switch refreshKind {
		case AnimationRefreshRenderOnly:
			deferred = append(deferred, queuedCommand{
				kind:     queuedCommandWindowFn,
				windowFn: commandMarkRenderOnlyRefresh,
			})
		case AnimationRefreshLayout:
			deferred = append(deferred, queuedCommand{
				kind:     queuedCommandWindowFn,
				windowFn: commandMarkLayoutRefresh,
			})
		}
		w.queueCommandsBatch(deferred)
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

func updateAnimate(a *Animate, deferred *[]queuedCommand) bool {
	if a.stopped {
		return false
	}
	if a.Callback == nil {
		a.stopped = true
		return false
	}
	if time.Since(a.start) > a.Delay {
		*deferred = append(*deferred, queuedCommand{
			kind:      queuedCommandAnimateFn,
			animateFn: a.Callback,
			animate:   a,
		})
		if a.Repeat {
			a.start = a.start.Add(a.Delay)
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
	if time.Since(b.start) > blinkCursorAnimationDelay {
		w.viewState.inputCursorOn = !w.viewState.inputCursorOn
		b.start = b.start.Add(blinkCursorAnimationDelay)
		return true
	}
	return false
}
