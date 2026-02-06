// Package metrics defines and registers Prometheus metrics for the NIST SP 800-22
// test service. All collectors are auto-registered via promauto and are safe for
// concurrent use.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TestsTotal counts individual statistical test executions, labelled by test
	// name and outcome ("pass" or "fail").
	TestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nist_tests_total",
			Help: "Total number of NIST statistical tests run",
		},
		[]string{"test", "status"},
	)

	// TestDuration records the wall-clock duration of individual statistical tests
	// in seconds.
	TestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nist_test_duration_seconds",
			Help:    "Duration of individual NIST tests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"test"},
	)

	// OverallDuration records the wall-clock duration of the entire 15-test suite
	// in seconds.
	OverallDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "nist_overall_duration_seconds",
			Help:    "Duration of the entire NIST test suite in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// LastOverallPassRate tracks the most recently computed overall pass rate as a
	// value between 0.0 and 1.0.
	LastOverallPassRate = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "nist_last_overall_pass_rate",
			Help: "Last overall pass rate of NIST tests (0.0-1.0)",
		},
	)

	// PValue tracks the most recently observed p-value for each statistical test.
	PValue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nist_p_value",
			Help: "P-value of individual NIST tests",
		},
		[]string{"test"},
	)

	// RequestsTotal counts gRPC requests, labelled by method name and outcome
	// ("success" or "error").
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nist_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)
)

// RecordTestDuration records the duration of a test
func RecordTestDuration(testName string, durationSeconds float64) {
	TestDuration.WithLabelValues(testName).Observe(durationSeconds)
}

// IncrementTestsTotal increments the total tests counter
func IncrementTestsTotal(testName, status string) {
	TestsTotal.WithLabelValues(testName, status).Inc()
}

// RecordPValue records the p-value of a test
func RecordPValue(testName string, pValue float64) {
	PValue.WithLabelValues(testName).Set(pValue)
}

// IncrementRequestsTotal increments the total requests counter
func IncrementRequestsTotal(method, status string) {
	RequestsTotal.WithLabelValues(method, status).Inc()
}
