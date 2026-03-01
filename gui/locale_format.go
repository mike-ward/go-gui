package gui

import (
	"fmt"
	"strings"
	"time"
)

// LocaleFormatDate formats a date using locale-aware month
// substitution. MMMM -> full month, MMM -> short month.
// Other tokens use a simple V-style token replacement:
//   YYYY->year, M->month, D->day, HH->hour, mm->minute,
//   ss->second.
func LocaleFormatDate(t time.Time, format string) string {
	monthIdx := int(t.Month()) - 1
	if monthIdx < 0 || monthIdx >= 12 {
		return localeDateReplace(t, format)
	}
	result := format
	hasFull := strings.Contains(result, "MMMM")
	hasShort := !hasFull && strings.Contains(result, "MMM")
	if hasFull {
		result = strings.Replace(result, "MMMM", "\x01\x01\x01\x01", 1)
	} else if hasShort {
		result = strings.Replace(result, "MMM", "\x01\x01\x01", 1)
	}
	result = localeDateReplace(t, result)
	if hasFull {
		result = strings.Replace(result, "\x01\x01\x01\x01",
			guiLocale.MonthsFull[monthIdx], 1)
	} else if hasShort {
		result = strings.Replace(result, "\x01\x01\x01",
			guiLocale.MonthsShort[monthIdx], 1)
	}
	return result
}

// localeDateReplace performs simple V-style date token
// substitution. Tokens: YYYY, MM (zero-padded month),
// M (month), DD (zero-padded day), D (day),
// HH, mm, ss.
func localeDateReplace(t time.Time, format string) string {
	r := format
	r = strings.ReplaceAll(r, "YYYY", fmt.Sprintf("%04d", t.Year()))
	// MM must be replaced before M to avoid double-replace.
	// Use sentinel to protect.
	if strings.Contains(r, "MM") {
		r = strings.ReplaceAll(r, "MM",
			fmt.Sprintf("%02d", int(t.Month())))
	} else {
		r = strings.ReplaceAll(r, "M",
			fmt.Sprintf("%d", int(t.Month())))
	}
	if strings.Contains(r, "DD") {
		r = strings.ReplaceAll(r, "DD",
			fmt.Sprintf("%02d", t.Day()))
	} else {
		r = strings.ReplaceAll(r, "D",
			fmt.Sprintf("%d", t.Day()))
	}
	r = strings.ReplaceAll(r, "HH",
		fmt.Sprintf("%02d", t.Hour()))
	r = strings.ReplaceAll(r, "mm",
		fmt.Sprintf("%02d", t.Minute()))
	r = strings.ReplaceAll(r, "ss",
		fmt.Sprintf("%02d", t.Second()))
	return r
}

// localeRowsFmt formats "Rows start-end/total".
func localeRowsFmt(start, end, total int) string {
	return fmt.Sprintf("%s %d-%d/%d",
		guiLocale.StrRows, start, end, total)
}

// localePageFmt formats "Page current/total".
func localePageFmt(page, total int) string {
	return fmt.Sprintf("%s %d/%d",
		guiLocale.StrPage, page, total)
}

// localeMatchesFmt formats "Matches count/total".
func localeMatchesFmt(count int, total string) string {
	return fmt.Sprintf("%s %d/%s",
		guiLocale.StrMatches, count, total)
}

// LocaleT looks up a translation key in the current locale.
// Returns the key itself when not found.
func LocaleT(key string) string {
	if v, ok := guiLocale.Translations[key]; ok {
		return v
	}
	return key
}
