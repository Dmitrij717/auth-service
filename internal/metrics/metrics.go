package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPMetric struct {
	requestTotal           *prometheus.CounterVec
	requestDurationSeconds *prometheus.HistogramVec
}

type Metrics struct {
	http     *HTTPMetric
	registry *prometheus.Registry
}

func NewMetrics() (*Metrics, error) {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request execution time",
		},
		[]string{"method", "route"},
	)

	m := &HTTPMetric{
		requestTotal:           counter,
		requestDurationSeconds: duration,
	}

	metrics := &Metrics{
		http:     m,
		registry: prometheus.NewRegistry(),
	}

	metrics.registry.MustRegister(
	// collectors.NewGoCollector(),
	// collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	if err := metrics.registry.Register(metrics.http.requestTotal); err != nil {
		return nil, fmt.Errorf("register HTTP requests counter: %w", err)
	}
	if err := metrics.registry.Register(metrics.http.requestDurationSeconds); err != nil {
		return nil, fmt.Errorf("register HTTP duration metric: %w", err)
	}

	return metrics, nil
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Metrics) ObserveHTTPRequest(
	method string,
	route string,
	statusCode int,
	duration time.Duration,
) {

	m.http.requestTotal.WithLabelValues(
		method,
		route,
		strconv.Itoa(statusCode),
	).Inc()

	m.http.requestDurationSeconds.WithLabelValues(
		method,
		route,
	).Observe(duration.Seconds())
}
