package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	CacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "weather_cache_hits_total", Help: "Total cache hits",
	})
	CacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "weather_cache_misses_total", Help: "Total cache misses",
	})
	CacheSets = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "weather_cache_sets_total", Help: "Total cache sets",
	})
	CacheDeletes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "weather_cache_deletes_total", Help: "Total cache deletes",
	})
)

func RegisterCacheMetrics() {
	prometheus.MustRegister(CacheHits, CacheMisses, CacheSets, CacheDeletes)
}
