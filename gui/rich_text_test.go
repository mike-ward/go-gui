package gui

import (
	"testing"

	"github.com/mike-ward/go-glyph"
)

func TestRichRunConstructors(t *testing.T) {
	s := TextStyle{Size: 14, Color: RGB(0, 0, 0)}

	r := RichRun("hello", s)
	if r.Text != "hello" || r.Style.Size != 14 {
		t.Fatalf("RichRun: got %q size=%v", r.Text, r.Style.Size)
	}

	br := RichBr()
	if br.Text != "\n" {
		t.Fatalf("RichBr: got %q", br.Text)
	}
}

func TestRichLinkSetsUnderline(t *testing.T) {
	s := TextStyle{Size: 14}
	r := RichLink("click", "https://example.com", s)
	if !r.Style.Underline {
		t.Fatal("RichLink should set Underline")
	}
	if r.Link != "https://example.com" {
		t.Fatalf("link: got %q", r.Link)
	}
}

func TestRichAbbrSetsBoldTypeface(t *testing.T) {
	s := TextStyle{Size: 14}
	r := RichAbbr("HTML", "HyperText Markup Language", s)
	if r.Style.Typeface != glyph.TypefaceBold {
		t.Fatal("RichAbbr should set TypefaceBold")
	}
	if r.Tooltip != "HyperText Markup Language" {
		t.Fatalf("tooltip: got %q", r.Tooltip)
	}
}

func TestRichFootnoteReducesSize(t *testing.T) {
	s := TextStyle{Size: 20}
	r := RichFootnote("1", "footnote text", s)
	if r.Style.Size >= 20 {
		t.Fatalf("size should be reduced: got %v", r.Style.Size)
	}
	if r.Tooltip != "footnote text" {
		t.Fatalf("tooltip: got %q", r.Tooltip)
	}
}

func TestRichTextToGlyphConversion(t *testing.T) {
	rt := RichText{
		Runs: []RichTextRun{
			{Text: "hello ", Style: TextStyle{Size: 14}},
			{Text: "world", Style: TextStyle{Size: 16}},
		},
	}
	grt := rt.toGlyphRichText()
	if len(grt.Runs) != 2 {
		t.Fatalf("expected 2 glyph runs, got %d", len(grt.Runs))
	}
	if grt.Runs[0].Text != "hello " {
		t.Fatalf("run 0 text: got %q", grt.Runs[0].Text)
	}
	if grt.Runs[1].Text != "world" {
		t.Fatalf("run 1 text: got %q", grt.Runs[1].Text)
	}
}

func TestRichTextPlain(t *testing.T) {
	rt := RichText{
		Runs: []RichTextRun{
			{Text: "hello "},
			{Text: "world"},
		},
	}
	got := richTextPlain(rt)
	if got != "hello world" {
		t.Fatalf("plain text: got %q", got)
	}
}
