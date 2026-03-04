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
	OnInit          func(*Window)
	OnEvent         func(*Event, *Window)
	BgColor         Color
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
	}
	go w.animationLoop()
	return w
}
