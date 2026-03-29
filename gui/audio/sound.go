//go:build !js && !android && !ios

package audio

import (
	"fmt"

	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

// Sound is a loaded sound effect (wraps [*mix.Chunk]).
//
// Sounds play on numbered mixing channels. Pass channel -1 to
// auto-select the first free channel.
type Sound struct {
	chunk *mix.Chunk
	// keep holds a reference to byte slices passed to
	// [LoadSoundBytes] so the GC does not collect the backing
	// memory while SDL_mixer still references it.
	keep []byte
}

// LoadSound loads a sound effect from a file path.
// Supports WAV, OGG, FLAC, AIFF, VOC and other formats
// depending on which decoders were initialized.
func LoadSound(path string) (*Sound, error) {
	chunk, err := mix.LoadWAV(path)
	if err != nil {
		return nil, fmt.Errorf("audio: load sound %q: %w", path, err)
	}
	return &Sound{chunk: chunk}, nil
}

// LoadSoundBytes loads a sound effect from in-memory bytes.
// The caller must not modify data after this call.
func LoadSoundBytes(data []byte) (*Sound, error) {
	rw, err := sdl.RWFromMem(data)
	if err != nil {
		return nil, fmt.Errorf(
			"audio: create RWops from bytes: %w", err)
	}
	chunk, err := mix.LoadWAVRW(rw, true)
	if err != nil {
		return nil, fmt.Errorf(
			"audio: load sound from bytes: %w", err)
	}
	return &Sound{chunk: chunk, keep: data}, nil
}

// Play plays the sound on the given channel (-1 = first free).
// loops is the number of extra loops (0 = play once,
// -1 = loop forever). Returns the channel number used.
func (s *Sound) Play(channel, loops int) (int, error) {
	ch, err := s.chunk.Play(channel, loops)
	if err != nil {
		return -1, fmt.Errorf("audio: play sound: %w", err)
	}
	return ch, nil
}

// PlayOnce plays the sound once on the first free channel.
func (s *Sound) PlayOnce() (int, error) {
	return s.Play(-1, 0)
}

// FadeIn plays the sound with a fade-in over ms milliseconds.
func (s *Sound) FadeIn(channel, loops, ms int) (int, error) {
	ch, err := s.chunk.FadeIn(channel, loops, ms)
	if err != nil {
		return -1, fmt.Errorf("audio: fade-in sound: %w", err)
	}
	return ch, nil
}

// SetVolume sets this sound's volume. v is clamped to [0.0, 1.0].
// This is the chunk-level volume, mixed with the channel volume.
func (s *Sound) SetVolume(v float32) {
	s.chunk.Volume(toMixVol(v))
}

// Volume returns the sound's current volume (0.0–1.0).
func (s *Sound) Volume() float32 {
	return fromMixVol(s.chunk.Volume(-1))
}

// Free releases the underlying SDL_mixer chunk. The Sound must not
// be used after calling Free. Do not Free a Sound that is still
// playing — halt the channel first. Safe to call on a nil Sound.
func (s *Sound) Free() {
	if s == nil || s.chunk == nil {
		return
	}
	s.chunk.Free()
	s.chunk = nil
	s.keep = nil
}

// --- Channel-level helpers ---

// HaltChannel stops playback on the given channel (-1 = all).
func HaltChannel(channel int) { mix.HaltChannel(channel) }

// FadeOutChannel fades out the given channel over ms milliseconds.
func FadeOutChannel(channel, ms int) { mix.FadeOutChannel(channel, ms) }

// PauseChannel pauses the given channel (-1 = all).
func PauseChannel(channel int) { mix.Pause(channel) }

// ResumeChannel resumes the given channel (-1 = all).
func ResumeChannel(channel int) { mix.Resume(channel) }

// IsPlaying reports whether the given channel is currently playing.
func IsPlaying(channel int) bool { return mix.Playing(channel) > 0 }
