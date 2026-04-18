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
	wasEmpty := len(w.animations) == 0
	w.animations[a.ID()] = a
	if wasEmpty {
		w.ensureAnimationLoop()
		w.animationResume()
	}
}

// ensureAnimationLoop starts the animation goroutine on first use.
// No-op for windows without lifecycle channels (unit-test stubs).
func (w *Window) ensureAnimationLoop() {
	if w.animationStop == nil {
		return
	}
	w.animationStartOnce.Do(func() {
		w.animationStarted = true
		go w.animationLoop()
	})
}

// animationResume signals the animation loop to restart its
// ticker. Safe to call when already running (buffered channel).
func (w *Window) animationResume() {
	select {
	case w.animationResumeCh <- struct{}{}:
	default:
	}
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
// The ticker starts paused and resumes when animationAdd signals
// via animationResumeCh. It pauses again when all animations stop.
func (w *Window) animationLoop() {
	if w.animationDone != nil {
		defer close(w.animationDone)
	}

	dt := float32(animationCycle) / float32(time.Second)
	deferred := make([]queuedCommand, 0, 8)
	stoppedIDs := make([]string, 0, 4)

	var ticker *time.Ticker
	var tickCh <-chan time.Time

	for {
		select {
		case <-tickCh:
		case <-w.animationResumeCh:
			if ticker == nil {
				ticker = time.NewTicker(animationCycle)
				tickCh = ticker.C
			}
			continue
		case <-w.animationStop:
			if ticker != nil {
				ticker.Stop()
			}
			return
		}

		refreshKind := AnimationRefreshNone
		deferred = deferred[:0]
		stoppedIDs = stoppedIDs[:0]

		w.mu.Lock()
		ac := newAnimationCommands(&deferred)
		for _, a := range w.animations {
			updated := a.Update(w, dt, &ac)
			if updated {
				refreshKind = maxAnimationRefreshKind(
					refreshKind, a.RefreshKind())
			}
			if a.IsStopped() {
				stoppedIDs = append(stoppedIDs, a.ID())
			}
		}
		for _, id := range stoppedIDs {
			delete(w.animations, id)
		}
		idle := len(w.animations) == 0
		w.mu.Unlock()

		if idle && ticker != nil {
			ticker.Stop()
			ticker = nil
			tickCh = nil
		}

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
		if len(deferred) > 0 {
			w.wakeMain()
		}
	}
}

// wakeMain calls the backend's wake function to unblock the
// main event loop from WaitEventTimeout. Nil-safe.
func (w *Window) wakeMain() {
	if fn := w.wakeMainFn; fn != nil {
		fn()
	}
}

func (w *Window) stopAnimationLoop() {
	if w.animationStop == nil || !w.animationStarted {
		return
	}
	w.animationStopOnce.Do(func() {
		close(w.animationStop)
		if w.animationDone != nil {
			<-w.animationDone
		}
	})
}

func updateAnimate(a *Animate, ac *AnimationCommands) bool {
	if a.stopped {
		return false
	}
	if a.Callback == nil {
		a.stopped = true
		return false
	}
	if time.Since(a.start) > a.Delay {
		ac.appendAnimate(a.Callback, a)
		if a.Repeat {
			// Zero delay with repeat fires every tick (~16ms).
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
