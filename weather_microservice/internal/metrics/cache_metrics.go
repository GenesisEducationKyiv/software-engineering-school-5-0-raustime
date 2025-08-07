package metrics

import "github.com/prometheus/client_golang/prometheus"

var CacheOperationsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cache_operations_total",
		Help: "Total cache operations (hits, misses, sets, deletes)",
	},
	[]string{"cache", "status", "engine"},
)

func RegisterCacheMetrics() {
	prometheus.MustRegister(CacheOperationsTotal)
}
