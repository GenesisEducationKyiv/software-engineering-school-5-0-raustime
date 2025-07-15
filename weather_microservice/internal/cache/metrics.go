package cache

type Metrics interface {
	IncCacheHits()
	IncCacheMisses()
	IncCacheSets()
	IncCacheDeletes()
}

// NoopMetrics – пуста реалізація, якщо метрики не потрібні.
type NoopMetrics struct{}

func (NoopMetrics) IncCacheHits()    {}
func (NoopMetrics) IncCacheMisses()  {}
func (NoopMetrics) IncCacheSets()    {}
func (NoopMetrics) IncCacheDeletes() {}
