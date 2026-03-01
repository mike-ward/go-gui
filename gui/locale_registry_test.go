package gui

import "testing"

func TestLocaleRegistryInit(t *testing.T) {
	names := LocaleRegisteredNames()
	want := map[string]bool{
		"en-US": true, "de-DE": true, "ar-SA": true,
	}
	for id := range want {
		found := false
		for _, n := range names {
			if n == id {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing registered locale: %s", id)
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
