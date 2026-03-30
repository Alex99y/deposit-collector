package utils

import "errors"

type CustomError struct {
	error
	isRetryable bool
}

func (e *CustomError) IsRetryable() bool {
	return e.isRetryable
}

func NewCustomError(message string, isRetryable bool) *CustomError {
	return &CustomError{
		error:       errors.New(message),
		isRetryable: isRetryable,
	}
}

func IsCustomError(err error) (*CustomError, bool) {
	if err == nil {
		return nil, false
	}
	customError, ok := err.(*CustomError)
	return customError, ok
}
