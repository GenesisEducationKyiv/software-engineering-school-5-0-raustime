package cache

import "github.com/prometheus/client_golang/prometheus"

type PrometheusMetrics struct {
	cacheHits    prometheus.Counter
	cacheMisses  prometheus.Counter
	cacheSets    prometheus.Counter
	cacheDeletes prometheus.Counter
}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		cacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_hits_total",
			Help: "Total number of successful cache hits",
		}),
		cacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_misses_total",
			Help: "Total number of cache misses",
		}),
		cacheSets: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_sets_total",
			Help: "Total number of cache sets",
		}),
		cacheDeletes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_deletes_total",
			Help: "Total number of cache deletes",
		}),
	}
}

func (m *PrometheusMetrics) Register() {
	prometheus.MustRegister(
		m.cacheHits,
		m.cacheMisses,
		m.cacheSets,
		m.cacheDeletes,
	)
}

func (m *PrometheusMetrics) IncCacheHits()    { m.cacheHits.Inc() }
func (m *PrometheusMetrics) IncCacheMisses()  { m.cacheMisses.Inc() }
func (m *PrometheusMetrics) IncCacheSets()    { m.cacheSets.Inc() }
func (m *PrometheusMetrics) IncCacheDeletes() { m.cacheDeletes.Inc() }
