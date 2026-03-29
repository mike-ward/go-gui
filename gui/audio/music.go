//go:build !js && !android && !ios

package audio

import (
	"fmt"

	"github.com/veandco/go-sdl2/mix"
)

// Music is a loaded music track (wraps [*mix.Music]).
//
// SDL_mixer supports only ONE music track playing at a time.
// Starting a new track halts the previous one. For layered or
// simultaneous audio, use [Sound] on separate channels.
type Music struct {
	mus *mix.Music
}

// LoadMusic loads a music track from a file path.
// Supports WAV, MOD, MIDI, OGG, MP3, FLAC depending on
// initialized decoders.
func LoadMusic(path string) (*Music, error) {
	mus, err := mix.LoadMUS(path)
	if err != nil {
		return nil, fmt.Errorf("audio: load music %q: %w", path, err)
	}
	return &Music{mus: mus}, nil
}

// Play starts music playback. loops is the number of extra loops
// (0 = play once, -1 = loop forever). Any currently playing music
// is halted first.
func (m *Music) Play(loops int) error {
	if err := m.mus.Play(loops); err != nil {
		return fmt.Errorf("audio: play music: %w", err)
	}
	return nil
}

// FadeIn starts music with a fade-in over ms milliseconds.
func (m *Music) FadeIn(loops, ms int) error {
	if err := m.mus.FadeIn(loops, ms); err != nil {
		return fmt.Errorf("audio: fade-in music: %w", err)
	}
	return nil
}

// Free releases the underlying SDL_mixer music. The Music must not
// be used after calling Free. Safe to call on a nil Music.
func (m *Music) Free() {
	if m == nil || m.mus == nil {
		return
	}
	m.mus.Free()
	m.mus = nil
}

// --- Global music controls (single music channel) ---

// HaltMusic stops the currently playing music immediately.
func HaltMusic() { mix.HaltMusic() }

// FadeOutMusic fades out the current music over ms milliseconds.
func FadeOutMusic(ms int) { mix.FadeOutMusic(ms) }

// PauseMusic pauses music playback.
func PauseMusic() { mix.PauseMusic() }

// ResumeMusic resumes paused music.
func ResumeMusic() { mix.ResumeMusic() }

// IsMusicPlaying reports whether music is currently playing.
func IsMusicPlaying() bool { return mix.PlayingMusic() }

// IsMusicPaused reports whether music is currently paused.
func IsMusicPaused() bool { return mix.PausedMusic() }

// RewindMusic rewinds to the beginning.
func RewindMusic() { mix.RewindMusic() }
