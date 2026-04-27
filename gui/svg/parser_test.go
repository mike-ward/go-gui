package svg

import (
	"testing"

	"github.com/mike-ward/go-gui/gui"
)

func TestMixOptsHashFieldOrderMatters(t *testing.T) {
	base := uint64(0xcbf29ce484222325)

	hAOnly := mixOptsHash(base,
		gui.SvgParseOpts{HoveredElementID: "a"})
	fAOnly := mixOptsHash(base,
		gui.SvgParseOpts{FocusedElementID: "a"})
	if hAOnly == fAOnly {
		t.Errorf("hovered=a and focused=a must hash differently")
	}

	flat := mixOptsHash(base,
		gui.SvgParseOpts{FlatnessTolerance: 0.5})
	if flat == base {
		t.Errorf("FlatnessTolerance must perturb the hash")
	}

	rm := mixOptsHash(base,
		gui.SvgParseOpts{PrefersReducedMotion: true})
	none := mixOptsHash(base, gui.SvgParseOpts{})
	if rm == none {
		t.Errorf("PrefersReducedMotion must perturb the hash")
	}
}

func TestMixOptsHashEmptyIsStable(t *testing.T) {
	base := uint64(42)
	a := mixOptsHash(base, gui.SvgParseOpts{})
	b := mixOptsHash(base, gui.SvgParseOpts{})
	if a != b {
		t.Errorf("empty opts must hash deterministically: %x vs %x", a, b)
	}
}
