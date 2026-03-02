package gui

import (
	"fmt"
	"log"
	"time"
)

// SvgCfg configures an SVG view component.
type SvgCfg struct {
	ID       string
	FileName string  // SVG file path
	SvgData  string  // OR inline SVG string
	Width    float32 // display width
	Height   float32 // display height
	Color    Color   // override fill (for monochrome icons)
	Animated bool    // enable SMIL animation (default true)
	Sizing   Sizing
	Padding  Padding
	OnClick  func(*Layout, *Event, *Window)

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// svgView implements View for SVG rendering.
type svgView struct {
	cfg SvgCfg
}

// Svg creates an SVG view from file or inline data.
func Svg(cfg SvgCfg) View {
	// Default: animated = true (Go zero value is false).
	// Callers must explicitly set Animated=false to disable.
	return &svgView{cfg: cfg}
}

func (sv *svgView) Content() []View { return nil }

func (sv *svgView) GenerateLayout(w *Window) Layout {
	c := &sv.cfg
	svgSrc := c.FileName
	if svgSrc == "" {
		svgSrc = c.SvgData
	}

	width := c.Width
	height := c.Height

	if width <= 0 || height <= 0 {
		natW, natH, err := w.GetSvgDimensions(svgSrc)
		if err != nil {
			log.Printf("svg: %v", err)
			return svgErrorLayout(svgSrc, w)
		}
		if width <= 0 {
			width = natW
		}
		if height <= 0 {
			height = natH
		}
	}

	cached, err := w.LoadSvg(svgSrc, width, height)
	if err != nil {
		log.Printf("svg: %v", err)
		return svgErrorLayout(svgSrc, w)
	}

	// Register animation loop for animated SVGs.
	if cached.HasAnimations && c.Animated {
		animHash := fmt.Sprintf("%x", hashString(svgSrc))
		nowNs := time.Now().UnixNano()
		animSeen := StateMap[string, int64](
			w, nsSvgAnimSeen, capModerate)
		animStart := StateMap[string, int64](
			w, nsSvgAnimStart, capModerate)
		animSeen.Set(animHash, nowNs)
		if !animStart.Contains(animHash) {
			animStart.Set(animHash, nowNs)
		}
		animID := "svg_anim:" + animHash
		if !w.HasAnimation(animID) {
			w.AnimationAdd(&Animate{
				AnimateID: animID,
				Delay:     animationCycle,
				Repeat:    true,
				Callback: func(an *Animate, w *Window) {
					seenMap := StateMap[string, int64](
						w, nsSvgAnimSeen, capModerate)
					seen, ok := seenMap.Get(animHash)
					if !ok {
						an.stopped = true
						return
					}
					elapsed := time.Now().UnixNano() - seen
					if elapsed > 200_000_000 {
						an.stopped = true
						return
					}
					w.UpdateWindow()
				},
			})
		}
	}

	var events *EventHandlers
	onClick := leftClickOnly(c.OnClick)
	if onClick != nil {
		events = &EventHandlers{
			OnClick: onClick,
		}
	}
	layout := Layout{
		Shape: &Shape{
			ShapeType: ShapeSVG,
			ID:        c.ID,
			A11YRole:  AccessRoleImage,
			A11Y: makeA11YInfo(
				a11yLabel(c.A11YLabel, c.ID),
				c.A11YDescription,
			),
			Resource: svgSrc,
			Width:    width,
			Height:   height,
			Color:    c.Color,
			Sizing:   c.Sizing,
			Padding:  c.Padding,
			Events:   events,
		},
	}
	ApplyFixedSizingConstraints(layout.Shape)
	return layout
}

// svgErrorLayout returns a magenta error text for missing SVGs.
func svgErrorLayout(src string, w *Window) Layout {
	ts := guiTheme.TextStyleDef
	ts.Color = Magenta
	tv := Text(TextCfg{
		Text:      fmt.Sprintf("[missing: %s]", src),
		TextStyle: ts,
	})
	return tv.GenerateLayout(w)
}
