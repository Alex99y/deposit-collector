package observability

import (
	"testing"
	"time"
)

func TestMeasureDurationReportsElapsedTime(t *testing.T) {
	var measured time.Duration

	done := MeasureDuration(func(duration time.Duration) {
		measured = duration
	})

	time.Sleep(10 * time.Millisecond)
	done()

	if measured <= 0 {
		t.Fatalf("expected positive measured duration, got: %v", measured)
	}
}

func TestMeasureDurationWithNilObserverIsNoop(t *testing.T) {
	done := MeasureDuration(nil)
	done()
}
