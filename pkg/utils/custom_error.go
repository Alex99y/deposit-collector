package utils

type CustomError struct {
	message     string
	isRetryable bool
}

func (e *CustomError) ErrorMessage() string {
	return e.message
}

func (e *CustomError) IsRetryable() bool {
	return e.isRetryable
}

func NewCustomError(message string, isRetryable bool) *CustomError {
	return &CustomError{message: message, isRetryable: isRetryable}
}
