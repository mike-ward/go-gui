package svg

import "testing"

func TestResolveCalc_BasicAddPx(t *testing.T) {
	if got := resolveCalcRefs("calc(10px + 4px)"); got != "14px" {
		t.Errorf("calc(10px + 4px) = %q, want 14px", got)
	}
}

func TestResolveCalc_SubMulDiv(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"calc(20px - 5px)", "15px"},
		{"calc(2 * 8px)", "16px"},
		{"calc(8px * 2)", "16px"},
		{"calc(20px / 4)", "5px"},
		{"calc(10 + 2)", "12"},
		{"calc(2 + 3 * 4)", "14"},
		{"calc((2 + 3) * 4)", "20"},
	}
	for _, tc := range cases {
		if got := resolveCalcRefs(tc.in); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestResolveCalc_MixedUnitsRejected(t *testing.T) {
	// calc(10px + 50%) — % unit unsupported, tokenizer rejects.
	if got := resolveCalcRefs("calc(10px + 50%)"); got != "" {
		t.Errorf("mixed units should drop: got %q", got)
	}
	// Unit mismatch on add: cannot mix unitless + px.
	if got := resolveCalcRefs("calc(10 + 5px)"); got != "" {
		t.Errorf("unitless + px should drop: got %q", got)
	}
}

func TestResolveCalc_DivByZero(t *testing.T) {
	if got := resolveCalcRefs("calc(10px / 0)"); got != "" {
		t.Errorf("div by zero should drop: got %q", got)
	}
}

func TestResolveCalc_NestedAndEmbedded(t *testing.T) {
	// Embedded inside a longer value.
	got := resolveCalcRefs("rgba(calc(100 + 28), 0, 0, 1)")
	want := "rgba(128, 0, 0, 1)"
	if got != want {
		t.Errorf("embedded calc: got %q want %q", got, want)
	}
	// Nested calc.
	if got := resolveCalcRefs("calc(calc(2px + 3px) * 2)"); got != "10px" {
		t.Errorf("nested calc: got %q", got)
	}
}

func TestResolveCalc_NoChangeWithoutCalc(t *testing.T) {
	if got := resolveCalcRefs("red"); got != "red" {
		t.Errorf("non-calc value mangled: %q", got)
	}
}

func TestResolveCalc_InfRejected(t *testing.T) {
	// ParseFloat parses "1e10000" as +Inf — must invalidate the calc.
	if got := resolveCalcRefs("calc(1e10000 + 1px)"); got != "" {
		t.Errorf("Inf literal should drop calc: got %q", got)
	}
	// Overflow during multiplication.
	if got := resolveCalcRefs("calc(1e300 * 1e300)"); got != "" {
		t.Errorf("Inf result should drop calc: got %q", got)
	}
}

func TestResolveCalc_RejectsUnsupportedUnits(t *testing.T) {
	cases := []string{
		"calc(1em + 1em)",
		"calc(2rem)",
		"calc(50% + 5px)",
		"calc(10vh)",
	}
	for _, in := range cases {
		if got := resolveCalcRefs(in); got != "" {
			t.Errorf("%s should drop (unsupported unit): got %q", in, got)
		}
	}
}

func TestResolveCalc_MismatchedParens(t *testing.T) {
	// Unmatched OPEN paren inside calc: matchParen fails → entire
	// resolveCalcRefs returns "" (declaration drops).
	if got := resolveCalcRefs("calc((1 + 2)"); got != "" {
		t.Errorf("unmatched open paren should drop: got %q", got)
	}
	// Empty body — shunt produces empty RPN, evalCalcRPN sees empty
	// stack → ok=false.
	if got := resolveCalcRefs("calc()"); got != "" {
		t.Errorf("empty calc() should drop: got %q", got)
	}
	// Lone operator — eval stack underflow.
	if got := resolveCalcRefs("calc(+)"); got != "" {
		t.Errorf("lone operator should drop: got %q", got)
	}
}

func TestResolveCalc_LeadingDotLiteral(t *testing.T) {
	if got := resolveCalcRefs("calc(.5px + .5px)"); got != "1px" {
		t.Errorf(".5px + .5px = %q, want 1px", got)
	}
	if got := resolveCalcRefs("calc(.25 * 8px)"); got != "2px" {
		t.Errorf(".25 * 8px = %q, want 2px", got)
	}
}

func TestResolveCalc_SubtractionOverflow(t *testing.T) {
	// 1e300 - -1e300 → 2e300 (still finite). Push to ±Inf via doubling.
	if got := resolveCalcRefs("calc(1e300 - -1e300)"); got == "" {
		t.Errorf("legitimate finite subtraction should not drop: got empty")
	}
	// Force +Inf via mult overflow then check downstream rejects.
	if got := resolveCalcRefs("calc(1e300 * 1e300 - 1)"); got != "" {
		t.Errorf("Inf-propagated subtraction should drop: got %q", got)
	}
}

func TestResolveCalc_UnaryMinus(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"calc(-5px + 10px)", "5px"},
		{"calc(2 * -3px)", "-6px"},
		{"calc(-1 * 4px)", "-4px"},
		{"calc(10px - -2px)", "12px"},
		{"calc(+3px + 4px)", "7px"},
	}
	for _, tc := range cases {
		if got := resolveCalcRefs(tc.in); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.in, got, tc.want)
		}
	}
}
