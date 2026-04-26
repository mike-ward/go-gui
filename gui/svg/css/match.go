package css

import (
	"slices"
	"strings"
)

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
	for _, a := range c.Attrs {
		if !matchAttr(a, el.Attrs) {
			return false
		}
	}
	if c.Root && !el.IsRoot {
		return false
	}
	if c.NthChild != nil && !c.NthChild.Matches(el.Index) {
		return false
	}
	if c.HoverPseudo && !el.State.Hover {
		return false
	}
	if c.FocusPseudo && !el.State.Focus {
		return false
	}
	if c.Not != nil && c.Not.Matches(el) {
		return false
	}
	return true
}

// matchAttr reports whether the (op, value) pair holds for el's named
// attribute. Missing attributes never match (even AttrOpExists is
// false when the attr is absent). Empty selector value is rejected
// for operators that require a non-empty needle.
func matchAttr(a AttrSel, attrs map[string]string) bool {
	v, ok := attrs[a.Name]
	if !ok {
		return false
	}
	switch a.Op {
	case AttrOpExists:
		return true
	case AttrOpEqual:
		return v == a.Value
	case AttrOpInclude:
		if a.Value == "" || strings.ContainsAny(a.Value, " \t\n\r\f") {
			return false
		}
		return slices.Contains(strings.Fields(v), a.Value)
	case AttrOpDashMatch:
		return v == a.Value || strings.HasPrefix(v, a.Value+"-")
	case AttrOpPrefix:
		return a.Value != "" && strings.HasPrefix(v, a.Value)
	case AttrOpSuffix:
		return a.Value != "" && strings.HasSuffix(v, a.Value)
	case AttrOpSubstring:
		return a.Value != "" && strings.Contains(v, a.Value)
	}
	return false
}

// Matches reports whether the complex selector matches el under the
// given ancestor stack and preceding sibling list. Ancestors are
// ordered root-first; ancestors[len-1] is the immediate parent.
// Siblings are ordered first-to-last with siblings[len-1] = el's
// immediate previous sibling. Both slices may be nil.
func (cs ComplexSelector) Matches(
	el ElementInfo, ancestors, siblings []ElementInfo,
) bool {
	parts := cs.Parts
	if len(parts) == 0 {
		return false
	}
	last := parts[len(parts)-1]
	if !last.Compound.Matches(el) {
		return false
	}
	ai := len(ancestors) - 1
	si := len(siblings) - 1
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
		case CombAdjacent:
			if si < 0 || !parts[i].Compound.Matches(siblings[si]) {
				return false
			}
			si--
		case CombGeneralSibling:
			matched := false
			for j := si; j >= 0; j-- {
				if parts[i].Compound.Matches(siblings[j]) {
					si = j - 1
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
// least one selector matching el under ancestors and preceding
// siblings. The result is in source order; callers run SortCascade
// to apply spec ordering.
func Match(
	rules []Rule, el ElementInfo, ancestors, siblings []ElementInfo,
) []MatchedDecl {
	if len(rules) == 0 {
		return nil
	}
	var out []MatchedDecl
	for ri := range rules {
		r := &rules[ri]
		var spec Specificity
		matched := false
		for _, sel := range r.Selectors {
			if !sel.Matches(el, ancestors, siblings) {
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
