package gui

import "context"

// WindowCfg configures a new Window.
type WindowCfg struct {
	State  any
	Title  string
	Width  int
	Height int
	// FixedSize disables user-driven window resizing when supported
	// by the active backend.
	FixedSize bool
	// AllowedSvgRoots restricts file-based SVG loads to these paths.
	// Empty means allow any local SVG path.
	AllowedSvgRoots []string
	// AllowedImageRoots restricts file-based image loads to these
	// paths. Empty means allow any local image path.
	AllowedImageRoots []string
	// MaxImageBytes caps source image file size for decoded image
	// loads. Zero or negative selects backend defaults.
	MaxImageBytes int64
	// MaxImagePixels caps decoded image dimensions (width*height).
	// Zero or negative selects backend defaults.
	MaxImagePixels int64
	// IconPNG is optional PNG-encoded icon data for the window.
	// The backend sets this as the window icon when supported.
	IconPNG []byte
	OnInit  func(*Window)
	OnEvent func(*Event, *Window)
	// OnCloseRequest runs when the OS reports a window-close (title
	// bar button, Cmd-W, etc.) before the window is destroyed. If
	// nil, the backend proceeds with destroy as before. If set, the
	// callback owns the decision: call Window.Close() to proceed, or
	// do nothing to cancel. Use for save/discard/cancel prompts.
	// Re-clicking the close control is required to retry after a veto
	// since the original SDL event is already drained.
	OnCloseRequest func(*Window)
	BgColor        Color
	// Timings enables per-frame pipeline timing instrumentation.
	Timings bool
	// DebugTimeTravel enables time-travel snapshot capture and
	// auto-spawns a scrubber window alongside the app window.
	// Requires multi-window mode (App + App.OpenWindow) and a
	// user state that implements the Snapshotter interface.
	// Leave off in release builds — the nil-history hot path
	// short-circuits with zero cost when disabled.
	DebugTimeTravel bool
	// HistoryBytes caps time-travel snapshot memory. Evicts
	// oldest entries when exceeded. Zero or negative selects
	// a default (64 MiB). Only consulted when DebugTimeTravel
	// is true.
	HistoryBytes int
}

// NewWindow creates a Window from the given configuration.
func NewWindow(cfg WindowCfg) *Window {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Window{
		state:         cfg.State,
		windowWidth:   cfg.Width,
		windowHeight:  cfg.Height,
		focused:       true,
		refreshLayout: true,
		OnEvent:       cfg.OnEvent,
		Config:        cfg,
		scratch:       newScratchPools(),
		ctx:           ctx,
		cancelCtx:     cancel,
		windowAnimation: windowAnimation{
			animationStop:     make(chan struct{}),
			animationDone:     make(chan struct{}),
			animationResumeCh: make(chan struct{}, 1),
		},
	}
	if cfg.DebugTimeTravel {
		w.EnableHistory(cfg.HistoryBytes)
	}
	return w
}
