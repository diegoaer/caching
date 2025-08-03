package lru

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lru_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type", "operation"},
	)
	cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lru_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type", "operation"},
	)
	totalItems = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lru_cache_total_items",
			Help: "Total number of items in the cache",
		},
		[]string{"cache_type", "operation"},
	)
	evictionCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lru_cache_evictions_total",
			Help: "Total number of items evicted from the cache",
		},
		[]string{"cache_type", "operation", "reason"},
	)
	expirationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "lru_cache_item_expiration_duration_seconds",
			Help:    "Histogram of item expiration durations in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"cache_type"},
	)
)

const (
	metricCacheTypeLRU     = "lru"
	metricCacheTypeSafeLRU = "safe_lru"

	metricOpGet    = "get"
	metricOpSet    = "set"
	metricOpRemove = "remove"

	metricReasonManual  = "manual"
	metricReasonExpired = "expired"
	metricReasonEvicted = "evicted"
)

func init() {
	prometheus.MustRegister(cacheHits)
	prometheus.MustRegister(cacheMisses)
	prometheus.MustRegister(totalItems)
	prometheus.MustRegister(evictionCount)
	prometheus.MustRegister(expirationHistogram)
}
