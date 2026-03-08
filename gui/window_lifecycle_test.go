package gui

import (
	"testing"

	"github.com/mike-ward/go-glyph"
)

func TestNewWindowSetsFields(t *testing.T) {
	type S struct{ X int }
	st := &S{X: 42}
	w := NewWindow(WindowCfg{
		State:  st,
		Title:  "test",
		Width:  800,
		Height: 600,
	})
	if w.windowWidth != 800 {
		t.Errorf("width = %d, want 800", w.windowWidth)
	}
	if w.windowHeight != 600 {
		t.Errorf("height = %d, want 600", w.windowHeight)
	}
	if !w.focused {
		t.Error("want focused=true")
	}
	if !w.refreshLayout {
		t.Error("want refreshLayout=true")
	}
	if State[S](w).X != 42 {
		t.Errorf("state.X = %d, want 42", State[S](w).X)
	}
	if w.Config.Title != "test" {
		t.Errorf("Config.Title = %q, want test", w.Config.Title)
	}
}

func TestUpdateViewSetsGenerator(t *testing.T) {
	w := NewWindow(WindowCfg{Width: 100, Height: 100})
	called := false
	w.UpdateView(func(_ *Window) View {
		called = true
		return Text(TextCfg{Text: "hi"})
	})
	if w.viewGenerator == nil {
		t.Fatal("viewGenerator nil after UpdateView")
	}
	if !w.refreshLayout {
		t.Error("want refreshLayout=true after UpdateView")
	}
	// Call generator to verify it works.
	w.viewGenerator(w)
	if !called {
		t.Error("generator not called")
	}
}

func TestFrameFnCallsUpdate(t *testing.T) {
	w := NewWindow(WindowCfg{
		State:  new(int),
		Width:  100,
		Height: 100,
	})
	updated := false
	w.viewGenerator = func(_ *Window) View {
		updated = true
		return Text(TextCfg{Text: "x"})
	}
	w.refreshLayout = true
	w.FrameFn()
	if !updated {
		t.Error("FrameFn did not call Update")
	}
	if w.refreshLayout {
		t.Error("refreshLayout should be cleared")
	}
}

func TestFrameFnNoopWhenNoRefresh(t *testing.T) {
	w := NewWindow(WindowCfg{
		State:  new(int),
		Width:  100,
		Height: 100,
	})
	called := false
	w.viewGenerator = func(_ *Window) View {
		called = true
		return Text(TextCfg{Text: "x"})
	}
	w.refreshLayout = false
	w.refreshRenderOnly = false
	w.FrameFn()
	if called {
		t.Error("FrameFn should not call generator when no refresh")
	}
}

func TestRenderTextEmitsCommand(t *testing.T) {
	w := NewWindow(WindowCfg{
		State:  new(int),
		Width:  200,
		Height: 200,
	})
	w.viewGenerator = func(_ *Window) View {
		return Text(TextCfg{
			Text: "hello",
			TextStyle: TextStyle{
				Color: RGB(255, 255, 255),
				Size:  16,
			},
		})
	}
	w.refreshLayout = true
	w.FrameFn()

	found := false
	for _, r := range w.renderers {
		if r.Kind == RenderText && r.Text == "hello" {
			found = true
			if r.FontSize != 16 {
				t.Errorf("FontSize = %f, want 16", r.FontSize)
			}
			break
		}
	}
	if !found {
		t.Error("no RenderText command with text 'hello'")
	}
}

func TestTextFallbackMeasurement(t *testing.T) {
	w := NewWindow(WindowCfg{
		State:  new(int),
		Width:  200,
		Height: 200,
	})
	// No textMeasurer set — should use placeholder.
	tv := Text(TextCfg{
		Text: "test",
		TextStyle: TextStyle{
			Size: 20,
		},
	})
	layout := tv.GenerateLayout(w)
	wantW := float32(len("test")) * 20 * 0.6
	if layout.Shape.Width != wantW {
		t.Errorf("width = %f, want %f", layout.Shape.Width, wantW)
	}
	wantH := float32(20 * 1.4)
	if layout.Shape.Height != wantH {
		t.Errorf("height = %f, want %f", layout.Shape.Height, wantH)
	}
}

type mockTextMeasurer struct{}

func (m *mockTextMeasurer) TextWidth(text string, _ TextStyle) float32 {
	return float32(len(text)) * 10
}
func (m *mockTextMeasurer) TextHeight(_ string, _ TextStyle) float32 {
	return 20
}
func (m *mockTextMeasurer) FontAscent(s TextStyle) float32 { return s.Size * 0.8 }
func (m *mockTextMeasurer) FontHeight(_ TextStyle) float32 {
	return 22
}
func (m *mockTextMeasurer) LayoutText(_ string, _ TextStyle, _ float32) (glyph.Layout, error) {
	return glyph.Layout{Height: 22}, nil
}

func TestTextWithMeasurer(t *testing.T) {
	w := NewWindow(WindowCfg{
		State:  new(int),
		Width:  200,
		Height: 200,
	})
	w.SetTextMeasurer(&mockTextMeasurer{})

	tv := Text(TextCfg{
		Text: "abc",
		TextStyle: TextStyle{
			Size: 16,
		},
	})
	layout := tv.GenerateLayout(w)
	if layout.Shape.Width != 30 {
		t.Errorf("width = %f, want 30", layout.Shape.Width)
	}
	if layout.Shape.Height != 22 {
		t.Errorf("height = %f, want 22", layout.Shape.Height)
	}
}

func TestRenderersAccessor(t *testing.T) {
	w := NewWindow(WindowCfg{Width: 50, Height: 50})
	w.renderers = append(w.renderers, RenderCmd{Kind: RenderRect})
	r := w.Renderers()
	if len(r) != 1 || r[0].Kind != RenderRect {
		t.Error("Renderers() mismatch")
	}
}

func TestMouseCursorStateAccessor(t *testing.T) {
	w := NewWindow(WindowCfg{Width: 50, Height: 50})
	w.SetMouseCursor(CursorIBeam)
	if w.MouseCursorState() != CursorIBeam {
		t.Errorf("got %d, want CursorIBeam", w.MouseCursorState())
	}
}

func TestPasswordMaskInRenderText(t *testing.T) {
	w := NewWindow(WindowCfg{
		State:  new(int),
		Width:  200,
		Height: 200,
	})
	shape := &Shape{
		ShapeType: ShapeText,
		Width:     100,
		Height:    20,
		Opacity:   1.0,
		TC: &ShapeTextConfig{
			Text:           "secret",
			TextIsPassword: true,
			TextStyle:      &TextStyle{Color: RGB(255, 255, 255), Size: 16},
		},
	}
	clip := DrawClip{X: 0, Y: 0, Width: 200, Height: 200}
	renderText(shape, clip, w)

	found := false
	for _, r := range w.renderers {
		if r.Kind == RenderText {
			found = true
			for _, ch := range r.Text {
				if ch != '•' {
					t.Errorf("expected password char, got %c", ch)
				}
			}
		}
	}
	if !found {
		t.Error("no RenderText command emitted")
	}
}

func TestRenderTextWrapSetsWidth(t *testing.T) {
	w := NewWindow(WindowCfg{
		State:  new(int),
		Width:  200,
		Height: 200,
	})
	shape := &Shape{
		ShapeType: ShapeText,
		Width:     250,
		Height:    20,
		Opacity:   1.0,
		TC: &ShapeTextConfig{
			Text:     "wrap me",
			TextMode: TextModeWrap,
			TextStyle: &TextStyle{
				Color: RGB(255, 255, 255), Size: 16,
			},
		},
	}
	clip := DrawClip{X: 0, Y: 0, Width: 400, Height: 400}
	renderText(shape, clip, w)

	found := false
	for _, r := range w.renderers {
		if r.Kind == RenderText && r.Text == "wrap me" {
			found = true
			if r.W != 250 {
				t.Errorf("W = %f, want 250", r.W)
			}
		}
	}
	if !found {
		t.Error("no RenderText command emitted")
	}
}

func TestRenderTextNoWrapOmitsWidth(t *testing.T) {
	w := NewWindow(WindowCfg{
		State:  new(int),
		Width:  200,
		Height: 200,
	})
	shape := &Shape{
		ShapeType: ShapeText,
		Width:     250,
		Height:    20,
		Opacity:   1.0,
		TC: &ShapeTextConfig{
			Text: "no wrap",
			TextStyle: &TextStyle{
				Color: RGB(255, 255, 255), Size: 16,
			},
		},
	}
	clip := DrawClip{X: 0, Y: 0, Width: 400, Height: 400}
	renderText(shape, clip, w)

	for _, r := range w.renderers {
		if r.Kind == RenderText && r.Text == "no wrap" {
			if r.W != 0 {
				t.Errorf("W = %f, want 0 for non-wrap text",
					r.W)
			}
		}
	}
}

func TestSetClipboard(t *testing.T) {
	w := &Window{}
	var got string
	w.SetClipboardFn(func(s string) { got = s })
	w.SetClipboard("hello")
	if got != "hello" {
		t.Errorf("clipboard = %q, want hello", got)
	}
}

func TestSetClipboardNilSafe(t *testing.T) {
	_ = t
	w := &Window{}
	// Should not panic when no fn set.
	w.SetClipboard("ignored")
}
