package svg

import (
	"strconv"
	"strings"
	"testing"
)

// --- hasUseElement short-circuit ---

func TestHasUseElementAbsent(t *testing.T) {
	root, err := decodeSvgTree(
		`<svg viewBox="0 0 10 10"><rect width="5" height="5"/></svg>`)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if hasUseElement(root) {
		t.Fatalf("expected hasUseElement=false")
	}
}

func TestHasUseElementPresent(t *testing.T) {
	root, err := decodeSvgTree(
		`<svg viewBox="0 0 10 10"><defs><rect id="r" width="5" height="5"/></defs><use href="#r"/></svg>`)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !hasUseElement(root) {
		t.Fatalf("expected hasUseElement=true")
	}
}

// --- expandUseElements basic ---

func TestExpandUseSymbol(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100">
		<defs>
			<symbol id="s"><rect width="10" height="10" fill="red"/></symbol>
		</defs>
		<use href="#s" x="5" y="6"/>
	</svg>`
	vg, err := parseSvg(svg)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(vg.Paths) == 0 {
		t.Fatalf("expected at least one path from <use>")
	}
}

func TestExpandUseTranslateComposesTransform(t *testing.T) {
	root, err := decodeSvgTree(
		`<svg viewBox="0 0 100 100">` +
			`<defs><symbol id="s"><rect width="1" height="1"/></symbol></defs>` +
			`<use href="#s" x="3" y="4" transform="rotate(90)"/>` +
			`</svg>`)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	expandUseElements(root)
	useG := findFirstByName(root, "g")
	if useG == nil {
		t.Fatalf("expected synthesized <g> wrapper")
	}
	tr := useG.AttrMap["transform"]
	if !strings.Contains(tr, "rotate(90)") || !strings.Contains(tr, "translate(3,4)") {
		t.Fatalf("expected composed transform, got %q", tr)
	}
	if strings.Index(tr, "rotate(90)") > strings.Index(tr, "translate(3,4)") {
		t.Fatalf("expected rotate before translate, got %q", tr)
	}
}

// --- malformed href ---

func TestExpandUseMalformedHref(t *testing.T) {
	cases := []string{
		`<svg><defs><rect id="a" width="1" height="1"/></defs><use href=""/></svg>`,
		`<svg><defs><rect id="a" width="1" height="1"/></defs><use href="#"/></svg>`,
		`<svg><defs><rect id="a" width="1" height="1"/></defs><use href="a"/></svg>`,
		`<svg><defs><rect id="a" width="1" height="1"/></defs><use href="#missing"/></svg>`,
	}
	for i, svg := range cases {
		root, err := decodeSvgTree(svg)
		if err != nil {
			t.Fatalf("case %d decode: %v", i, err)
		}
		expandUseElements(root)
		// <use> remains as-is when unresolvable; parse must not loop or crash.
		if findFirstByName(root, "use") == nil {
			t.Fatalf("case %d: <use> should remain when href is unresolvable", i)
		}
	}
}

// --- cycle guard ---

func TestExpandUseCycleGuard(t *testing.T) {
	svg := `<svg>
		<defs>
			<g id="a"><use href="#b"/></g>
			<g id="b"><use href="#a"/></g>
		</defs>
		<use href="#a"/>
	</svg>`
	root, err := decodeSvgTree(svg)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	expandUseElements(root)
	// no panic, no infinite expansion
}

// --- depth cap ---

func TestExpandUseDepthCap(t *testing.T) {
	var b strings.Builder
	b.WriteString(`<svg><defs>`)
	for i := range 20 {
		// chain: id_i contains <use href="#id_{i+1}">
		// last id has no use
		if i < 19 {
			b.WriteString(`<g id="g`)
			b.WriteString(strconv.Itoa(i))
			b.WriteString(`"><use href="#g`)
			b.WriteString(strconv.Itoa(i + 1))
			b.WriteString(`"/></g>`)
		} else {
			b.WriteString(`<g id="g`)
			b.WriteString(strconv.Itoa(i))
			b.WriteString(`"><rect width="1" height="1"/></g>`)
		}
	}
	b.WriteString(`</defs><use href="#g0"/></svg>`)

	root, err := decodeSvgTree(b.String())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	expandUseElements(root)
	// passes if no stack blow-up; depth cap stops chain.
}

// --- fanout budget ---

func TestExpandUseFanoutBudget(t *testing.T) {
	// Build symbol with a single rect, then a parent containing
	// many <use href="#s"> siblings. Each surfaces 1 child clone, so
	// total clones <= budget.
	var b strings.Builder
	b.WriteString(`<svg><defs><symbol id="s"><rect width="1" height="1"/></symbol></defs>`)
	const fan = 5
	for range fan {
		b.WriteString(`<use href="#s"/>`)
	}
	b.WriteString(`</svg>`)
	root, err := decodeSvgTree(b.String())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	expandUseElements(root)
	// Each <use> becomes a <g> wrapping the symbol's rect child.
	gs := countByName(root, "g")
	if gs < fan {
		t.Fatalf("expected at least %d <g> wrappers, got %d", fan, gs)
	}
}

func TestExpandUseFanoutBudgetTruncates(t *testing.T) {
	// Force budget exhaustion: clone deeply nested target many times.
	// Use a target with multiple children so each <use> consumes
	// several budget units.
	var sym strings.Builder
	sym.WriteString(`<symbol id="s">`)
	for range 100 {
		sym.WriteString(`<rect width="1" height="1"/>`)
	}
	sym.WriteString(`</symbol>`)

	var b strings.Builder
	b.WriteString(`<svg><defs>`)
	b.WriteString(sym.String())
	b.WriteString(`</defs>`)
	for range 2000 {
		b.WriteString(`<use href="#s"/>`)
	}
	b.WriteString(`</svg>`)
	root, err := decodeSvgTree(b.String())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Must complete without OOM/panic and budget must clamp output.
	expandUseElements(root)
	// Without the budget, expansion would produce 2000 × 100 = 200000
	// cloned rects on top of the 100 in the source <symbol>. The
	// budget caps total clones at maxUseExpandClones; the surviving
	// rect count must be well below the unbounded result.
	rectCount := countByName(root, "rect")
	if rectCount > maxUseExpandClones+200 {
		t.Fatalf("rectCount=%d exceeds budget %d (+slack)",
			rectCount, maxUseExpandClones)
	}
	if rectCount >= 200000 {
		t.Fatalf("budget did not truncate fanout: rectCount=%d", rectCount)
	}
}

// --- ID stripping after clone ---

func TestExpandUseStripsClonedID(t *testing.T) {
	root, err := decodeSvgTree(
		`<svg><defs><symbol id="s"><rect id="inner" width="1" height="1"/></symbol></defs>` +
			`<use href="#s"/></svg>`)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	expandUseElements(root)
	// At most one rect retains id="inner" (the original inside
	// <symbol>); cloned rects under the synthesized <g> must not.
	withID := 0
	for _, r := range collectByName(root, "rect") {
		if r.AttrMap["id"] == "inner" {
			withID++
		}
	}
	if withID > 1 {
		t.Fatalf("expected at most 1 rect with id=inner, got %d", withID)
	}
}

// --- helpers ---

func findFirstByName(n *xmlNode, name string) *xmlNode {
	for i := range n.Children {
		c := &n.Children[i]
		if c.Name == name {
			return c
		}
		if r := findFirstByName(c, name); r != nil {
			return r
		}
	}
	return nil
}

func collectByName(n *xmlNode, name string) []*xmlNode {
	var out []*xmlNode
	walk(n, func(c *xmlNode) {
		if c.Name == name {
			out = append(out, c)
		}
	})
	return out
}

func countByName(n *xmlNode, name string) int {
	count := 0
	walk(n, func(c *xmlNode) {
		if c.Name == name {
			count++
		}
	})
	return count
}

func walk(n *xmlNode, fn func(*xmlNode)) {
	for i := range n.Children {
		c := &n.Children[i]
		fn(c)
		walk(c, fn)
	}
}
