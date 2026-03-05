package gui

// WindowCfg configures a new Window.
type WindowCfg struct {
	State  any
	Title  string
	Width  int
	Height int
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
	OnInit         func(*Window)
	OnEvent        func(*Event, *Window)
	BgColor        Color
	// Timings enables per-frame pipeline timing instrumentation.
	Timings bool
}

// NewWindow creates a Window from the given configuration.
func NewWindow(cfg WindowCfg) *Window {
	w := &Window{
		state:         cfg.State,
		windowWidth:   cfg.Width,
		windowHeight:  cfg.Height,
		focused:       true,
		refreshLayout: true,
		OnEvent:       cfg.OnEvent,
		Config:        cfg,
		animationStop: make(chan struct{}),
		animationDone: make(chan struct{}),
	}
	go w.animationLoop()
	return w
}
