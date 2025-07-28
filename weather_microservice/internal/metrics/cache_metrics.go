package metrics

import "github.com/prometheus/client_golang/prometheus"

type CacheMetrics struct {
	Hits    prometheus.Counter
	Misses  prometheus.Counter
	Sets    prometheus.Counter
	Deletes prometheus.Counter
}

func NewCacheMetrics() *CacheMetrics {
	return &CacheMetrics{
		Hits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_hits_total", Help: "Total cache hits",
		}),
		Misses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_misses_total", Help: "Total cache misses",
		}),
		Sets: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_sets_total", Help: "Total cache sets",
		}),
		Deletes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_deletes_total", Help: "Total cache deletes",
		}),
	}
}

func (m *CacheMetrics) Register() {
	prometheus.MustRegister(m.Hits, m.Misses, m.Sets, m.Deletes)
}
