package metrics

type CacheMetricsAdapter struct {
	cacheName string
	engine    string
}

func NewCacheMetricsAdapter(cacheName, engine string) *CacheMetricsAdapter {
	return &CacheMetricsAdapter{
		cacheName: cacheName,
		engine:    engine,
	}
}

func (a *CacheMetricsAdapter) IncCacheHits() {
	CacheOperationsTotal.WithLabelValues(a.cacheName, "hit", a.engine).Inc()
}

func (a *CacheMetricsAdapter) IncCacheMisses() {
	CacheOperationsTotal.WithLabelValues(a.cacheName, "miss", a.engine).Inc()
}

func (a *CacheMetricsAdapter) IncCacheSets() {
	CacheOperationsTotal.WithLabelValues(a.cacheName, "set", a.engine).Inc()
}

func (a *CacheMetricsAdapter) IncCacheDeletes() {
	CacheOperationsTotal.WithLabelValues(a.cacheName, "delete", a.engine).Inc()
}
