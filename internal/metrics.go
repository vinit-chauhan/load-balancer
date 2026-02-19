package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpRequestsTotal counts the total number of HTTP requests.
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "path", "method", "code"},
	)

	// HttpRequestDurationSeconds measures the latency of HTTP requests.
	HttpRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "path", "method", "code"},
	)

	// ActiveConnections measures the number of currently active connections to each backend.
	ActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections to backend services",
		},
		[]string{"service", "backend_url"},
	)
)

// InitMetrics initializes and registers Prometheus metrics. This function is called once at startup.
func InitMetrics() {
	// All metrics are auto-registered with the default registry when promauto.New* is called.
	// We can explicitly call any initialization here if needed, but for simple metrics, it's often not required.
}
