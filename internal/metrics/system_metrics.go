package metrics

import (
	errors "errors"
	time "time"

	observability "deposit-collector/pkg/observability"
)

type SystemMetrics struct {
	metrics *observability.PrometheusMetrics
}

const (
	SYSTEM_DB_QUERY_TOTAL            = "system_db_query_total"
	SYSTEM_DB_QUERY_DURATION_SECONDS = "system_db_query_duration_seconds"
)

func (m *SystemMetrics) IncrementSystemDBQueryTotal(
	operation string,
	status string,
) error {
	counter, _ := m.metrics.GetCounter(SYSTEM_DB_QUERY_TOTAL)
	return counter.Inc(observability.Labels{
		"operation": operation,
		"status":    status,
	})
}

func (m *SystemMetrics) ObserveSystemDBQueryDuration(
	operation string,
	duration time.Duration,
) error {
	histogram, _ := m.metrics.GetHistogram(SYSTEM_DB_QUERY_DURATION_SECONDS)
	return histogram.Observe(
		duration.Seconds(),
		observability.Labels{"operation": operation},
	)
}

func NewSystemMetrics(
	metrics *observability.PrometheusMetrics,
) (*SystemMetrics, error) {
	if metrics == nil {
		return nil, errors.New("metrics is nil")
	}

	_, err := metrics.RegisterCounter(observability.CounterDefinition{
		Name:      SYSTEM_DB_QUERY_TOTAL,
		Help:      "Total number of system database queries",
		LabelKeys: []string{"operation", "status"},
	})
	if IgnoreAlreadyRegisteredError(err) != nil {
		return nil, err
	}

	_, err = metrics.RegisterHistogram(observability.HistogramDefinition{
		Name:      SYSTEM_DB_QUERY_DURATION_SECONDS,
		Help:      "Duration in seconds of system database queries",
		LabelKeys: []string{"operation"},
	})
	if IgnoreAlreadyRegisteredError(err) != nil {
		return nil, err
	}

	return &SystemMetrics{metrics: metrics}, nil
}
