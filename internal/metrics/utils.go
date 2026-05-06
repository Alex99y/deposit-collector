package metrics

import (
	errors "errors"

	observability "deposit-collector/pkg/observability"
)

func IgnoreAlreadyRegisteredError(err error) error {
	if errors.Is(err, observability.ErrMetricAlreadyRegistered) {
		return nil
	}
	return err
}
