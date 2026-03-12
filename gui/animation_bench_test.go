package gui

import (
	"testing"
	"time"
)

var benchAnimationValueSink float32

func benchAnimationOnValue(v float32, _ *Window) {
	benchAnimationValueSink = v
}

func benchAnimationOnDone(_ *Window) {}

func benchAnimationOnAnimate(_ *Animate, _ *Window) {}

func BenchmarkUpdateAnimate(b *testing.B) {
	a := &Animate{
		AnimID: "bench:animate",
		Callback:  benchAnimationOnAnimate,
		Delay:     0,
		Repeat:    true,
		start:     time.Now().Add(-time.Second),
	}
	w := &Window{}
	deferred := make([]queuedCommand, 0, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deferred = deferred[:0]
		_ = updateAnimate(a, &deferred)
		w.queueCommandsBatch(deferred)
		w.flushCommands()
	}
}

func BenchmarkUpdateTween(b *testing.B) {
	tw := NewTweenAnimation("bench:tween", 0, 1, benchAnimationOnValue)
	tw.OnDone = benchAnimationOnDone
	tw.Duration = time.Hour
	tw.start = time.Now().Add(-time.Second)
	w := &Window{}
	deferred := make([]queuedCommand, 0, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deferred = deferred[:0]
		_ = updateTween(tw, &deferred)
		w.queueCommandsBatch(deferred)
		w.flushCommands()
	}
}

func BenchmarkUpdateKeyframe(b *testing.B) {
	kf := NewKeyframeAnimation("bench:keyframe",
		[]Keyframe{
			{At: 0, Value: 0},
			{At: 0.5, Value: 0.5, Easing: EaseLinear},
			{At: 1, Value: 1, Easing: EaseLinear},
		},
		benchAnimationOnValue,
	)
	kf.OnDone = benchAnimationOnDone
	kf.Duration = time.Hour
	kf.start = time.Now().Add(-time.Second)
	w := &Window{}
	deferred := make([]queuedCommand, 0, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deferred = deferred[:0]
		_ = updateKeyframe(kf, &deferred)
		w.queueCommandsBatch(deferred)
		w.flushCommands()
	}
}

func BenchmarkUpdateSpring(b *testing.B) {
	sp := NewSpringAnimation("bench:spring", benchAnimationOnValue)
	sp.OnDone = benchAnimationOnDone
	sp.Config = SpringGentle
	sp.SpringTo(0, 100)
	sp.start = time.Now().Add(-time.Second)
	w := &Window{}
	deferred := make([]queuedCommand, 0, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if sp.stopped {
			sp.SpringTo(0, 100)
			sp.start = time.Now().Add(-time.Second)
		}
		deferred = deferred[:0]
		_ = updateSpring(sp, 0.016, &deferred)
		w.queueCommandsBatch(deferred)
		w.flushCommands()
	}
}
