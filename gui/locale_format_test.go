package gui

import (
	"testing"
	"time"
)

func TestLocaleFormatDateShort(t *testing.T) {
	old := guiLocale
	defer func() { guiLocale = old }()
	guiLocale = LocaleEnUS

	dt := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	got := LocaleFormatDate(dt, "M/D/YYYY")
	if got != "3/15/2025" {
		t.Fatalf("got %q, want 3/15/2025", got)
	}
}

func TestLocaleFormatDateLongMonth(t *testing.T) {
	old := guiLocale
	defer func() { guiLocale = old }()
	guiLocale = LocaleEnUS

	dt := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	got := LocaleFormatDate(dt, "MMMM D, YYYY")
	if got != "March 15, 2025" {
		t.Fatalf("got %q, want March 15, 2025", got)
	}
}

func TestLocaleFormatDateShortMonth(t *testing.T) {
	old := guiLocale
	defer func() { guiLocale = old }()
	guiLocale = LocaleEnUS

	dt := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	got := LocaleFormatDate(dt, "MMM YYYY")
	if got != "Dec 2025" {
		t.Fatalf("got %q, want Dec 2025", got)
	}
}

func TestLocaleFormatDateGerman(t *testing.T) {
	old := guiLocale
	defer func() { guiLocale = old }()
	guiLocale = LocaleDeDE

	dt := time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC)
	got := LocaleFormatDate(dt, "D. MMMM YYYY")
	want := "5. M\u00E4rz 2025"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLocaleFormatDateTime(t *testing.T) {
	old := guiLocale
	defer func() { guiLocale = old }()
	guiLocale = LocaleEnUS

	dt := time.Date(2025, 1, 2, 14, 5, 9, 0, time.UTC)
	got := LocaleFormatDate(dt, "YYYY-MM-DD HH:mm:ss")
	if got != "2025-01-02 14:05:09" {
		t.Fatalf("got %q, want 2025-01-02 14:05:09", got)
	}
}

func TestLocaleRowsFmt(t *testing.T) {
	old := guiLocale
	defer func() { guiLocale = old }()
	guiLocale = LocaleEnUS

	got := localeRowsFmt(1, 50, 200)
	if got != "Rows 1-50/200" {
		t.Fatalf("got %q", got)
	}
}

func TestLocalePageFmt(t *testing.T) {
	old := guiLocale
	defer func() { guiLocale = old }()
	guiLocale = LocaleEnUS

	got := localePageFmt(3, 10)
	if got != "Page 3/10" {
		t.Fatalf("got %q", got)
	}
}

func TestLocaleMatchesFmt(t *testing.T) {
	old := guiLocale
	defer func() { guiLocale = old }()
	guiLocale = LocaleEnUS

	got := localeMatchesFmt(5, "100")
	if got != "Matches 5/100" {
		t.Fatalf("got %q", got)
	}
}

func TestLocaleT(t *testing.T) {
	old := guiLocale
	defer func() { guiLocale = old }()

	guiLocale = Locale{
		Translations: map[string]string{
			"greeting": "hello",
		},
	}
	if got := LocaleT("greeting"); got != "hello" {
		t.Fatalf("got %q, want hello", got)
	}
	if got := LocaleT("missing"); got != "missing" {
		t.Fatalf("got %q, want missing", got)
	}
}
