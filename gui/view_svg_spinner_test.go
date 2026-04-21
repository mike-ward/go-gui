package gui

import (
	"strings"
	"testing"
)

func TestSvgSpinnerFactory(t *testing.T) {
	v := SvgSpinner(SvgSpinnerCfg{
		ID:   "s1",
		Kind: SvgSpinner90Ring,
	})
	if v == nil {
		t.Fatal("SvgSpinner returned nil")
	}
	if _, ok := v.(*svgView); !ok {
		t.Fatalf("expected *svgView, got %T", v)
	}
}

func TestSvgSpinnerCountMatchesData(t *testing.T) {
	if SvgSpinnerCount() == 0 {
		t.Fatal("expected at least one embedded spinner")
	}
	if SvgSpinnerCount() != len(svgSpinnerData) {
		t.Fatalf("count %d != data len %d",
			SvgSpinnerCount(), len(svgSpinnerData))
	}
	if SvgSpinnerCount() != len(svgSpinnerName) {
		t.Fatalf("count %d != name len %d",
			SvgSpinnerCount(), len(svgSpinnerName))
	}
}

func TestSvgSpinnerAllAssetsNonEmpty(t *testing.T) {
	for i, data := range svgSpinnerData {
		if !strings.Contains(data, "<svg") {
			t.Errorf("kind %d (%s): missing <svg", i, svgSpinnerName[i])
		}
	}
}

func TestSvgSpinnerNameLookup(t *testing.T) {
	if got := SvgSpinnerName(SvgSpinner90Ring); got != "90-ring" {
		t.Fatalf("expected '90-ring', got %q", got)
	}
	if got := SvgSpinnerName(SvgSpinnerKind(svgSpinnerCount)); got != "" {
		t.Fatalf("expected empty for out-of-range, got %q", got)
	}
}

func TestSvgSpinnerOutOfRangeKindFallsBack(t *testing.T) {
	v := SvgSpinner(SvgSpinnerCfg{Kind: SvgSpinnerKind(svgSpinnerCount + 10)})
	sv, ok := v.(*svgView)
	if !ok {
		t.Fatalf("expected *svgView, got %T", v)
	}
	if sv.cfg.SvgData != svgSpinnerData[0] {
		t.Fatal("expected out-of-range kind to fall back to kind 0")
	}
}

func TestSvgSpinner90RingAssetShapeValid(t *testing.T) {
	data := svgSpinnerData[SvgSpinner90Ring]
	if !strings.Contains(data, `type="rotate"`) {
		t.Fatal("expected rotate animateTransform in 90-ring asset")
	}
	if !strings.Contains(data, `values=`) {
		t.Fatal("expected values= form in 90-ring asset")
	}
}
