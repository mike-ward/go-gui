package gui

import "fmt"

// TabItemCfg configures one tab in a TabControl.
type TabItemCfg struct {
	ID       string
	Label    string
	Content  []View
	Disabled bool
}

// NewTabItem creates a TabItemCfg.
func NewTabItem(id, label string, content []View) TabItemCfg {
	return TabItemCfg{ID: id, Label: label, Content: content}
}

// TabControlCfg configures a tab control. Controlled component:
// Selected is owned by app state and updated through OnSelect.
type TabControlCfg struct {
	ID                  string
	Items               []TabItemCfg
	Selected            string
	OnSelect            func(string, *Event, *Window)
	Sizing              Sizing
	Color               Color
	ColorBorder         Color
	ColorHeader         Color
	ColorHeaderBorder   Color
	ColorContent        Color
	ColorContentBorder  Color
	ColorTab            Color
	ColorTabHover       Color
	ColorTabFocus       Color
	ColorTabClick       Color
	ColorTabSelected    Color
	ColorTabDisabled    Color
	ColorTabBorder      Color
	ColorTabBorderFocus Color
	Padding             Padding
	PaddingHeader       Padding
	PaddingContent      Padding
	PaddingTab          Padding
	SizeBorder          float32
	SizeHeaderBorder    float32
	SizeContentBorder   float32
	SizeTabBorder       float32
	Radius              float32
	RadiusHeader        float32
	RadiusContent       float32
	RadiusTab           float32
	RadiusTabBorder     float32
	Spacing             float32
	SpacingHeader       float32
	TextStyle           TextStyle
	TextStyleSelected   TextStyle
	TextStyleDisabled   TextStyle
	IDFocus             uint32
	Disabled            bool
	Invisible           bool

	A11YLabel       string
	A11YDescription string
}

func applyTabControlDefaults(cfg *TabControlCfg) {
	s := &DefaultTabControlStyle
	if cfg.Sizing == (Sizing{}) {
		cfg.Sizing = FillFill
	}
	if !cfg.Color.IsSet() {
		cfg.Color = s.Color
	}
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = s.ColorBorder
	}
	if !cfg.ColorHeader.IsSet() {
		cfg.ColorHeader = s.ColorHeader
	}
	if !cfg.ColorHeaderBorder.IsSet() {
		cfg.ColorHeaderBorder = s.ColorHeaderBorder
	}
	if !cfg.ColorContent.IsSet() {
		cfg.ColorContent = s.ColorContent
	}
	if !cfg.ColorContentBorder.IsSet() {
		cfg.ColorContentBorder = s.ColorContentBorder
	}
	if !cfg.ColorTab.IsSet() {
		cfg.ColorTab = s.ColorTab
	}
	if !cfg.ColorTabHover.IsSet() {
		cfg.ColorTabHover = s.ColorTabHover
	}
	if !cfg.ColorTabFocus.IsSet() {
		cfg.ColorTabFocus = s.ColorTabFocus
	}
	if !cfg.ColorTabClick.IsSet() {
		cfg.ColorTabClick = s.ColorTabClick
	}
	if !cfg.ColorTabSelected.IsSet() {
		cfg.ColorTabSelected = s.ColorTabSelected
	}
	if !cfg.ColorTabDisabled.IsSet() {
		cfg.ColorTabDisabled = s.ColorTabDisabled
	}
	if !cfg.ColorTabBorder.IsSet() {
		cfg.ColorTabBorder = s.ColorTabBorder
	}
	if !cfg.ColorTabBorderFocus.IsSet() {
		cfg.ColorTabBorderFocus = s.ColorTabBorderFocus
	}
	if cfg.Padding == (Padding{}) {
		cfg.Padding = s.Padding
	}
	if cfg.PaddingHeader == (Padding{}) {
		cfg.PaddingHeader = s.PaddingHeader
	}
	if cfg.PaddingContent == (Padding{}) {
		cfg.PaddingContent = s.PaddingContent
	}
	if cfg.PaddingTab == (Padding{}) {
		cfg.PaddingTab = s.PaddingTab
	}
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = s.SizeBorder
	}
	if cfg.SizeContentBorder == 0 {
		cfg.SizeContentBorder = s.SizeContentBorder
	}
	if cfg.SizeTabBorder == 0 {
		cfg.SizeTabBorder = s.SizeTabBorder
	}
	if cfg.Radius == 0 {
		cfg.Radius = s.Radius
	}
	if cfg.RadiusHeader == 0 {
		cfg.RadiusHeader = s.RadiusHeader
	}
	if cfg.RadiusContent == 0 {
		cfg.RadiusContent = s.RadiusContent
	}
	if cfg.RadiusTab == 0 {
		cfg.RadiusTab = s.RadiusTab
	}
	if cfg.RadiusTabBorder == 0 {
		cfg.RadiusTabBorder = s.RadiusTabBorder
	}
	if cfg.Spacing == 0 {
		cfg.Spacing = s.Spacing
	}
	if cfg.SpacingHeader == 0 {
		cfg.SpacingHeader = s.SpacingHeader
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = s.TextStyle
	}
	if cfg.TextStyleSelected == (TextStyle{}) {
		cfg.TextStyleSelected = s.TextStyleSelected
	}
	if cfg.TextStyleDisabled == (TextStyle{}) {
		cfg.TextStyleDisabled = s.TextStyleDisabled
	}
}

// Tabs is an alias for TabControl.
func Tabs(cfg TabControlCfg) View {
	return TabControl(cfg)
}

// TabControl creates a tab control with header row and content.
func TabControl(cfg TabControlCfg) View {
	applyTabControlDefaults(&cfg)

	tabNavIDs := make([]string, len(cfg.Items))
	tabNavDisabled := make([]bool, len(cfg.Items))
	for i, item := range cfg.Items {
		tabNavIDs[i] = item.ID
		tabNavDisabled[i] = item.Disabled
	}

	selectedIdx := tabSelectedIndex(tabNavIDs, tabNavDisabled, cfg.Selected)

	headerItems := make([]View, 0, len(cfg.Items))
	for i, item := range cfg.Items {
		isSelected := i == selectedIdx
		isDisabled := cfg.Disabled || item.Disabled

		tabColor := cfg.ColorTab
		hoverColor := cfg.ColorTabHover
		focusColor := cfg.ColorTabFocus
		clickColor := cfg.ColorTabClick
		borderColor := cfg.ColorTabBorder

		if isDisabled {
			tabColor = cfg.ColorTabDisabled
			hoverColor = cfg.ColorTabDisabled
			focusColor = cfg.ColorTabDisabled
			clickColor = cfg.ColorTabDisabled
		} else if isSelected {
			tabColor = cfg.ColorTabSelected
			hoverColor = cfg.ColorTabSelected
			focusColor = cfg.ColorTabSelected
			clickColor = cfg.ColorTabSelected
			borderColor = cfg.ColorTabBorderFocus
		}

		ts := cfg.TextStyle
		if isDisabled {
			ts = cfg.TextStyleDisabled
		} else if isSelected {
			ts = cfg.TextStyleSelected
		}

		a11yState := AccessStateNone
		if isSelected {
			a11yState = AccessStateSelected
		}

		var onClick func(*Layout, *Event, *Window)
		if !isDisabled {
			onClick = makeTabOnClick(cfg.OnSelect, item.ID, cfg.IDFocus)
		}

		headerItems = append(headerItems, Button(ButtonCfg{
			ID:               tabButtonID(cfg.ID, item.ID),
			A11YRole:         AccessRoleTabItem,
			A11YState:        a11yState,
			A11YLabel:        item.Label,
			Color:            tabColor,
			ColorHover:       hoverColor,
			ColorFocus:       focusColor,
			ColorClick:       clickColor,
			ColorBorder:      borderColor,
			ColorBorderFocus: cfg.ColorTabBorderFocus,
			Padding:          cfg.PaddingTab,
			SizeBorder:       Some(cfg.SizeTabBorder),
			Radius:           Some(cfg.RadiusTab),
			Disabled:         isDisabled,
			OnClick:          onClick,
			Content: []View{
				Text(TextCfg{Text: item.Label, TextStyle: ts}),
			},
		}))
	}

	var activeContent []View
	if selectedIdx >= 0 && selectedIdx < len(cfg.Items) {
		src := cfg.Items[selectedIdx].Content
		activeContent = make([]View, len(src))
		copy(activeContent, src)
	}

	// Closure captures.
	disabled := cfg.Disabled
	selected := cfg.Selected
	onSelect := cfg.OnSelect
	idFocus := cfg.IDFocus

	return Column(ContainerCfg{
		ID:              cfg.ID,
		IDFocus:         cfg.IDFocus,
		A11YRole:        AccessRoleTab,
		A11YLabel:       a11yLabel(cfg.A11YLabel, cfg.ID),
		A11YDescription: cfg.A11YDescription,
		Sizing:          cfg.Sizing,
		Color:           cfg.Color,
		ColorBorder:     cfg.ColorBorder,
		SizeBorder:      Some(cfg.SizeBorder),
		Radius:          Some(cfg.Radius),
		Padding:         Some(cfg.Padding),
		Spacing:         Some(cfg.Spacing),
		Disabled:        cfg.Disabled,
		Invisible:       cfg.Invisible,
		OnKeyDown: func(_ *Layout, e *Event, w *Window) {
			tabControlOnKeydown(disabled, tabNavIDs, tabNavDisabled,
				selected, onSelect, idFocus, e, w)
		},
		Content: []View{
			Row(ContainerCfg{
				Color:       cfg.ColorHeader,
				ColorBorder: cfg.ColorHeaderBorder,
				SizeBorder:  Some(cfg.SizeHeaderBorder),
				Radius:      Some(cfg.RadiusHeader),
				Padding:     Some(cfg.PaddingHeader),
				Spacing:     Some(cfg.SpacingHeader),
				Sizing:      FillFit,
				Content:     headerItems,
			}),
			Column(ContainerCfg{
				Color:       cfg.ColorContent,
				ColorBorder: cfg.ColorContentBorder,
				SizeBorder:  Some(cfg.SizeContentBorder),
				Radius:      Some(cfg.RadiusContent),
				Padding:     Some(cfg.PaddingContent),
				Sizing:      FillFill,
				Content:     activeContent,
			}),
		},
	})
}

func makeTabOnClick(
	onSelect func(string, *Event, *Window),
	id string, idFocus uint32,
) func(*Layout, *Event, *Window) {
	return func(_ *Layout, e *Event, w *Window) {
		if onSelect != nil {
			onSelect(id, e, w)
		}
		if idFocus > 0 {
			w.SetIDFocus(idFocus)
		}
		e.IsHandled = true
	}
}

func tabControlOnKeydown(
	disabled bool,
	tabNavIDs []string,
	tabNavDisabled []bool,
	selected string,
	onSelect func(string, *Event, *Window),
	idFocus uint32,
	e *Event,
	w *Window,
) {
	if disabled || len(tabNavIDs) == 0 || e.Modifiers != ModNone {
		return
	}

	selectedIdx := tabSelectedIndex(tabNavIDs, tabNavDisabled, selected)
	targetIdx := -1

	switch e.KeyCode {
	case KeyLeft, KeyUp:
		if selectedIdx >= 0 {
			targetIdx = tabPrevEnabledIndex(tabNavDisabled, selectedIdx)
		} else {
			targetIdx = tabLastEnabledIndex(tabNavDisabled)
		}
	case KeyRight, KeyDown:
		if selectedIdx >= 0 {
			targetIdx = tabNextEnabledIndex(tabNavDisabled, selectedIdx)
		} else {
			targetIdx = tabFirstEnabledIndex(tabNavDisabled)
		}
	case KeyHome:
		targetIdx = tabFirstEnabledIndex(tabNavDisabled)
	case KeyEnd:
		targetIdx = tabLastEnabledIndex(tabNavDisabled)
	case KeyEnter:
		if selectedIdx >= 0 {
			targetIdx = selectedIdx
		} else {
			targetIdx = tabFirstEnabledIndex(tabNavDisabled)
		}
	default:
		if e.CharCode == CharSpace {
			if selectedIdx >= 0 {
				targetIdx = selectedIdx
			} else {
				targetIdx = tabFirstEnabledIndex(tabNavDisabled)
			}
		} else {
			return
		}
	}

	if targetIdx < 0 || targetIdx >= len(tabNavIDs) {
		return
	}
	targetID := tabNavIDs[targetIdx]
	if len(targetID) == 0 {
		return
	}

	refire := e.KeyCode == KeyEnter || e.CharCode == CharSpace
	if targetID != selected || refire {
		if onSelect != nil {
			onSelect(targetID, e, w)
		}
	}
	if idFocus > 0 {
		w.SetIDFocus(idFocus)
	}
	e.IsHandled = true
}

func tabSelectedIndex(ids []string, disabled []bool, selected string) int {
	if len(selected) > 0 {
		for i, id := range ids {
			if id == selected && !disabled[i] {
				return i
			}
		}
	}
	return tabFirstEnabledIndex(disabled)
}

func tabFirstEnabledIndex(disabled []bool) int {
	for i, d := range disabled {
		if !d {
			return i
		}
	}
	return -1
}

func tabLastEnabledIndex(disabled []bool) int {
	for i := len(disabled) - 1; i >= 0; i-- {
		if !disabled[i] {
			return i
		}
	}
	return -1
}

func tabNextEnabledIndex(disabled []bool, selectedIdx int) int {
	n := len(disabled)
	if n == 0 {
		return -1
	}
	idx := selectedIdx
	if idx < 0 || idx >= n {
		idx = -1
	}
	for range n {
		idx = (idx + 1 + n) % n
		if !disabled[idx] {
			return idx
		}
	}
	return -1
}

func tabPrevEnabledIndex(disabled []bool, selectedIdx int) int {
	n := len(disabled)
	if n == 0 {
		return -1
	}
	idx := selectedIdx
	if idx < 0 || idx >= n {
		idx = 0
	}
	for range n {
		idx = (idx - 1 + n) % n
		if !disabled[idx] {
			return idx
		}
	}
	return -1
}

func tabButtonID(controlID, tabID string) string {
	return fmt.Sprintf("tc_%s_%s", controlID, tabID)
}
