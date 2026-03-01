package gui

import "testing"

func TestContainerIDScrollAutoAppendsScrollbars(t *testing.T) {
	v := Column(ContainerCfg{
		IDScroll: 1,
		Content:  []View{Rectangle(RectangleCfg{})},
	})
	cv := v.(*containerView)
	// 1 user child + 2 scrollbars = 3
	if len(cv.content) != 3 {
		t.Errorf("expected 3 children, got %d", len(cv.content))
	}
}

func TestContainerScrollbarCfgXHiddenSuppresses(t *testing.T) {
	v := Column(ContainerCfg{
		IDScroll:      2,
		ScrollbarCfgX: &ScrollbarCfg{Overflow: ScrollbarHidden},
		Content:       []View{Rectangle(RectangleCfg{})},
	})
	cv := v.(*containerView)
	// 1 user child + 0 horizontal + 1 vertical = 2
	if len(cv.content) != 2 {
		t.Errorf("expected 2 children, got %d", len(cv.content))
	}
}

func TestContainerScrollbarCfgYOverrides(t *testing.T) {
	customThumb := RGB(255, 0, 0)
	v := Column(ContainerCfg{
		IDScroll:      3,
		ScrollbarCfgY: &ScrollbarCfg{ColorThumb: customThumb},
		Content:       []View{Rectangle(RectangleCfg{})},
	})
	cv := v.(*containerView)
	// 1 user child + 1 horizontal (default) + 1 vertical (custom) = 3
	if len(cv.content) != 3 {
		t.Errorf("expected 3 children, got %d", len(cv.content))
	}
}

func TestContainerNoIDScrollNoScrollbars(t *testing.T) {
	v := Column(ContainerCfg{
		Content: []View{Rectangle(RectangleCfg{})},
	})
	cv := v.(*containerView)
	if len(cv.content) != 1 {
		t.Errorf("expected 1 child, got %d", len(cv.content))
	}
}
