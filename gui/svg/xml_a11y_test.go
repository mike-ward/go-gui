package svg

import (
	"strings"
	"testing"
)

func TestParseRootA11y(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		svg  string
		want struct {
			title, desc, ariaLabel, ariaRoleDesc string
			ariaHidden                           bool
		}
	}{
		{
			name: "empty",
			svg:  `<svg xmlns="http://www.w3.org/2000/svg"></svg>`,
		},
		{
			name: "title and desc",
			svg: `<svg xmlns="http://www.w3.org/2000/svg">
				<title>Logo</title>
				<desc>Company brand mark</desc>
			</svg>`,
			want: struct {
				title, desc, ariaLabel, ariaRoleDesc string
				ariaHidden                           bool
			}{title: "Logo", desc: "Company brand mark"},
		},
		{
			name: "aria attrs",
			svg: `<svg xmlns="http://www.w3.org/2000/svg"
				aria-label="Close" aria-roledescription="button"
				aria-hidden="true"></svg>`,
			want: struct {
				title, desc, ariaLabel, ariaRoleDesc string
				ariaHidden                           bool
			}{ariaLabel: "Close", ariaRoleDesc: "button", ariaHidden: true},
		},
		{
			name: "first title wins; nested ignored",
			svg: `<svg xmlns="http://www.w3.org/2000/svg">
				<title>First</title>
				<title>Second</title>
				<g><title>Tooltip</title></g>
			</svg>`,
			want: struct {
				title, desc, ariaLabel, ariaRoleDesc string
				ariaHidden                           bool
			}{title: "First"},
		},
		{
			name: "aria-hidden false",
			svg: `<svg xmlns="http://www.w3.org/2000/svg"
				aria-hidden="false"></svg>`,
		},
		{
			name: "whitespace trimmed",
			svg: `<svg xmlns="http://www.w3.org/2000/svg">
				<title>  spaced  </title>
			</svg>`,
			want: struct {
				title, desc, ariaLabel, ariaRoleDesc string
				ariaHidden                           bool
			}{title: "spaced"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vg, err := parseSvg(tt.svg)
			if err != nil {
				t.Fatalf("parseSvg: %v", err)
			}
			if vg.A11y.Title != tt.want.title {
				t.Errorf("Title = %q, want %q", vg.A11y.Title, tt.want.title)
			}
			if vg.A11y.Desc != tt.want.desc {
				t.Errorf("Desc = %q, want %q", vg.A11y.Desc, tt.want.desc)
			}
			if vg.A11y.AriaLabel != tt.want.ariaLabel {
				t.Errorf("AriaLabel = %q, want %q",
					vg.A11y.AriaLabel, tt.want.ariaLabel)
			}
			if vg.A11y.AriaRoleDesc != tt.want.ariaRoleDesc {
				t.Errorf("AriaRoleDesc = %q, want %q",
					vg.A11y.AriaRoleDesc, tt.want.ariaRoleDesc)
			}
			if vg.A11y.AriaHidden != tt.want.ariaHidden {
				t.Errorf("AriaHidden = %v, want %v",
					vg.A11y.AriaHidden, tt.want.ariaHidden)
			}
		})
	}
}

func TestParseRootA11yNilRoot(t *testing.T) {
	t.Parallel()
	a := parseRootA11y(nil)
	if a != (struct {
		Title, Desc, AriaLabel, AriaRoleDesc string
		AriaHidden                           bool
	}{}) {
		// Compare via field-wise checks since SvgA11y is from gui pkg.
		if a.Title != "" || a.Desc != "" || a.AriaLabel != "" ||
			a.AriaRoleDesc != "" || a.AriaHidden {
			t.Errorf("nil root: got %+v, want zero value", a)
		}
	}
}

func TestParseRootA11yClampsLongFields(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("x", maxA11yFieldLen+500)
	src := `<svg xmlns="http://www.w3.org/2000/svg"
		aria-label="` + long + `" aria-roledescription="` + long + `"
		aria-hidden="` + strings.Repeat("Y", maxA11yFieldLen+10) + `">
		<title>` + long + `</title>
		<desc>` + long + `</desc>
	</svg>`
	vg, err := parseSvg(src)
	if err != nil {
		t.Fatalf("parseSvg: %v", err)
	}
	if got := len(vg.A11y.Title); got > maxA11yFieldLen {
		t.Errorf("Title len = %d, exceeds cap %d", got, maxA11yFieldLen)
	}
	if got := len(vg.A11y.Desc); got > maxA11yFieldLen {
		t.Errorf("Desc len = %d, exceeds cap %d", got, maxA11yFieldLen)
	}
	if got := len(vg.A11y.AriaLabel); got > maxA11yFieldLen {
		t.Errorf("AriaLabel len = %d, exceeds cap %d",
			got, maxA11yFieldLen)
	}
	if got := len(vg.A11y.AriaRoleDesc); got > maxA11yFieldLen {
		t.Errorf("AriaRoleDesc len = %d, exceeds cap %d",
			got, maxA11yFieldLen)
	}
	// Long aria-hidden value with non-true content stays false.
	if vg.A11y.AriaHidden {
		t.Error("AriaHidden = true, expected false for non-'true' bulk")
	}
}
