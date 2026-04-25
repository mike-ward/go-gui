package css

import (
	"sort"
	"strconv"
	"strings"

	tdcss "github.com/tdewolff/parse/v2/css"
)

// isKeyframesAtRule reports whether at-rule data names @keyframes
// (or a vendor-prefixed variant — tdewolff strips well-known
// prefixes for us, but we still lowercase to be defensive).
func isKeyframesAtRule(data []byte) bool {
	s := strings.ToLower(string(data))
	return s == "@keyframes" ||
		s == "@-webkit-keyframes" ||
		s == "@-moz-keyframes" ||
		s == "@-o-keyframes"
}

// keyframesName extracts the animation name token from a
// @keyframes prelude. Whitespace tokens are skipped; the first
// IdentToken (or quoted-string identifier) wins.
func keyframesName(toks []tdcss.Token) string {
	for _, t := range toks {
		if t.TokenType == tdcss.WhitespaceToken {
			continue
		}
		if t.TokenType == tdcss.IdentToken {
			return string(t.Data)
		}
		if t.TokenType == tdcss.StringToken {
			s := string(t.Data)
			if len(s) >= 2 &&
				(s[0] == '"' || s[0] == '\'') &&
				s[len(s)-1] == s[0] {
				return s[1 : len(s)-1]
			}
			return s
		}
	}
	return ""
}

// parseKeyframeSelectors converts the keyframe stop selectors
// ("0%", "from", "50%, 100%", etc.) into a list of resolved
// [0,1] offsets. Returns ok=false when nothing parses cleanly.
func parseKeyframeSelectors(toks []tdcss.Token) ([]float32, bool) {
	var groups [][]tdcss.Token
	start := 0
	for i, t := range toks {
		if t.TokenType == tdcss.CommaToken {
			groups = append(groups, toks[start:i])
			start = i + 1
		}
	}
	groups = append(groups, toks[start:])
	out := make([]float32, 0, len(groups))
	for _, g := range groups {
		v, ok := parseKeyframeStopSelector(g)
		if !ok {
			continue
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

// parseKeyframeStopSelector resolves one keyframe stop selector to
// a [0,1] offset. Recognized: "from"=0, "to"=1, percentage tokens
// (PercentageToken or DimensionToken with a "%" unit), bare numbers
// (treated as percentages — non-spec but tolerant of authoring
// shortcuts).
func parseKeyframeStopSelector(toks []tdcss.Token) (float32, bool) {
	for _, t := range toks {
		if t.TokenType == tdcss.WhitespaceToken {
			continue
		}
		switch t.TokenType {
		case tdcss.IdentToken:
			s := strings.ToLower(string(t.Data))
			switch s {
			case "from":
				return 0, true
			case "to":
				return 1, true
			}
			return 0, false
		case tdcss.PercentageToken:
			s := strings.TrimSuffix(string(t.Data), "%")
			f, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return 0, false
			}
			return clampOffset(float32(f) / 100), true
		case tdcss.NumberToken:
			f, err := strconv.ParseFloat(string(t.Data), 32)
			if err != nil {
				return 0, false
			}
			return clampOffset(float32(f) / 100), true
		case tdcss.DimensionToken:
			s := string(t.Data)
			if !strings.HasSuffix(s, "%") {
				return 0, false
			}
			f, err := strconv.ParseFloat(s[:len(s)-1], 32)
			if err != nil {
				return 0, false
			}
			return clampOffset(float32(f) / 100), true
		}
	}
	return 0, false
}

func clampOffset(v float32) float32 {
	// NaN comparison always false, so explicit guard required.
	if v != v {
		return 0
	}
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// sortKeyframeStops sorts stops ascending by Offset. Stable so
// duplicate offsets keep source order — later entries win during
// timeline materialization.
func sortKeyframeStops(stops []KeyframeStop) {
	sort.SliceStable(stops, func(i, j int) bool {
		return stops[i].Offset < stops[j].Offset
	})
}
