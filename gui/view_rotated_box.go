package gui

// RotatedBoxCfg configures a RotatedBox view.
type RotatedBoxCfg struct {
	QuarterTurns int  // 1=90° CW, 2=180°, 3=270° CW
	Content      View // single child
}

// RotatedBox rotates its child content by quarter turns.
// Returns the child directly when turns == 0 (no rotation).
func RotatedBox(cfg RotatedBoxCfg) View {
	turns := ((cfg.QuarterTurns % 4) + 4) % 4
	if turns == 0 || cfg.Content == nil {
		if cfg.Content != nil {
			return cfg.Content
		}
		return &rotatedBoxView{turns: 0}
	}
	return &rotatedBoxView{
		turns:   uint8(turns),
		content: cfg.Content,
	}
}

type rotatedBoxView struct {
	turns   uint8
	content View
}

func (v *rotatedBoxView) Content() []View {
	if v.content == nil {
		return nil
	}
	return []View{v.content}
}

func (v *rotatedBoxView) GenerateLayout(_ *Window) Layout {
	return Layout{
		Shape: &Shape{
			ShapeType:    ShapeRectangle,
			Axis:         AxisTopToBottom,
			QuarterTurns: v.turns,
			Clip:         true,
			Sizing:       FitFit,
			Opacity:      1.0,
		},
	}
}
