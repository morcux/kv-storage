package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kv_requests_total",
			Help: "Total number of KV requests processed",
		},
		[]string{"method", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kv_request_duration_seconds",
			Help:    "Time taken to process request",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)
