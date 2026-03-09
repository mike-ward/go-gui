package gui

// MenubarCfg configures a horizontal menubar or standalone
// menu.
type MenubarCfg struct {
	ID                string
	TextStyle         TextStyle
	TextStyleSubtitle TextStyle
	Color             Color
	ColorBorder       Color
	ColorSelect       Color
	Sizing            Sizing
	Padding           Opt[Padding]
	PaddingMenuItem   Opt[Padding]
	SizeBorder        Opt[float32]
	PaddingSubmenu    Opt[Padding]
	PaddingSubtitle   Opt[Padding]
	Action            func(string, *Event, *Window)
	Items             []MenuItemCfg
	WidthSubmenuMin   Opt[float32]
	WidthSubmenuMax   Opt[float32]
	Radius            Opt[float32]
	RadiusBorder      Opt[float32]
	RadiusSubmenu     Opt[float32]
	RadiusMenuItem    Opt[float32]
	Spacing           Opt[float32]
	SpacingSubmenu    Opt[float32]
	IDFocus           uint32
	FloatAnchor       FloatAttach
	FloatTieOff       FloatAttach
	FloatOffsetX      float32
	FloatOffsetY      float32
	FloatZIndex       int
	Disabled          bool
	Invisible         bool
	Float             bool
	FloatAutoFlip     bool
}

// Menubar creates a horizontal menubar with keyboard
// navigation.
func Menubar(w *Window, cfg MenubarCfg) View {
	applyMenubarDefaults(&cfg)
	if cfg.IDFocus == 0 {
		cfg.IDFocus = fnvSum32("menubar_" + cfg.ID)
	}
	checkForDuplicateMenuIDs(cfg.Items)

	// On focus with no selection, select first item.
	if w.IsFocus(cfg.IDFocus) {
		sel := StateReadOr[uint32, string](
			w, nsMenu, cfg.IDFocus, "")
		if sel == "" {
			if first, ok := firstSelectable(cfg.Items); ok {
				sm := StateMap[uint32, string](
					w, nsMenu, capModerate)
				sm.Set(cfg.IDFocus, first.ID)
			}
		}
	}

	return Row(ContainerCfg{
		ID:            cfg.ID,
		IDFocus:       cfg.IDFocus,
		Color:         cfg.Color,
		ColorBorder:   cfg.ColorBorder,
		SizeBorder:    cfg.SizeBorder,
		Radius:        cfg.RadiusBorder,
		Spacing:       cfg.Spacing,
		Padding:       cfg.Padding,
		Sizing:        cfg.Sizing,
		Float:         cfg.Float,
		FloatAutoFlip: cfg.FloatAutoFlip,
		FloatAnchor:   cfg.FloatAnchor,
		FloatTieOff:   cfg.FloatTieOff,
		FloatOffsetX:  cfg.FloatOffsetX,
		FloatOffsetY:  cfg.FloatOffsetY,
		FloatZIndex:   cfg.FloatZIndex,
		Disabled:      cfg.Disabled,
		Invisible:     cfg.Invisible,
		A11YRole:      AccessRoleMenuBar,
		OnKeyDown:     makeMenubarOnKeyDown(cfg),
		AmendLayout:   makeMenuAmendLayout(cfg.IDFocus),
		Content:       menuBuild(cfg, 0, cfg.Items, w),
	})
}

func applyMenubarDefaults(cfg *MenubarCfg) {
	d := &DefaultMenubarStyle
	if !cfg.Color.IsSet() {
		cfg.Color = d.Color
	}
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = d.ColorBorder
	}
	if !cfg.ColorSelect.IsSet() {
		cfg.ColorSelect = d.ColorSelect
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = d.TextStyle
	}
	if cfg.TextStyleSubtitle == (TextStyle{}) {
		cfg.TextStyleSubtitle = d.TextStyleSubtitle
	}
	if cfg.Sizing == (Sizing{}) {
		cfg.Sizing = FillFit
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(d.Padding)
	}
	if !cfg.PaddingMenuItem.IsSet() {
		cfg.PaddingMenuItem = Some(d.PaddingMenuItem)
	}
	if !cfg.PaddingSubmenu.IsSet() {
		cfg.PaddingSubmenu = Some(d.PaddingSubmenu)
	}
	if !cfg.PaddingSubtitle.IsSet() {
		cfg.PaddingSubtitle = Some(d.PaddingSubtitle)
	}
	if !cfg.SpacingSubmenu.IsSet() {
		cfg.SpacingSubmenu = Some(d.SpacingSubmenu)
	}
	if cfg.Action == nil {
		cfg.Action = func(_ string, e *Event, _ *Window) {
			e.IsHandled = true
		}
	}
}

// MenuIdMap maps menu item IDs to directional nav nodes.
type MenuIdMap map[string]MenuIdNode

// MenuIdNode stores directional navigation targets.
type MenuIdNode struct {
	Left  string
	Right string
	Up    string
	Down  string
}

func makeMenubarOnKeyDown(cfg MenubarCfg) func(*Layout, *Event, *Window) {
	return func(layout *Layout, e *Event, w *Window) {
		menubarOnKeyDown(cfg, layout, e, w)
	}
}

func menubarOnKeyDown(cfg MenubarCfg, _ *Layout, e *Event, w *Window) {
	sm := StateMap[uint32, string](w, nsMenu, capModerate)

	switch e.KeyCode {
	case KeyEscape:
		w.SetIDFocus(0)
		sm.Delete(cfg.IDFocus)
		e.IsHandled = true

	case KeySpace, KeyEnter:
		sel, _ := sm.Get(cfg.IDFocus)
		if sel == "" {
			return
		}
		item, found := findMenuItemCfg(cfg.Items, sel)
		if !found {
			return
		}
		if item.Action != nil {
			item.Action(&item, e, w)
		}
		if cfg.Action != nil {
			cfg.Action(sel, e, w)
		}
		// Close menu on leaf item selection.
		if len(item.Submenu) == 0 {
			w.SetIDFocus(0)
			sm.Delete(cfg.IDFocus)
		}
		e.IsHandled = true

	case KeyLeft, KeyRight, KeyUp, KeyDown:
		sel, _ := sm.Get(cfg.IDFocus)
		if sel == "" {
			return
		}
		idMap := menuMapper(cfg.Items)
		node, ok := idMap[sel]
		if !ok {
			return
		}
		var target string
		switch e.KeyCode {
		case KeyLeft:
			target = node.Left
		case KeyRight:
			target = node.Right
		case KeyUp:
			target = node.Up
		case KeyDown:
			target = node.Down
		}
		if target != "" && target != sel {
			sm.Set(cfg.IDFocus, target)
			w.viewState.menuKeyNav = true
		}
		e.IsHandled = true
	}
}

// menuMapperVertical builds a directional navigation graph
// for a vertical standalone menu (context menu). Top-level
// items use Up/Down for siblings, Right to enter submenus.
func menuMapperVertical(items []MenuItemCfg) MenuIdMap {
	m := make(MenuIdMap)
	selectables := make([]MenuItemCfg, 0, len(items))
	for _, item := range items {
		if isSelectableMenuID(item.ID) {
			selectables = append(selectables, item)
		}
	}
	if len(selectables) == 0 {
		return m
	}

	for i, item := range selectables {
		node := MenuIdNode{
			Up:    menuItemUp(i, selectables),
			Down:  menuItemDown(i, selectables),
			Right: menuItemRight(item, ""),
		}
		m[item.ID] = node

		if len(item.Submenu) > 0 {
			submenuMapper(item.Submenu, item.ID, node,
				node, "", m)
		}
	}
	return m
}

// menuMapper builds a directional navigation graph for all
// menu items.
func menuMapper(items []MenuItemCfg) MenuIdMap {
	m := make(MenuIdMap)
	selectables := make([]MenuItemCfg, 0, len(items))
	for _, item := range items {
		if isSelectableMenuID(item.ID) {
			selectables = append(selectables, item)
		}
	}
	if len(selectables) == 0 {
		return m
	}

	for i, item := range selectables {
		leftIdx := (i - 1 + len(selectables)) % len(selectables)
		rightIdx := (i + 1) % len(selectables)

		node := MenuIdNode{
			Left:  selectables[leftIdx].ID,
			Right: selectables[rightIdx].ID,
			Up:    item.ID,
			Down:  item.ID,
		}

		// Down goes to first submenu child.
		if len(item.Submenu) > 0 {
			if first, ok := firstSelectable(item.Submenu); ok {
				node.Down = first.ID
			}
		}

		m[item.ID] = node

		// Build submenu mappings.
		if len(item.Submenu) > 0 {
			rightID := selectables[rightIdx].ID
			submenuMapper(item.Submenu, item.ID, node,
				node, rightID, m)
		}
	}
	return m
}

// submenuMapper recursively builds navigation for submenu
// items.
func submenuMapper(items []MenuItemCfg, parentID string,
	_, rootNode MenuIdNode, rootRight string,
	m MenuIdMap) {

	selectables := make([]MenuItemCfg, 0, len(items))
	for _, item := range items {
		if isSelectableMenuID(item.ID) {
			selectables = append(selectables, item)
		}
	}
	if len(selectables) == 0 {
		return
	}

	for i, item := range selectables {
		node := MenuIdNode{
			Left:  parentID,
			Right: menuItemRight(item, rootRight),
			Up:    menuItemUp(i, selectables),
			Down:  menuItemDown(i, selectables),
		}
		m[item.ID] = node

		if len(item.Submenu) > 0 {
			submenuMapper(item.Submenu, item.ID, node,
				rootNode, rootRight, m)
		}
	}
}

// isSelectableMenuID returns true if the ID is neither a
// separator nor subtitle sentinel.
func isSelectableMenuID(id string) bool {
	return id != MenuSeparatorID && id != MenuSubtitleID
}

// findMenuByID recursively searches items for a matching ID.
func findMenuByID(items []MenuItemCfg, id string) (MenuItemCfg, bool) {
	return findMenuItemCfg(items, id)
}

func nextSelectable(idx int, items []MenuItemCfg) (MenuItemCfg, bool) {
	for i := idx + 1; i < len(items); i++ {
		if isSelectableMenuID(items[i].ID) {
			return items[i], true
		}
	}
	return MenuItemCfg{}, false
}

func previousSelectable(idx int, items []MenuItemCfg) (MenuItemCfg, bool) {
	for i := idx - 1; i >= 0; i-- {
		if isSelectableMenuID(items[i].ID) {
			return items[i], true
		}
	}
	return MenuItemCfg{}, false
}

func firstSelectable(items []MenuItemCfg) (MenuItemCfg, bool) {
	for _, item := range items {
		if isSelectableMenuID(item.ID) {
			return item, true
		}
	}
	return MenuItemCfg{}, false
}

func lastSelectable(items []MenuItemCfg) (MenuItemCfg, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		if isSelectableMenuID(items[i].ID) {
			return items[i], true
		}
	}
	return MenuItemCfg{}, false
}

// menuItemRight returns the right-nav target: first submenu
// child if present, else idRight (root-level right sibling).
func menuItemRight(item MenuItemCfg, idRight string) string {
	if len(item.Submenu) > 0 {
		if first, ok := firstSelectable(item.Submenu); ok {
			return first.ID
		}
	}
	return idRight
}

// menuItemUp returns the up-nav target: previous selectable
// sibling or wrap to last.
func menuItemUp(idx int, items []MenuItemCfg) string {
	if prev, ok := previousSelectable(idx, items); ok {
		return prev.ID
	}
	if last, ok := lastSelectable(items); ok {
		return last.ID
	}
	return items[idx].ID
}

// menuItemDown returns the down-nav target: next selectable
// sibling or wrap to first.
func menuItemDown(idx int, items []MenuItemCfg) string {
	if next, ok := nextSelectable(idx, items); ok {
		return next.ID
	}
	if first, ok := firstSelectable(items); ok {
		return first.ID
	}
	return items[idx].ID
}
