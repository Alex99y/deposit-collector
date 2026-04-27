package metrics

import (
	errors "errors"
	time "time"

	observability "deposit-collector/pkg/observability"
)

type Metrics struct {
	metrics *observability.PrometheusMetrics
}

const (
	API_REQUESTS_TOTAL            = "api_requests_total"
	API_REQUESTS_DURATION_SECONDS = "api_requests_duration_seconds"
	API_REQUESTS_STATUS           = "api_requests_status"
)

func (m *Metrics) IncrementAPIRequestsCount(method string, path string) error {
	counter, _ := m.metrics.GetCounter(API_REQUESTS_TOTAL)
	return counter.Inc(observability.Labels{"method": method, "path": path})
}

func (m *Metrics) ObserveAPIRequestsDuration(
	path string,
	status string,
	duration time.Duration,
) error {
	histogram, _ := m.metrics.GetHistogram(API_REQUESTS_DURATION_SECONDS)
	return histogram.Observe(
		duration.Seconds(),
		observability.Labels{"path": path, "status": status},
	)
}

func (m *Metrics) IncrementAPIRequestsStatus(
	method string,
	path string,
	status string,
) error {
	counter, _ := m.metrics.GetCounter(API_REQUESTS_STATUS)
	return counter.Inc(
		observability.Labels{"method": method, "path": path, "status": status},
	)
}

func NewMetrics(metrics *observability.PrometheusMetrics) (*Metrics, error) {
	if metrics == nil {
		return nil, errors.New("metrics is nil")
	}

	_, err := metrics.RegisterCounter(observability.CounterDefinition{
		Name:      API_REQUESTS_TOTAL,
		Help:      "Total number of API requests",
		LabelKeys: []string{"method", "path"},
	})
	if IgnoreAlreadyRegisteredError(err) != nil {
		return nil, err
	}
	_, err = metrics.RegisterHistogram(observability.HistogramDefinition{
		Name:      API_REQUESTS_DURATION_SECONDS,
		Help:      "Duration of API requests",
		LabelKeys: []string{"path", "status"},
	})
	if IgnoreAlreadyRegisteredError(err) != nil {
		return nil, err
	}
	_, err = metrics.RegisterCounter(observability.CounterDefinition{
		Name:      API_REQUESTS_STATUS,
		Help:      "Total number of API requests status",
		LabelKeys: []string{"method", "path", "status"},
	})
	if IgnoreAlreadyRegisteredError(err) != nil {
		return nil, err
	}
	return &Metrics{metrics: metrics}, nil
}

func IgnoreAlreadyRegisteredError(err error) error {
	if errors.Is(err, observability.ErrMetricAlreadyRegistered) {
		return nil
	}
	return err
}
