package gui

// SystemTrayCfg configures a system tray icon and menu.
type SystemTrayCfg struct {
	Tooltip  string
	IconPNG  []byte // PNG icon data
	Menu     []NativeMenuItemCfg
	OnAction func(id string) // menu item callback
}

// SystemTrayHandle identifies an active system tray entry.
type SystemTrayHandle struct{ id int }
