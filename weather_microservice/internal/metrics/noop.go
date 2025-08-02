package metrics

// NoopMetrics — реалізація інтерфейсу Metrics, яка нічого не робить.
type NoopMetrics struct{}

func (NoopMetrics) IncCacheHits()    {}
func (NoopMetrics) IncCacheMisses()  {}
func (NoopMetrics) IncCacheSets()    {}
func (NoopMetrics) IncCacheDeletes() {}
