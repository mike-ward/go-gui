package gui

import "testing"

func TestInputMaskPresets(t *testing.T) {
	assertEqual(t, len(InputMaskFromPreset(MaskNone)), 0)
	if InputMaskFromPreset(MaskPhoneUS) != "(999) 999-9999" {
		t.Fatal("phone_us preset mismatch")
	}
	if InputMaskFromPreset(MaskCreditCard16) != "9999 9999 9999 9999" {
		t.Fatal("credit_card_16 preset mismatch")
	}
	if InputMaskFromPreset(MaskCreditCardAmex) != "9999 999999 99999" {
		t.Fatal("credit_card_amex preset mismatch")
	}
	if InputMaskFromPreset(MaskExpiryMMYY) != "99/99" {
		t.Fatal("expiry_mm_yy preset mismatch")
	}
	if InputMaskFromPreset(MaskCVC) != "999" {
		t.Fatal("cvc preset mismatch")
	}
}

func TestInputMaskSanitizePastePhone(t *testing.T) {
	compiled, err := CompileInputMask(InputMaskFromPreset(MaskPhoneUS), nil)
	if err != nil {
		t.Fatal(err)
	}
	res := InputMaskInsert("", 0, 0, 0, "abc555-123-4567xyz", &compiled)
	if !res.Changed {
		t.Fatal("expected changed")
	}
	if res.Text != "(555) 123-4567" {
		t.Fatalf("got %q, want %q", res.Text, "(555) 123-4567")
	}
	if res.CursorPos != len([]rune(res.Text)) {
		t.Fatalf("cursor %d, want %d", res.CursorPos, len([]rune(res.Text)))
	}
}

func TestInputMaskRejectInvalidChar(t *testing.T) {
	compiled, err := CompileInputMask("99", nil)
	if err != nil {
		t.Fatal(err)
	}
	res := InputMaskInsert("", 0, 0, 0, "a", &compiled)
	if res.Changed {
		t.Fatal("expected no change")
	}
	if res.Text != "" {
		t.Fatalf("got %q, want empty", res.Text)
	}
	assertEqual(t, res.CursorPos, 0)
}

func TestInputMaskDeleteSkipsLiterals(t *testing.T) {
	compiled, err := CompileInputMask(InputMaskFromPreset(MaskPhoneUS), nil)
	if err != nil {
		t.Fatal(err)
	}
	text := ""
	cursor := 0
	for _, ch := range "5551234" {
		res := InputMaskInsert(text, cursor, 0, 0, string(ch), &compiled)
		text = res.Text
		cursor = res.CursorPos
	}
	if text != "(555) 123-4" {
		t.Fatalf("got %q, want %q", text, "(555) 123-4")
	}
	del := InputMaskDelete(text, 4, 0, 0, &compiled)
	if !del.Changed {
		t.Fatal("expected changed")
	}
	if del.Text != "(555) 234" {
		t.Fatalf("got %q, want %q", del.Text, "(555) 234")
	}
}

func TestInputMaskBackspaceRemovesEditableSlot(t *testing.T) {
	compiled, err := CompileInputMask(InputMaskFromPreset(MaskPhoneUS), nil)
	if err != nil {
		t.Fatal(err)
	}
	start := InputMaskInsert("", 0, 0, 0, "5551234", &compiled)
	if start.Text != "(555) 123-4" {
		t.Fatalf("got %q, want %q", start.Text, "(555) 123-4")
	}
	back := InputMaskBackspace(start.Text, start.CursorPos, 0, 0, &compiled)
	if !back.Changed {
		t.Fatal("expected changed")
	}
	if back.Text != "(555) 123" {
		t.Fatalf("got %q, want %q", back.Text, "(555) 123")
	}
}

func TestInputMaskCustomTokenTransform(t *testing.T) {
	isUpperLetter := func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
	}
	toUpper := func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 32
		}
		return r
	}
	custom := []MaskTokenDef{
		{Symbol: 'A', Matcher: isUpperLetter, Transform: toUpper},
	}
	compiled, err := CompileInputMask("AA-99", custom)
	if err != nil {
		t.Fatal(err)
	}
	res := InputMaskInsert("", 0, 0, 0, "ab12", &compiled)
	if !res.Changed {
		t.Fatal("expected changed")
	}
	if res.Text != "AB-12" {
		t.Fatalf("got %q, want %q", res.Text, "AB-12")
	}
}
