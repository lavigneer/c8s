// Package metrics provides Prometheus metrics for the C8S API server and controller.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal tracks total HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestDuration tracks HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "c8s_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
		},
		[]string{"method", "path"},
	)

	// WebhooksReceived tracks webhooks received from git providers
	WebhooksReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_webhooks_received_total",
			Help: "Total number of webhooks received",
		},
		[]string{"provider", "event_type", "result"},
	)

	// CacheHits tracks cache hit/miss ratio
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_cache_requests_total",
			Help: "Total number of cache requests",
		},
		[]string{"cache_type", "result"}, // result: hit, miss
	)

	// SSEConnections tracks active SSE connections
	SSEConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "c8s_sse_connections_active",
			Help: "Number of active SSE connections",
		},
	)

	// APIErrors tracks API errors by type
	APIErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_api_errors_total",
			Help: "Total number of API errors",
		},
		[]string{"error_code", "handler"},
	)
)
