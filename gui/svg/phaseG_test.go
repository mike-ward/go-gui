package svg

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Phase G — golden fingerprint over 12 CSS-only spinner fixtures plus
// 3 mixed SMIL+CSS fixtures. Locks the full pipeline (parse → cascade
// → @keyframes compile → SMIL parse → tessellate → animated re-tess)
// against fixtures that exercise every Phase B-F surface: selector
// kinds, specificity + !important, custom properties, color tween,
// alternate / fill-mode, cubic-bezier / steps, compound transforms,
// multiple animations, presentation-attr cascade, display:none /
// visibility:hidden, transform-origin defaults.
//
// Run `go test ./gui/svg/ -run TestPhaseG -phaseG-update` to regen the
// goldens after a deliberate change.

var phaseGUpdate = flag.Bool("phaseG-update", false,
	"regenerate phase G CSS spinner goldens")

// phaseGSpinners is the canonical fixture set. Order matters — the
// goldens file is written in this order.
var phaseGSpinners = []string{
	"css-rotate.svg",
	"css-pulse-opacity.svg",
	"css-color-tween.svg",
	"css-alternate.svg",
	"css-fill-modes.svg",
	"css-cubic-bezier.svg",
	"css-steps.svg",
	"css-compound-transform.svg",
	"css-multi-anim.svg",
	"css-custom-prop.svg",
	"css-iteration-count.svg",
	"css-id-class-cascade.svg",
	"mixed-smil-css.svg",
	"mixed-presentation-cascade.svg",
	"mixed-display-none.svg",
}

func TestPhaseGCssSpinnerFingerprint(t *testing.T) {
	got := buildPhaseGFingerprints(t)
	goldenPath := filepath.Join("testdata", "phaseG_css_goldens.txt")
	if *phaseGUpdate {
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write goldens: %v", err)
		}
		t.Logf("wrote %s (%d bytes)", goldenPath, len(got))
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read goldens: %v (run with -phaseG-update to seed)", err)
	}
	if string(want) != got {
		t.Fatalf("phase G CSS fingerprint drift:\n--- want\n%s\n--- got\n%s\n"+
			"re-run with -phaseG-update if the change is intentional",
			string(want), got)
	}
}

// buildPhaseGFingerprints emits per-fixture lines covering geometry,
// static tessellation, and the parsed animation list (sorted to be
// order-stable). CSS animations are transform / opacity / color
// kinds, none of which flip VectorPath.Animated, so re-tess via
// TessellateAnimated is not exercised here — phase 0 already pins
// that path for SMIL primitive-attr animations. Fingerprint
// sensitivity matches phase 0 helpers (hashTessellated /
// hashAnimations) so a regression diff highlights what shifted.
func buildPhaseGFingerprints(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("testdata", "css-spinners")
	var out strings.Builder
	for _, name := range phaseGSpinners {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		p := New()
		parsed, err := p.ParseSvg(string(data))
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		fmt.Fprintf(&out, "%s vg     w=%g h=%g vb=%g,%g paths=%d anims=%d\n",
			name, parsed.Width, parsed.Height,
			parsed.ViewBoxX, parsed.ViewBoxY,
			len(parsed.Paths), len(parsed.Animations))

		staticTris := p.Tessellate(parsed, 1)
		fmt.Fprintf(&out, "%s static tris=%d hash=%s\n",
			name, len(staticTris), hashTessellated(staticTris))

		fmt.Fprintf(&out, "%s anims  hash=%s\n",
			name, hashAnimations(parsed.Animations))
	}
	return out.String()
}
