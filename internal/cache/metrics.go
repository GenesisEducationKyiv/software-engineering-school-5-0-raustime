package cache

import "github.com/prometheus/client_golang/prometheus"

var (
	cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_cache_hits_total",
			Help: "Total number of successful cache hits",
		},
		[]string{"city"},
	)

	cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"city"},
	)

	cacheSets = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_cache_sets_total",
			Help: "Total number of cache sets",
		},
		[]string{"city"},
	)

	cacheDeletes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_cache_deletes_total",
			Help: "Total number of cache deletes",
		},
		[]string{"city"},
	)
)

// RegisterCacheMetrics registers all cache metrics with Prometheus
func RegisterCacheMetrics() {
	prometheus.MustRegister(cacheHits, cacheMisses, cacheSets, cacheDeletes)
}
