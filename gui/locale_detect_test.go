package gui

import "testing"

func TestNormalizeLocaleEnv(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"en_US.UTF-8", "en-US"},
		{"de_DE.utf8", "de-DE"},
		{"fr_FR", "fr-FR"},
		{"C", "en-US"},
		{"POSIX", "en-US"},
		{"", "en-US"},
		{"ja_JP.eucJP", "ja-JP"},
		{"pt-BR", "pt-BR"},
	}
	for _, tt := range tests {
		got := normalizeLocaleEnv(tt.in)
		if got != tt.want {
			t.Errorf("normalizeLocaleEnv(%q) = %q, want %q",
				tt.in, got, tt.want)
		}
	}
}

func TestLocaleAutoDetect(t *testing.T) {
	saved := guiLocale
	defer func() { guiLocale = saved }()

	// Exact match: de-DE is registered.
	guiLocale = localeDefaults()
	if l, ok := LocaleGet("de-DE"); ok {
		SetLocale(l)
		if guiLocale.ID != "de-DE" {
			t.Errorf("SetLocale de-DE: got %s", guiLocale.ID)
		}
	} else {
		t.Fatal("de-DE not registered")
	}

	// Reset and verify LocaleAutoDetect doesn't panic.
	guiLocale = localeDefaults()
	LocaleAutoDetect()
}
