package metrics

type CacheMetricsAdapter struct{}

func (CacheMetricsAdapter) IncCacheHits()    { CacheHits.Inc() }
func (CacheMetricsAdapter) IncCacheMisses()  { CacheMisses.Inc() }
func (CacheMetricsAdapter) IncCacheSets()    { CacheSets.Inc() }
func (CacheMetricsAdapter) IncCacheDeletes() { CacheDeletes.Inc() }
