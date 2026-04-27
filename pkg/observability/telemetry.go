package observability

import "time"

type ObserveDuration func(time.Duration)

// MeasureDuration starts a timer and returns a function that reports
// elapsed time through the provided observer.
func MeasureDuration(observer ObserveDuration) func() {
	if observer == nil {
		return func() {}
	}

	stopTimer := StartTimer()

	return func() {
		observer(stopTimer())
	}
}
