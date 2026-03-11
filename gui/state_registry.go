package gui

import (
	"fmt"
	"log"
	"reflect"
)

// StateRegistry stores per-widget BoundedMap instances keyed by
// namespace string.
type StateRegistry struct {
	maps map[string]any
	meta map[string]stateMapMeta
}

type stateMapMeta struct {
	typeTag stateMapTypeTag
	maxSize int
}

type stateMapTypeTag struct {
	keyType   reflect.Type
	valueType reflect.Type
}

func stateMapTypeTagOf[K comparable, V any]() stateMapTypeTag {
	return stateMapTypeTag{
		keyType:   reflect.TypeFor[K](),
		valueType: reflect.TypeFor[V](),
	}
}

func stateMapTypeCheck[K comparable, V any](r *StateRegistry, ns string) error {
	if r.meta == nil {
		return nil
	}
	if m, ok := r.meta[ns]; ok {
		tag := stateMapTypeTagOf[K, V]()
		if m.typeTag != tag {
			return fmt.Errorf(
				"state_map type mismatch: %s expected %s got %s",
				ns,
				m.typeTag.keyType.String()+":"+m.typeTag.valueType.String(),
				tag.keyType.String()+":"+tag.valueType.String(),
			)
		}
	}
	return nil
}

// StateMap returns (or lazily creates) a *BoundedMap[K, V] for the
// given namespace.
func StateMap[K comparable, V any](w *Window, ns string, maxSize int) *BoundedMap[K, V] {
	if err := stateMapTypeCheck[K, V](&w.viewState.registry, ns); err != nil {
		panic(err.Error())
	}
	if ptr, ok := w.viewState.registry.maps[ns]; ok {
		if m, ok := w.viewState.registry.meta[ns]; ok {
			if m.maxSize != maxSize {
				log.Printf("state_map max_size mismatch: %s registered %d, requested %d",
					ns, m.maxSize, maxSize)
			}
		}
		return ptr.(*BoundedMap[K, V])
	}
	m := NewBoundedMap[K, V](maxSize)
	if w.viewState.registry.maps == nil {
		w.viewState.registry.maps = make(map[string]any)
		w.viewState.registry.meta = make(map[string]stateMapMeta)
	}
	w.viewState.registry.maps[ns] = m
	w.viewState.registry.meta[ns] = stateMapMeta{
		typeTag: stateMapTypeTagOf[K, V](),
		maxSize: maxSize,
	}
	return m
}

// StateMapRead returns a *BoundedMap[K, V] for read-only access.
// Returns nil if namespace not initialized.
func StateMapRead[K comparable, V any](w *Window, ns string) *BoundedMap[K, V] {
	if err := stateMapTypeCheck[K, V](&w.viewState.registry, ns); err != nil {
		panic(err.Error())
	}
	if ptr, ok := w.viewState.registry.maps[ns]; ok {
		return ptr.(*BoundedMap[K, V])
	}
	return nil
}

// StateReadOr returns the value for key in namespace, or defaultVal
// if not found.
func StateReadOr[K comparable, V any](w *Window, ns string, key K, defaultVal V) V {
	sm := StateMapRead[K, V](w, ns)
	if sm == nil {
		return defaultVal
	}
	v, ok := sm.Get(key)
	if !ok {
		return defaultVal
	}
	return v
}

// Clear drops all registry references.
func (r *StateRegistry) Clear() {
	clear(r.maps)
	clear(r.meta)
}

// entryCount returns the number of entries in the BoundedMap
// for the given namespace, or 0 if not found.
func (r *StateRegistry) entryCount(ns string) int {
	if ptr, ok := r.maps[ns]; ok && ptr != nil {
		// Use type assertion to get length — we stored as any
		type lenner interface{ Len() int }
		if l, ok := ptr.(lenner); ok {
			return l.Len()
		}
	}
	return 0
}

// Namespace constants for internal gui state maps.
const (
	nsOverflow            = "gui.overflow"
	nsScrollX             = "gui.scroll.x"
	nsScrollY             = "gui.scroll.y"
	nsSelect              = "gui.select"
	nsInput               = "gui.input"
	nsInputFocus          = "gui.input.focus"
	nsSelectHL            = "gui.select.highlight"
	nsListBoxFocus        = "gui.listbox.focus"
	nsListBoxCache        = "gui.listbox.cache"
	nsProgress            = "gui.progress"
	nsSidebar             = "gui.sidebar"
	nsCombobox            = "gui.combobox"
	nsComboboxQuery       = "gui.combobox.query"
	nsComboboxHighlight   = "gui.combobox.highlight"
	nsComboboxItems       = "gui.combobox.items"
	nsCmdPalette          = "gui.cmd_palette"
	nsCmdPaletteQuery     = "gui.cmd_palette.query"
	nsCmdPaletteHighlight = "gui.cmd_palette.highlight"
	nsCmdPaletteItems     = "gui.cmd_palette.items"
	nsTreeExpanded        = "gui.tree.expanded"
	nsTreeFocus           = "gui.tree.focus"
	nsTreeLazy            = "gui.tree.lazy"
	nsInspector           = "gui.inspector"
	nsInspectorWidth      = "gui.inspector.w"
	nsDrawCanvas          = "gui.draw_canvas"
	nsMenu                = "gui.menu"
	nsDatePicker          = "gui.date_picker"
	nsColorPicker         = "gui.color_picker"
	nsRangeSliderPress    = "gui.range_slider.press"
	nsInputDate           = "gui.input_date"
	nsDgColWidths         = "gui.dg.col_widths"
	nsDgPresentation      = "gui.dg.presentation"
	nsDgResize            = "gui.dg.resize"
	nsDgHeaderHover       = "gui.dg.header_hover"
	nsDgRange             = "gui.dg.range"
	nsDgChooserOpen       = "gui.dg.chooser_open"
	nsDgEdit              = "gui.dg.edit"
	nsDgCrud              = "gui.dg.crud"
	nsDgJump              = "gui.dg.jump"
	nsDgPendingJump       = "gui.dg.pending_jump"
	nsDgSource            = "gui.dg.source"
	nsActiveDownloads     = "gui.active_downloads"
	nsSvgCache            = "gui.svg_cache"
	nsSvgDimCache         = "gui.svg_dim_cache"
	nsSvgAnimSeen         = "gui.svg_anim_seen"
	nsSvgAnimStart        = "gui.svg_anim_start"
	nsDragReorder         = "gui.drag_reorder"
	nsDragReorderIDsMeta  = "gui.drag_reorder.ids_meta"
	nsTableColWidths      = "gui.table.col_widths"
	nsDockDrag            = "gui.dock_drag"
	nsContextMenu         = "gui.context_menu"
	nsRtfLinkMenu         = "gui.rtf_link_menu"
	nsForm                = "gui.form"
)

// Capacity tiers.
const (
	capFew      = 20
	capModerate = 50
	capMany     = 100
	capScroll   = 200
)
