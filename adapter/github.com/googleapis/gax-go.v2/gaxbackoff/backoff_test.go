package gaxbackoff

import (
	"math"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	initialBackoff := 250 * time.Millisecond
	maxBackoff := 5 * time.Second
	backoffMultiplier := 2.0

	backoffConfig := &BackoffPolicy{
		Initial:    initialBackoff,
		Max:        maxBackoff,
		Multiplier: backoffMultiplier,
	}

	b := backoffConfig.New()

	var val time.Duration

	for i := 0; i < 5; i++ {
		max := initialBackoff * time.Duration(math.Pow(2, float64(i)))

		val = b.Pause()
		if val > max {
			t.Errorf("expected %v to be less than %v", val, max)
		}
	}

	for i := 0; i < 100_000; i++ {
		val = b.Pause()
		if val > maxBackoff {
			t.Errorf("expected %v to be less than %v", val, maxBackoff)
		}
	}
}
