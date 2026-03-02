package gui

import "testing"

func TestIconLookupContainsAllConstants(t *testing.T) {
	if len(IconLookup) != 255 {
		t.Errorf("IconLookup: got %d entries, want 255", len(IconLookup))
	}
}

func TestIconLookupKnownKeys(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"icon_arrow_down", IconArrowDown},
		{"icon_check", IconCheck},
		{"icon_home", IconHome},
		{"icon_star", IconStar},
		{"icon_yaki_dango", IconYakiDango},
	}
	for _, tt := range tests {
		if got, ok := IconLookup[tt.key]; !ok {
			t.Errorf("IconLookup missing key %q", tt.key)
		} else if got != tt.want {
			t.Errorf("IconLookup[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestIconLookupMissingKey(t *testing.T) {
	if _, ok := IconLookup["nonexistent"]; ok {
		t.Error("expected missing key to return false")
	}
}

func TestFontVariantsStruct(t *testing.T) {
	fv := FontVariants{Normal: "a.ttf", Bold: "b.ttf", Italic: "i.ttf", Mono: "m.ttf"}
	if fv.Normal != "a.ttf" || fv.Mono != "m.ttf" {
		t.Error("FontVariants fields not set correctly")
	}
}

func TestIconFontName(t *testing.T) {
	if IconFontName != "feathericon" {
		t.Errorf("IconFontName = %q, want %q", IconFontName, "feathericon")
	}
}
