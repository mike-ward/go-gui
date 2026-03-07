package gui

import (
	"path/filepath"
	"sort"
	"sync"
)

var (
	localeRegistryMu sync.RWMutex
	localeRegistry   = map[string]Locale{}
)

func init() {
	LocaleRegister(LocaleEnUS)
	LocaleRegister(LocaleDeDE)
	LocaleRegister(LocaleArSA)
	LocaleRegister(LocaleFrFR)
	LocaleRegister(LocaleEsES)
	LocaleRegister(LocalePtBR)
	LocaleRegister(LocaleJaJP)
	LocaleRegister(LocaleZhCN)
	LocaleRegister(LocaleKoKR)
	LocaleRegister(LocaleHeIL)
}

// LocaleRegister adds a locale to the global registry by ID.
// Overwrites any existing entry with the same ID.
func LocaleRegister(l Locale) {
	localeRegistryMu.Lock()
	localeRegistry[l.ID] = l
	localeRegistryMu.Unlock()
}

// LocaleGet retrieves a registered locale by ID.
func LocaleGet(id string) (Locale, bool) {
	localeRegistryMu.RLock()
	l, ok := localeRegistry[id]
	localeRegistryMu.RUnlock()
	return l, ok
}

// LocaleRegisteredNames returns sorted IDs of all registered
// locales.
func LocaleRegisteredNames() []string {
	localeRegistryMu.RLock()
	names := make([]string, 0, len(localeRegistry))
	for k := range localeRegistry {
		names = append(names, k)
	}
	localeRegistryMu.RUnlock()
	sort.Strings(names)
	return names
}

// LocaleLoadDir loads all *.json files from a directory and
// registers each as a locale.
func LocaleLoadDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	for _, f := range files {
		l, err := LocaleLoad(f)
		if err != nil {
			return err
		}
		LocaleRegister(l)
	}
	return nil
}
