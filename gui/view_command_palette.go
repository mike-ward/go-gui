package gui

// CommandPaletteItem represents one action in the palette.
type CommandPaletteItem struct {
	ID       string
	Label    string
	Detail   string
	Icon     string
	Group    string
	Disabled bool
}

// CommandPaletteCfg configures a command palette view.
type CommandPaletteCfg struct {
	ID             string
	Items          []CommandPaletteItem
	OnAction       func(string, *Event, *Window)
	OnDismiss      func(*Window)
	Placeholder    string
	TextStyle      TextStyle
	DetailStyle    TextStyle
	Color          Color
	ColorBorder    Color
	ColorHighlight Color
	SizeBorder     Opt[float32]
	Radius         Opt[float32]
	Width          float32
	MaxHeight      float32
	BackdropColor  Color
	IDFocus        uint32
	IDScroll       uint32
}

// commandPaletteView implements View for command palette.
type commandPaletteView struct {
	cfg CommandPaletteCfg
}

// CommandPalette creates the palette view. Include in view tree;
// hidden unless CommandPaletteShow was called.
func CommandPalette(cfg CommandPaletteCfg) View {
	applyCommandPaletteDefaults(&cfg)
	return &commandPaletteView{cfg: cfg}
}

func (cp *commandPaletteView) Content() []View { return nil }

func (cp *commandPaletteView) GenerateLayout(w *Window) Layout {
	cfg := &cp.cfg
	dn := &DefaultCommandPaletteStyle
	sizeBorder := cfg.SizeBorder.Get(dn.SizeBorder)
	radius := cfg.Radius.Get(dn.Radius)
	visible := StateReadOr[string, bool](w, nsCmdPalette, cfg.ID, false)
	if !visible {
		return GenerateViewLayout(Row(ContainerCfg{}), w)
	}

	query := StateReadOr[string, string](w, nsCmdPaletteQuery, cfg.ID, "")
	highlighted := StateReadOr[string, int](w, nsCmdPaletteHighlight, cfg.ID, 0)

	// Convert to core items.
	coreItems := make([]ListCoreItem, len(cfg.Items))
	for i, item := range cfg.Items {
		coreItems[i] = cmdPaletteItemToCore(item)
	}

	// Filter + rank.
	prepared := listCorePrepare(coreItems, query, highlighted)
	filtered := prepared.Items
	filteredIDs := prepared.IDs
	hl := prepared.HL

	// Virtualization.
	rowH := listCoreRowHeightEstimate(cfg.TextStyle, PaddingTwoFive)
	var scrollY float32
	if cfg.IDScroll > 0 {
		scrollY = StateReadOr[uint32, float32](w, nsScrollY, cfg.IDScroll, 0)
	}
	first, last := listCoreVisibleRange(len(filtered), rowH, cfg.MaxHeight, scrollY)

	onAction := cfg.OnAction
	paletteID := cfg.ID
	onDismiss := cfg.OnDismiss

	coreCfg := ListCoreCfg{
		TextStyle:      cfg.TextStyle,
		DetailStyle:    cfg.DetailStyle,
		ColorHighlight: cfg.ColorHighlight,
		ColorHover:     cfg.ColorHighlight,
		ColorSelected:  cfg.ColorHighlight,
		PaddingItem:    PaddingTwoFive,
		ShowDetails:    true,
		ShowIcons:      true,
		OnItemClick: func(itemID string, _ int, e *Event, w *Window) {
			if onAction != nil {
				onAction(itemID, e, w)
			}
			CommandPaletteDismiss(paletteID, w)
			if onDismiss != nil {
				onDismiss(w)
			}
		},
	}

	resultViews := listCoreViews(filtered, coreCfg, first, last, hl, nil, rowH)

	// Build layout: backdrop column with centered card.
	return GenerateViewLayout(Column(ContainerCfg{
		Color:   cfg.BackdropColor,
		Sizing:  FillFill,
		Float:   true,
		VAlign:  VAlignTop,
		HAlign:  HAlignCenter,
		Padding: NewPadding(60, 0, 0, 0),
		OnClick: func(_ *Layout, e *Event, w *Window) {
			CommandPaletteDismiss(paletteID, w)
			if onDismiss != nil {
				onDismiss(w)
			}
			e.IsHandled = true
		},
		Content: []View{
			Column(ContainerCfg{
				ID:          cfg.ID,
				IDFocus:     cfg.IDFocus,
				Color:       cfg.Color,
				ColorBorder: cfg.ColorBorder,
				SizeBorder:  Some(sizeBorder),
				Radius:      Some(radius),
				Width:       cfg.Width,
				Padding:     PaddingNone,
				Spacing:     Some(float32(0)),
				Sizing:      FixedFit,
				OnKeyDown:   makePaletteOnKeyDown(paletteID, onAction, onDismiss, filteredIDs),
				OnClick: func(_ *Layout, e *Event, _ *Window) {
					// Prevent backdrop click when clicking card.
					e.IsHandled = true
				},
				Content: []View{
					Row(ContainerCfg{
						Padding: PaddingSmall,
						Sizing:  FillFit,
						Content: []View{
							Input(InputCfg{
								ID:            cfg.ID + ".input",
								Text:          query,
								Placeholder:   cfg.Placeholder,
								TextStyle:     cfg.TextStyle,
								IDFocus:       cfg.IDFocus,
								Sizing:        FillFit,
								OnTextChanged: makePaletteOnTextChanged(cfg.ID),
							}),
						},
					}),
					Column(ContainerCfg{
						IDScroll:  cfg.IDScroll,
						MaxHeight: cfg.MaxHeight,
						Sizing:    FillFit,
						Padding:   PaddingNone,
						Spacing:   Some(float32(0)),
						Clip:      true,
						Content:   resultViews,
					}),
				},
			}),
		},
	}), w)
}

// CommandPaletteShow makes the palette visible and focuses input.
func CommandPaletteShow(id string, idFocus uint32, w *Window) {
	ss := StateMap[string, bool](w, nsCmdPalette, capModerate)
	ss.Set(id, true)
	sq := StateMap[string, string](w, nsCmdPaletteQuery, capModerate)
	sq.Set(id, "")
	sh := StateMap[string, int](w, nsCmdPaletteHighlight, capModerate)
	sh.Set(id, 0)
	w.SetIDFocus(idFocus)
	w.UpdateWindow()
}

// CommandPaletteDismiss hides the palette.
func CommandPaletteDismiss(id string, w *Window) {
	ss := StateMap[string, bool](w, nsCmdPalette, capModerate)
	ss.Set(id, false)
	sq := StateMap[string, string](w, nsCmdPaletteQuery, capModerate)
	sq.Set(id, "")
	sh := StateMap[string, int](w, nsCmdPaletteHighlight, capModerate)
	sh.Set(id, 0)
	w.UpdateWindow()
}

// CommandPaletteToggle toggles palette visibility.
func CommandPaletteToggle(id string, idFocus uint32, w *Window) {
	visible := StateReadOr[string, bool](w, nsCmdPalette, id, false)
	if visible {
		CommandPaletteDismiss(id, w)
	} else {
		CommandPaletteShow(id, idFocus, w)
	}
}

// CommandPaletteIsVisible returns whether the palette is showing.
func CommandPaletteIsVisible(w *Window, id string) bool {
	return StateReadOr[string, bool](w, nsCmdPalette, id, false)
}

func cmdPaletteItemToCore(item CommandPaletteItem) ListCoreItem {
	return ListCoreItem{
		ID:       item.ID,
		Label:    item.Label,
		Detail:   item.Detail,
		Icon:     item.Icon,
		Group:    item.Group,
		Disabled: item.Disabled,
	}
}

func makePaletteOnTextChanged(paletteID string) func(*Layout, string, *Window) {
	return func(_ *Layout, newText string, w *Window) {
		sq := StateMap[string, string](w, nsCmdPaletteQuery, capModerate)
		sq.Set(paletteID, newText)
		sh := StateMap[string, int](w, nsCmdPaletteHighlight, capModerate)
		sh.Set(paletteID, 0)
		w.UpdateWindow()
	}
}

func makePaletteOnKeyDown(paletteID string, onAction func(string, *Event, *Window), onDismiss func(*Window), filteredIDs []string) func(*Layout, *Event, *Window) {
	return func(_ *Layout, e *Event, w *Window) {
		paletteOnKeyDown(paletteID, onAction, onDismiss, filteredIDs, e, w)
	}
}

func paletteOnKeyDown(paletteID string, onAction func(string, *Event, *Window), onDismiss func(*Window), filteredIDs []string, e *Event, w *Window) {
	if e.KeyCode == KeyEscape {
		CommandPaletteDismiss(paletteID, w)
		if onDismiss != nil {
			onDismiss(w)
		}
		e.IsHandled = true
		return
	}

	itemCount := len(filteredIDs)
	sh := StateMap[string, int](w, nsCmdPaletteHighlight, capModerate)
	cur, _ := sh.Get(paletteID)
	action := listCoreNavigate(e.KeyCode, itemCount)

	if action == ListCoreSelectItem {
		if cur >= 0 && cur < itemCount && onAction != nil {
			onAction(filteredIDs[cur], e, w)
			CommandPaletteDismiss(paletteID, w)
			if onDismiss != nil {
				onDismiss(w)
			}
		}
		e.IsHandled = true
		return
	}
	next, changed := listCoreApplyNav(action, cur, itemCount)
	if changed {
		sh.Set(paletteID, next)
		w.UpdateWindow()
		e.IsHandled = true
	}
}

func applyCommandPaletteDefaults(cfg *CommandPaletteCfg) {
	d := &DefaultCommandPaletteStyle
	if cfg.ID == "" {
		cfg.ID = "__cmd_palette__"
	}
	if cfg.Placeholder == "" {
		cfg.Placeholder = "Type a command..."
	}
	if cfg.Color == (Color{}) {
		cfg.Color = d.Color
	}
	if cfg.ColorBorder == (Color{}) {
		cfg.ColorBorder = d.ColorBorder
	}
	if cfg.ColorHighlight == (Color{}) {
		cfg.ColorHighlight = d.ColorHighlight
	}
	if cfg.Width == 0 {
		cfg.Width = d.Width
	}
	if cfg.MaxHeight == 0 {
		cfg.MaxHeight = d.MaxHeight
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = d.TextStyle
	}
	if cfg.DetailStyle == (TextStyle{}) {
		cfg.DetailStyle = d.DetailStyle
	}
	if cfg.BackdropColor == (Color{}) {
		cfg.BackdropColor = d.BackdropColor
	}
}
