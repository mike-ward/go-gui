package gui

// Menu item sentinel IDs.
const (
	MenuSeparatorID = "__separator__"
	MenuSubtitleID  = "__subtitle__"
)

// MenuItemCfg configures a single menu item. Items may be
// text, separators, subtitles, or submenus.
type MenuItemCfg struct {
	// Internal — set by menuBuild from theme/context.
	colorSelect Color
	textStyle   TextStyle
	sizing      Sizing
	radius      float32
	spacing     float32
	disabled    bool
	selected    bool

	// Public configuration.
	ID        string
	Text      string
	Padding   Padding
	Action    func(*MenuItemCfg, *Event, *Window)
	Submenu   []MenuItemCfg
	CustomView View
	Separator bool
}

// MenuItemText creates a simple text menu item.
func MenuItemText(id, text string) MenuItemCfg {
	return MenuItemCfg{
		ID:   id,
		Text: text,
	}
}

// MenuSeparator creates a separator line.
func MenuSeparator() MenuItemCfg {
	return MenuItemCfg{
		ID:        MenuSeparatorID,
		Separator: true,
	}
}

// MenuSubtitle creates a disabled subtitle item.
func MenuSubtitle(text string) MenuItemCfg {
	return MenuItemCfg{
		ID:       MenuSubtitleID,
		Text:     text,
		disabled: true,
	}
}

// MenuSubmenu creates an item with a submenu. A "›" indicator
// is appended.
func MenuSubmenu(id, text string, submenu []MenuItemCfg) MenuItemCfg {
	return MenuItemCfg{
		ID:      id,
		Text:    text,
		Submenu: submenu,
	}
}

// menuItem builds the View for a single menu item.
func menuItem(menubarCfg MenubarCfg, itemCfg MenuItemCfg) View {
	if itemCfg.Separator {
		return Column(ContainerCfg{
			Sizing:  FillFit,
			Padding: Some(NewPadding(2, 0, 2, 0)),
			Content: []View{
				Rectangle(RectangleCfg{
					Height: 1,
					Sizing: FillFit,
					Color:  menubarCfg.ColorBorder,
				}),
			},
		})
	}

	itemColor := ColorTransparent
	if itemCfg.selected {
		itemColor = itemCfg.colorSelect
	}

	var content View
	if itemCfg.CustomView != nil {
		content = itemCfg.CustomView
	} else {
		textContent := itemCfg.Text
		if len(itemCfg.Submenu) > 0 {
			textContent += "  \u203A"
		}
		content = Text(TextCfg{
			Text:      textContent,
			TextStyle: itemCfg.textStyle,
		})
	}

	itemID := itemCfg.ID
	cfgIDFocus := menubarCfg.IDFocus

	var onHover func(*Layout, *Event, *Window)
	if !itemCfg.disabled {
		onHover = func(_ *Layout, _ *Event, w *Window) {
			if !w.IsFocus(cfgIDFocus) {
				return
			}
			if w.viewState.menuKeyNav {
				return
			}
			w.SetMouseCursor(CursorPointingHand)
			sm := StateMap[uint32, string](
				w, nsMenu, capModerate)
			cur, _ := sm.Get(cfgIDFocus)
			if cur != itemID {
				sm.Set(cfgIDFocus, itemID)
			}
		}
	}

	return Column(ContainerCfg{
		Color:   itemColor,
		Sizing:  itemCfg.sizing,
		Padding: Some(itemCfg.Padding),
		Radius:  Some(itemCfg.radius),
		OnClick: menuItemClick(menubarCfg, itemCfg),
		OnHover: onHover,
		Content: []View{content},
	})
}

// menuItemClick returns the OnClick handler for a menu item.
func menuItemClick(cfg MenubarCfg, itemCfg MenuItemCfg) func(*Layout, *Event, *Window) {
	return leftClickOnly(func(_ *Layout, e *Event, w *Window) {
		w.SetIDFocus(cfg.IDFocus)

		if !isSelectableMenuID(itemCfg.ID) {
			return
		}

		sm := StateMap[uint32, string](
			w, nsMenu, capModerate)
		sm.Set(cfg.IDFocus, itemCfg.ID)

		if itemCfg.Action != nil {
			itemCfg.Action(&itemCfg, e, w)
		}
		if cfg.Action != nil {
			cfg.Action(itemCfg.ID, e, w)
		}

		// Close menu if leaf item (no submenu).
		if len(itemCfg.Submenu) == 0 {
			w.SetIDFocus(0)
			sm.Delete(cfg.IDFocus)
		}

		e.IsHandled = true
	})
}
