package gui

type listBoxView struct {
	cfg ListBoxCfg
}

type listBoxCache struct {
	dataHash uint64
	itemIDs  []string
}

// ListBoxOption represents one row in a ListBox.
type ListBoxOption struct {
	ID           string
	Name         string
	Value        string
	IsSubheading bool
}

// ListBoxCfg configures a list box view.
type ListBoxCfg struct {
	ID              string
	Sizing          Sizing
	TextStyle       TextStyle
	SubheadingStyle TextStyle
	Color           Color
	ColorHover      Color
	ColorBorder     Color
	ColorSelect     Color
	Padding         Padding
	SelectedIDs     []string
	Data            []ListBoxOption
	OnSelect        func(ids []string, e *Event, w *Window)
	Height          float32
	MinWidth        float32
	MaxWidth        float32
	MinHeight       float32
	MaxHeight       float32
	Radius          Opt[float32]
	SizeBorder      Opt[float32]
	IDScroll        uint32
	IDFocus         uint32
	Multiple        bool
	Disabled        bool
	Invisible       bool

	A11YLabel       string
	A11YDescription string
}

// ListBoxOption helpers.

// NewListBoxOption constructs a ListBoxOption.
func NewListBoxOption(id, name, value string) ListBoxOption {
	return ListBoxOption{ID: id, Name: name, Value: value}
}

// NewListBoxSubheading constructs a subheading row.
func NewListBoxSubheading(id, title string) ListBoxOption {
	return ListBoxOption{ID: id, Name: title, IsSubheading: true}
}

// ListBox creates a list box view.
func ListBox(cfg ListBoxCfg) View {
	applyListBoxDefaults(&cfg)
	if listBoxCanVirtualize(&cfg) {
		return &listBoxView{cfg: cfg}
	}

	dn := &DefaultListBoxStyle
	sizeBorder := cfg.SizeBorder.Get(dn.SizeBorder)
	radius := cfg.Radius.Get(dn.Radius)

	selectedSet := listCoreSelectedSet(cfg.SelectedIDs)
	list := make([]View, 0, len(cfg.Data))
	for i := range cfg.Data {
		list = append(list, listBoxItemView(cfg.Data[i], cfg, selectedSet))
	}

	listBoxID := cfg.ID
	isMultiple := cfg.Multiple
	onSelect := cfg.OnSelect
	selectedIDs := cfg.SelectedIDs
	itemIDs := make([]string, 0, len(cfg.Data))
	for i := range cfg.Data {
		if !cfg.Data[i].IsSubheading {
			itemIDs = append(itemIDs, cfg.Data[i].ID)
		}
	}

	return Column(ContainerCfg{
		ID:        cfg.ID,
		A11YRole:  AccessRoleList,
		A11YLabel: a11yLabel(cfg.A11YLabel, cfg.ID),
		IDFocus:   cfg.IDFocus,
		IDScroll:  cfg.IDScroll,
		OnKeyDown: func(_ *Layout, e *Event, w *Window) {
			listBoxOnKeyDown(listBoxID, itemIDs,
				isMultiple, onSelect, selectedIDs, e, w)
		},
		Width:       cfg.MaxWidth,
		Height:      cfg.Height,
		MinWidth:    cfg.MinWidth,
		MaxWidth:    cfg.MaxWidth,
		MinHeight:   cfg.MinHeight,
		MaxHeight:   cfg.MaxHeight,
		Color:       cfg.Color,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  Some(sizeBorder),
		Radius:      Some(radius),
		Padding:     Some(cfg.Padding),
		Sizing:      cfg.Sizing,
		Spacing:     Some(float32(0)),
		Disabled:    cfg.Disabled,
		Invisible:   cfg.Invisible,
		Content:     list,
	})
}

func listBoxCanVirtualize(cfg *ListBoxCfg) bool {
	if cfg == nil || cfg.IDScroll == 0 {
		return false
	}
	return cfg.Height > 0 || cfg.MaxHeight > 0
}

func (lv *listBoxView) Content() []View { return nil }

func (lv *listBoxView) GenerateLayout(w *Window) Layout {
	cfg := &lv.cfg

	dn := &DefaultListBoxStyle
	sizeBorder := cfg.SizeBorder.Get(dn.SizeBorder)
	radius := cfg.Radius.Get(dn.Radius)

	cacheMap := StateMap[string, *listBoxCache](w, nsListBoxCache, capModerate)
	cache, ok := cacheMap.Get(cfg.ID)
	if !ok || cache == nil {
		cache = &listBoxCache{}
		cacheMap.Set(cfg.ID, cache)
	}
	dataHash := listBoxDataHash(cfg.Data)
	if cache.dataHash != dataHash || len(cache.itemIDs) == 0 {
		itemIDs := make([]string, 0, len(cfg.Data))
		for i := range cfg.Data {
			if !cfg.Data[i].IsSubheading {
				itemIDs = append(itemIDs, cfg.Data[i].ID)
			}
		}
		cache.itemIDs = itemIDs
		cache.dataHash = dataHash
	}

	selectedSet := listCoreSelectedSet(cfg.SelectedIDs)

	first, last := 0, len(cfg.Data)-1
	virtualize := cfg.IDScroll > 0
	listH := cfg.Height
	if listH <= 0 {
		listH = cfg.MaxHeight
	}
	if virtualize && listH > 0 && len(cfg.Data) > 0 {
		rowH := listCoreRowHeightEstimate(cfg.TextStyle, PaddingTwoFive)
		scrollY := StateReadOr[uint32, float32](w, nsScrollY, cfg.IDScroll, 0)
		first, last = listCoreVisibleRange(len(cfg.Data), rowH, listH, scrollY)
	} else {
		virtualize = false
	}

	listBoxID := cfg.ID
	isMultiple := cfg.Multiple
	onSelect := cfg.OnSelect
	selectedIDs := cfg.SelectedIDs
	itemIDs := cache.itemIDs

	listCap := len(cfg.Data)
	if virtualize && last >= first {
		listCap = last - first + 3
	}
	list := make([]View, 0, listCap)

	if virtualize && first > 0 {
		rowH := listCoreRowHeightEstimate(cfg.TextStyle, PaddingTwoFive)
		list = append(list, Rectangle(RectangleCfg{
			Color:  ColorTransparent,
			Height: float32(first) * rowH,
			Sizing: FillFixed,
		}))
	}

	for idx := first; idx <= last; idx++ {
		if idx < 0 || idx >= len(cfg.Data) {
			continue
		}
		list = append(list, listBoxItemView(cfg.Data[idx], *cfg, selectedSet))
	}

	if virtualize && last < len(cfg.Data)-1 {
		rowH := listCoreRowHeightEstimate(cfg.TextStyle, PaddingTwoFive)
		remaining := len(cfg.Data) - 1 - last
		list = append(list, Rectangle(RectangleCfg{
			Color:  ColorTransparent,
			Height: float32(remaining) * rowH,
			Sizing: FillFixed,
		}))
	}

	return GenerateViewLayout(Column(ContainerCfg{
		ID:        cfg.ID,
		A11YRole:  AccessRoleList,
		A11YLabel: a11yLabel(cfg.A11YLabel, cfg.ID),
		IDFocus:   cfg.IDFocus,
		IDScroll:  cfg.IDScroll,
		OnKeyDown: func(_ *Layout, e *Event, w *Window) {
			listBoxOnKeyDown(listBoxID, itemIDs,
				isMultiple, onSelect, selectedIDs, e, w)
		},
		Width:       cfg.MaxWidth,
		Height:      cfg.Height,
		MinWidth:    cfg.MinWidth,
		MaxWidth:    cfg.MaxWidth,
		MinHeight:   cfg.MinHeight,
		MaxHeight:   cfg.MaxHeight,
		Color:       cfg.Color,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  Some(sizeBorder),
		Radius:      Some(radius),
		Padding:     Some(cfg.Padding),
		Sizing:      cfg.Sizing,
		Spacing:     Some(float32(0)),
		Disabled:    cfg.Disabled,
		Invisible:   cfg.Invisible,
		Content:     list,
	}), w)
}

func listBoxItemView(dat ListBoxOption, cfg ListBoxCfg, selectedSet map[string]struct{}) View {
	color := ColorTransparent
	if listCoreContainsSelected(selectedSet, cfg.SelectedIDs, dat.ID) {
		color = cfg.ColorSelect
	}
	isSub := dat.IsSubheading
	content := listBoxItemContent(dat, cfg)

	datID := dat.ID
	isMultiple := cfg.Multiple
	onSelect := cfg.OnSelect
	hasOnSelect := onSelect != nil
	selectedIDs := cfg.SelectedIDs
	colorHover := cfg.ColorHover

	a11yState := AccessStateNone
	if listCoreContainsSelected(selectedSet, cfg.SelectedIDs, dat.ID) {
		a11yState = AccessStateSelected
	}

	return Row(ContainerCfg{
		A11YRole:  AccessRoleListItem,
		A11YLabel: dat.Name,
		A11YState: a11yState,
		Color:     color,
		Padding:   Some(PaddingTwoFive),
		Sizing:    FillFit,
		Content:   []View{content},
		OnClick: func(_ *Layout, e *Event, w *Window) {
			if hasOnSelect && !isSub {
				ids := listBoxNextSelectedIDs(
					selectedIDs, datID, isMultiple)
				onSelect(ids, e, w)
			}
		},
		OnHover: func(layout *Layout, _ *Event, w *Window) {
			if hasOnSelect && !isSub {
				w.SetMouseCursor(CursorPointingHand)
				if layout.Shape.Color == ColorTransparent {
					layout.Shape.Color = colorHover
				}
			}
		},
	})
}

func listBoxItemContent(dat ListBoxOption, cfg ListBoxCfg) View {
	if dat.IsSubheading {
		return Column(ContainerCfg{
			Spacing: Some[float32](1),
			Padding: Some(PaddingNone),
			Sizing:  FillFit,
			Content: []View{
				Text(TextCfg{
					Text:      dat.Name,
					TextStyle: cfg.SubheadingStyle,
				}),
				Row(ContainerCfg{
					Padding: Some(PaddingNone),
					Sizing:  FillFit,
					Content: []View{
						Rectangle(RectangleCfg{
							Width:  1,
							Height: 1,
							Sizing: FillFit,
							Color:  cfg.SubheadingStyle.Color,
						}),
					},
				}),
			},
		})
	}
	return Text(TextCfg{
		Text:      dat.Name,
		Mode:      TextModeMultiline,
		TextStyle: cfg.TextStyle,
	})
}

func listBoxOnKeyDown(
	listBoxID string,
	itemIDs []string,
	isMultiple bool,
	onSelect func([]string, *Event, *Window),
	selectedIDs []string,
	e *Event,
	w *Window,
) {
	if len(itemIDs) == 0 || onSelect == nil {
		return
	}

	action := listCoreNavigate(e.KeyCode, len(itemIDs))
	if e.KeyCode == KeySpace {
		action = ListCoreSelectItem
	}
	if action == ListCoreNone {
		return
	}
	e.IsHandled = true

	lbf := StateMap[string, int](w, nsListBoxFocus, capModerate)
	curIdx, _ := lbf.Get(listBoxID)

	if action == ListCoreSelectItem {
		if curIdx >= 0 && curIdx < len(itemIDs) {
			datID := itemIDs[curIdx]
			ids := listBoxNextSelectedIDs(
				selectedIDs, datID, isMultiple)
			onSelect(ids, e, w)
		}
		return
	}

	next, changed := listCoreApplyNav(action, curIdx, len(itemIDs))
	if changed {
		lbf.Set(listBoxID, next)
	}
}

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func applyListBoxDefaults(cfg *ListBoxCfg) {
	d := &DefaultButtonStyle
	if cfg.Color == (Color{}) {
		cfg.Color = d.Color
	}
	if cfg.ColorHover == (Color{}) {
		cfg.ColorHover = d.ColorHover
	}
	if cfg.ColorBorder == (Color{}) {
		cfg.ColorBorder = d.ColorBorder
	}
	if cfg.ColorSelect == (Color{}) {
		cfg.ColorSelect = colorSelectDark
	}
	if cfg.Padding == (Padding{}) {
		cfg.Padding = PaddingTwo
	}

	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
	if cfg.SubheadingStyle == (TextStyle{}) {
		cfg.SubheadingStyle = TextStyle{
			Color: RGB(180, 180, 180),
			Size:  SizeTextSmall,
		}
	}
}

func listBoxDataHash(items []ListBoxOption) uint64 {
	const offset uint64 = 1469598103934665603
	const prime uint64 = 1099511628211
	h := offset
	for i := range items {
		it := items[i]
		for j := 0; j < len(it.ID); j++ {
			h ^= uint64(it.ID[j])
			h *= prime
		}
		h ^= 0xff
		h *= prime

		for j := 0; j < len(it.Name); j++ {
			h ^= uint64(it.Name[j])
			h *= prime
		}
		h ^= 0xff
		h *= prime

		for j := 0; j < len(it.Value); j++ {
			h ^= uint64(it.Value[j])
			h *= prime
		}
		h ^= 0xff
		h *= prime

		if it.IsSubheading {
			h ^= 1
		}
		h *= prime
	}
	return h
}
