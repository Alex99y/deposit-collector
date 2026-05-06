package metrics

import (
	errors "errors"

	observability "deposit-collector/pkg/observability"
)

const (
	QUERY_STATUS_SUCCESS QueryStatus = "success"
	QUERY_STATUS_FAILED  QueryStatus = "failed"
)

type QueryStatus string

func IgnoreAlreadyRegisteredError(err error) error {
	if errors.Is(err, observability.ErrMetricAlreadyRegistered) {
		return nil
	}
	return err
}
