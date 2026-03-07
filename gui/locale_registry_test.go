package gui

import "testing"

func TestLocaleRegistryInit(t *testing.T) {
	names := LocaleRegisteredNames()
	want := []string{
		"ar-SA", "de-DE", "en-US", "es-ES", "fr-FR",
		"he-IL", "ja-JP", "ko-KR", "pt-BR", "zh-CN",
	}
	for _, id := range want {
		found := false
		for _, n := range names {
			if n == id {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing registered locale: %s (have %v)",
				id, names)
		}
	}
}

func TestLocaleGetKnown(t *testing.T) {
	l, ok := LocaleGet("de-DE")
	if !ok {
		t.Fatal("LocaleGet(de-DE) returned false")
	}
	if l.ID != "de-DE" {
		t.Fatalf("ID = %q", l.ID)
	}
}

func TestLocaleGetUnknown(t *testing.T) {
	_, ok := LocaleGet("xx-XX")
	if ok {
		t.Fatal("LocaleGet(xx-XX) should return false")
	}
}

func TestLocaleRegisterOverwrite(t *testing.T) {
	custom := localeDefaults()
	custom.ID = "test-overwrite"
	custom.StrOK = "first"
	LocaleRegister(custom)

	custom.StrOK = "second"
	LocaleRegister(custom)

	l, ok := LocaleGet("test-overwrite")
	if !ok {
		t.Fatal("not found")
	}
	if l.StrOK != "second" {
		t.Fatalf("StrOK = %q, want second", l.StrOK)
	}
}

func TestLocaleRegisteredNamesSorted(t *testing.T) {
	names := LocaleRegisteredNames()
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Fatalf("not sorted: %v", names)
		}
	}
}

func TestLocalePresetFields(t *testing.T) {
	tests := []struct {
		id      string
		dir     TextDirection
		decSep  rune
		curCode string
	}{
		{"en-US", TextDirLTR, '.', "USD"},
		{"de-DE", TextDirAuto, ',', "EUR"},
		{"ar-SA", TextDirRTL, '.', "SAR"},
		{"fr-FR", TextDirAuto, ',', "EUR"},
		{"es-ES", TextDirAuto, ',', "EUR"},
		{"pt-BR", TextDirAuto, ',', "BRL"},
		{"ja-JP", TextDirAuto, '.', "JPY"},
		{"zh-CN", TextDirAuto, '.', "CNY"},
		{"ko-KR", TextDirAuto, '.', "KRW"},
		{"he-IL", TextDirRTL, '.', "ILS"},
	}
	for _, tt := range tests {
		l, ok := LocaleGet(tt.id)
		if !ok {
			t.Errorf("%s: not registered", tt.id)
			continue
		}
		if l.TextDir != tt.dir {
			t.Errorf("%s: TextDir = %v, want %v",
				tt.id, l.TextDir, tt.dir)
		}
		if l.Number.DecimalSep != tt.decSep {
			t.Errorf("%s: DecimalSep = %c, want %c",
				tt.id, l.Number.DecimalSep, tt.decSep)
		}
		if l.Currency.Code != tt.curCode {
			t.Errorf("%s: Currency.Code = %s, want %s",
				tt.id, l.Currency.Code, tt.curCode)
		}
		if l.StrOK == "" {
			t.Errorf("%s: StrOK empty", tt.id)
		}
		if l.StrCancel == "" {
			t.Errorf("%s: StrCancel empty", tt.id)
		}
		if l.WeekdaysFull[0] == "" {
			t.Errorf("%s: WeekdaysFull[0] empty", tt.id)
		}
		if l.MonthsFull[0] == "" {
			t.Errorf("%s: MonthsFull[0] empty", tt.id)
		}
	}
}
