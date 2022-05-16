package lestrratbackoff

import (
	"testing"
	"time"

	"github.com/lestrrat-go/backoff/v2"
)

func Test_ExponentialPolicy(t *testing.T) {
	backoffPolicy := NewExponentialPolicy(
		backoff.WithMinInterval(100*time.Millisecond),
		backoff.WithMaxInterval(1*time.Second),
		backoff.WithJitterFactor(0.1),
		backoff.WithMultiplier(2.0),
	)

	b := backoffPolicy.New()

	val := b.Pause()

	{
		min := 90 * time.Millisecond
		max := 110 * time.Millisecond

		if val < min {
			t.Errorf("expected %v to be greater than %v", val, min)
		}
		if val > max {
			t.Errorf("expected %v to be less than %v", val, max)
		}
	}

	for i := 0; i < 3; i++ {
		interval := val * 2.0
		min := time.Duration(float64(interval) * 0.9)
		max := time.Duration(float64(interval) * 1.1)

		val = b.Pause()
		if val < min {
			t.Errorf("expected %v to be greater than %v", val, min)
		}
		if val > max {
			t.Errorf("expected %v to be less than %v", val, max)
		}
	}

	min := 900 * time.Millisecond
	max := 1100 * time.Millisecond

	for i := 0; i < 100_000; i++ {
		val := b.Pause()
		if val < min {
			t.Errorf("expected %v to be greater than %v", val, min)
		}
		if val > max {
			t.Errorf("expected %v to be less than %v", val, max)
		}
	}
}
