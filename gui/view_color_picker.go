package gui

import (
	"fmt"
	"strconv"
)

// colorPickerState stores persistent HSV to preserve hue when
// saturation or value goes to zero.
type colorPickerState struct {
	H float32
	S float32
	V float32
}

// ColorPickerCfg configures a color picker view.
type ColorPickerCfg struct {
	ID            string
	Color         Color
	OnColorChange func(Color, *Event, *Window)
	Style         ColorPickerStyle
	IDFocus       uint32
	ShowHSV       bool
	Sizing        Sizing
	Width         float32
	Height        float32

	A11YLabel       string
	A11YDescription string
}

type colorPickerView struct {
	cfg ColorPickerCfg
}

// ColorPicker creates a color picker view with SV area, hue slider,
// alpha slider, hex input, and RGBA/HSV channel inputs.
func ColorPicker(cfg ColorPickerCfg) View {
	applyColorPickerDefaults(&cfg)
	return &colorPickerView{cfg: cfg}
}

func (cv *colorPickerView) Content() []View { return nil }

func (cv *colorPickerView) GenerateLayout(w *Window) Layout {
	cfg := &cv.cfg
	style := cfg.Style

	// Get or init HSV state.
	sm := StateMap[string, colorPickerState](
		w, nsColorPicker, capModerate)
	hsv, ok := sm.Get(cfg.ID)
	if !ok {
		h, s, v := cfg.Color.ToHSV()
		hsv = colorPickerState{H: h, S: s, V: v}
		sm.Set(cfg.ID, hsv)
	}

	svSize := style.SVSize
	sliderH := style.SliderHeight
	indicatorSize := style.IndicatorSize

	var content []View

	// SV area + hue slider side by side.
	content = append(content, cpSVAndHueRow(cfg, hsv,
		svSize, sliderH, indicatorSize))

	// Alpha slider.
	content = append(content, cpAlphaSlider(cfg))

	// Preview + hex row.
	content = append(content, cpPreviewRow(cfg))

	// RGBA inputs.
	content = append(content, cpRGBAInputs(cfg))

	// Optional HSV inputs.
	if cfg.ShowHSV {
		content = append(content, cpHSVInputs(cfg, hsv))
	}

	col := &containerView{
		cfg: ContainerCfg{
			ID:          cfg.ID,
			IDFocus:     cfg.IDFocus,
			A11YLabel:   a11yLabel(cfg.A11YLabel, "Color Picker"),
			Color:       style.Color,
			ColorBorder: style.ColorBorder,
			SizeBorder:  Some(style.SizeBorder),
			Radius:      Some(style.Radius),
			Padding:     style.Padding,
			Spacing:     Some(SpacingSmall),
			Sizing:      cfg.Sizing,
			Width:       cfg.Width,
			Height:      cfg.Height,
			Content:     content,
			axis:        AxisTopToBottom,
		},
		content:   content,
		shapeType: ShapeRectangle,
	}
	return GenerateViewLayout(col, w)
}

// cpSVAndHueRow builds the SV area and hue slider row.
func cpSVAndHueRow(
	cfg *ColorPickerCfg, hsv colorPickerState,
	svSize, sliderH, indicatorSize float32,
) View {
	return Row(ContainerCfg{
		Padding: PaddingNone,
		Spacing: Some(SpacingSmall),
		Content: []View{
			cpSVArea(cfg, hsv, svSize, indicatorSize),
			cpHueSlider(cfg, hsv, sliderH, svSize,
				indicatorSize),
		},
	})
}

// cpSVArea builds the saturation/value area with gradients.
func cpSVArea(
	cfg *ColorPickerCfg, hsv colorPickerState,
	size, indicatorSize float32,
) View {
	pureHue := HueColor(hsv.H)
	cfgID := cfg.ID
	cfgColor := cfg.Color
	onChange := cfg.OnColorChange

	return container(ContainerCfg{
		ID:     cfg.ID + ".sv",
		Width:  size,
		Height: size,
		axis:   AxisTopToBottom,
		Gradient: &GradientDef{
			Direction: GradientToRight,
			Stops: []GradientStop{
				{Color: White, Pos: 0},
				{Color: pureHue, Pos: 1},
			},
		},
		Padding:    PaddingNone,
		SizeBorder: Some(float32(0)),
		Radius:     Some(cfg.Style.Radius),
		Content: []View{
			container(ContainerCfg{
				Sizing: FillFill,
				Gradient: &GradientDef{
					Direction: GradientToBottom,
					Stops: []GradientStop{
						{Color: ColorTransparent, Pos: 0},
						{Color: Black, Pos: 1},
					},
				},
				Padding: PaddingNone,
				Content: []View{
					Circle(ContainerCfg{
						Width:       indicatorSize,
						Height:      indicatorSize,
						Color:       cfgColor,
						ColorBorder: White,
						SizeBorder:  Some(float32(2)),
						Padding:     PaddingNone,
						AmendLayout: func(
							layout *Layout, _ *Window,
						) {
							cpAmendSVIndicator(layout, hsv,
								size, indicatorSize)
						},
					}),
				},
				OnClick: func(
					layout *Layout, e *Event, w *Window,
				) {
					ev := *e
					ev.MouseX += layout.Shape.X
					ev.MouseY += layout.Shape.Y
					cpSVMouseAction(cfgID, cfgColor,
						onChange, layout.Shape, &ev, w)
					w.MouseLock(MouseLockCfg{
						MouseMove: func(
							layout *Layout, e *Event,
							w *Window,
						) {
							sv, ok := layout.FindByID(
								cfgID + ".sv")
							if !ok {
								return
							}
							cpSVMouseAction(cfgID, cfgColor,
								onChange, sv.Shape, e, w)
						},
						MouseUp: func(
							_ *Layout, _ *Event,
							w *Window,
						) {
							w.MouseUnlock()
							w.SetMouseCursorArrow()
						},
					})
					e.IsHandled = true
				},
			}),
		},
	})
}

// cpHueSlider builds the vertical hue slider.
func cpHueSlider(
	cfg *ColorPickerCfg, hsv colorPickerState,
	sliderWidth, sliderHeight, indicatorSize float32,
) View {
	cfgID := cfg.ID
	cfgColor := cfg.Color
	onChange := cfg.OnColorChange

	return container(ContainerCfg{
		ID:     cfg.ID + ".hue",
		Width:  sliderWidth,
		Height: sliderHeight,
		Gradient: &GradientDef{
			Direction: GradientToBottom,
			Stops: []GradientStop{
				{Color: RGB(255, 0, 0), Pos: 0.0},
				{Color: RGB(255, 255, 0), Pos: 1.0 / 6},
				{Color: RGB(0, 255, 0), Pos: 2.0 / 6},
				{Color: RGB(0, 255, 255), Pos: 3.0 / 6},
				{Color: RGB(0, 0, 255), Pos: 4.0 / 6},
				{Color: RGB(255, 0, 255), Pos: 5.0 / 6},
				{Color: RGB(255, 0, 0), Pos: 1.0},
			},
		},
		Padding: PaddingNone,
		Radius:  Some(cfg.Style.Radius),
		Content: []View{
			Circle(ContainerCfg{
				Width:       indicatorSize,
				Height:      indicatorSize,
				Color:       HueColor(hsv.H),
				ColorBorder: White,
				SizeBorder:  Some(float32(2)),
				Padding:     PaddingNone,
				AmendLayout: func(
					layout *Layout, _ *Window,
				) {
					cpAmendHueIndicator(layout, hsv,
						sliderHeight, indicatorSize)
				},
			}),
		},
		OnClick: func(
			layout *Layout, e *Event, w *Window,
		) {
			ev := *e
			ev.MouseX += layout.Shape.X
			ev.MouseY += layout.Shape.Y
			cpHueMouseAction(cfgID, cfgColor, onChange,
				layout.Shape, &ev, w)
			w.MouseLock(MouseLockCfg{
				MouseMove: func(
					layout *Layout, e *Event, w *Window,
				) {
					hue, ok := layout.FindByID(
						cfgID + ".hue")
					if !ok {
						return
					}
					cpHueMouseAction(cfgID, cfgColor,
						onChange, hue.Shape, e, w)
				},
				MouseUp: func(
					_ *Layout, _ *Event, w *Window,
				) {
					w.MouseUnlock()
				},
			})
			e.IsHandled = true
		},
	})
}

// cpAlphaSlider builds the alpha channel slider.
func cpAlphaSlider(cfg *ColorPickerCfg) View {
	onChange := cfg.OnColorChange
	c := cfg.Color
	thumbSize := cfg.Style.IndicatorSize
	trackSize := float32(6)
	return RangeSlider(RangeSliderCfg{
		ID:           cfg.ID + ".alpha",
		Value:        float32(c.A),
		Min:          0,
		Max:          255,
		Step:         1,
		Sizing:       FillFit,
		Size:         trackSize,
		SizeBorder:   1,
		ThumbSize:    thumbSize,
		Height:       thumbSize,
		RadiusBorder: trackSize / 2,
		OnChange: func(v float32, e *Event, w *Window) {
			if onChange != nil {
				nc := c
				nc.A = uint8(f32Clamp(v, 0, 255))
				onChange(nc, e, w)
			}
		},
		RoundValue: true,
	})
}

// cpPreviewRow builds the color swatch + hex input row.
func cpPreviewRow(cfg *ColorPickerCfg) View {
	c := cfg.Color
	onChange := cfg.OnColorChange
	cfgID := cfg.ID

	return Row(ContainerCfg{
		Padding: PaddingNone,
		Spacing: Some(SpacingSmall),
		VAlign:  VAlignMiddle,
		Content: []View{
			Rectangle(RectangleCfg{
				Width:  32,
				Height: 32,
				Color:  c,
				Radius: cfg.Style.Radius,
			}),
			Input(InputCfg{
				ID:        cfgID + ".hex",
				Text:      c.ToHexString(),
				TextStyle: cfg.Style.TextStyle,
				Width:     100,
				OnTextCommit: func(
					_ *Layout, text string,
					_ InputCommitReason, w *Window,
				) {
					nc, ok := ColorFromHexString(text)
					if ok && onChange != nil {
						h, s, v := nc.ToHSV()
						sm := StateMap[string, colorPickerState](
							w, nsColorPicker, capModerate)
						sm.Set(cfgID, colorPickerState{
							H: h, S: s, V: v,
						})
						// Use a nil event for commit-triggered
						// changes; caller should handle.
						onChange(nc, nil, w)
					}
				},
			}),
		},
	})
}

// cpRGBAInputs builds the R/G/B channel inputs row.
func cpRGBAInputs(cfg *ColorPickerCfg) View {
	c := cfg.Color
	return Row(ContainerCfg{
		Padding: PaddingNone,
		Spacing: Some(SpacingSmall),
		Content: []View{
			cpChannelInput(cfg, "R", c.R, 0),
			cpChannelInput(cfg, "G", c.G, 1),
			cpChannelInput(cfg, "B", c.B, 2),
		},
	})
}

// cpHSVInputs builds the H/S/V channel inputs row.
func cpHSVInputs(
	cfg *ColorPickerCfg, hsv colorPickerState,
) View {
	return Row(ContainerCfg{
		Padding: PaddingNone,
		Spacing: Some(SpacingSmall),
		Content: []View{
			cpHSVChannelInput(cfg, "H", int(hsv.H), 360, 0),
			cpHSVChannelInput(cfg, "S",
				int(hsv.S*100), 100, 1),
			cpHSVChannelInput(cfg, "V",
				int(hsv.V*100), 100, 2),
		},
	})
}

// cpChannelInput builds a labeled RGB channel input.
func cpChannelInput(
	cfg *ColorPickerCfg, ch string, val uint8, idx int,
) View {
	onChange := cfg.OnColorChange
	cfgID := cfg.ID
	c := cfg.Color

	return Column(ContainerCfg{
		Padding: PaddingNone,
		Spacing: Some(float32(2)),
		Content: []View{
			Text(TextCfg{
				Text: ch,
				TextStyle: TextStyle{
					Color: cfg.Style.TextStyle.Color,
					Size:  cfg.Style.TextStyle.Size,
					Align: TextAlignCenter,
				},
			}),
			Input(InputCfg{
				ID:        fmt.Sprintf("%s.%s", cfgID, ch),
				Text:      fmt.Sprintf("%d", val),
				TextStyle: cfg.Style.TextStyle,
				Width:     50,
				OnTextCommit: func(
					_ *Layout, text string,
					_ InputCommitReason, w *Window,
				) {
					v, ok := cpParseUint8(text)
					if !ok || onChange == nil {
						return
					}
					nc := c
					switch idx {
					case 0:
						nc.R = v
					case 1:
						nc.G = v
					case 2:
						nc.B = v
					}
					h, s, vv := nc.ToHSV()
					sm := StateMap[string, colorPickerState](
						w, nsColorPicker, capModerate)
					sm.Set(cfgID, colorPickerState{
						H: h, S: s, V: vv,
					})
					onChange(nc, nil, w)
				},
			}),
		},
	})
}

// cpHSVChannelInput builds a labeled HSV channel input.
func cpHSVChannelInput(
	cfg *ColorPickerCfg, ch string,
	val, maxVal int, idx int,
) View {
	onChange := cfg.OnColorChange
	cfgID := cfg.ID

	return Column(ContainerCfg{
		Padding: PaddingNone,
		Spacing: Some(float32(2)),
		Content: []View{
			Text(TextCfg{
				Text: ch,
				TextStyle: TextStyle{
					Color: cfg.Style.TextStyle.Color,
					Size:  cfg.Style.TextStyle.Size,
					Align: TextAlignCenter,
				},
			}),
			Input(InputCfg{
				ID:        fmt.Sprintf("%s.hsv.%s", cfgID, ch),
				Text:      fmt.Sprintf("%d", val),
				TextStyle: cfg.Style.TextStyle,
				Width:     50,
				OnTextCommit: func(
					_ *Layout, text string,
					_ InputCommitReason, w *Window,
				) {
					n, err := strconv.Atoi(text)
					if err != nil || onChange == nil {
						return
					}
					if n < 0 {
						n = 0
					}
					if n > maxVal {
						n = maxVal
					}
					sm := StateMap[string, colorPickerState](
						w, nsColorPicker, capModerate)
					hsv, _ := sm.Get(cfgID)
					switch idx {
					case 0:
						hsv.H = float32(n)
					case 1:
						hsv.S = float32(n) / 100
					case 2:
						hsv.V = float32(n) / 100
					}
					sm.Set(cfgID, hsv)
					nc := ColorFromHSVA(
						hsv.H, hsv.S, hsv.V, cfg.Color.A)
					onChange(nc, nil, w)
				},
			}),
		},
	})
}

// cpSVMouseAction handles mouse interaction in the SV area.
func cpSVMouseAction(
	id string, color Color,
	onChange func(Color, *Event, *Window),
	shape *Shape, e *Event, w *Window,
) {
	if onChange == nil {
		return
	}
	s := f32Clamp((e.MouseX-shape.X)/shape.Width, 0, 1)
	v := 1 - f32Clamp((e.MouseY-shape.Y)/shape.Height, 0, 1)

	sm := StateMap[string, colorPickerState](
		w, nsColorPicker, capModerate)
	hsv, _ := sm.Get(id)
	hsv.S = s
	hsv.V = v
	sm.Set(id, hsv)

	nc := ColorFromHSVA(hsv.H, s, v, color.A)
	onChange(nc, e, w)
}

// cpHueMouseAction handles mouse interaction on the hue slider.
func cpHueMouseAction(
	id string, color Color,
	onChange func(Color, *Event, *Window),
	shape *Shape, e *Event, w *Window,
) {
	if onChange == nil {
		return
	}
	h := f32Clamp(
		(e.MouseY-shape.Y)/shape.Height, 0, 1) * 360

	sm := StateMap[string, colorPickerState](
		w, nsColorPicker, capModerate)
	hsv, _ := sm.Get(id)
	hsv.H = h
	sm.Set(id, hsv)

	nc := ColorFromHSVA(h, hsv.S, hsv.V, color.A)
	onChange(nc, e, w)
}

// cpAmendSVIndicator positions the SV indicator circle.
func cpAmendSVIndicator(
	layout *Layout, hsv colorPickerState,
	svSize, indicatorSize float32,
) {
	if layout.Parent == nil {
		return
	}
	parent := layout.Parent
	radius := indicatorSize / 2
	layout.Shape.X = parent.Shape.X +
		hsv.S*svSize - radius
	layout.Shape.Y = parent.Shape.Y +
		(1-hsv.V)*svSize - radius
}

// cpAmendHueIndicator positions the hue slider indicator.
func cpAmendHueIndicator(
	layout *Layout, hsv colorPickerState,
	sliderHeight, indicatorSize float32,
) {
	if layout.Parent == nil {
		return
	}
	parent := layout.Parent
	radius := indicatorSize / 2
	y := (hsv.H / 360) * sliderHeight
	layout.Shape.X = parent.Shape.X +
		parent.Shape.Width/2 - radius
	layout.Shape.Y = parent.Shape.Y + y - radius
}

func applyColorPickerDefaults(cfg *ColorPickerCfg) {
	d := &DefaultColorPickerStyle
	if cfg.Style == (ColorPickerStyle{}) {
		cfg.Style = *d
	}
	if !cfg.Color.IsSet() {
		cfg.Color = Red
	}
}

// cpParseUint8 parses a string as a uint8 value.
func cpParseUint8(s string) (uint8, bool) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > 255 {
		return 0, false
	}
	return uint8(n), true
}
