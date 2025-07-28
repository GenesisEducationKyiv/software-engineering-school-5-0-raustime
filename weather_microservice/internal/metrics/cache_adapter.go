package metrics

// CacheMetricsAdapter реалізує інтерфейс cache.Metrics
type CacheMetricsAdapter struct {
	metrics *CacheMetrics
}

func NewCacheMetricsAdapter(metrics *CacheMetrics) *CacheMetricsAdapter {
	return &CacheMetricsAdapter{metrics: metrics}
}

func (a *CacheMetricsAdapter) IncCacheHits()    { a.metrics.Hits.Inc() }
func (a *CacheMetricsAdapter) IncCacheMisses()  { a.metrics.Misses.Inc() }
func (a *CacheMetricsAdapter) IncCacheSets()    { a.metrics.Sets.Inc() }
func (a *CacheMetricsAdapter) IncCacheDeletes() { a.metrics.Deletes.Inc() }
