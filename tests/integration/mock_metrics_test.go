package integration

import "sync/atomic"

// MockMetrics — тестовий мок-об'єкт для перевірки викликів метрик.
type MockMetrics struct {
	Hits    int32
	Misses  int32
	Sets    int32
	Deletes int32
}

func (m *MockMetrics) IncCacheHits() {
	atomic.AddInt32(&m.Hits, 1)
}

func (m *MockMetrics) IncCacheMisses() {
	atomic.AddInt32(&m.Misses, 1)
}

func (m *MockMetrics) IncCacheSets() {
	atomic.AddInt32(&m.Sets, 1)
}

func (m *MockMetrics) IncCacheDeletes() {
	atomic.AddInt32(&m.Deletes, 1)
}

// Reset очищає всі лічильники (опціонально для багаторазового використання в одному тесті).
func (m *MockMetrics) Reset() {
	atomic.StoreInt32(&m.Hits, 0)
	atomic.StoreInt32(&m.Misses, 0)
	atomic.StoreInt32(&m.Sets, 0)
	atomic.StoreInt32(&m.Deletes, 0)
}
