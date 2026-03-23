package gui

// DrawCanvasCfg configures a draw canvas view.
type DrawCanvasCfg struct {
	ID              string
	A11YLabel       string
	A11YDescription string
	Version         uint64
	Sizing          Sizing
	Width           float32
	Height          float32
	MinWidth        float32
	MaxWidth        float32
	MinHeight       float32
	MaxHeight       float32
	Padding         Opt[Padding]
	Clip            bool
	Color           Color
	Radius          float32
	OnDraw          func(*DrawContext)
	OnClick         func(*Layout, *Event, *Window)
	OnHover         func(*Layout, *Event, *Window)
	OnGesture       func(*Layout, *Event, *Window)
	OnMouseScroll   func(*Layout, *Event, *Window)
}

// drawCanvasView implements View for user-drawn canvas content.
type drawCanvasView struct {
	cfg DrawCanvasCfg
}

// DrawCanvas creates a canvas with user-drawn geometry.
func DrawCanvas(cfg DrawCanvasCfg) View {
	if cfg.Sizing == (Sizing{}) {
		cfg.Sizing = FixedFixed
	}
	if !cfg.Color.IsSet() {
		cfg.Color = ColorTransparent
	}
	return &drawCanvasView{cfg: cfg}
}

func (dv *drawCanvasView) Content() []View { return nil }

func (dv *drawCanvasView) GenerateLayout(_ *Window) Layout {
	c := &dv.cfg

	var events *EventHandlers
	if c.OnClick != nil || c.OnHover != nil || c.OnGesture != nil ||
		c.OnMouseScroll != nil || c.OnDraw != nil {
		events = &EventHandlers{
			OnClick:       leftClickOnly(c.OnClick),
			OnHover:       c.OnHover,
			OnGesture:     c.OnGesture,
			OnMouseScroll: c.OnMouseScroll,
			OnDraw:        c.OnDraw,
		}
	}

	layout := Layout{
		Shape: &Shape{
			ShapeType: ShapeDrawCanvas,
			ID:        c.ID,
			Version:   c.Version,
			A11YRole:  AccessRoleImage,
			A11Y:      makeA11YInfo(c.A11YLabel, c.A11YDescription),
			Width:     c.Width,
			Height:    c.Height,
			MinWidth:  c.MinWidth,
			MaxWidth:  c.MaxWidth,
			MinHeight: c.MinHeight,
			MaxHeight: c.MaxHeight,
			Sizing:    c.Sizing,
			Padding:   c.Padding.Get(Padding{}),
			Clip:      c.Clip,
			Color:     c.Color,
			Radius:    c.Radius,
			Events:    events,
		},
	}
	ApplyFixedSizingConstraints(layout.Shape)
	return layout
}
