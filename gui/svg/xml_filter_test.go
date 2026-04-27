package svg

import "testing"

// Two non-contiguous elements that share the same filter must produce
// two distinct svgFilteredGroups in document order — not be merged into
// one offscreen buffer (which would composite them at the wrong z) and
// not be ordered by map iteration (nondeterministic across runs).
func TestParseSvg_SameFilterOnTwoElementsYieldsTwoGroups(t *testing.T) {
	vg, err := parseSvg(`<svg viewBox="0 0 100 100">
		<defs>
			<filter id="b"><feGaussianBlur stdDeviation="1"/></filter>
		</defs>
		<rect id="a" x="0" y="0" width="10" height="10" filter="url(#b)"/>
		<rect id="mid" x="20" y="0" width="10" height="10"/>
		<rect id="c" x="40" y="0" width="10" height="10" filter="url(#b)"/>
	</svg>`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := len(vg.FilteredGroups); got != 2 {
		t.Fatalf("FilteredGroups=%d; want 2 (one per filter occurrence)", got)
	}
	for i, g := range vg.FilteredGroups {
		if g.FilterID != "b" {
			t.Errorf("group %d FilterID=%q; want %q", i, g.FilterID, "b")
		}
		if len(g.Paths) != 1 {
			t.Errorf("group %d paths=%d; want 1", i, len(g.Paths))
		}
	}
	if vg.FilteredGroups[0].GroupKey == vg.FilteredGroups[1].GroupKey {
		t.Fatalf("group keys collide: %d", vg.FilteredGroups[0].GroupKey)
	}
	if vg.FilteredGroups[0].GroupKey >= vg.FilteredGroups[1].GroupKey {
		t.Fatalf("groups out of document order: keys %d,%d",
			vg.FilteredGroups[0].GroupKey, vg.FilteredGroups[1].GroupKey)
	}
}
