package gui

// contextMenuState is the internal state for a ContextMenu widget.
type contextMenuState struct {
	Open bool
	X    float32
	Y    float32
}

// ContextMenuCfg configures a ContextMenu widget.
type ContextMenuCfg struct {
	ID      string
	Items   []MenuItemCfg
	Action  func(string, *Event, *Window)
	IDFocus uint32

	// Container passthrough (outer wrapper).
	Sizing  Sizing
	Width   float32
	Height  float32
	HAlign  HorizontalAlign
	VAlign  VerticalAlign
	Padding Opt[Padding]

	// Menu styling — optional, defaults from theme.
	Color             Color
	ColorBorder       Color
	ColorSelect       Color
	SizeBorder        Opt[float32]
	Radius            Opt[float32]
	RadiusMenuItem    Opt[float32]
	TextStyle         TextStyle
	TextStyleSubtitle TextStyle
	PaddingMenuItem   Padding
	PaddingSubmenu    Padding
	SpacingSubmenu    Opt[float32]
	WidthSubmenuMin   Opt[float32]
	WidthSubmenuMax   Opt[float32]

	// User click handler — fires before context menu logic.
	OnAnyClick func(*Layout, *Event, *Window)

	Content []View
}

// ContextMenu creates a container that opens a floating menu
// on right-click at the cursor position.
func ContextMenu(w *Window, cfg ContextMenuCfg) View {
	if cfg.IDFocus == 0 {
		cfg.IDFocus = fnvSum32("context_menu_" + cfg.ID)
	}
	checkForDuplicateMenuIDs(cfg.Items)
	applyContextMenuDefaults(&cfg)

	st := StateReadOr[string, contextMenuState](
		w, nsContextMenu, cfg.ID, contextMenuState{})

	content := make([]View, 0, len(cfg.Content)+1)
	content = append(content, cfg.Content...)

	if st.Open && w.IsFocus(cfg.IDFocus) {
		content = append(content,
			contextMenuPopup(w, cfg, st.X, st.Y))
	}

	idFocus := cfg.IDFocus
	return Column(ContainerCfg{
		ID:      cfg.ID,
		Sizing:  cfg.Sizing,
		Width:   cfg.Width,
		Height:  cfg.Height,
		HAlign:  cfg.HAlign,
		VAlign:  cfg.VAlign,
		Padding: cfg.Padding,
		IDFocus: cfg.IDFocus,
		OnAnyClick: func(l *Layout, e *Event, w *Window) {
			if cfg.OnAnyClick != nil {
				cfg.OnAnyClick(l, e, w)
				if e.IsHandled {
					return
				}
			}
			sm := StateMap[string, contextMenuState](
				w, nsContextMenu, capFew)
			if e.MouseButton == MouseRight {
				sm.Set(cfg.ID, contextMenuState{
					Open: true,
					X:    e.MouseX + l.Shape.X,
					Y:    e.MouseY + l.Shape.Y,
				})
				w.SetIDFocus(idFocus)
			} else {
				sm.Set(cfg.ID, contextMenuState{})
			}
			e.IsHandled = true
		},
		AmendLayout: func(_ *Layout, w *Window) {
			if !w.IsFocus(idFocus) {
				sm := StateMapRead[string, contextMenuState](
					w, nsContextMenu)
				if sm != nil {
					sm.Delete(cfg.ID)
				}
			}
		},
		Content: content,
	})
}

// contextMenuPopup builds the floating menu popup.
func contextMenuPopup(w *Window, cfg ContextMenuCfg, mx, my float32) View {
	action := func(id string, e *Event, w *Window) {
		sm := StateMap[string, contextMenuState](
			w, nsContextMenu, capFew)
		sm.Set(cfg.ID, contextMenuState{})
		if cfg.Action != nil {
			cfg.Action(id, e, w)
		}
	}

	return Menu(w, MenubarCfg{
		ID:                cfg.ID + "_popup",
		IDFocus:           cfg.IDFocus,
		Items:             cfg.Items,
		Action:            action,
		Color:             cfg.Color,
		ColorBorder:       cfg.ColorBorder,
		ColorSelect:       cfg.ColorSelect,
		SizeBorder:        cfg.SizeBorder,
		RadiusMenuItem:    cfg.RadiusMenuItem,
		TextStyle:         cfg.TextStyle,
		TextStyleSubtitle: cfg.TextStyleSubtitle,
		PaddingMenuItem:   cfg.PaddingMenuItem,
		PaddingSubmenu:    cfg.PaddingSubmenu,
		SpacingSubmenu:    cfg.SpacingSubmenu,
		WidthSubmenuMin:   cfg.WidthSubmenuMin,
		WidthSubmenuMax:   cfg.WidthSubmenuMax,
		Float:             true,
		FloatAnchor:       FloatTopLeft,
		FloatTieOff:       FloatTopLeft,
		FloatOffsetX:      mx,
		FloatOffsetY:      my,
	})
}

// applyContextMenuDefaults fills zero-value styling fields
// from DefaultMenubarStyle.
func applyContextMenuDefaults(cfg *ContextMenuCfg) {
	d := &DefaultMenubarStyle
	if cfg.Color == (Color{}) {
		cfg.Color = d.Color
	}
	if cfg.ColorBorder == (Color{}) {
		cfg.ColorBorder = d.ColorBorder
	}
	if cfg.ColorSelect == (Color{}) {
		cfg.ColorSelect = d.ColorSelect
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = d.TextStyle
	}
	if cfg.TextStyleSubtitle == (TextStyle{}) {
		cfg.TextStyleSubtitle = d.TextStyleSubtitle
	}
	if cfg.PaddingMenuItem == (Padding{}) {
		cfg.PaddingMenuItem = d.PaddingMenuItem
	}
	if cfg.PaddingSubmenu == (Padding{}) {
		cfg.PaddingSubmenu = d.PaddingSubmenu
	}
}
