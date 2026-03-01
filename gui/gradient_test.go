package gui

import "testing"

func TestGradientDefDefaults(t *testing.T) {
	g := GradientDef{}
	if g.Type != GradientLinear {
		t.Error("default type should be linear")
	}
	if g.Direction != GradientToTop {
		t.Error("default direction should be top")
	}
}

func TestGradientStops(t *testing.T) {
	g := GradientDef{
		Stops: []GradientStop{
			{Color: Red, Pos: 0},
			{Color: Blue, Pos: 1},
		},
		Direction: GradientToRight,
	}
	if len(g.Stops) != 2 {
		t.Error("expected 2 stops")
	}
	if g.Stops[0].Pos != 0 || g.Stops[1].Pos != 1 {
		t.Error("stop positions wrong")
	}
}
