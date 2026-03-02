package gui

import (
	"log"
	"math"
)

// RangeSliderCfg configures a range slider view.
type RangeSliderCfg struct {
	ID               string
	Sizing           Sizing
	Color            Color
	ColorBorder      Color
	ColorThumb       Color
	ColorFocus       Color
	ColorHover       Color
	ColorLeft        Color
	ColorClick       Color
	Padding          Padding
	SizeBorder       float32
	OnChange         func(float32, *Event, *Window)
	Value            float32
	Min              float32
	Max              float32
	Step             float32
	Width            float32
	Height           float32
	Size             float32
	ThumbSize        float32
	Radius           float32
	RadiusBorder     float32
	IDFocus          uint32
	RoundValue       bool
	Vertical         bool
	Disabled         bool
	Invisible        bool

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// RangeSlider creates a range slider view.
func RangeSlider(cfg RangeSliderCfg) View {
	if cfg.Color == (Color{}) {
		cfg.Color = guiTheme.RangeSliderStyle.Color
	}
	if cfg.ColorBorder == (Color{}) {
		cfg.ColorBorder = guiTheme.RangeSliderStyle.ColorBorder
	}
	if cfg.ColorThumb == (Color{}) {
		cfg.ColorThumb = guiTheme.RangeSliderStyle.ColorThumb
	}
	if cfg.ColorFocus == (Color{}) {
		cfg.ColorFocus = guiTheme.RangeSliderStyle.ColorFocus
	}
	if cfg.ColorHover == (Color{}) {
		cfg.ColorHover = guiTheme.RangeSliderStyle.ColorHover
	}
	if cfg.ColorLeft == (Color{}) {
		cfg.ColorLeft = guiTheme.RangeSliderStyle.ColorLeft
	}
	if cfg.ColorClick == (Color{}) {
		cfg.ColorClick = guiTheme.RangeSliderStyle.ColorClick
	}
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = guiTheme.RangeSliderStyle.SizeBorder
	}
	if cfg.Size == 0 {
		cfg.Size = guiTheme.RangeSliderStyle.Size
	}
	if cfg.ThumbSize == 0 {
		cfg.ThumbSize = guiTheme.RangeSliderStyle.ThumbSize
	}
	if cfg.Radius == 0 {
		cfg.Radius = guiTheme.RangeSliderStyle.Radius
	}
	if cfg.Max == 0 && cfg.Min == 0 {
		cfg.Max = 100
	}
	if cfg.Step == 0 {
		cfg.Step = 1
	}

	if cfg.Min >= cfg.Max {
		log.Printf("range_slider: min (%f) >= max (%f); adjusting",
			cfg.Min, cfg.Max)
		cfg.Max = cfg.Min + 1
	}

	wrapperWidth := cfg.Size
	wrapperHeight := f32Max(cfg.Size, cfg.ThumbSize)
	trackWidth := float32(0)
	trackHeight := cfg.Size

	if cfg.Vertical {
		wrapperWidth = f32Max(cfg.Size, cfg.ThumbSize)
		wrapperHeight = cfg.Size
		trackWidth = cfg.Size
		trackHeight = 0
	}
	if cfg.Width > 0 {
		wrapperWidth = cfg.Width
	}
	if cfg.Height > 0 {
		wrapperHeight = cfg.Height
	}

	sliderID := cfg.ID
	onChange := cfg.OnChange
	value := cfg.Value
	min := cfg.Min
	max := cfg.Max
	step := cfg.Step
	vertical := cfg.Vertical
	roundValue := cfg.RoundValue
	size := cfg.Size
	szBorder := cfg.SizeBorder
	thumbSize := cfg.ThumbSize
	colorFocus := cfg.ColorFocus
	colorHover := cfg.ColorHover
	colorClick := cfg.ColorClick
	disabled := cfg.Disabled
	idFocus := cfg.IDFocus

	trackSizing := FillFixed
	if cfg.Vertical {
		trackSizing = Sizing{SizingFixed, SizingFill}
	}

	trackAxis := AxisLeftToRight
	if cfg.Vertical {
		trackAxis = AxisTopToBottom
	}

	wrapperAxis := AxisLeftToRight
	if cfg.Vertical {
		wrapperAxis = AxisTopToBottom
	}

	return container(ContainerCfg{
		ID:       cfg.ID,
		IDFocus:  cfg.IDFocus,
		A11YRole: AccessRoleSlider,
		A11Y: &AccessInfo{
			Label:       a11yLabel(cfg.A11YLabel, cfg.ID),
			Description: cfg.A11YDescription,
			ValueNum:    cfg.Value,
			ValueMin:    cfg.Min,
			ValueMax:    cfg.Max,
		},
		Width:     wrapperWidth,
		Height:    wrapperHeight,
		Disabled:  cfg.Disabled,
		Invisible: cfg.Invisible,
		Padding:   Some(PaddingNone),
		Sizing:    cfg.Sizing,
		HAlign:    HAlignCenter,
		VAlign:    VAlignMiddle,
		axis:      wrapperAxis,
		OnClick: func(layout *Layout, e *Event, w *Window) {
			ev := *e
			ev.MouseX = e.MouseX + layout.Shape.X
			ev.MouseY = e.MouseY + layout.Shape.Y
			rangeSliderMouseMove(layout, &ev, w,
				sliderID, onChange, value,
				min, max, vertical, roundValue)
			w.MouseLock(MouseLockCfg{
				MouseMove: func(
					layout *Layout, e *Event, w *Window,
				) {
					rangeSliderMouseMove(layout, e, w,
						sliderID, onChange, value,
						min, max, vertical, roundValue)
				},
				MouseUp: func(
					_ *Layout, _ *Event, w *Window,
				) {
					w.MouseUnlock()
				},
			})
			e.IsHandled = true
		},
		AmendLayout: func(layout *Layout, w *Window) {
			rangeSliderAmendLayoutSlide(layout, w,
				onChange, value, min, max, size, szBorder,
				vertical, colorFocus, disabled, idFocus,
				roundValue)
		},
		OnHover: func(layout *Layout, e *Event, w *Window) {
			w.SetMouseCursorPointingHand()
			if len(layout.Children) > 0 {
				layout.Children[0].Shape.ColorBorder = colorHover
				if e.MouseButton == MouseLeft &&
					len(layout.Children[0].Children) > 1 {
					layout.Children[0].Children[1].Shape.ColorBorder = colorClick
				}
			}
		},
		OnKeyDown: func(layout *Layout, e *Event, w *Window) {
			rangeSliderOnKeyDown(layout, e, w,
				onChange, value, min, max, step, roundValue)
		},
		Content: []View{
			container(ContainerCfg{
				Width:       trackWidth,
				Height:      trackHeight,
				Sizing:      trackSizing,
				Color:       cfg.Color,
				ColorBorder: cfg.ColorBorder,
				SizeBorder:  Some(cfg.SizeBorder),
				Radius:      Some(cfg.RadiusBorder),
				Padding:     Some(PaddingNone),
				axis:        trackAxis,
				Content: []View{
					Rectangle(RectangleCfg{
						Sizing:      FillFill,
						Color:       cfg.ColorLeft,
						ColorBorder: cfg.ColorLeft,
					}),
					Circle(ContainerCfg{
						Width:       cfg.ThumbSize,
						Height:      cfg.ThumbSize,
						Color:       cfg.ColorThumb,
						ColorBorder: cfg.ColorBorder,
						SizeBorder:  Some[float32](1.5),
						Padding:     Some(PaddingNone),
						AmendLayout: func(
							layout *Layout, w *Window,
						) {
							rangeSliderAmendLayoutThumb(
								layout, w, value,
								min, max, thumbSize,
								vertical)
						},
					}),
				},
			}),
		},
	})
}

func rangeSliderAmendLayoutSlide(
	layout *Layout, w *Window,
	onChange func(float32, *Event, *Window),
	value, min, max, size, sizeBorder float32,
	vertical bool, colorFocus Color,
	disabled bool, idFocus uint32, roundValue bool,
) {
	if layout.Shape.Events == nil {
		layout.Shape.Events = &EventHandlers{}
	}
	layout.Shape.Events.OnMouseScroll = func(
		_ *Layout, e *Event, w *Window,
	) {
		rangeSliderOnMouseScroll(e, w, onChange,
			value, min, max, roundValue)
	}

	if len(layout.Children) == 0 {
		return
	}
	track := &layout.Children[0]
	if len(track.Children) < 2 {
		return
	}
	leftBar := &track.Children[0]
	thumb := &track.Children[1]

	clamped := f32Clamp(value, min, max)
	percent := float32(math.Abs(float64(clamped / (max - min))))

	if vertical {
		h := track.Shape.Height
		y := f32Min(h*percent, h)
		leftBar.Shape.Height = y
		leftBar.Shape.Width = size - sizeBorder*2
	} else {
		wd := track.Shape.Width
		x := f32Min(wd*percent, wd)
		leftBar.Shape.Width = x
		leftBar.Shape.Height = size - sizeBorder*2
	}

	if disabled {
		return
	}
	if w.IsFocus(idFocus) {
		thumb.Shape.Color = colorFocus
		thumb.Shape.ColorBorder = colorFocus
	}
}

func rangeSliderAmendLayoutThumb(
	layout *Layout, _ *Window,
	value, min, max, thumbSize float32, vertical bool,
) {
	clamped := f32Clamp(value, min, max)
	percent := float32(math.Abs(float64(clamped / (max - min))))
	radius := thumbSize / 2

	if vertical {
		h := layout.Parent.Shape.Height
		y := f32Min(h*percent, h)
		layout.Shape.Y = layout.Parent.Shape.Y + y - radius
		layout.Shape.X = layout.Parent.Shape.X +
			layout.Parent.Shape.Width/2 - radius
	} else {
		wd := layout.Parent.Shape.Width
		x := f32Min(wd*percent, wd)
		layout.Shape.X = layout.Parent.Shape.X + x - radius
		layout.Shape.Y = layout.Parent.Shape.Y +
			layout.Parent.Shape.Height/2 - radius
	}
}

func rangeSliderMouseMove(
	layout *Layout, e *Event, w *Window,
	sliderID string,
	onChange func(float32, *Event, *Window),
	curValue, min, max float32,
	vertical, roundValue bool,
) {
	if onChange == nil {
		return
	}
	sl, ok := layout.FindLayout(func(n Layout) bool {
		return n.Shape.ID == sliderID
	})
	if !ok {
		return
	}
	w.SetMouseCursorPointingHand()
	shape := sl.Shape
	if vertical {
		h := shape.Height
		pct := f32Clamp((e.MouseY-shape.Y)/h, 0, 1)
		val := (max - min) * pct
		v := f32Clamp(val, min, max)
		if roundValue {
			v = float32(math.Round(float64(v)))
		}
		onChange(v, e, w)
	} else {
		wd := shape.Width
		pct := f32Clamp((e.MouseX-shape.X)/wd, 0, 1)
		val := (max - min) * pct
		v := f32Clamp(val, min, max)
		if roundValue {
			v = float32(math.Round(float64(v)))
		}
		if v != curValue {
			onChange(v, e, w)
		}
	}
}

func rangeSliderOnKeyDown(
	_ *Layout, e *Event, w *Window,
	onChange func(float32, *Event, *Window),
	curValue, min, max, step float32, roundValue bool,
) {
	if onChange == nil || e.Modifiers != ModNone {
		return
	}
	v := curValue
	switch e.KeyCode {
	case KeyHome:
		v = min
	case KeyEnd:
		v = max
	case KeyLeft, KeyUp:
		v = f32Clamp(v-step, min, max)
	case KeyRight, KeyDown:
		v = f32Clamp(v+step, min, max)
	default:
		return
	}
	if roundValue {
		v = float32(math.Round(float64(v)))
	}
	if v != curValue {
		onChange(v, e, w)
	}
}

func rangeSliderOnMouseScroll(
	e *Event, w *Window,
	onChange func(float32, *Event, *Window),
	curValue, min, max float32, roundValue bool,
) {
	e.IsHandled = true
	if onChange == nil || e.Modifiers != ModNone {
		return
	}
	v := f32Clamp(curValue+e.ScrollY, min, max)
	if roundValue {
		v = float32(math.Round(float64(v)))
	}
	if v != curValue {
		onChange(v, e, w)
	}
}
