package gui

import "testing"

func BenchmarkBoundedMapSetUniqueToCapacity(b *testing.B) {
	const capSize = 1024
	m := NewBoundedMap[int, int](capSize)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if m.Len() == capSize {
			m.Clear()
		}
		m.Set(i, i)
	}
}

func BenchmarkBoundedMapSetWithEvictions(b *testing.B) {
	const capSize = 1024
	m := NewBoundedMap[int, int](capSize)
	for i := 0; i < capSize; i++ {
		m.Set(i, i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(capSize+i, i)
	}
}

func BenchmarkBoundedMapDeleteInsertChurn(b *testing.B) {
	m := NewBoundedMap[int, int](64)
	m.Set(0, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		m.Set(i, i)
		m.Delete(i)
	}
}

func BenchmarkBoundedMapKeysUnderChurn(b *testing.B) {
	m := NewBoundedMap[int, int](128)
	for i := 0; i < 2000; i++ {
		m.Set(i, i)
		if i%2 == 0 {
			m.Delete(i)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Keys()
	}
}

func BenchmarkBoundedMapRangeKeysUnderChurn(b *testing.B) {
	m := NewBoundedMap[int, int](128)
	for i := 0; i < 2000; i++ {
		m.Set(i, i)
		if i%2 == 0 {
			m.Delete(i)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		m.RangeKeys(func(int) bool {
			count++
			return true
		})
		if count == 0 {
			b.Fatal("unexpected empty map")
		}
	}
}
