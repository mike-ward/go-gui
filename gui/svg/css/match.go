package css

import "slices"

// Matches reports whether a compound selector matches el.
func (c Compound) Matches(el ElementInfo) bool {
	if c.Tag != "" && c.Tag != "*" && c.Tag != el.Tag {
		return false
	}
	if c.ID != "" && c.ID != el.ID {
		return false
	}
	for _, cls := range c.Classes {
		if !slices.Contains(el.Classes, cls) {
			return false
		}
	}
	if c.Root && !el.IsRoot {
		return false
	}
	if c.NthChild != nil && !c.NthChild.Matches(el.Index) {
		return false
	}
	return true
}

// Matches reports whether the complex selector matches el under the
// given ancestor stack. Ancestors are ordered root-first so
// ancestors[len-1] is the immediate parent.
func (cs ComplexSelector) Matches(el ElementInfo, ancestors []ElementInfo) bool {
	parts := cs.Parts
	if len(parts) == 0 {
		return false
	}
	last := parts[len(parts)-1]
	if !last.Compound.Matches(el) {
		return false
	}
	ai := len(ancestors) - 1
	for i := len(parts) - 2; i >= 0; i-- {
		comb := parts[i+1].Combinator
		switch comb {
		case CombChild:
			if ai < 0 || !parts[i].Compound.Matches(ancestors[ai]) {
				return false
			}
			ai--
		case CombDescendant:
			matched := false
			for j := ai; j >= 0; j-- {
				if parts[i].Compound.Matches(ancestors[j]) {
					ai = j - 1
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// MatchedDecl pairs a declaration with the cascade context (origin,
// specificity, source order, !important) needed to rank it against
// other declarations during the cascade walk.
type MatchedDecl struct {
	Decl
	Origin Origin
	Spec   Specificity
	Source int
}

// Match returns every author-rule declaration whose rule has at
// least one selector matching el under ancestors. The result is in
// source order; callers run SortCascade to apply spec ordering.
func Match(rules []Rule, el ElementInfo, ancestors []ElementInfo) []MatchedDecl {
	if len(rules) == 0 {
		return nil
	}
	var out []MatchedDecl
	for ri := range rules {
		r := &rules[ri]
		var spec Specificity
		matched := false
		for _, sel := range r.Selectors {
			if !sel.Matches(el, ancestors) {
				continue
			}
			if !matched || spec.Less(sel.Spec) {
				spec = sel.Spec
				matched = true
			}
		}
		if !matched {
			continue
		}
		for _, d := range r.Decls {
			out = append(out, MatchedDecl{
				Decl:   d,
				Origin: OriginRule,
				Spec:   spec,
				Source: r.Source,
			})
		}
	}
	return out
}

// SortCascade orders matched declarations by the SVG-CSS cascade.
// The comparator gives a single rank per decl that bakes Origin and
// !important into a uint8 layer (low layer applies first, high layer
// last so it wins under "last write wins"):
//
//	0  pres-attr normal
//	1  rule normal
//	2  inline normal
//	3  pres-attr important (rare; SVG pres-attrs cannot carry
//	   !important per spec, but slot is reserved for symmetry)
//	4  rule important
//	5  inline important
//
// Within a layer: specificity ascending, then source order. Stable
// so callers can iterate and apply "last write wins" per property.
func SortCascade(decls []MatchedDecl) {
	for i := 1; i < len(decls); i++ {
		j := i
		for j > 0 && cascadeLess(decls[j], decls[j-1]) {
			decls[j], decls[j-1] = decls[j-1], decls[j]
			j--
		}
	}
}

func cascadeLayer(d MatchedDecl) uint8 {
	base := uint8(d.Origin)
	if d.Important {
		return base + numOrigins
	}
	return base
}

func cascadeLess(a, b MatchedDecl) bool {
	la, lb := cascadeLayer(a), cascadeLayer(b)
	if la != lb {
		return la < lb
	}
	if a.Spec != b.Spec {
		return a.Spec.Less(b.Spec)
	}
	return a.Source < b.Source
}
