package agents

import "sync"

type QueuedMap[K comparable, V any] struct {
	entries map[K]V
	lock    sync.RWMutex
}

func NewQueuedMap[K comparable, V any](size int) *QueuedMap[K, V] {
	return &QueuedMap[K, V]{
		entries: make(map[K]V, size),
	}
}

func (m *QueuedMap[K, V]) Get(key K) (V, bool) {
	m.lock.RLock()
	v, ok := m.entries[key]
	m.lock.RUnlock()

	return v, ok
}

func (m *QueuedMap[K, V]) Has(key K) bool {
	m.lock.RLock()
	_, ok := m.entries[key]
	m.lock.RUnlock()

	return ok
}

func (m *QueuedMap[K, V]) Set(key K, value V) {
	m.lock.Lock()
	m.entries[key] = value
	m.lock.Unlock()
}

func (m *QueuedMap[K, V]) Delete(key K) {
	m.lock.Lock()
	delete(m.entries, key)
	m.lock.Unlock()
}
