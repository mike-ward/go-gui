package css

import (
	"strings"

	tdcss "github.com/tdewolff/parse/v2/css"
)

// advanceSkippedMedia steps the media-skip state machine while the
// outer @media block's query did not match. It only updates the
// depth counter on nested @-rule begin/end events so the skip ends
// at the matching outer EndAtRule. Any other grammar event is a
// no-op (the caller drops it).
func advanceSkippedMedia(
	gt tdcss.GrammarType, data []byte, depth *int, skip *bool,
) {
	switch gt {
	case tdcss.BeginAtRuleGrammar:
		if isMediaAtRule(data) {
			*depth++
		}
	case tdcss.EndAtRuleGrammar:
		if *depth > 0 {
			*depth--
			if *depth == 0 {
				*skip = false
			}
		}
	}
}

// isMediaAtRule reports whether at-rule data names @media. tdewolff
// keeps the leading '@'; lowercase before comparing because authors
// occasionally use @MEDIA in source.
func isMediaAtRule(data []byte) bool {
	return strings.EqualFold(string(data), "@media")
}

// mediaMatches evaluates the prelude of an @media block against the
// snapshotted ParseOptions. The only supported query is
// `(prefers-reduced-motion: reduce)`; every other prelude is treated
// as an unmatched query and drops its block. A bare `@media` with no
// prelude (rare; non-spec) is also treated as unmatched. Comma-
// separated query lists succeed if any branch matches.
func mediaMatches(toks []tdcss.Token, opts ParseOptions) bool {
	for _, group := range splitByComma(toks) {
		if mediaQueryMatches(group, opts) {
			return true
		}
	}
	return false
}

// mediaQueryMatches evaluates one comma-delimited media query branch.
// Recognized only:
//
//	(prefers-reduced-motion: reduce)   → opts.PrefersReducedMotion
//	(prefers-reduced-motion: no-preference)
//	                                    → !opts.PrefersReducedMotion
//
// Anything else drops the branch.
func mediaQueryMatches(toks []tdcss.Token, opts ParseOptions) bool {
	feature, value, ok := extractMediaFeature(toks)
	if !ok {
		return false
	}
	if !strings.EqualFold(feature, "prefers-reduced-motion") {
		return false
	}
	switch strings.ToLower(value) {
	case "reduce":
		return opts.PrefersReducedMotion
	case "no-preference":
		return !opts.PrefersReducedMotion
	}
	return false
}

// extractMediaFeature parses a `(name: value)` block from the token
// stream. Returns ok=false if the parens are malformed or the block
// is not a single name:value pair.
func extractMediaFeature(toks []tdcss.Token) (string, string, bool) {
	toks = trimWS(toks)
	if len(toks) < 2 {
		return "", "", false
	}
	if toks[0].TokenType != tdcss.LeftParenthesisToken {
		return "", "", false
	}
	if toks[len(toks)-1].TokenType != tdcss.RightParenthesisToken {
		return "", "", false
	}
	inner := trimWS(toks[1 : len(toks)-1])
	if len(inner) == 0 {
		return "", "", false
	}
	colon := -1
	for i, t := range inner {
		if t.TokenType == tdcss.ColonToken {
			colon = i
			break
		}
	}
	if colon <= 0 || colon == len(inner)-1 {
		return "", "", false
	}
	nameToks := trimWS(inner[:colon])
	if len(nameToks) != 1 || nameToks[0].TokenType != tdcss.IdentToken {
		return "", "", false
	}
	valToks := trimWS(inner[colon+1:])
	if len(valToks) == 0 {
		return "", "", false
	}
	var b strings.Builder
	for _, t := range valToks {
		b.Write(t.Data)
	}
	return string(nameToks[0].Data), strings.TrimSpace(b.String()), true
}
