package gui

import (
	"math"
	"testing"
)

func assertF64Near(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) >= 0.000001 {
		t.Errorf("got %f, want %f", got, want)
	}
}

func TestNumericParseENLocale(t *testing.T) {
	loc := numericLocaleNormalize(NumericLocaleCfg{})
	v, ok := numericParse("1,234.50", loc)
	if !ok {
		t.Fatal("parse failed")
	}
	assertF64Near(t, v, 1234.5)
}

func TestNumericParseDELocale(t *testing.T) {
	loc := numericLocaleNormalize(NumericLocaleCfg{DecimalSep: ',', GroupSep: '.'})
	v, ok := numericParse("1.234,50", loc)
	if !ok {
		t.Fatal("parse failed")
	}
	assertF64Near(t, v, 1234.5)
}

func TestNumericParseInvalidString(t *testing.T) {
	loc := numericLocaleNormalize(NumericLocaleCfg{})
	if _, ok := numericParse("1,,234", loc); ok {
		t.Fatal("expected failure")
	}
	if _, ok := numericParse("abc", loc); ok {
		t.Fatal("expected failure")
	}
}

func TestNumericParseInvalidGrouping(t *testing.T) {
	loc := numericLocaleNormalize(NumericLocaleCfg{GroupSizes: []int{3, 2}})
	if _, ok := numericParse("12,345,67", loc); ok {
		t.Fatal("expected failure")
	}
}

func TestNumericFormatENLocale(t *testing.T) {
	got := numericFormat(1234.5, 2, NumericLocaleCfg{})
	if got != "1,234.50" {
		t.Fatalf("got %q, want %q", got, "1,234.50")
	}
}

func TestNumericFormatDELocale(t *testing.T) {
	locale := NumericLocaleCfg{DecimalSep: ',', GroupSep: '.'}
	got := numericFormat(1234.5, 2, locale)
	if got != "1.234,50" {
		t.Fatalf("got %q, want %q", got, "1.234,50")
	}
}

func TestNumericFormatGroupSizes(t *testing.T) {
	locale := NumericLocaleCfg{GroupSep: ',', GroupSizes: []int{3, 2}}
	got := numericFormat(1234567, 0, locale)
	if got != "12,34,567" {
		t.Fatalf("got %q, want %q", got, "12,34,567")
	}
}

func TestNumericClampUnboundedAllowsLargeValues(t *testing.T) {
	v := 1.0e308
	assertF64Near(t, numericClamp(v, Opt[float64]{}, Opt[float64]{}), v)
}

func TestNumericCommitResultClamps(t *testing.T) {
	val, text := numericInputCommitResult(
		"1,250.30", Opt[float64]{}, Some(0.0), Some(1000.0), 2, NumericLocaleCfg{})
	if text != "1,000.00" {
		t.Fatalf("got %q, want %q", text, "1,000.00")
	}
	if !val.IsSet() {
		t.Fatal("expected set value")
	}
	assertF64Near(t, val.Get(0), 1000.0)
}

func TestNumericCommitResultInvalidFallbackToValue(t *testing.T) {
	val, text := numericInputCommitResult(
		"abc", Some(12.5), Opt[float64]{}, Opt[float64]{}, 1, NumericLocaleCfg{})
	if text != "12.5" {
		t.Fatalf("got %q, want %q", text, "12.5")
	}
	if !val.IsSet() {
		t.Fatal("expected set value")
	}
	assertF64Near(t, val.Get(0), 12.5)
}

func TestNumericCommitResultInvalidWithoutValue(t *testing.T) {
	val, text := numericInputCommitResult(
		"abc", Opt[float64]{}, Opt[float64]{}, Opt[float64]{}, 2, NumericLocaleCfg{})
	if val.IsSet() {
		t.Fatal("expected unset value")
	}
	if text != "" {
		t.Fatalf("got %q, want empty", text)
	}
}

func TestNumericStepResultUsesMinSeed(t *testing.T) {
	val, text := numericInputStepResult(
		"", Opt[float64]{}, Some(10.0), Opt[float64]{}, 2,
		NumericStepCfg{}, NumericLocaleCfg{}, 1.0, ModNone)
	if text != "11.00" {
		t.Fatalf("got %q, want %q", text, "11.00")
	}
	if !val.IsSet() {
		t.Fatal("expected set value")
	}
	assertF64Near(t, val.Get(0), 11.0)
}

func TestNumericStepResultModifiers(t *testing.T) {
	cfg := NumericStepCfg{
		Step:            1.0,
		ShiftMultiplier: 10.0,
		AltMultiplier:   0.1,
	}
	valShift, textShift := numericInputStepResult(
		"5", Opt[float64]{}, Opt[float64]{}, Opt[float64]{}, 2, cfg, NumericLocaleCfg{}, 1.0, ModShift)
	if textShift != "15.00" {
		t.Fatalf("shift: got %q, want %q", textShift, "15.00")
	}
	if !valShift.IsSet() {
		t.Fatal("expected set value")
	}
	assertF64Near(t, valShift.Get(0), 15.0)

	valAlt, textAlt := numericInputStepResult(
		"5", Opt[float64]{}, Opt[float64]{}, Opt[float64]{}, 2, cfg, NumericLocaleCfg{}, 1.0, ModAlt)
	if textAlt != "5.10" {
		t.Fatalf("alt: got %q, want %q", textAlt, "5.10")
	}
	if !valAlt.IsSet() {
		t.Fatal("expected set value")
	}
	assertF64Near(t, valAlt.Get(0), 5.1)
}

func TestNumericCurrencyCommitResultPrefixSymbol(t *testing.T) {
	mc := numericModeCfg{
		mode:              NumericCurrency,
		affix:             "$",
		affixPosition:     AffixPrefix,
		displayMultiplier: 1.0,
	}
	val, text := numericInputCommitResultMode(
		"-$1,234.5", Opt[float64]{}, Opt[float64]{}, Opt[float64]{}, 2, NumericLocaleCfg{}, mc)
	if text != "-$1,234.50" {
		t.Fatalf("got %q, want %q", text, "-$1,234.50")
	}
	if !val.IsSet() {
		t.Fatal("expected set value")
	}
	assertF64Near(t, val.Get(0), -1234.5)
}

func TestNumericCurrencyCommitResultSuffixSymbol(t *testing.T) {
	locale := NumericLocaleCfg{DecimalSep: ',', GroupSep: '.'}
	mc := numericModeCfg{
		mode:              NumericCurrency,
		affix:             "EUR",
		affixPosition:     AffixSuffix,
		affixSpacing:      true,
		displayMultiplier: 1.0,
	}
	val, text := numericInputCommitResultMode(
		"1.234,5 EUR", Opt[float64]{}, Opt[float64]{}, Opt[float64]{}, 2, locale, mc)
	if text != "1.234,50 EUR" {
		t.Fatalf("got %q, want %q", text, "1.234,50 EUR")
	}
	if !val.IsSet() {
		t.Fatal("expected set value")
	}
	assertF64Near(t, val.Get(0), 1234.5)
}

func TestNumericPercentCommitRatioValue(t *testing.T) {
	mc := numericModeCfg{
		mode:              NumericPercent,
		affix:             "%",
		affixPosition:     AffixSuffix,
		displayMultiplier: 100.0,
	}
	val, text := numericInputCommitResultMode(
		"12.5%", Opt[float64]{}, Opt[float64]{}, Opt[float64]{}, 2, NumericLocaleCfg{}, mc)
	if text != "12.50%" {
		t.Fatalf("got %q, want %q", text, "12.50%")
	}
	if !val.IsSet() {
		t.Fatal("expected set value")
	}
	assertF64Near(t, val.Get(0), 0.125)
}

func TestNumericPercentStepResultUsesDisplayUnits(t *testing.T) {
	mc := numericModeCfg{
		mode:              NumericPercent,
		affix:             "%",
		affixPosition:     AffixSuffix,
		displayMultiplier: 100.0,
	}
	val, text := numericInputStepResultMode(
		"12.50%", Opt[float64]{}, Opt[float64]{}, Opt[float64]{}, 2,
		NumericStepCfg{}, NumericLocaleCfg{}, 1.0, ModNone, mc)
	if text != "13.50%" {
		t.Fatalf("got %q, want %q", text, "13.50%")
	}
	if !val.IsSet() {
		t.Fatal("expected set value")
	}
	assertF64Near(t, val.Get(0), 0.135)
}

func TestNumericPercentRoundTripCanonical(t *testing.T) {
	mc := numericModeCfg{
		mode:              NumericPercent,
		affix:             "%",
		affixPosition:     AffixSuffix,
		displayMultiplier: 100.0,
	}
	loc := numericLocaleNormalize(NumericLocaleCfg{})
	source := -0.125
	formatted := numericModeFormatValue(source, 2, loc, mc)
	if formatted != "-12.50%" {
		t.Fatalf("formatted: got %q, want %q", formatted, "-12.50%")
	}
	parsed, ok := numericModeParseValue(formatted, 2, loc, mc)
	if !ok {
		t.Fatal("parse failed")
	}
	assertF64Near(t, parsed, source)
}

func TestNumericPreCommitRejectsInvalidDelta(t *testing.T) {
	mc := numericModeCfg{displayMultiplier: 1.0}
	_, ok := numericInputPreCommitTransformMode(
		"12", "12a", 2, NumericLocaleCfg{}, mc)
	if ok {
		t.Fatal("expected rejection")
	}
}

func TestNumericPreCommitAcceptsTransientForms(t *testing.T) {
	mc := numericModeCfg{displayMultiplier: 1.0}
	got, ok := numericInputPreCommitTransformMode(
		"", "-", 2, NumericLocaleCfg{}, mc)
	if !ok {
		t.Fatal("expected accept")
	}
	if got != "-" {
		t.Fatalf("got %q, want %q", got, "-")
	}
	got, ok = numericInputPreCommitTransformMode(
		"12", "12.", 2, NumericLocaleCfg{}, mc)
	if !ok {
		t.Fatal("expected accept")
	}
	if got != "12." {
		t.Fatalf("got %q, want %q", got, "12.")
	}
}

func TestNumericPreCommitAcceptsTransientCurrencyAffix(t *testing.T) {
	mc := numericModeCfg{
		mode:              NumericCurrency,
		affix:             "$",
		affixPosition:     AffixPrefix,
		displayMultiplier: 1.0,
	}
	got, ok := numericInputPreCommitTransformMode(
		"", "$", 2, NumericLocaleCfg{}, mc)
	if !ok {
		t.Fatal("expected accept")
	}
	if got != "$" {
		t.Fatalf("got %q, want %q", got, "$")
	}
	got, ok = numericInputPreCommitTransformMode(
		"", "-$", 2, NumericLocaleCfg{}, mc)
	if !ok {
		t.Fatal("expected accept")
	}
	if got != "-$" {
		t.Fatalf("got %q, want %q", got, "-$")
	}
}

func TestNumericPreCommitAcceptsTransientPercentAffix(t *testing.T) {
	mc := numericModeCfg{
		mode:              NumericPercent,
		affix:             "%",
		affixPosition:     AffixSuffix,
		displayMultiplier: 100.0,
	}
	got, ok := numericInputPreCommitTransformMode(
		"", "%", 2, NumericLocaleCfg{}, mc)
	if !ok {
		t.Fatal("expected accept")
	}
	if got != "%" {
		t.Fatalf("got %q, want %q", got, "%")
	}
	got, ok = numericInputPreCommitTransformMode(
		"12", "12.%", 2, NumericLocaleCfg{}, mc)
	if !ok {
		t.Fatal("expected accept")
	}
	if got != "12.%" {
		t.Fatalf("got %q, want %q", got, "12.%")
	}
}

func TestNumericLargeFloat64Overflow(t *testing.T) {
	val := 1.5e20
	formatted := numericFormat(val, 2, NumericLocaleCfg{})
	if formatted != "150,000,000,000,000,000,000.00" {
		t.Fatalf("got %q", formatted)
	}
}

func TestNumericEmptyPrefixSpacing(t *testing.T) {
	mc := numericModeCfg{
		mode:              NumericCurrency,
		affix:             "$",
		affixPosition:     AffixPrefix,
		affixSpacing:      true,
		displayMultiplier: 1.0,
	}
	got := numericApplyAffix("", numericLocaleNormalize(NumericLocaleCfg{}), mc)
	if got != "$" {
		t.Fatalf("got %q, want %q", got, "$")
	}
}

func TestNumericFormatNegativeZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	got := numericFormat(negZero, 2, NumericLocaleCfg{})
	if got != "0.00" {
		t.Fatalf("got %q, want %q", got, "0.00")
	}
}

func TestNumericGroupIntegerPartNonStandard(t *testing.T) {
	got := numericGroupIntegerPart("100", ',', []int{2})
	if got != "1,00" {
		t.Fatalf("got %q, want %q", got, "1,00")
	}
}

func TestNumericParseCollidingSeparators(t *testing.T) {
	loc := numericLocaleNormalize(NumericLocaleCfg{
		DecimalSep: '.', GroupSep: '.',
	})
	// When separators collide, groupSep is disabled; '.' is decimal.
	v, ok := numericParse("1.234", loc)
	if !ok {
		t.Fatal("parse failed")
	}
	assertF64Near(t, v, 1.234)
}

func TestNumericParseIndianNumbering(t *testing.T) {
	loc := numericLocaleNormalize(NumericLocaleCfg{
		GroupSizes: []int{3, 2},
	})
	v, ok := numericParse("12,34,567", loc)
	if !ok {
		t.Fatal("parse failed")
	}
	assertF64Near(t, v, 1234567)
}
