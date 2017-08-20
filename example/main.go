package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rolandhawk/dynamicvector"
)

var (
	responseTime = dynamicvector.NewHistogram(dynamicvector.HistogramOpts{
		Name:        "response_time_seconds",
		Help:        "Application response time",
		Buckets:     []float64{1, 10, 100},
		ConstLabels: map[string]string{"key": "value"},
	})
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(responseTime)
}

func main() {
	// Add any labels that you want.
	responseTime.With(prometheus.Labels{"url": "/index"}).Observe(0.1)
	responseTime.With(prometheus.Labels{"url": "/test"}).Observe(1.1)
	responseTime.With(prometheus.Labels{"url": "/test", "user": "1"}).Observe(19.1)

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
