//go:build !js && !android && !ios

// Package audio provides opt-in audio playback using SDL_mixer.
//
// Call [Init] before loading or playing audio. Call [Quit] when done.
// Initialization is independent of the GUI backend — it uses
// [sdl.InitSubSystem] to add audio without modifying backend init.
//
// SDL_mixer supports one music track and N sound-effect channels
// (default 16).
package audio

import (
	"cmp"
	"fmt"

	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

// Cfg configures the audio subsystem. Zero value selects sensible
// defaults.
type Cfg struct {
	// Frequency is the output sample rate in Hz. Default: 44100.
	Frequency int
	// Format is the SDL audio format. Default: mix.DEFAULT_FORMAT.
	Format uint16
	// OutputChannels is the number of output channels
	// (1=mono, 2=stereo). Default: 2.
	OutputChannels int
	// ChunkSize is the audio buffer size in bytes. Smaller values
	// reduce latency but increase CPU. Default: 2048.
	ChunkSize int
	// MixChannels is the number of mixing channels for sound
	// effects. Default: 16.
	MixChannels int
	// Formats is a bitmask of format decoders to load
	// (mix.INIT_OGG, mix.INIT_MP3, etc.).
	// Default: mix.INIT_OGG | mix.INIT_MP3.
	Formats int
}

var initialized bool

// Init initializes the audio subsystem. It is opt-in: calling Init
// adds audio via [sdl.InitSubSystem] without changing the backend's
// initial [sdl.Init] call.
//
// Pass zero or one [Cfg]; additional values are ignored.
// Call from the main goroutine (same thread as the SDL event loop).
// Idempotent — repeated calls return nil.
func Init(cfg ...Cfg) error {
	if initialized {
		return nil
	}
	var c Cfg
	if len(cfg) > 0 {
		c = cfg[0]
	}
	freq := cmp.Or(c.Frequency, 44100)
	format := cmp.Or(c.Format, uint16(mix.DEFAULT_FORMAT))
	outCh := cmp.Or(c.OutputChannels, 2)
	chunk := cmp.Or(c.ChunkSize, 2048)
	mixCh := cmp.Or(c.MixChannels, 16)
	formats := cmp.Or(c.Formats, mix.INIT_OGG|mix.INIT_MP3)

	if err := sdl.InitSubSystem(sdl.INIT_AUDIO); err != nil {
		return fmt.Errorf("audio: init SDL audio subsystem: %w", err)
	}
	if err := mix.Init(formats); err != nil {
		sdl.QuitSubSystem(sdl.INIT_AUDIO)
		return fmt.Errorf("audio: init mixer formats: %w", err)
	}
	if err := mix.OpenAudio(freq, format, outCh, chunk); err != nil {
		mix.Quit()
		sdl.QuitSubSystem(sdl.INIT_AUDIO)
		return fmt.Errorf("audio: open audio device: %w", err)
	}
	mix.AllocateChannels(mixCh)
	initialized = true
	return nil
}

// Quit shuts down the audio subsystem. All playing sounds and music
// are halted. Safe to call even if [Init] was never called.
func Quit() {
	if !initialized {
		return
	}
	mix.HaltChannel(-1)
	mix.HaltMusic()
	mix.CloseAudio()
	mix.Quit()
	sdl.QuitSubSystem(sdl.INIT_AUDIO)
	initialized = false
}

// SetMasterVolume sets the volume for all sound channels.
// v is clamped to [0.0, 1.0]. Affects sound effects only; use
// [SetMusicVolume] for the music channel.
func SetMasterVolume(v float32) {
	mix.Volume(-1, toMixVol(v))
}

// MasterVolume returns the current master sound volume (0.0–1.0).
func MasterVolume() float32 {
	return fromMixVol(mix.Volume(-1, -1))
}

// SetMusicVolume sets the global music volume.
// v is clamped to [0.0, 1.0].
func SetMusicVolume(v float32) {
	mix.VolumeMusic(toMixVol(v))
}

// MusicVolume returns the current music volume (0.0–1.0).
func MusicVolume() float32 {
	return fromMixVol(mix.VolumeMusic(-1))
}

// --- internal helpers ---

func toMixVol(v float32) int {
	return int(clamp01(v) * float32(mix.MAX_VOLUME))
}

func fromMixVol(v int) float32 {
	return float32(v) / float32(mix.MAX_VOLUME)
}

func clamp01(v float32) float32 {
	return max(0, min(1, v))
}
