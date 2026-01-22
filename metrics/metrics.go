package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TokenRequestsTotal counts incoming token requests.
	// Used with promhttp.InstrumentHandlerCounter.
	TokenRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "speedproxy_token_requests_total",
			Help: "Total number of token requests.",
		},
		[]string{"code", "method"},
	)

	// TokenRequestDuration measures incoming token request latency.
	// Used with promhttp.InstrumentHandlerDuration.
	TokenRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedproxy_token_request_duration_seconds",
			Help:    "Duration of token requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"code", "method"},
	)

	// UpstreamRequestsTotal counts upstream token exchange requests.
	UpstreamRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "speedproxy_upstream_requests_total",
			Help: "Total number of upstream token exchange requests.",
		},
		[]string{"code"},
	)

	// UpstreamRequestDuration measures upstream token exchange latency.
	UpstreamRequestDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "speedproxy_upstream_request_duration_seconds",
			Help:    "Duration of upstream token exchange requests.",
			Buckets: prometheus.DefBuckets,
		},
	)
)
