package gui

// ContainerCfg configures container views (column, row, canvas,
// circle, wrap). Containers layout children vertically,
// horizontally, or freely with sizing, alignment, scrolling,
// floating, borders, and event handling.
type ContainerCfg struct {
	// Identity
	ID string

	// Sizing
	Sizing    Sizing
	Width     float32
	Height    float32
	MinWidth  float32
	MaxWidth  float32
	MinHeight float32
	MaxHeight float32

	// Layout
	Spacing  float32
	Padding  Padding
	HAlign   HorizontalAlign
	VAlign   VerticalAlign
	TextDir  TextDirection
	Wrap     bool
	Overflow bool

	// Appearance
	Color          Color
	ColorBorder    Color
	SizeBorder     float32
	Radius         float32
	BlurRadius     float32
	Opacity        float32
	Shadow         *BoxShadow
	Gradient       *GradientDef
	BorderGradient *GradientDef

	// Behavior
	IDFocus    uint32
	IDScroll   uint32
	ScrollMode ScrollMode
	Clip       bool
	FocusSkip  bool
	Disabled   bool
	Invisible  bool
	OverDraw   bool
	Hero       bool

	// Floating
	Float        bool
	FloatAnchor  FloatAttach
	FloatTieOff  FloatAttach
	FloatOffsetX float32
	FloatOffsetY float32

	// Accessibility
	A11YRole        AccessRole
	A11YState       AccessState
	A11YLabel       string
	A11YDescription string
	A11Y            *AccessInfo

	// Event handlers
	OnClick     func(*Layout, *Event, *Window)
	OnAnyClick  func(*Layout, *Event, *Window)
	OnChar      func(*Layout, *Event, *Window)
	OnKeyDown   func(*Layout, *Event, *Window)
	OnMouseMove func(*Layout, *Event, *Window)
	OnMouseUp   func(*Layout, *Event, *Window)
	OnScroll    func(*Layout, *Window)
	AmendLayout func(*Layout, *Window)
	OnHover     func(*Layout, *Event, *Window)
	OnIMECommit func(*Layout, string, *Window)

	// Position
	X float32
	Y float32

	// Content
	Content []View

	// Internal — set by factory functions.
	axis                 Axis
	scrollbarOrientation ScrollbarOrientation
}

// containerView implements View for container-based layouts.
type containerView struct {
	cfg       ContainerCfg
	content   []View
	shapeType ShapeType
}

func (cv *containerView) Content() []View { return cv.content }

func (cv *containerView) GenerateLayout(w *Window) Layout {
	c := &cv.cfg
	layout := Layout{
		Shape: &Shape{
			ShapeType:            cv.shapeType,
			ID:                   c.ID,
			IDFocus:              c.IDFocus,
			Axis:                 c.axis,
			ScrollbarOrientation: c.scrollbarOrientation,
			X:                    c.X,
			Y:                    c.Y,
			Width:                c.Width,
			MinWidth:             c.MinWidth,
			MaxWidth:             c.MaxWidth,
			Height:               c.Height,
			MinHeight:            c.MinHeight,
			MaxHeight:            c.MaxHeight,
			Clip:                 c.Clip,
			FocusSkip:            c.FocusSkip,
			Spacing:              c.Spacing,
			Sizing:               c.Sizing,
			Padding:              c.Padding,
			HAlign:               c.HAlign,
			VAlign:               c.VAlign,
			TextDir:              c.TextDir,
			Radius:               c.Radius,
			Color:                c.Color,
			FX:                   cv.makeEffects(),
			SizeBorder:           c.SizeBorder,
			ColorBorder:          c.ColorBorder,
			Disabled:             c.Disabled,
			Float:                c.Float,
			FloatAnchor:          c.FloatAnchor,
			FloatTieOff:          c.FloatTieOff,
			FloatOffsetX:         c.FloatOffsetX,
			FloatOffsetY:         c.FloatOffsetY,
			IDScroll:             c.IDScroll,
			OverDraw:             c.OverDraw,
			ScrollMode:           c.ScrollMode,
			Events:               cv.makeEvents(),
			Hero:                 c.Hero,
			Wrap:                 c.Wrap,
			Overflow:             c.Overflow,
			Opacity:              c.Opacity,
			A11YRole:             cv.deriveA11YRole(),
			A11YState:            c.A11YState,
			A11Y:                 cv.makeA11Y(),
		},
	}
	ApplyFixedSizingConstraints(layout.Shape)
	return layout
}

func (cv *containerView) makeEffects() *ShapeEffects {
	c := &cv.cfg
	if c.Shadow == nil && c.Gradient == nil &&
		c.BorderGradient == nil && c.BlurRadius == 0 {
		return nil
	}
	return &ShapeEffects{
		Shadow:         c.Shadow,
		Gradient:       c.Gradient,
		BorderGradient: c.BorderGradient,
		BlurRadius:     c.BlurRadius,
	}
}

func (cv *containerView) makeEvents() *EventHandlers {
	c := &cv.cfg
	if c.OnClick == nil && c.OnChar == nil &&
		c.OnKeyDown == nil && c.OnMouseMove == nil &&
		c.OnMouseUp == nil && c.OnHover == nil &&
		c.OnIMECommit == nil && c.OnScroll == nil &&
		c.AmendLayout == nil {
		return nil
	}
	return &EventHandlers{
		OnClick:     c.OnClick,
		OnChar:      c.OnChar,
		OnKeyDown:   c.OnKeyDown,
		OnMouseMove: c.OnMouseMove,
		OnMouseUp:   c.OnMouseUp,
		OnHover:     c.OnHover,
		OnIMECommit: c.OnIMECommit,
		OnScroll:    c.OnScroll,
		AmendLayout: c.AmendLayout,
	}
}

func (cv *containerView) makeA11Y() *AccessInfo {
	if cv.cfg.A11Y != nil {
		return cv.cfg.A11Y
	}
	return makeA11YInfo(cv.cfg.A11YLabel, cv.cfg.A11YDescription)
}

func (cv *containerView) deriveA11YRole() AccessRole {
	if cv.cfg.A11YRole != AccessRoleNone {
		return cv.cfg.A11YRole
	}
	if cv.cfg.IDScroll > 0 {
		return AccessRoleScrollArea
	}
	return AccessRoleNone
}

// container is the fundamental layout builder. Factory
// functions (Column, Row, etc.) set axis then delegate here.
func container(cfg ContainerCfg) View {
	if cfg.Invisible {
		return invisibleContainerView()
	}
	// Resolve click handler.
	if cfg.OnAnyClick != nil {
		cfg.OnClick = cfg.OnAnyClick
	} else {
		cfg.OnClick = leftClickOnly(cfg.OnClick)
	}
	// Default opacity.
	if cfg.Opacity == 0 {
		cfg.Opacity = 1.0
	}
	return &containerView{
		cfg:       cfg,
		content:   cfg.Content,
		shapeType: ShapeRectangle,
	}
}

// Column arranges content top to bottom.
func Column(cfg ContainerCfg) View {
	cfg.axis = AxisTopToBottom
	return container(cfg)
}

// Row arranges content left to right.
func Row(cfg ContainerCfg) View {
	cfg.axis = AxisLeftToRight
	return container(cfg)
}

// Wrap arranges content left to right, flowing to the next
// line when container width is exceeded.
func Wrap(cfg ContainerCfg) View {
	cfg.axis = AxisLeftToRight
	cfg.Wrap = true
	return container(cfg)
}

// Canvas does not arrange or layout its content.
func Canvas(cfg ContainerCfg) View {
	return container(cfg)
}

// Circle creates a circular container.
func Circle(cfg ContainerCfg) View {
	cfg.axis = AxisTopToBottom
	cv := container(cfg).(*containerView)
	cv.shapeType = ShapeCircle
	return cv
}

func invisibleContainerView() *containerView {
	return &containerView{
		cfg: ContainerCfg{
			Disabled: true,
			OverDraw: true,
			Padding:  PaddingNone,
		},
		shapeType: ShapeRectangle,
	}
}
