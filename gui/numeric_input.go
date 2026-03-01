package gui

import (
	"math"
	"strconv"
	"strings"
)

// NumericInputMode determines how numeric values are displayed.
type NumericInputMode uint8

const (
	NumericNumber   NumericInputMode = iota
	NumericCurrency
	NumericPercent
)

// NumericAffixPosition determines prefix/suffix placement.
type NumericAffixPosition uint8

const (
	AffixPrefix NumericAffixPosition = iota
	AffixSuffix
)

// NumericLocaleCfg defines symbols for parse/format.
type NumericLocaleCfg struct {
	DecimalSep rune
	GroupSep   rune
	GroupSizes []int
	MinusSign  rune
	PlusSign   rune
}

// NumericStepCfg configures stepping interactions.
type NumericStepCfg struct {
	Step            float64
	ShiftMultiplier float64
	AltMultiplier   float64
	MouseWheel      bool
	Keyboard        bool
	ShowButtons     bool
}

// NumericCurrencyModeCfg defines currency symbol placement.
type NumericCurrencyModeCfg struct {
	Symbol        string
	Position      NumericAffixPosition
	SymbolSpacing bool
}

// NumericPercentModeCfg defines percent symbol placement.
type NumericPercentModeCfg struct {
	Symbol        string
	Position      NumericAffixPosition
	SymbolSpacing bool
}

// numericModeCfg is an internal config for mode-aware
// parse/format.
type numericModeCfg struct {
	mode               NumericInputMode
	affix              string
	affixPosition      NumericAffixPosition
	affixSpacing       bool
	displayMultiplier  float64
}

func numericLocaleNormalize(cfg NumericLocaleCfg) NumericLocaleCfg {
	sizes := make([]int, 0)
	for _, s := range cfg.GroupSizes {
		if s > 0 {
			sizes = append(sizes, s)
		}
	}
	if len(sizes) == 0 {
		sizes = append(sizes, 3)
	}
	groupSep := cfg.GroupSep
	if groupSep == 0 {
		groupSep = ','
	}
	decimalSep := cfg.DecimalSep
	if decimalSep == 0 {
		decimalSep = '.'
	}
	if groupSep == decimalSep {
		groupSep = 0
	}
	minus := cfg.MinusSign
	if minus == 0 {
		minus = '-'
	}
	plus := cfg.PlusSign
	if plus == 0 {
		plus = '+'
	}
	return NumericLocaleCfg{
		DecimalSep: decimalSep,
		GroupSep:   groupSep,
		GroupSizes: sizes,
		MinusSign:  minus,
		PlusSign:   plus,
	}
}

func numericStepCfgNormalize(cfg NumericStepCfg) NumericStepCfg {
	step := cfg.Step
	if step <= 0 {
		step = 1.0
	}
	shift := cfg.ShiftMultiplier
	if shift <= 0 {
		shift = 10.0
	}
	alt := cfg.AltMultiplier
	if alt <= 0 {
		alt = 0.1
	}
	return NumericStepCfg{
		Step:            step,
		ShiftMultiplier: shift,
		AltMultiplier:   alt,
		MouseWheel:      cfg.MouseWheel,
		Keyboard:        cfg.Keyboard,
		ShowButtons:     cfg.ShowButtons,
	}
}

func numericDecimalsClamped(decimals int) int {
	if decimals < 0 {
		return 0
	}
	if decimals > 9 {
		return 9
	}
	return decimals
}

func numericScaleFactor(decimals int) int64 {
	d := numericDecimalsClamped(decimals)
	factor := int64(1)
	for range d {
		factor *= 10
	}
	return factor
}

func numericToScaled(value float64, decimals int) int64 {
	factor := float64(numericScaleFactor(decimals))
	return int64(math.Round(value * factor))
}

func numericScaledToValue(scaled int64, decimals int) float64 {
	factor := float64(numericScaleFactor(decimals))
	return float64(scaled) / factor
}

func numericGroupSize(groupSizes []int, idx int) int {
	if idx >= 0 && idx < len(groupSizes) && groupSizes[idx] > 0 {
		return groupSizes[idx]
	}
	return 3
}

func numericGroupIntegerPart(raw string, groupSep rune, groupSizes []int) string {
	if groupSep == 0 {
		return raw
	}
	digits := []rune(raw)
	if len(digits) <= 3 {
		return raw
	}
	var reversed []rune
	count := 0
	groupIdx := 0
	gs := numericGroupSize(groupSizes, groupIdx)
	for i := len(digits) - 1; i >= 0; i-- {
		reversed = append(reversed, digits[i])
		count++
		if i > 0 && count == gs {
			reversed = append(reversed, groupSep)
			count = 0
			if groupIdx+1 < len(groupSizes) {
				groupIdx++
			}
			gs = numericGroupSize(groupSizes, groupIdx)
		}
	}
	// Reverse in place.
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return string(reversed)
}

func numericFormatScaled(scaled int64, decimals int, locale NumericLocaleCfg) string {
	loc := numericLocaleNormalize(locale)
	d := numericDecimalsClamped(decimals)
	scale := numericScaleFactor(d)
	magnitude := scaled
	sign := ""
	if scaled < 0 {
		magnitude = -scaled
		sign = string(loc.MinusSign)
	}
	intPart := strconv.FormatInt(magnitude/scale, 10)
	grouped := numericGroupIntegerPart(intPart, loc.GroupSep, loc.GroupSizes)
	if d == 0 {
		return sign + grouped
	}
	fracPart := strconv.FormatInt(magnitude%scale, 10)
	if len(fracPart) < d {
		fracPart = strings.Repeat("0", d-len(fracPart)) + fracPart
	}
	return sign + grouped + string(loc.DecimalSep) + fracPart
}

// numericFormat formats a value with the given decimals and
// locale.
func numericFormat(value float64, decimals int, locale NumericLocaleCfg) string {
	scaled := numericToScaled(value, decimals)
	return numericFormatScaled(scaled, decimals, locale)
}

func numericIntegerGroupsValid(intSegment []rune, groupSep rune, groupSizes []int) bool {
	var groupLengths []int
	count := 0
	for i := len(intSegment) - 1; i >= 0; i-- {
		ch := intSegment[i]
		if ch == groupSep {
			if count == 0 {
				return false
			}
			groupLengths = append(groupLengths, count)
			count = 0
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
		count++
	}
	if count == 0 {
		return false
	}
	groupLengths = append(groupLengths, count)
	for idx := 0; idx < len(groupLengths); idx++ {
		length := groupLengths[idx]
		expected := numericGroupSize(groupSizes, idx)
		if idx == len(groupLengths)-1 {
			if length <= expected {
				continue
			}
			return false
		}
		if length != expected {
			return false
		}
	}
	return true
}

// numericParse parses a locale-formatted number string.
func numericParse(raw string, locale NumericLocaleCfg) (float64, bool) {
	loc := numericLocaleNormalize(locale)
	if len(raw) == 0 {
		return 0, false
	}
	rs := []rune(raw)
	start := 0
	var normalized []rune
	seenDigit := false
	seenDecimal := false
	prevGroup := false
	sawGroupSep := false
	decimalIndex := -1

	if rs[0] == loc.MinusSign {
		normalized = append(normalized, '-')
		start = 1
	} else if rs[0] == loc.PlusSign {
		start = 1
	}
	for i := start; i < len(rs); i++ {
		ch := rs[i]
		if ch >= '0' && ch <= '9' {
			normalized = append(normalized, ch)
			seenDigit = true
			prevGroup = false
			continue
		}
		if ch == loc.DecimalSep {
			if seenDecimal || prevGroup {
				return 0, false
			}
			normalized = append(normalized, '.')
			seenDecimal = true
			prevGroup = false
			decimalIndex = i
			continue
		}
		if loc.GroupSep != 0 && ch == loc.GroupSep {
			if seenDecimal || !seenDigit || prevGroup {
				return 0, false
			}
			prevGroup = true
			sawGroupSep = true
			continue
		}
		return 0, false
	}
	if !seenDigit || prevGroup {
		return 0, false
	}
	if loc.GroupSep != 0 && sawGroupSep {
		intEnd := len(rs)
		if decimalIndex >= 0 {
			intEnd = decimalIndex
		}
		if intEnd < start {
			return 0, false
		}
		intSeg := rs[start:intEnd]
		if len(intSeg) == 0 {
			return 0, false
		}
		if !numericIntegerGroupsValid(intSeg, loc.GroupSep, loc.GroupSizes) {
			return 0, false
		}
	}
	number, err := strconv.ParseFloat(string(normalized), 64)
	if err != nil {
		return 0, false
	}
	return number, true
}

func numericParseScaled(raw string, decimals int, locale NumericLocaleCfg) (int64, bool) {
	parsed, ok := numericParse(raw, locale)
	if !ok {
		return 0, false
	}
	return numericToScaled(parsed, decimals), true
}

// numericClamp clamps value between optional min and max. nil
// pointers mean unbounded.
func numericClamp(value float64, minVal, maxVal *float64) float64 {
	lo := math.Inf(-1)
	hi := math.Inf(1)
	if minVal != nil {
		lo = *minVal
	}
	if maxVal != nil {
		hi = *maxVal
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	if value < lo {
		return lo
	}
	if value > hi {
		return hi
	}
	return value
}

// --- mode-aware functions ---

func makeNumericModeCfg(mode NumericInputMode, currency NumericCurrencyModeCfg, percent NumericPercentModeCfg) numericModeCfg {
	switch mode {
	case NumericCurrency:
		return numericModeCfg{
			mode:              NumericCurrency,
			affix:             strings.TrimSpace(currency.Symbol),
			affixPosition:     currency.Position,
			affixSpacing:      currency.SymbolSpacing,
			displayMultiplier: 1.0,
		}
	case NumericPercent:
		return numericModeCfg{
			mode:              NumericPercent,
			affix:             strings.TrimSpace(percent.Symbol),
			affixPosition:     percent.Position,
			affixSpacing:      percent.SymbolSpacing,
			displayMultiplier: 100.0,
		}
	default:
		return numericModeCfg{
			mode:              NumericNumber,
			displayMultiplier: 1.0,
		}
	}
}

func numericModeToDisplay(value float64, mc numericModeCfg) float64 {
	return value * mc.displayMultiplier
}

func numericModeFromDisplay(value float64, mc numericModeCfg) float64 {
	if mc.displayMultiplier == 0 {
		return value
	}
	return value / mc.displayMultiplier
}

func numericModeStepDelta(stepDisplay float64, mc numericModeCfg) float64 {
	return numericModeFromDisplay(stepDisplay, mc)
}

func numericStripAffix(raw string, locale NumericLocaleCfg, mc numericModeCfg) (string, bool) {
	loc := numericLocaleNormalize(locale)
	text := strings.TrimSpace(raw)
	if len(text) == 0 {
		return "", false
	}
	sign := ""
	minus := string(loc.MinusSign)
	plus := string(loc.PlusSign)
	if len(minus) > 0 && strings.HasPrefix(text, minus) {
		sign = minus
		text = strings.TrimLeft(text[len(minus):], " \t")
	} else if len(plus) > 0 && strings.HasPrefix(text, plus) {
		sign = plus
		text = strings.TrimLeft(text[len(plus):], " \t")
	}
	if len(mc.affix) > 0 {
		switch mc.affixPosition {
		case AffixPrefix:
			if strings.HasPrefix(text, mc.affix) {
				text = strings.TrimLeft(text[len(mc.affix):], " \t")
			}
		case AffixSuffix:
			right := strings.TrimRight(text, " \t")
			if strings.HasSuffix(right, mc.affix) {
				right = strings.TrimRight(right[:len(right)-len(mc.affix)], " \t")
				text = right
			}
		}
	}
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return "", false
	}
	return sign + text, true
}

func numericApplyAffix(formatted string, locale NumericLocaleCfg, mc numericModeCfg) string {
	if len(mc.affix) == 0 {
		return formatted
	}
	loc := numericLocaleNormalize(locale)
	minus := string(loc.MinusSign)
	plus := string(loc.PlusSign)
	sign := ""
	number := formatted
	if len(minus) > 0 && strings.HasPrefix(number, minus) {
		sign = minus
		number = number[len(minus):]
	} else if len(plus) > 0 && strings.HasPrefix(number, plus) {
		sign = plus
		number = number[len(plus):]
	}
	space := ""
	if mc.affixSpacing {
		space = " "
	}
	switch mc.affixPosition {
	case AffixPrefix:
		return sign + mc.affix + space + number
	case AffixSuffix:
		return sign + number + space + mc.affix
	}
	return formatted
}

func numericModeParseValue(raw string, decimals int, locale NumericLocaleCfg, mc numericModeCfg) (float64, bool) {
	plain, ok := numericStripAffix(strings.TrimSpace(raw), locale, mc)
	if !ok {
		return 0, false
	}
	displayScaled, ok := numericParseScaled(plain, decimals, locale)
	if !ok {
		return 0, false
	}
	displayValue := numericScaledToValue(displayScaled, decimals)
	return numericModeFromDisplay(displayValue, mc), true
}

func numericModeIsTransientInput(raw string, locale NumericLocaleCfg, mc numericModeCfg) bool {
	loc := numericLocaleNormalize(locale)
	text := strings.TrimSpace(raw)
	if len(text) == 0 {
		return true
	}
	minus := string(loc.MinusSign)
	plus := string(loc.PlusSign)
	if len(minus) > 0 && text == minus {
		return true
	}
	if len(plus) > 0 && text == plus {
		return true
	}
	if len(minus) > 0 && strings.HasPrefix(text, minus) {
		text = strings.TrimLeft(text[len(minus):], " \t")
	} else if len(plus) > 0 && strings.HasPrefix(text, plus) {
		text = strings.TrimLeft(text[len(plus):], " \t")
	}
	if len(text) == 0 {
		return true
	}
	if len(mc.affix) > 0 {
		switch mc.affixPosition {
		case AffixPrefix:
			if text == mc.affix {
				return true
			}
			if strings.HasPrefix(text, mc.affix) {
				text = strings.TrimLeft(text[len(mc.affix):], " \t")
				if len(text) == 0 {
					return true
				}
			}
		case AffixSuffix:
			if text == mc.affix {
				return true
			}
			right := strings.TrimRight(text, " \t")
			if strings.HasSuffix(right, mc.affix) {
				right = strings.TrimRight(
					right[:len(right)-len(mc.affix)], " \t")
				text = right
				if len(text) == 0 {
					return true
				}
			}
		}
	}
	decSep := string(loc.DecimalSep)
	if len(decSep) == 0 {
		return false
	}
	if text == decSep {
		return true
	}
	if !strings.HasSuffix(text, decSep) {
		return false
	}
	prefix := text[:len(text)-len(decSep)]
	if len(prefix) == 0 {
		return true
	}
	if _, ok := numericParse(prefix, loc); ok {
		return true
	}
	return false
}

func numericModeFormatValue(value float64, decimals int, locale NumericLocaleCfg, mc numericModeCfg) string {
	displayValue := numericModeToDisplay(value, mc)
	scaled := numericToScaled(displayValue, decimals)
	formatted := numericFormatScaled(scaled, decimals, locale)
	return numericApplyAffix(formatted, locale, mc)
}

func numericStepDelta(cfg NumericStepCfg, modifiers Modifier) float64 {
	step := cfg.Step
	if modifiers.Has(ModShift) {
		step *= cfg.ShiftMultiplier
	}
	if modifiers.Has(ModAlt) {
		step *= cfg.AltMultiplier
	}
	if step < 0 {
		return -step
	}
	return step
}

func numericStepSeedMode(text string, value *float64, minVal *float64, decimals int, locale NumericLocaleCfg, mc numericModeCfg) float64 {
	if value != nil {
		return *value
	}
	if parsed, ok := numericModeParseValue(text, decimals, locale, mc); ok {
		return parsed
	}
	if minVal != nil {
		return *minVal
	}
	return 0.0
}

// numericInputCommitResult resolves text → (value, formatted).
func numericInputCommitResult(text string, value *float64, minVal, maxVal *float64, decimals int, locale NumericLocaleCfg) (*float64, string) {
	return numericInputCommitResultMode(text, value, minVal, maxVal, decimals, locale, numericModeCfg{displayMultiplier: 1.0})
}

func numericInputCommitResultMode(text string, value *float64, minVal, maxVal *float64, decimals int, locale NumericLocaleCfg, mc numericModeCfg) (*float64, string) {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 {
		return nil, ""
	}
	if parsed, ok := numericModeParseValue(trimmed, decimals, locale, mc); ok {
		clamped := numericClamp(parsed, minVal, maxVal)
		return &clamped, numericModeFormatValue(clamped, decimals, locale, mc)
	}
	if value != nil {
		clamped := numericClamp(*value, minVal, maxVal)
		return &clamped, numericModeFormatValue(clamped, decimals, locale, mc)
	}
	return nil, ""
}

// numericInputStepResult steps a value in the given direction.
// numericInputStepResult steps a value in the given direction.
func numericInputStepResult(text string, value *float64, minVal, maxVal *float64, decimals int, stepCfg NumericStepCfg, locale NumericLocaleCfg, direction float64, modifiers Modifier) (*float64, string) {
	return numericInputStepResultMode(text, value, minVal, maxVal, decimals, stepCfg, locale, direction, modifiers, numericModeCfg{displayMultiplier: 1.0})
}

func numericInputStepResultMode(text string, value *float64, minVal, maxVal *float64, decimals int, stepCfg NumericStepCfg, locale NumericLocaleCfg, direction float64, modifiers Modifier, mc numericModeCfg) (*float64, string) {
	if direction == 0 {
		return numericInputCommitResultMode(text, value, minVal, maxVal, decimals, locale, mc)
	}
	normalized := numericStepCfgNormalize(stepCfg)
	stepDisplay := numericStepDelta(normalized, modifiers)
	delta := numericModeStepDelta(stepDisplay, mc)
	seed := numericStepSeedMode(text, value, minVal, decimals, locale, mc)
	clamped := numericClamp(seed+(delta*direction), minVal, maxVal)
	return &clamped, numericModeFormatValue(clamped, decimals, locale, mc)
}

func numericInputPreCommitTransformMode(current, proposed string, decimals int, locale NumericLocaleCfg, mc numericModeCfg) (string, bool) {
	if proposed == current {
		return proposed, true
	}
	trimmed := strings.TrimSpace(proposed)
	if len(trimmed) == 0 {
		return "", true
	}
	if _, ok := numericModeParseValue(trimmed, decimals, locale, mc); ok {
		return proposed, true
	}
	if numericModeIsTransientInput(proposed, locale, mc) {
		return proposed, true
	}
	return "", false
}
