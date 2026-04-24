package svg

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// seedFuzzFromSpinners adds every .svg under gui/assets/svg-spinners
// as a seed corpus entry. Fuzz mutator then derives hostile inputs.
func seedFuzzFromSpinners(f *testing.F) {
	f.Helper()
	dir := filepath.Join("..", "assets", "svg-spinners")
	entries, err := os.ReadDir(dir)
	if err != nil {
		f.Logf("seed dir unreadable: %v", err)
		return
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".svg" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		f.Add(string(b))
	}
}

// FuzzParseSvg pushes arbitrary byte sequences through parseSvg and
// asserts no panic plus finite viewBox / dims / path transforms.
// Hostile input must not propagate NaN/Inf into downstream
// tessellation.
func FuzzParseSvg(f *testing.F) {
	seedFuzzFromSpinners(f)
	f.Add(`<svg viewBox="0 0 24 24"><rect x="0" y="0" width="24" height="24"/></svg>`)
	f.Add(`<svg><circle cx="50" cy="50" r="40"/></svg>`)
	f.Add(`<svg><path d="M10 80 Q 95 10 180 80"/></svg>`)
	f.Add(`<svg><g transform="rotate(45)"><rect width="10" height="10"/></g></svg>`)
	f.Add(`<svg><defs><linearGradient id="g"><stop offset="0%" stop-color="red"/></linearGradient></defs></svg>`)
	f.Add(`<svg><filter id="f"><feGaussianBlur stdDeviation="5"/></filter></svg>`)
	f.Add(`<svg><text x="10" y="20">Hello</text></svg>`)
	f.Add(`<svg><textPath href="#p">Along path</textPath></svg>`)
	f.Add(`<svg viewBox="NaN Inf 10 10"><rect width="-5" height="1e40"/></svg>`)
	f.Add(`<svg><circle r="NaN" cx="Inf" cy="-Inf"/></svg>`)
	f.Add(`<svg viewBox="0 0 1e38 1e38"><path d="M 0 0 L 1e38 1e38 Z"/></svg>`)
	f.Add(`<svg><g transform="matrix(NaN 0 0 NaN 0 0)"><rect/></g></svg>`)
	f.Add("")
	f.Add("<not-svg>")

	f.Fuzz(func(t *testing.T, data string) {
		vg, err := parseSvg(data)
		if err != nil || vg == nil {
			return
		}
		if !finiteF32(vg.Width) || !finiteF32(vg.Height) ||
			!finiteF32(vg.ViewBoxX) || !finiteF32(vg.ViewBoxY) {
			t.Fatalf("non-finite dims: w=%v h=%v vbx=%v vby=%v",
				vg.Width, vg.Height, vg.ViewBoxX, vg.ViewBoxY)
		}
		for i := range vg.Paths {
			for k, v := range vg.Paths[i].Transform {
				if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
					t.Fatalf("path %d transform[%d]=%v not finite",
						i, k, v)
				}
			}
		}
	})
}

func FuzzParseSvgDimensions(f *testing.F) {
	f.Add(`<svg viewBox="0 0 100 100"></svg>`)
	f.Add(`<svg width="24" height="24"></svg>`)
	f.Add(`<svg viewBox="0 0 0 0"></svg>`)
	f.Add("")
	f.Fuzz(func(t *testing.T, data string) {
		w, h := parseSvgDimensions(data)
		if w < 0 {
			t.Errorf("negative width: %f", w)
		}
		if h < 0 {
			t.Errorf("negative height: %f", h)
		}
	})
}
