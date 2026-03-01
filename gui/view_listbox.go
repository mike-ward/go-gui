package gui


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
	Radius          float32
	SizeBorder      float32
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

	list := make([]View, 0, len(cfg.Data))
	for _, dat := range cfg.Data {
		list = append(list,
			listBoxItemView(dat, cfg))
	}

	// Build a11y value text.
	var valParts []string
	for _, dat := range cfg.Data {
		if !containsStr(cfg.SelectedIDs, dat.ID) {
			continue
		}
		valParts = append(valParts, dat.Name)
	}

	listBoxID := cfg.ID
	isMultiple := cfg.Multiple
	onSelect := cfg.OnSelect
	selectedIDs := cfg.SelectedIDs

	// Collect selectable item IDs for keyboard nav.
	itemIDs := make([]string, 0, len(cfg.Data))
	for _, dat := range cfg.Data {
		if !dat.IsSubheading {
			itemIDs = append(itemIDs, dat.ID)
		}
	}

	return Column(ContainerCfg{
		ID:          cfg.ID,
		A11YRole:    AccessRoleList,
		A11YLabel:   a11yLabel(cfg.A11YLabel, cfg.ID),
		IDFocus:     cfg.IDFocus,
		IDScroll:    cfg.IDScroll,
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
		SizeBorder:  cfg.SizeBorder,
		Radius:      cfg.Radius,
		Padding:     cfg.Padding,
		Sizing:      cfg.Sizing,
		Spacing:     0,
		Disabled:    cfg.Disabled,
		Invisible:   cfg.Invisible,
		Content:     list,
	})
}

func listBoxItemView(dat ListBoxOption, cfg ListBoxCfg) View {
	color := ColorTransparent
	if containsStr(cfg.SelectedIDs, dat.ID) {
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
	if containsStr(cfg.SelectedIDs, dat.ID) {
		a11yState = AccessStateSelected
	}

	return Row(ContainerCfg{
		A11YRole:  AccessRoleListItem,
		A11YLabel: dat.Name,
		A11YState: a11yState,
		Color:     color,
		Padding:   PaddingTwoFive,
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
			Spacing: 1,
			Padding: PaddingNone,
			Sizing:  FillFit,
			Content: []View{
				Text(TextCfg{
					Text:      dat.Name,
					TextStyle: cfg.SubheadingStyle,
				}),
				Row(ContainerCfg{
					Padding: PaddingNone,
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
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = SizeBorderDef
	}
	if cfg.Radius == 0 {
		cfg.Radius = RadiusSmall
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
