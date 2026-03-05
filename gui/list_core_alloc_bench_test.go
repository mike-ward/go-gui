package gui

import (
	"strconv"
	"testing"
)

func benchmarkListCoreItems(n int) []ListCoreItem {
	items := make([]ListCoreItem, n)
	for i := 0; i < n; i++ {
		s := "Item " + strconv.Itoa(i)
		items[i] = ListCoreItem{ID: s, Label: s}
	}
	return items
}

func benchmarkOptions(n int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = "Option " + strconv.Itoa(i)
	}
	return out
}

func BenchmarkListCorePrepare(b *testing.B) {
	items := benchmarkListCoreItems(2000)
	b.ReportAllocs()

	b.Run("empty_query", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = listCorePrepare(items, "", 25)
		}
	})

	b.Run("fuzzy_query", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = listCorePrepare(items, "199", 25)
		}
	})
}

func BenchmarkComboboxGenerateLayout(b *testing.B) {
	options := benchmarkOptions(500)
	w := newTestWindow()
	cfg := ComboboxCfg{
		ID:       "bench-cb",
		Options:  options,
		OnSelect: func(_ string, _ *Event, _ *Window) {},
		IDScroll: 9901,
	}

	b.Run("closed", func(b *testing.B) {
		ss := StateMap[string, bool](w, nsCombobox, capModerate)
		ss.Set(cfg.ID, false)
		v := Combobox(cfg)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = GenerateViewLayout(v, w)
		}
	})

	b.Run("open_query", func(b *testing.B) {
		ss := StateMap[string, bool](w, nsCombobox, capModerate)
		ss.Set(cfg.ID, true)
		sq := StateMap[string, string](w, nsComboboxQuery, capModerate)
		sq.Set(cfg.ID, "49")
		v := Combobox(cfg)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = GenerateViewLayout(v, w)
		}
	})
}

func BenchmarkCommandPaletteGenerateLayout(b *testing.B) {
	items := make([]CommandPaletteItem, 500)
	for i := 0; i < len(items); i++ {
		s := strconv.Itoa(i)
		items[i] = CommandPaletteItem{
			ID:     "cmd-" + s,
			Label:  "Command " + s,
			Detail: "Detail " + s,
		}
	}

	w := newTestWindow()
	id := "bench-cp"
	CommandPaletteShow(id, 77, w)
	StateMap[string, string](w, nsCmdPaletteQuery, capModerate).
		Set(id, "49")

	v := CommandPalette(CommandPaletteCfg{
		ID:       id,
		Items:    items,
		OnAction: func(_ string, _ *Event, _ *Window) {},
		IDFocus:  77,
		IDScroll: 9902,
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateViewLayout(v, w)
	}
}

func BenchmarkListBoxGenerateLayout(b *testing.B) {
	data := make([]ListBoxOption, 500)
	for i := 0; i < len(data); i++ {
		s := strconv.Itoa(i)
		data[i] = NewListBoxOption("id-"+s, "Name "+s, "")
	}
	selected := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		selected = append(selected, "id-"+strconv.Itoa(i))
	}

	b.Run("unbounded", func(b *testing.B) {
		w := newTestWindow()
		v := ListBox(ListBoxCfg{
			ID:          "bench-lb",
			Data:        data,
			SelectedIDs: selected,
			OnSelect:    func(_ []string, _ *Event, _ *Window) {},
		})

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = GenerateViewLayout(v, w)
		}
	})

	b.Run("bounded_virtualized", func(b *testing.B) {
		w := newTestWindow()
		scrollID := uint32(9903)
		StateMap[uint32, float32](w, nsScrollY, capScroll).Set(scrollID, 1000)
		v := ListBox(ListBoxCfg{
			ID:          "bench-lb-v",
			IDScroll:    scrollID,
			MaxHeight:   220,
			Data:        data,
			SelectedIDs: selected,
			OnSelect:    func(_ []string, _ *Event, _ *Window) {},
		})

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = GenerateViewLayout(v, w)
		}
	})
}
