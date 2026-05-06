package metrics

import (
	errors "errors"
	time "time"

	observability "deposit-collector/pkg/observability"
)

type RepositoryMetrics struct {
	metrics *observability.PrometheusMetrics
}

const (
	DB_QUERY_TOTAL            = "db_query_total"
	DB_QUERY_DURATION_SECONDS = "db_query_duration_seconds"
)

func (m *RepositoryMetrics) IncrementDBQueryTotal(
	operation string,
	status string,
) error {
	counter, _ := m.metrics.GetCounter(DB_QUERY_TOTAL)
	return counter.Inc(observability.Labels{
		"operation": operation,
		"status":    status,
	})
}

func (m *RepositoryMetrics) ObserveDBQueryDuration(
	operation string,
	duration time.Duration,
) error {
	histogram, _ := m.metrics.GetHistogram(DB_QUERY_DURATION_SECONDS)
	return histogram.Observe(
		duration.Seconds(),
		observability.Labels{"operation": operation},
	)
}

func NewRepositoryMetrics(
	metrics *observability.PrometheusMetrics,
) (*RepositoryMetrics, error) {
	if metrics == nil {
		return nil, errors.New("metrics is nil")
	}

	_, err := metrics.RegisterCounter(observability.CounterDefinition{
		Name:      DB_QUERY_TOTAL,
		Help:      "Total number of database queries",
		LabelKeys: []string{"operation", "status"},
	})
	if IgnoreAlreadyRegisteredError(err) != nil {
		return nil, err
	}

	_, err = metrics.RegisterHistogram(observability.HistogramDefinition{
		Name:      DB_QUERY_DURATION_SECONDS,
		Help:      "Duration in seconds of database queries",
		LabelKeys: []string{"operation"},
	})
	if IgnoreAlreadyRegisteredError(err) != nil {
		return nil, err
	}

	return &RepositoryMetrics{metrics: metrics}, nil
}
