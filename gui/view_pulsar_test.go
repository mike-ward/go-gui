package gui

import "testing"

func TestPulsarShowsText1WhenCursorOn(t *testing.T) {
	w := &Window{}
	w.viewState.inputCursorOn = true
	v := Pulsar(PulsarCfg{}, w)
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) != 1 {
		t.Fatalf("children = %d, want 1", len(layout.Children))
	}
	tc := layout.Children[0].Shape.TC
	if tc == nil || tc.Text != "..." {
		t.Errorf("text = %v, want ...", tc)
	}
}

func TestPulsarShowsText2WhenCursorOff(t *testing.T) {
	w := &Window{}
	w.viewState.inputCursorOn = false
	v := Pulsar(PulsarCfg{}, w)
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) != 1 {
		t.Fatalf("children = %d, want 1", len(layout.Children))
	}
	tc := layout.Children[0].Shape.TC
	if tc == nil || tc.Text != ".." {
		t.Errorf("text = %v, want ..", tc)
	}
}

func TestPulsarCustomText(t *testing.T) {
	w := &Window{}
	w.viewState.inputCursorOn = true
	v := Pulsar(PulsarCfg{Text1: "ON", Text2: "OFF"}, w)
	layout := GenerateViewLayout(v, w)
	tc := layout.Children[0].Shape.TC
	if tc == nil || tc.Text != "ON" {
		t.Errorf("text = %v, want ON", tc)
	}
}
