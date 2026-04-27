package svg

import (
	"strings"
	"testing"
)

// Adversarial: a value containing a single quote that opens a fake
// embedded attribute would let findAttr return the injected value.
// writeAttrEscaped must escape ' so the cascade sees the real attr.
func TestAttrInjection_SingleQuoteEscaped(t *testing.T) {
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">` +
		`<rect note=" x='99' " x="1" y="1" width="2" height="2" fill="red"/>` +
		`</svg>`
	root, err := decodeSvgTree(src)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	var rect *xmlNode
	var walk func(n *xmlNode)
	walk = func(n *xmlNode) {
		if n.Name == "rect" {
			rect = n
			return
		}
		for i := range n.Children {
			walk(&n.Children[i])
			if rect != nil {
				return
			}
		}
	}
	walk(root)
	if rect == nil {
		t.Fatalf("rect not found")
	}
	if strings.Contains(rect.OpenTag, "x='99'") {
		t.Fatalf("single quote not escaped, OpenTag=%q", rect.OpenTag)
	}
	got, ok := findAttr(rect.OpenTag, "x")
	if !ok || got != "1" {
		t.Fatalf("expected x=1, got x=%q ok=%v openTag=%q",
			got, ok, rect.OpenTag)
	}
}

// Adversarial: input larger than maxSvgFileSize must be rejected by
// parseSvg directly, not only by the file loader.
func TestParseSvg_RejectsOversizeContent(t *testing.T) {
	huge := strings.Repeat("a", maxSvgFileSize+1)
	_, err := parseSvg(huge)
	if err == nil {
		t.Fatalf("expected oversize error, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Fatalf("expected too-large error, got %v", err)
	}
}

// Adversarial: many paths sharing a single complex clipPath should
// not re-tessellate the clip geometry per path. Cache hit is asserted
// indirectly: tessellate completes for a moderately complex case in a
// single call without producing duplicate distinct triangle slices.
func TestClipPathCache_ReusesTriangles(t *testing.T) {
	const rectCount = 8
	const subpathCount = 2
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">`)
	b.WriteString(`<defs><clipPath id="c">`)
	b.WriteString(`<path d="M0,0 L50,0 L50,50 L0,50 Z"/>`)
	b.WriteString(`<path d="M50,50 L100,50 L100,100 L50,100 Z"/>`)
	b.WriteString(`</clipPath></defs>`)
	for range rectCount {
		b.WriteString(`<rect x="0" y="0" width="100" height="100" ` +
			`fill="red" clip-path="url(#c)"/>`)
	}
	b.WriteString(`</svg>`)
	vg, err := parseSvg(b.String())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	tris := vg.tessellatePaths(vg.Paths, 1.0)
	// Each rect emits subpathCount clip-mask entries; expect masks to
	// share underlying triangle slices via cache (one per subpath).
	var masks [][]float32
	for _, tp := range tris {
		if tp.IsClipMask {
			masks = append(masks, tp.Triangles)
		}
	}
	wantMasks := rectCount * subpathCount
	if len(masks) < wantMasks {
		t.Fatalf("expected >=%d mask entries, got %d",
			wantMasks, len(masks))
	}
	// Verify slice header reuse: at least two mask entries (across
	// different paths) must point to the same backing array.
	shared := false
	for i := range masks {
		for j := i + 1; j < len(masks); j++ {
			if len(masks[i]) > 0 && len(masks[j]) > 0 &&
				&masks[i][0] == &masks[j][0] {
				shared = true
				break
			}
		}
		if shared {
			break
		}
	}
	if !shared {
		t.Fatalf("clip mask triangles not shared across paths " +
			"(cache miss)")
	}
}

// Locks in the writeAttrEscaped escape table. Each pair is the raw
// rune in the attribute value and its expected encoding in the
// reconstructed OpenTag. Single quote and angle brackets matter for
// blocking findAttr injection; double quote, ampersand, less-than
// preserve XML attribute-value validity.
func TestWriteAttrEscaped_EscapesAllSpecialChars(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"doubleQuote", `a"b`, "a&quot;b"},
		{"singleQuote", `a'b`, "a&#39;b"},
		{"ampersand", `a&b`, "a&amp;b"},
		{"lessThan", `a<b`, "a&lt;b"},
		{"greaterThan", `a>b`, "a&gt;b"},
		{"plain", `abc`, "abc"},
		{"empty", ``, ``},
		{"allSpecials", `<&">'`, "&lt;&amp;&quot;&gt;&#39;"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			writeAttrEscaped(&b, tc.in)
			if got := b.String(); got != tc.want {
				t.Fatalf("escape(%q): got %q want %q",
					tc.in, got, tc.want)
			}
		})
	}
}

// parseSvgDimensions caps caller-supplied input to maxSvgFileSize.
// Truncated tail must not affect dimensions extracted from the root
// <svg> tag, and oversize input must not panic or stall.
func TestParseSvgDimensions_OversizeContentTruncated(t *testing.T) {
	head := `<svg xmlns="http://www.w3.org/2000/svg" ` +
		`viewBox="0 0 42 24" width="42" height="24">`
	pad := strings.Repeat("x", maxSvgFileSize)
	src := head + pad + `</svg>`
	w, h := parseSvgDimensions(src)
	if w != 42 || h != 24 {
		t.Fatalf("expected 42x24, got %vx%v", w, h)
	}
	// Boundary: open tag straddling the cap. Cap to a length that
	// chops mid-attribute; dimension probe should fall back to
	// defaults rather than crashing.
	short := head[:5] + strings.Repeat("y", maxSvgFileSize+1)
	w2, h2 := parseSvgDimensions(short)
	if w2 <= 0 || h2 <= 0 {
		t.Fatalf("non-positive fallback dims: %vx%v", w2, h2)
	}
}

// Distinct ClipPathIDs must not collide in the cache. Each ID gets
// its own tessellated triangle set; mixing must not return one ID's
// geometry under another's key.
func TestClipPathCache_DistinctIDsTessellateIndependently(t *testing.T) {
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">` +
		`<defs>` +
		`<clipPath id="a"><path d="M0,0 L10,0 L10,10 L0,10 Z"/></clipPath>` +
		`<clipPath id="b"><path d="M0,0 L80,0 L80,80 L0,80 Z"/></clipPath>` +
		`</defs>` +
		`<rect x="0" y="0" width="100" height="100" fill="red" clip-path="url(#a)"/>` +
		`<rect x="0" y="0" width="100" height="100" fill="blue" clip-path="url(#b)"/>` +
		`<rect x="0" y="0" width="100" height="100" fill="green" clip-path="url(#a)"/>` +
		`</svg>`
	vg, err := parseSvg(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	tris := vg.tessellatePaths(vg.Paths, 1.0)

	// Group masks by ClipGroup. Three rects → three groups, but
	// groups 1 and 3 share clipPath "a" and must point to the same
	// underlying triangle slice; group 2 is "b" and must differ.
	groupTris := map[int][]float32{}
	for _, tp := range tris {
		if !tp.IsClipMask {
			continue
		}
		if _, ok := groupTris[tp.ClipGroup]; !ok {
			groupTris[tp.ClipGroup] = tp.Triangles
		}
	}
	if len(groupTris) < 3 {
		t.Fatalf("expected 3 clip groups, got %d", len(groupTris))
	}
	a1, a2, b := groupTris[1], groupTris[3], groupTris[2]
	if len(a1) == 0 || len(a2) == 0 || len(b) == 0 {
		t.Fatalf("empty mask tris: a1=%d a2=%d b=%d",
			len(a1), len(a2), len(b))
	}
	if &a1[0] != &a2[0] {
		t.Fatalf("clipPath 'a' not cached across groups (different backing)")
	}
	if &a1[0] == &b[0] {
		t.Fatalf("clipPath 'a' and 'b' collided in cache")
	}
}
