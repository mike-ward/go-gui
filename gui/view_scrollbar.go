package gui

// ScrollbarOverflow determines when scrollbars are shown.
type ScrollbarOverflow uint8

const (
	ScrollbarAuto    ScrollbarOverflow = iota
	ScrollbarHidden
	ScrollbarVisible
	ScrollbarOnHover
)

// ScrollbarCfg configures the style of a scrollbar.
type ScrollbarCfg struct {
	ID              string
	ColorThumb      Color
	ColorBackground Color
	Size            float32
	MinThumbSize    float32
	Radius          float32
	RadiusThumb     float32
	GapEdge         float32
	GapEnd          float32
	IDScroll        uint32
	Overflow        ScrollbarOverflow
	Orientation     ScrollbarOrientation
}

// Scrollbar layout constants.
const (
	scrollExtend  = 10
	scrollSnapMin = float32(0.03)
	scrollSnapMax = float32(0.97)
	thumbIndex    = 0
)

func applyScrollbarDefaults(cfg *ScrollbarCfg) {
	if cfg.ColorThumb == (Color{}) {
		cfg.ColorThumb = RGB(140, 140, 140)
	}
	if cfg.ColorBackground == (Color{}) {
		cfg.ColorBackground = ColorTransparent
	}
	if cfg.Size == 0 {
		cfg.Size = 8
	}
	if cfg.MinThumbSize == 0 {
		cfg.MinThumbSize = 20
	}
	if cfg.Radius == 0 {
		cfg.Radius = RadiusSmall
	}
	if cfg.RadiusThumb == 0 {
		cfg.RadiusThumb = RadiusSmall
	}
	if cfg.GapEdge == 0 {
		cfg.GapEdge = 2
	}
	if cfg.GapEnd == 0 {
		cfg.GapEnd = 2
	}
}

// Scrollbar creates a scrollbar overlay view.
func Scrollbar(cfg ScrollbarCfg) View {
	applyScrollbarDefaults(&cfg)

	thumbView := scrollbarThumb(cfg)

	if cfg.Orientation == ScrollbarHorizontal {
		return Row(ContainerCfg{
			ID:                   cfg.ID,
			A11YRole:             AccessRoleScrollBar,
			Color:                cfg.ColorBackground,
			OverDraw:             true,
			Spacing:              0,
			Padding:              PaddingNone,
			scrollbarOrientation: ScrollbarHorizontal,
			AmendLayout:          makeScrollbarAmendLayout(cfg),
			OnHover:              makeScrollbarOnHover(cfg),
			Content:              []View{thumbView},
		})
	}
	return Column(ContainerCfg{
		ID:                   cfg.ID,
		A11YRole:             AccessRoleScrollBar,
		Color:                cfg.ColorBackground,
		OverDraw:             true,
		Spacing:              0,
		Padding:              PaddingNone,
		scrollbarOrientation: ScrollbarVertical,
		AmendLayout:          makeScrollbarAmendLayout(cfg),
		OnHover:              makeScrollbarOnHover(cfg),
		Content:              []View{thumbView},
	})
}

func scrollbarThumb(cfg ScrollbarCfg) View {
	return Column(ContainerCfg{
		Color:   cfg.ColorThumb,
		Radius:  cfg.RadiusThumb,
		Padding: PaddingNone,
		Spacing: 0,
	})
}

func makeScrollbarAmendLayout(cfg ScrollbarCfg) func(*Layout, *Window) {
	return func(layout *Layout, w *Window) {
		scrollbarAmendLayout(cfg, layout, w)
	}
}

func makeScrollbarOnHover(cfg ScrollbarCfg) func(*Layout, *Event, *Window) {
	return func(layout *Layout, _ *Event, w *Window) {
		if len(layout.Children) == 0 {
			return
		}
		if layout.Children[thumbIndex].Shape.Color != ColorTransparent ||
			cfg.Overflow == ScrollbarOnHover {
			layout.Children[thumbIndex].Shape.Color = cfg.ColorThumb
			w.SetMouseCursor(CursorArrow)
		}
	}
}

func scrollbarAmendLayout(cfg ScrollbarCfg, layout *Layout, w *Window) {
	if layout.Parent == nil || len(layout.Children) == 0 {
		return
	}
	parent := layout.Parent

	if cfg.Orientation == ScrollbarHorizontal {
		layout.Shape.X = parent.Shape.X + parent.Shape.Padding.Left
		layout.Shape.Y = parent.Shape.Y + parent.Shape.Height - cfg.Size
		layout.Shape.Width = parent.Shape.Width - parent.Shape.Padding.Width()
		layout.Shape.Height = cfg.Size

		cWidth := contentWidth(parent)
		if cWidth == 0 {
			return
		}
		tWidth := layout.Shape.Width * (layout.Shape.Width / cWidth)
		thumbWidth := f32Clamp(tWidth, cfg.MinThumbSize, layout.Shape.Width)
		availWidth := layout.Shape.Width - thumbWidth

		sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
		scrollOffset := float32(0)
		if v, ok := sx.Get(cfg.IDScroll); ok {
			scrollOffset = -v
		}

		layout.Shape.X -= cfg.GapEnd
		layout.Shape.Y -= cfg.GapEdge
		layout.Shape.Width -= cfg.GapEnd + cfg.GapEnd

		offset := float32(0)
		if availWidth > 0 {
			offset = f32Clamp(
				(scrollOffset/(cWidth-layout.Shape.Width))*availWidth,
				0, availWidth)
		}
		layout.Children[thumbIndex].Shape.X = layout.Shape.X + offset
		layout.Children[thumbIndex].Shape.Y = layout.Shape.Y
		layout.Children[thumbIndex].Shape.Width = thumbWidth - cfg.GapEnd - cfg.GapEnd
		layout.Children[thumbIndex].Shape.Height = cfg.Size

		if (cfg.Overflow != ScrollbarVisible && availWidth < 0.1) ||
			cfg.Overflow == ScrollbarOnHover {
			layout.Children[thumbIndex].Shape.Color = ColorTransparent
		}
	} else {
		layout.Shape.X = parent.Shape.X + parent.Shape.Width - cfg.Size
		layout.Shape.Y = parent.Shape.Y + parent.Shape.Padding.Top
		layout.Shape.Width = cfg.Size
		layout.Shape.Height = parent.Shape.Height - parent.Shape.Padding.Height()

		cHeight := contentHeight(parent)
		if cHeight == 0 {
			return
		}
		tHeight := layout.Shape.Height * (layout.Shape.Height / cHeight)
		thumbHeight := f32Clamp(tHeight, cfg.MinThumbSize, layout.Shape.Height)
		availHeight := layout.Shape.Height - thumbHeight

		sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
		scrollOffset := float32(0)
		if v, ok := sy.Get(cfg.IDScroll); ok {
			scrollOffset = -v
		}

		layout.Shape.X -= cfg.GapEdge
		layout.Shape.Y += cfg.GapEnd
		layout.Shape.Height -= cfg.GapEnd + cfg.GapEnd

		layout.Children[thumbIndex].Shape.X = layout.Shape.X
		offset := float32(0)
		if availHeight > 0 {
			offset = f32Clamp(
				(scrollOffset/(cHeight-layout.Shape.Height))*availHeight,
				0, availHeight)
		}
		layout.Children[thumbIndex].Shape.Y = layout.Shape.Y + offset
		layout.Children[thumbIndex].Shape.Height = thumbHeight - cfg.GapEnd - cfg.GapEnd
		layout.Children[thumbIndex].Shape.Width = cfg.Size

		if (cfg.Overflow != ScrollbarVisible && availHeight < 0.1) ||
			cfg.Overflow == ScrollbarOnHover {
			layout.Children[thumbIndex].Shape.Color = ColorTransparent
		}
	}
}

