package svg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// knownEmptyTessellation lists spinner assets whose geometry is
// gated on SVG features the renderer does not implement (currently:
// SVG <filter> primitives needed for the gooey blob composite).
// The smoke test still parses these — it just tolerates a zero-tris
// result rather than failing.
var knownEmptyTessellation = map[string]bool{
	"4-dots-goeey.svg":  true,
	"gooey-balls-1.svg": true,
	"gooey-balls-2.svg": true,
}

// TestSpinnerAssetsSmokeParse ensures every embedded spinner asset
// parses without error and produces non-empty geometry. New assets
// added under gui/assets/svg-spinners are picked up automatically;
// regressions in CSS/SMIL coverage surface as a parse error or a
// zero-path output for an asset that previously rendered.
func TestSpinnerAssetsSmokeParse(t *testing.T) {
	dir := filepath.Join("..", "assets", "svg-spinners")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("spinner asset dir unavailable: %v", err)
	}
	p := New()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".svg") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			parsed, err := p.ParseSvg(string(data))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if knownEmptyTessellation[name] {
				return
			}
			if len(parsed.Paths) == 0 {
				t.Fatalf("zero paths after parse")
			}
			if tris := p.Tessellate(parsed, 1); len(tris) == 0 {
				t.Fatalf("zero tris after tessellate")
			}
		})
	}
}
