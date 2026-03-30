package svg

import "testing"

func FuzzParseSvg(f *testing.F) {
	f.Add(`<svg viewBox="0 0 24 24"><rect x="0" y="0" width="24" height="24"/></svg>`)
	f.Add(`<svg><circle cx="50" cy="50" r="40"/></svg>`)
	f.Add(`<svg><path d="M10 80 Q 95 10 180 80"/></svg>`)
	f.Add(`<svg><g transform="rotate(45)"><rect width="10" height="10"/></g></svg>`)
	f.Add(`<svg><defs><linearGradient id="g"><stop offset="0%" stop-color="red"/></linearGradient></defs></svg>`)
	f.Add(`<svg><filter id="f"><feGaussianBlur stdDeviation="5"/></filter></svg>`)
	f.Add(`<svg><text x="10" y="20">Hello</text></svg>`)
	f.Add(`<svg><textPath href="#p">Along path</textPath></svg>`)
	f.Add("")
	f.Add("<not-svg>")
	f.Fuzz(func(t *testing.T, data string) {
		_, _ = parseSvg(data)
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
