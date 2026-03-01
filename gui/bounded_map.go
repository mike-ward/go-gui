package gui

const boundedOrderCompactMin = 64

// BoundedMap is a map with maximum size. When full, oldest entries
// are evicted (FIFO).
type BoundedMap[K comparable, V any] struct {
	data    map[K]V
	order   []K
	head    int
	maxSize int
}

// NewBoundedMap creates a BoundedMap with the given max size.
func NewBoundedMap[K comparable, V any](maxSize int) *BoundedMap[K, V] {
	return &BoundedMap[K, V]{
		data:    make(map[K]V),
		maxSize: maxSize,
	}
}

// Set adds or updates a key-value pair. Evicts oldest if at capacity.
func (m *BoundedMap[K, V]) Set(key K, value V) {
	if m.maxSize < 1 {
		return
	}
	if _, exists := m.data[key]; exists {
		m.data[key] = value
		return
	}
	if len(m.data) >= m.maxSize && len(m.order) > m.head {
		for m.head < len(m.order) {
			oldestKey := m.order[m.head]
			m.head++
			if _, exists := m.data[oldestKey]; exists {
				delete(m.data, oldestKey)
				break
			}
		}
	}
	m.order = append(m.order, key)
	m.data[key] = value
	m.compactOrder()
}

// Get returns value for key. Second return is false if not found.
func (m *BoundedMap[K, V]) Get(key K) (V, bool) {
	v, ok := m.data[key]
	return v, ok
}

// Delete removes key from map.
func (m *BoundedMap[K, V]) Delete(key K) {
	if _, exists := m.data[key]; !exists {
		return
	}
	delete(m.data, key)
	if len(m.data) == 0 {
		m.order = m.order[:0]
		m.head = 0
		return
	}
	m.compactOrder()
}

// Contains returns true if key exists.
func (m *BoundedMap[K, V]) Contains(key K) bool {
	_, ok := m.data[key]
	return ok
}

// Len returns number of entries.
func (m *BoundedMap[K, V]) Len() int {
	return len(m.data)
}

// Clear removes all entries.
func (m *BoundedMap[K, V]) Clear() {
	clear(m.data)
	m.order = m.order[:0]
	m.head = 0
}

// Keys returns all keys in insertion order.
func (m *BoundedMap[K, V]) Keys() []K {
	if len(m.data) == 0 || m.head >= len(m.order) {
		return nil
	}
	out := make([]K, 0, len(m.data))
	for _, k := range m.order[m.head:] {
		if _, exists := m.data[k]; exists {
			out = append(out, k)
		}
	}
	return out
}

func (m *BoundedMap[K, V]) compactOrder() {
	if m.head <= 0 {
		return
	}
	if m.head < boundedOrderCompactMin && m.head*2 < len(m.order) {
		return
	}
	compact := make([]K, 0, len(m.data))
	for _, k := range m.order[m.head:] {
		if _, exists := m.data[k]; exists {
			compact = append(compact, k)
		}
	}
	m.order = compact
	m.head = 0
}
