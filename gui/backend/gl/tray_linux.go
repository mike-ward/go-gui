//go:build linux

package gl

import (
	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/sni"
)

var trayBridge sni.Tray

func (n *nativePlatform) CreateSystemTray(
	cfg gui.SystemTrayCfg, actionCb func(string),
) (int, error) {
	return trayBridge.Create(cfg, actionCb)
}

func (n *nativePlatform) UpdateSystemTray(
	id int, cfg gui.SystemTrayCfg,
) {
	trayBridge.Update(id, cfg)
}

func (n *nativePlatform) RemoveSystemTray(id int) {
	trayBridge.Remove(id)
}
