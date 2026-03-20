package gui

import "testing"

func TestRenderCmdKindNameExhaustive(t *testing.T) {
	tests := []struct {
		kind RenderKind
		want string
	}{
		{RenderNone, "RenderNone"},
		{RenderClip, "RenderClip"},
		{RenderRect, "RenderRect"},
		{RenderStrokeRect, "RenderStrokeRect"},
		{RenderCircle, "RenderCircle"},
		{RenderImage, "RenderImage"},
		{RenderText, "RenderText"},
		{RenderLine, "RenderLine"},
		{RenderShadow, "RenderShadow"},
		{RenderBlur, "RenderBlur"},
		{RenderGradient, "RenderGradient"},
		{RenderGradientBorder, "RenderGradientBorder"},
		{RenderSvg, "RenderSvg"},
		{RenderLayout, "RenderLayout"},
		{RenderLayoutTransformed, "RenderLayoutTransformed"},
		{RenderLayoutPlaced, "RenderLayoutPlaced"},
		{RenderFilterBegin, "RenderFilterBegin"},
		{RenderFilterEnd, "RenderFilterEnd"},
		{RenderFilterComposite, "RenderFilterComposite"},
		{RenderCustomShader, "RenderCustomShader"},
		{RenderTextPath, "RenderTextPath"},
		{RenderRTF, "RenderRTF"},
		{RenderRotateBegin, "RenderRotateBegin"},
		{RenderRotateEnd, "RenderRotateEnd"},
		{RenderStencilBegin, "RenderStencilBegin"},
		{RenderStencilEnd, "RenderStencilEnd"},
		{RenderKind(255), "Unknown"},
	}
	for _, tt := range tests {
		got := renderCmdKindName(tt.kind)
		if got != tt.want {
			t.Errorf("renderCmdKindName(%d) = %q, want %q",
				tt.kind, got, tt.want)
		}
	}
}

func TestFindFilterBracketRangeFound(t *testing.T) {
	cmds := []RenderCmd{
		{Kind: RenderFilterBegin},
		{Kind: RenderRect},
		{Kind: RenderFilterEnd},
		{Kind: RenderRect},
	}
	r := findFilterBracketRange(cmds, 0)
	if !r.FoundEnd {
		t.Fatal("expected FoundEnd=true")
	}
	if r.StartIdx != 0 || r.EndIdx != 2 || r.NextIdx != 3 {
		t.Errorf("got start=%d end=%d next=%d, want 0/2/3",
			r.StartIdx, r.EndIdx, r.NextIdx)
	}
}

func TestFindFilterBracketRangeNotFound(t *testing.T) {
	cmds := []RenderCmd{
		{Kind: RenderFilterBegin},
		{Kind: RenderRect},
		{Kind: RenderRect},
	}
	r := findFilterBracketRange(cmds, 0)
	if r.FoundEnd {
		t.Fatal("expected FoundEnd=false")
	}
	if r.EndIdx != 3 || r.NextIdx != 3 {
		t.Errorf("got end=%d next=%d, want 3/3", r.EndIdx, r.NextIdx)
	}
}

func TestFindFilterBracketRangeMidSlice(t *testing.T) {
	cmds := []RenderCmd{
		{Kind: RenderRect},
		{Kind: RenderFilterBegin},
		{Kind: RenderRect},
		{Kind: RenderFilterEnd},
	}
	r := findFilterBracketRange(cmds, 1)
	if !r.FoundEnd || r.StartIdx != 1 || r.EndIdx != 3 || r.NextIdx != 4 {
		t.Errorf("got start=%d end=%d next=%d found=%v, want 1/3/4/true",
			r.StartIdx, r.EndIdx, r.NextIdx, r.FoundEnd)
	}
}

func TestFindFilterBracketRangeEmpty(t *testing.T) {
	r := findFilterBracketRange(nil, 0)
	if r.FoundEnd {
		t.Fatal("expected FoundEnd=false for empty slice")
	}
	if r.EndIdx != 0 || r.NextIdx != 0 {
		t.Errorf("got end=%d next=%d, want 0/0", r.EndIdx, r.NextIdx)
	}
}
