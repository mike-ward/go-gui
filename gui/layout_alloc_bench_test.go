package gui

import "testing"

func benchmarkArrangeLayout() Layout {
	root := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			Axis:      AxisTopToBottom,
			Sizing:    FillFill,
			Width:     1200,
			Height:    900,
			Spacing:   8,
		},
	}
	root.Children = make([]Layout, 0, 120)
	for i := 0; i < 120; i++ {
		ch := Layout{
			Shape: &Shape{
				ShapeType: ShapeRectangle,
				Axis:      AxisLeftToRight,
				Sizing:    FixedFit,
				Width:     200,
				Height:    40,
				Spacing:   4,
			},
		}
		if i%15 == 0 {
			ch.Shape.Float = true
		}
		root.Children = append(root.Children, ch)
	}
	return root
}

func BenchmarkLayoutArrange(b *testing.B) {
	w := &Window{scratch: newScratchPools()}
	w.windowWidth = 1200
	w.windowHeight = 900

	template := benchmarkArrangeLayout()
	layout := Layout{
		Shape:    &Shape{},
		Children: make([]Layout, len(template.Children)),
	}
	childShapes := make([]Shape, len(template.Children))

	b.ReportAllocs()
	for b.Loop() {
		*layout.Shape = *template.Shape
		layout.Parent = nil
		for j := range template.Children {
			childShapes[j] = *template.Children[j].Shape
			layout.Children[j] = Layout{
				Shape: &childShapes[j],
			}
		}
		layers := layoutArrange(&layout, w)
		w.scratch.layerLayouts.put(layers)
	}
}

func benchmarkWrapLayout() Layout {
	root := Layout{
		Shape: &Shape{
			ShapeType: ShapeRectangle,
			Axis:      AxisLeftToRight,
			Wrap:      true,
			Width:     600,
			Spacing:   6,
		},
		Children: make([]Layout, 0, 200),
	}
	for i := 0; i < 200; i++ {
		root.Children = append(root.Children, Layout{
			Shape: &Shape{
				ShapeType: ShapeRectangle,
				Width:     70,
				Height:    20,
			},
		})
	}
	return root
}

func BenchmarkLayoutWrapContainers(b *testing.B) {
	w := &Window{scratch: newScratchPools()}
	b.ReportAllocs()
	for b.Loop() {
		layout := benchmarkWrapLayout()
		layoutWrapContainers(&layout, w)
	}
}

func benchmarkFocusLayout() *Layout {
	root := &Layout{
		Shape: &Shape{ShapeType: ShapeRectangle},
	}
	root.Children = make([]Layout, 0, 200)
	for i := uint32(1); i <= 200; i++ {
		root.Children = append(root.Children, Layout{
			Shape: &Shape{
				ShapeType: ShapeRectangle,
				IDFocus:   i,
			},
		})
	}
	return root
}

func BenchmarkFocusTraversal(b *testing.B) {
	w := &Window{scratch: newScratchPools()}
	root := benchmarkFocusLayout()
	b.ReportAllocs()
	for b.Loop() {
		if s, ok := root.NextFocusable(w); ok {
			w.SetIDFocus(s.IDFocus)
		}
	}
}
