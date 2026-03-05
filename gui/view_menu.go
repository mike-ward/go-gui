package gui

import "log"

// Menu creates a standalone columnar menu (used by
// OverflowPanel and context menus).
func Menu(w *Window, cfg MenubarCfg) View {
	applyMenubarDefaults(&cfg)
	if cfg.IDFocus == 0 {
		cfg.IDFocus = fnvSum32("menu_" + cfg.ID)
		log.Printf("menu: auto-generated IDFocus=%d for %q",
			cfg.IDFocus, cfg.ID)
	}
	checkForDuplicateMenuIDs(cfg.Items)

	return Column(ContainerCfg{
		ID:           cfg.ID,
		Color:        cfg.Color,
		ColorBorder:  cfg.ColorBorder,
		SizeBorder:   cfg.SizeBorder,
		Radius:       cfg.RadiusBorder,
		MinWidth:     cfg.WidthSubmenuMin.Get(DefaultMenubarStyle.WidthSubmenuMin),
		MaxWidth:     cfg.WidthSubmenuMax.Get(DefaultMenubarStyle.WidthSubmenuMax),
		Spacing:      Some(cfg.SpacingSubmenu.Get(DefaultMenubarStyle.SpacingSubmenu)),
		Padding:      cfg.PaddingSubmenu,
		Float:        cfg.Float,
		FloatAnchor:  cfg.FloatAnchor,
		FloatTieOff:  cfg.FloatTieOff,
		FloatOffsetX: cfg.FloatOffsetX,
		FloatOffsetY: cfg.FloatOffsetY,
		IDFocus:      cfg.IDFocus,
		AmendLayout:  makeMenuAmendLayout(cfg.IDFocus),
		Content:      menuBuild(cfg, 1, cfg.Items, w),
	})
}

// menuBuild recursively builds menu item views.
func menuBuild(cfg MenubarCfg, level int, items []MenuItemCfg, w *Window) []View {
	sizing := FillFit
	if level == 0 {
		sizing = FitFit
	}

	selectedID := StateReadOr[uint32, string](
		w, nsMenu, cfg.IDFocus, "")

	views := make([]View, 0, len(items))
	for _, item := range items {
		// Determine padding.
		pad := item.Padding
		if !pad.IsSet() {
			if item.CustomView != nil {
				pad = PaddingNone
			} else if item.ID == MenuSubtitleID {
				pad = cfg.PaddingSubtitle
			} else {
				pad = cfg.PaddingMenuItem
			}
		}

		// Determine text style.
		ts := cfg.TextStyle
		if item.ID == MenuSubtitleID {
			ts = cfg.TextStyleSubtitle
		}

		// Build the configured item.
		configured := item
		configured.colorSelect = cfg.ColorSelect
		configured.Padding = pad
		configured.selected = (selectedID == item.ID)
		configured.sizing = sizing
		configured.radius = cfg.RadiusMenuItem.Get(DefaultMenubarStyle.RadiusMenuItem)
		configured.spacing = cfg.SpacingSubmenu.Get(DefaultMenubarStyle.SpacingSubmenu)
		configured.textStyle = ts

		// Attach submenu as child of menu item so float
		// positioning is relative to the item, not the bar.
		if len(item.Submenu) > 0 &&
			(selectedID == item.ID ||
				isMenuIDInTree(item.Submenu, selectedID)) {

			anchor := FloatTopRight
			tieOff := FloatTopLeft
			if level == 0 {
				anchor = FloatBottomLeft
				tieOff = FloatTopLeft
			}

			subViews := menuBuild(cfg, level+1,
				item.Submenu, w)

			submenu := Column(ContainerCfg{
				Color:       cfg.Color,
				ColorBorder: cfg.ColorBorder,
				SizeBorder:  cfg.SizeBorder,
				Radius:      cfg.RadiusSubmenu,
				MinWidth:    cfg.WidthSubmenuMin.Get(DefaultMenubarStyle.WidthSubmenuMin),
				MaxWidth:    cfg.WidthSubmenuMax.Get(DefaultMenubarStyle.WidthSubmenuMax),
				Spacing:     Some(cfg.SpacingSubmenu.Get(DefaultMenubarStyle.SpacingSubmenu)),
				Padding:     cfg.PaddingSubmenu,
				Float:         true,
				FloatAutoFlip: true,
				FloatAnchor: anchor,
				FloatTieOff: tieOff,
				Content:     subViews,
			})
			views = append(views, menuItem(cfg, configured, submenu))
		} else {
			views = append(views, menuItem(cfg, configured))
		}
	}
	return views
}

// makeMenuAmendLayout clears menu selection when the widget
// loses focus.
func makeMenuAmendLayout(idFocus uint32) func(*Layout, *Window) {
	return func(_ *Layout, w *Window) {
		if !w.IsFocus(idFocus) {
			sm := StateMapRead[uint32, string](w, nsMenu)
			if sm != nil {
				sm.Delete(idFocus)
			}
		}
	}
}


// findMenuItemCfg recursively searches for a menu item by ID.
func findMenuItemCfg(items []MenuItemCfg, id string) (MenuItemCfg, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
		if found, ok := findMenuItemCfg(item.Submenu, id); ok {
			return found, true
		}
	}
	return MenuItemCfg{}, false
}

// checkForDuplicateMenuIDs logs a warning if duplicate
// IDs exist (ignoring sentinel IDs).
func checkForDuplicateMenuIDs(items []MenuItemCfg) {
	seen := make(map[string]bool)
	if dup, ok := checkMenuIDs(items, seen); ok {
		log.Printf("menu: duplicate item ID %q", dup)
	}
}

func checkMenuIDs(items []MenuItemCfg, seen map[string]bool) (string, bool) {
	for _, item := range items {
		if item.ID == MenuSeparatorID || item.ID == MenuSubtitleID {
			continue
		}
		if seen[item.ID] {
			return item.ID, true
		}
		seen[item.ID] = true
		if dup, ok := checkMenuIDs(item.Submenu, seen); ok {
			return dup, true
		}
	}
	return "", false
}

// isMenuIDInTree checks if an ID exists in a submenu tree.
func isMenuIDInTree(submenu []MenuItemCfg, id string) bool {
	for _, item := range submenu {
		if item.ID == id {
			return true
		}
		if isMenuIDInTree(item.Submenu, id) {
			return true
		}
	}
	return false
}
